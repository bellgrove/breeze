# Project Research Summary

**Project:** Breeze — Go MQTT/TimescaleDB fruit grading daemon
**Domain:** Industrial IoT telemetry — MQTT ingestion, local re-grading, PostgreSQL/TimescaleDB persistence
**Researched:** 2026-03-07
**Confidence:** HIGH

## Executive Summary

Breeze is a focused Go daemon that bridges an OEM fruit grader (Tomra Spectrim) to a TimescaleDB analytics store via MQTT. The project has a working baseline; this milestone adds three tightly-coupled capabilities: grademap version history (audit log for grade configuration), timing reconciliation (associating each fruit record with the exact grademap that produced its grade), and write buffering (surviving PostgreSQL outages without losing fruit records). The correct architecture for all three is well-established and follows standard IoT edge buffering and SCD-2 temporal patterns — no experimental technology is required.

The recommended approach is to build in strict dependency order. Grademap persistence is the foundation that both timing reconciliation and audit queries depend on, and it can ship independently. The write buffer and reconnect loop must be built together as a unit — a buffer without automatic flush is useless, and a flush path without transactional CopyFrom creates duplicates. Timing reconciliation is the last step and depends on an empirical investigation of the OEM fruit payload to determine whether a version field exists; that investigation gates the entire reconciliation design.

The primary risks are all in the existing codebase rather than in new technology choices. Four bugs must be fixed before or alongside the new features: the paho dispatcher blocks on a full queue (can stall all MQTT processing during a DB outage), CopyFrom is not wrapped in a transaction (causes duplicate rows on retry), json.Unmarshal errors are silently ignored (writes zero-value rows), and the shutdown defer order closes the pool before the final flush completes. These are prerequisites, not optional cleanup. No new dependencies are needed — all required APIs exist in the current go.mod.

## Key Findings

### Recommended Stack

The existing stack (pgx/v5, pgxpool, TimescaleDB, paho MQTT) is the right stack. No new dependencies are required for this milestone. The key additions are API usage patterns within the existing dependencies: `pool.Ping(ctx)` for connectivity probing, `pgconn.ConnectError` vs `pgconn.PgError` for error classification, `pool.Reset()` for forcing post-outage reconnection, and transactional `tx.CopyFrom()` rather than bare `pool.CopyFrom()`.

The write buffer should be a `sync.Mutex`-protected `[]Fruit` slice — not a channel, not the already-imported `gammazero/deque`. Channels are destructive on read (cannot replay failed writes); `gammazero/deque` is not thread-safe and not bounded. A 20-line slice-based ring with `Drain`/`Prepend` methods is the correct, zero-dependency solution. The grademap table should be a plain PostgreSQL table with a B-tree index on `received_at` — not a hypertable. Grademap changes are rare (tens per day at most); TimescaleDB hypertable overhead is unwarranted for a configuration-history table.

**Core technologies:**
- `pgxpool` (v5.7.5, existing): connection pool — use `Ping` + `Reset` for reconnection probing; never recreate the pool after an outage
- `pgconn` (v5.7.5 sub-package, existing): error classification — `ConnectError` for network failures, `PgError` for PostgreSQL errors, `SafeToRetry` for pre-send failures
- `sync.Mutex` + `[]Fruit` (stdlib): write buffer — replay-safe, bounded, drop-oldest on overflow; no new dependency
- `breeze_grademap` plain table (TimescaleDB/PostgreSQL): append-only grademap version log — enables AS-OF temporal queries via `ORDER BY received_at DESC LIMIT 1`

### Expected Features

The three features from PROJECT.md map cleanly to research-validated patterns:

**Must have (table stakes for this milestone):**
- Grademap append-only INSERT on every MQTT `grademap` receipt — foundational for audit and all reconciliation; without this, nothing else can be built
- `received_at` timestamp recorded at MQTT receipt time (not from OEM payload) — anchors the audit log correctly
- In-memory write buffer with configurable size cap and drop-oldest overflow — closes the data-loss bug on DB outage
- Reconnect probe loop with exponential backoff and automatic buffer flush — the buffer is useless without this
- Transactional `CopyFrom` wrapper — required for safe retry without duplicates
- Non-blocking send in `OnMessage` with drop-and-warn — prerequisite to prevent paho dispatcher stall during outage

**Should have (raises analytical accuracy):**
- Grademap version FK (`grademap_id`) on each `breeze_fruit` row — enables exact per-fruit grademap attribution without timestamp range queries; low cost once grademap insert is done
- Timing reconciliation — match each fruit to its correct historical grademap accounting for OEM pipeline delay; Path A (version field in fruit JSON, clean) or Path B (configurable delay offset, approximate)
- Grademap diff on store — JSON diff between consecutive payloads; useful for audit UX, low code cost

**Defer (v2+):**
- Persistent disk buffer (SQLite/bbolt) — handles process-restart-during-outage; adds dependency and crash-consistency complexity; in-memory bounded buffer is sufficient for the documented scenario
- Configurable drop-policy (drop-oldest vs drop-newest) — only relevant if compliance requirements emerge

### Architecture Approach

The existing straight-line pipeline (MQTT callback → `queue` channel → `CopyFrom` → TimescaleDB) requires three targeted additions without structural reorganisation. The key design principle is that `run()` in `breeze.go` owns all DB interaction; `Processor` remains infrastructure-free. The write buffer lives in `run()`, not inside `Processor` — the `pgx.CopyFromSource` interface makes per-item retry impossible from within `Processor`. Grademap DB writes happen via a new `grademapUpdate chan []byte` channel from `Processor` to `run()`, not inline in `OnMessage` (which runs on the paho dispatcher goroutine and must never block).

**Major components (updated):**
1. `run()` in `breeze.go` — owns `pendingBuf []Fruit`, `dbHealthy atomic.Bool`, grademap DB writes; orchestrates the reconnect loop
2. `reconnectLoop` goroutine (new, in `breeze.go`) — `pool.Ping` with exponential backoff (1s→60s, jitter); signals main loop on DB recovery
3. `Processor` in `processor/processor.go` — adds `grademapUpdate chan []byte`; `OnMessage` uses non-blocking send for both fruit queue and grademap channel
4. `breeze_grademap` table (new DDL) — append-only version log; plain PostgreSQL table, B-tree index on `(received_at)`
5. `WriteBuffer` struct — `sync.Mutex` + `[]Fruit`; `Drain(n)` / `Prepend(items)` / `Len()` methods; capacity from config

### Critical Pitfalls

1. **pgxpool dead connections not detected proactively** — pool does not fire events on connection loss; `CopyFrom` fails on first post-outage write; implement `pool.Ping` probe goroutine with exponential backoff; never rely on `pool.Stat()` to assess connectivity

2. **paho dispatcher goroutine blocks when queue is full** — with `OrderMatters=true`, any blocking operation in `OnMessage` stalls all MQTT message delivery including grademap updates; `OnMessage` must use non-blocking sends (`select { case ch<-v: default: warn }`) everywhere

3. **Non-transactional `CopyFrom` causes partial commits** — PostgreSQL COPY auto-commits per statement; mid-copy errors leave partial batches in DB; retry replay then inserts duplicates; wrap every `CopyFrom` in `tx.Begin/CopyFrom/Commit/Rollback`

4. **Shutdown defer order closes pool before final flush** — current `defer pool.Close()` / `defer close(next_batch)` LIFO execution closes the pool while writes may still be in flight; replace with explicit shutdown sequence gated by `sync.WaitGroup` or done channel

5. **Silent zero-value fruit from ignored `json.Unmarshal` errors** — `OnMessage` discards the unmarshal error; zero-value `Fruit` structs are graded and written to TimescaleDB; check the error, log the raw payload, and return without enqueuing

6. **Grademap timing misattribution during OEM pipeline delay** — OEM applies new grademap after belt pipeline delay; fruit in transit is graded on old rules but Breeze immediately switches to new map; investigate fruit JSON for version/sequence field before implementing any timing fix

## Implications for Roadmap

Based on research, strict dependency ordering is required. The suggested phases follow the build order identified in ARCHITECTURE.md and the prerequisite gate identified in PITFALLS.md.

### Phase 1: Code Cleanup — Fix Prerequisites

**Rationale:** Four existing bugs become dangerous or unrecoverable once a write buffer is added on top of them. These must be fixed first or simultaneously with Phase 2. Building reconnection buffering on top of a non-transactional CopyFrom and a blocking OnMessage handler will compound the problems.

**Delivers:** A clean foundation — no silent data corruption, no paho dispatcher risk, no duplicate rows on future retry, no shutdown data loss

**Addresses:**
- Fix `json.Unmarshal` error check in `OnMessage` — prevents zero-value rows
- Fix counter/channel TOCTOU race in `Processor.Next()` / `Values()` — prevents silent row drops
- Make `OnMessage` queue send non-blocking — prerequisite for safe DB outage handling
- Fix shutdown defer order — use explicit WaitGroup-gated shutdown

**Avoids:** Pitfalls 4 (counter race), 7 (zero-value fruit), and sets up safe conditions for Pitfalls 2 and 3 fixes in Phase 2

**Research flag:** Standard patterns; no additional research needed

### Phase 2: Write Buffer and Reconnect Loop

**Rationale:** The data-loss bug on DB outage is the highest-value fix. The buffer, reconnect probe, and transactional CopyFrom are tightly coupled and must all ship together — a buffer without flush-on-reconnect is useless, and a flush path without transactional CopyFrom creates duplicates.

**Delivers:** Zero fruit loss during DB outages up to the configured buffer limit; automatic recovery without operator intervention

**Addresses:**
- `WriteBuffer` struct (`sync.Mutex` + `[]Fruit`, `Drain`/`Prepend`/`Len`, configurable `MaxItems` in config YAML)
- `reconnectLoop` goroutine with exponential backoff (`pool.Ping`, 1s→60s + jitter)
- `dbHealthy atomic.Bool` flag in `run()`; buffer-only mode when unhealthy
- Transactional `tx.CopyFrom` replacing bare `pool.CopyFrom`
- `pool.Reset()` call on confirmed reconnect before flush attempt
- Configurable `WriteBuffer.MaxItems` (default 10,000 items, ~15 MB heap)
- Drop-oldest policy on overflow with `slog.Warn` per dropped record

**Implements:** WriteBuffer component, reconnectLoop goroutine (Architecture section)

**Avoids:** Pitfalls 1 (dead pool detection), 2 (paho block — via Phase 1 non-blocking send), 3 (partial commit), 5 (shutdown before flush), 8 (unbounded buffer growth)

**Research flag:** Standard patterns; all API behaviour verified against pgx/v5 official docs. No additional research needed.

### Phase 3: Grademap Persistence

**Rationale:** Grademap persistence is independent of the write buffer and can ship in any order relative to Phase 2. It is placed third because Phase 1 cleanup is a prerequisite for safe operation and Phase 2 is the highest urgency fix. However, it could ship before or alongside Phase 2 if sequencing permits.

**Delivers:** Full audit trail of every grademap received; foundation for timing reconciliation and all future analytical queries

**Addresses:**
- New `breeze_grademap` table (plain PostgreSQL, append-only, `received_at TIMESTAMPTZ`, `name TEXT`, `payload JSONB`)
- Index on `(received_at)` for temporal lookups
- `grademapUpdate chan []byte` added to `Processor`; non-blocking send in `OnMessage`
- New `case <-p.GrademapUpdate():` in `run()` select loop; `pool.QueryRow` INSERT with `RETURNING id`
- Cache last-inserted `grademap_id` in `run()` for stamping onto fruit rows

**Uses:** `pool.QueryRow` with `RETURNING id`; `pgconn` error types for DB error handling on insert

**Avoids:** Anti-feature — mutable single-row grademap (destroys history); Anti-feature — writing in `OnMessage` (blocks paho dispatcher)

**Research flag:** Standard append-only audit log / SCD-2 pattern; well documented. No additional research needed.

### Phase 4: Timing Reconciliation

**Rationale:** Timing reconciliation depends on Phase 3 (grademap table must exist with live data) and on a prerequisite investigation step. The investigation must be completed and its outcome documented before any implementation begins — the two implementation paths (version field vs delay model) diverge significantly.

**Delivers:** Each fruit record attributed to the correct historical grademap accounting for OEM pipeline delay; closes the accuracy gap in analytical queries

**Investigation gate (must complete first):**
- Inspect real OEM fruit JSON payloads for any version, sequence, or configuration-reference field
- The architecture research confirms: `Sizer.Schema.Version` is an int and `Sizer.Grade.Id/Name` reflect output grade — neither is a grademap configuration version. However, the full fruit payload structure beyond what `UnmarshalJSON` currently maps may contain additional fields. This needs live traffic inspection, not assumption.

**If version field found (Path A — preferred):**
- Parse version from fruit JSON in `UnmarshalJSON`
- Look up matching `breeze_grademap` row; set `grademap_id` FK on fruit row
- Clean, exact attribution

**If no version field (Path B — fallback):**
- Add `GrademapPipelineDelay time.Duration` to config YAML (default 0)
- Adjust temporal lookup: `WHERE received_at <= ($fruit_time - delay) ORDER BY received_at DESC LIMIT 1`
- Flag ambiguous fruit in the transition window rather than silently misattributing

**Research flag:** NEEDS investigation — inspect live OEM fruit JSON payloads before planning this phase. The architecture research has already confirmed the currently-mapped fields do not contain a grademap version, but unmapped fields may exist.

### Phase Ordering Rationale

- Phase 1 before everything: existing bugs become compounding problems once a buffer is introduced; the counter/channel race and non-blocking send issues are cheap to fix now and expensive later
- Phase 2 urgency: data loss on DB outage is the active, observable production problem; highest business value
- Phase 3 before Phase 4: the grademap table must exist and be populated with live data before timing reconciliation can be designed, implemented, or empirically calibrated
- Phase 4 gated by investigation: implementing reconciliation without knowing whether a version field exists wastes significant effort and may produce the wrong architecture

### Research Flags

Phases needing deeper research or investigation during planning:
- **Phase 4 (Timing Reconciliation):** Requires live OEM traffic inspection to determine whether fruit JSON contains a grademap version field. This is a binary gate — the implementation approach changes entirely based on the finding. Do not plan implementation until investigation is complete and documented.

Phases with standard, well-documented patterns (no additional research needed):
- **Phase 1 (Code Cleanup):** All fixes are idiomatic Go corrections; no novel patterns
- **Phase 2 (Write Buffer + Reconnect):** pgxpool ping-based health check and slice-based ring buffer are well-documented; all API behaviour verified
- **Phase 3 (Grademap Persistence):** Append-only audit log / SCD-2 pattern is canonical PostgreSQL; no unknowns

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All APIs verified against pkg.go.dev official docs; import paths confirmed (pgx/v5/pgconn, not standalone pgconn); pool behaviour confirmed against upstream GitHub discussions |
| Features | HIGH | Feature set is grounded in direct codebase analysis and established IoT edge + SCD-2 patterns; only gap is OEM version-field investigation (proprietary, no public data) |
| Architecture | HIGH | Based on direct codebase read of breeze.go, processor.go, fruit.go, grademap.go; component placement decisions are well-reasoned and consistent across all research files |
| Pitfalls | HIGH | Most pitfalls verified against pgx source, paho issues, and PostgreSQL COPY documentation; timing section is MEDIUM due to OEM-specific domain uncertainty |

**Overall confidence:** HIGH

### Gaps to Address

- **OEM fruit JSON version field:** Architecture research confirmed that currently-mapped fields (`Sizer.Schema.Version`, `Sizer.Grade.Id`) are not grademap configuration versions. However, the full payload may contain unmapped fields. Inspect live traffic before Phase 4 planning. Document the finding — it is a hard gate.

- **OEM pipeline delay magnitude:** If no version field exists (Path B reconciliation), the `GrademapPipelineDelay` config value must be calibrated empirically by observing real grade distribution shifts after grademap changes. This cannot be determined from research alone. Set to 0 initially; measure and adjust.

- **Grademap change frequency in production:** Research assumes tens of grademap changes per day at most. If grademap changes are more frequent (e.g., per-batch reconfiguration), the plain-table vs hypertable decision should be revisited. Unlikely but worth confirming with operators.

## Sources

### Primary (HIGH confidence)
- [pgxpool pkg.go.dev](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) — Pool, Config, Ping, Reset, Stat, ShouldPing, PrepareConn lifecycle
- [pgconn pkg.go.dev](https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn) — ConnectError, PgError, SafeToRetry, Timeout error types
- [gammazero/deque pkg.go.dev](https://pkg.go.dev/github.com/gammazero/deque) — thread-safety and capacity behaviour (confirmed: not thread-safe, unbounded)
- [Temporal Tables — SQL Server (Microsoft Learn)](https://learn.microsoft.com/en-us/sql/relational-databases/tables/temporal-tables) — SQL:2011 temporal validity patterns
- [Go Channel Patterns - Drop — DEV Community](https://dev.to/b0r/go-channel-patterns-drop-4k19) — Go channel drop-newest vs non-blocking send behaviour
- Codebase direct analysis: `breeze.go`, `processor/processor.go`, `processor/fruit.go`, `processor/grademap.go`, `.planning/codebase/CONCERNS.md`

### Secondary (MEDIUM confidence)
- [pgx Discussion #1415 — Recovering from connection interruption](https://github.com/jackc/pgx/discussions/1415) — pool does not auto-retry; circuit-breaker recommendation (community + maintainer confirmed)
- [pgx Issue #891 — Detecting dead pgxpool](https://github.com/jackc/pgx/issues/891) — pool auto-recovery behaviour
- [paho Issue #474 — Deadlocks when not wrapping callbacks](https://github.com/eclipse-paho/paho.mqtt.golang/issues/474) — single dispatcher goroutine behaviour
- [How to query SCD-2 — Maja Ferle](https://majaferle.com/how-to-query-slowly-changing-dimensions-type-2/) — canonical SCD-2 temporal query pattern
- [Edge Buffering — QuestDB Glossary](https://questdb.com/glossary/edge-buffering/) — IoT edge buffer design
- [pgx transactions and COPY — danp.net](https://danp.net/posts/pgx-tx-copy/) — CopyFrom transactional semantics
- [TimescaleDB metadata table best practices](https://www.tigerdata.com/learn/best-practices-for-time-series-metadata-tables) — hypertable vs plain table decision for configuration history

---
*Research completed: 2026-03-07*
*Ready for roadmap: yes*
