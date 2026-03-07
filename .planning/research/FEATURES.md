# Feature Research

**Domain:** Industrial IoT telemetry daemon — MQTT ingest, local re-grading, TimescaleDB persistence
**Researched:** 2026-03-07
**Confidence:** HIGH (architecture and pattern decisions); MEDIUM (OEM grademap version-field investigation — proprietary protocol, no public data)

---

## Scope

This document covers the three active feature areas from PROJECT.md:

1. **Grademap versioning** — persist grademap history to TimescaleDB with full audit trail
2. **Timing reconciliation** — associate each fruit record with the correct historical grademap
3. **Write buffer** — survive PostgreSQL outages without losing fruit records

---

## Feature Landscape

### Table Stakes (Required for Correctness)

These are not optional. Without them, the analytics goal ("4% insect damage" not "10% class B") is unreliable or unknowingly wrong.

| Feature | Why Required | Complexity | Notes |
|---------|--------------|------------|-------|
| Grademap append-only insert on change | Every grademap change must be persisted; without it, historical re-grading is impossible | LOW | INSERT on MQTT `grademap` receipt; never UPDATE existing rows |
| `received_at` timestamp on every stored grademap | Anchors the audit log to wall-clock time so fruit records can be matched to the correct map | LOW | Use `time.Now()` at moment of MQTT receipt, not the grademap payload's own fields |
| "Active grademap at time T" lookup | The query that maps each fruit's processing timestamp to its grademap | MEDIUM | Standard SCD-2 / temporal validity pattern; see Architecture section below |
| Write buffer for DB outages | Fruit records graded on the old map must not be silently dropped when DB is down | MEDIUM | In-memory FIFO; size limit configured; FIFO flush on reconnect |
| Automatic flush on DB reconnect | Buffer is useless without automated drain | MEDIUM | Health-check loop or pgxpool event hook triggers replay |

### Differentiators (Raise Analytical Accuracy)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Grademap-version FK on each fruit row | Allows exact query: "show me all fruit graded under grademap version X" rather than relying on timestamp overlap | MEDIUM | Adds a `grademap_id` foreign key column to `breeze_fruit`; requires schema migration |
| OEM pipeline-delay model | Fruit in transit when a grademap change arrives are still graded on the old map; modelling the delay closes the accuracy gap | HIGH | Only needed if fruit JSON has no version field; investigate first |
| Grademap diff on store | Record what changed between consecutive grademaps (JSON diff stored alongside the raw payload) | LOW | Makes audit queries far more useful; minimal code cost |

### Anti-Features (Do Not Build These)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Mutable "current grademap" row (UPDATE in place) | Simpler schema | Destroys history; cannot answer "what map was active at 14:32?" | Append-only insert; mark active with `valid_to = NULL` or use separate `latest` view |
| HTTP endpoint to query grademap history | Operators want visibility | Out of scope; headless daemon; adds a whole HTTP layer and auth concern | TimescaleDB queries directly; Grafana/psql suffice |
| On-disk write buffer (SQLite, file) | Survives process restart | Adds a dependency, crash-consistency concerns, and significant complexity for what is typically a short outage | In-memory bounded FIFO is sufficient; process restart during DB outage is an acceptable edge case; log lost count |
| Unlimited in-memory buffer | "Never lose a record" | Unbounded memory growth; process OOMs and loses everything | Configurable size limit; log when dropping; accept bounded loss explicitly |
| Re-grading historical records on grademap change | Retroactive correction of old data | Correct grademap is already captured at ingest time if timing reconciliation works; retro re-grade is a read-time concern, not a write-time one | Query-time re-grade via stored grademap payloads if ever needed |

---

## Feature Detail: Grademap Versioning

### Pattern: Append-Only Audit Log (SCD Type 2 / Event Log)

**Approach A — Append-only with open-ended validity (recommended)**

```sql
CREATE TABLE breeze_grademap (
    id          BIGSERIAL PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name        TEXT NOT NULL,
    payload     JSONB NOT NULL           -- raw grademap JSON for full fidelity
);
```

- Every MQTT `grademap` message triggers an INSERT.
- No UPDATE, no DELETE.
- `received_at` is the system clock at MQTT receipt — the anchor for all temporal lookups.
- `payload` stores the raw bytes for auditability and future re-grading.

**Why not `valid_from` / `valid_to` pair:** A `valid_to` column requires updating the previous row when a new grademap arrives. This is a mutation, adds a two-step write (UPDATE + INSERT), and a race condition if the UPDATE succeeds but the INSERT fails. Instead, derive `valid_to` at query time using the `LEAD()` window function:

```sql
SELECT
    id,
    received_at                                              AS valid_from,
    LEAD(received_at) OVER (ORDER BY received_at)            AS valid_to,
    name,
    payload
FROM breeze_grademap
ORDER BY received_at;
```

This view or CTE gives the SCD-2 semantics without any mutation. `valid_to IS NULL` means the row is still active.

**Confidence:** HIGH — standard PostgreSQL/TimescaleDB pattern; backed by SCD-2 literature and PostgreSQL window function documentation.

### Pattern: Grademap Version FK on Fruit Rows (enhancement)

Add `grademap_id BIGINT REFERENCES breeze_grademap(id)` to `breeze_fruit`. Set it at the moment of grading by caching the last-inserted grademap row ID in memory alongside the in-memory `Grademap` struct. This avoids the timestamp-range lookup entirely for per-fruit attribution, and is the cleanest design if timing reconciliation is accurate.

**Confidence:** HIGH — straightforward FK relationship; no novel patterns required.

---

## Feature Detail: Timing Reconciliation

### Decision Gate: Does Fruit JSON Carry a Grademap Version Field?

This must be investigated by inspecting real payloads from the OEM before choosing an implementation approach. The two paths diverge significantly:

**Path A — Fruit JSON has a version field (clean)**

If the fruit payload contains a grademap ID or version token that matches a field in the grademap payload:

1. Parse the version field from the fruit JSON.
2. Look up the matching `breeze_grademap` row by that version.
3. Set `grademap_id` on the fruit row directly.

No timing model needed. This is the preferred outcome. **Investigate first.**

**Path B — No version field (delay model required)**

The OEM applies a new grademap after a pipeline delay (fruit already in transit are graded on the old map). Without a version field, reconciliation must estimate which grademap was active when the OEM processed a fruit, not when Breeze received the MQTT message.

Options:

| Option | How It Works | Accuracy | Complexity |
|--------|-------------|----------|------------|
| Timestamp overlap with configurable offset | Subtract a fixed `grademap_propagation_delay` from fruit's processing timestamp, then use temporal lookup | Medium — offset is approximate | Low |
| Observe empirically and hard-code | Record real lag between grademap MQTT receipt and first fruit using that map; use as config value | Medium-High if environment is stable | Low (investigation cost only) |
| No reconciliation (accept window) | Attribute all fruit to the map active at Breeze receipt time | Low — incorrect during transition windows | Zero |

**Recommendation:** Use "timestamp overlap with configurable offset" if no version field exists. Store `grademap_propagation_delay_ms` in YAML config. Default to 0 (immediate) and adjust after empirical observation.

### "Active Grademap at Time T" Query

Standard SCD-2 temporal lookup against the append-only table:

```sql
-- Given a fruit processed at :fruit_time, find its grademap
SELECT g.id, g.payload
FROM (
    SELECT
        id,
        received_at                                         AS valid_from,
        LEAD(received_at) OVER (ORDER BY received_at)       AS valid_to,
        payload
    FROM breeze_grademap
) g
WHERE :fruit_time >= g.valid_from
  AND (:fruit_time < g.valid_to OR g.valid_to IS NULL)
LIMIT 1;
```

Index required: `CREATE INDEX ON breeze_grademap (received_at);`

**Confidence:** HIGH — this is the textbook temporal validity query pattern for append-only history tables in PostgreSQL.

---

## Feature Detail: Write Buffer for DB Outages

### Required Behavior

When `pool.CopyFrom()` fails:

1. Do not drop the fruit records that were drained from the queue during the failed write.
2. Hold those records in a secondary buffer until the DB is reachable again.
3. Automatically flush the buffer (in FIFO order) on reconnect.
4. When the buffer is full, drop the **oldest** records first, log each drop with a count.

### Drop Policy: Drop-Oldest (Recommended)

**Why drop-oldest, not drop-newest:**

In a fruit-grading audit context, the most recent production data has the highest operational value. An operator investigating a current run needs recent records. Records from an outage that started 10 minutes ago are less actionable than records from 30 seconds ago. Drop-oldest preserves recency.

This aligns with how circular buffers work in comparable systems (MongoDB oplog, WAL-based systems, IoT edge buffers).

**Why not drop-newest:**

Drop-newest means a short burst of new data keeps the old stale data. The buffer fills and refuses new entries. This is appropriate when every historical record is equally precious (compliance logging, financial audit). For operational fruit grading, recency matters more.

**Why not persistent disk buffer:**

PROJECT.md explicitly constrains the tech stack to a single binary with no additional dependencies. A disk buffer (SQLite, bbolt, flat file) adds crash-consistency complexity for what the PROJECT.md models as a DB reconnection scenario (not a process restart scenario). Persist-to-disk is a future enhancement if requirements change.

### Buffer Design

```
In-memory circular buffer (drop-oldest):
  - Type: []Fruit (slice used as ring)
  - Capacity: configurable via YAML (e.g., buffer_limit: 10000)
  - On overflow: overwrite oldest slot, increment dropped_count counter
  - On flush: drain in FIFO order via CopyFrom
  - On error: re-enqueue or log and advance (do not re-enqueue infinitely)
```

**Existing channel queue note:** The current `Processor.queue` channel (capacity 50) is a FIFO buffer but is consumed synchronously at write time — drained items are gone if the write fails. The write buffer must sit *downstream* of the channel drain, holding records that survived dequeue but failed to persist.

### Reconnect Detection

pgxpool does not provide an event callback for reconnection. The recommended pattern:

- After a failed write, enter a retry loop with exponential backoff.
- On each iteration: `pool.Ping(ctx)` to test liveness.
- On success: flush buffer via a single `CopyFrom` call.
- On continued failure: sleep and retry; do not block the MQTT ingest path.

The MQTT goroutine and the write/flush goroutine must run concurrently so fruit continues to accumulate in the buffer during a DB outage. The existing architecture (MQTT callback enqueues to `queue` channel; main loop reads and writes) supports this without structural changes — the main loop retry just needs to not block on DB writes indefinitely.

**Confidence:** MEDIUM — pgxpool ping-based health check is documented; the buffer design pattern is standard for IoT edge buffering; specific drop-oldest implementation in Go requires a custom ring buffer (not a native channel).

---

## Feature Dependencies

```
[Grademap append-only insert]
    └──required by──> [Active grademap at time T lookup]
                          └──required by──> [Timing reconciliation]
                          └──required by──> [Grademap-version FK on fruit row]

[Write buffer]
    └──required by──> [Automatic flush on reconnect]

[OEM version field investigation]
    └──determines──> [Timing reconciliation approach]
                         Path A: version field → direct FK lookup (simpler)
                         Path B: no version field → delay model (complex)
```

### Dependency Notes

- **Grademap insert requires `received_at` precision:** Millisecond-precision timestamps are needed; Go's `time.Now()` at MQTT receipt provides this. Do not use the grademap payload's own date fields as the primary anchor — they reflect OEM time, not DB-insert time.
- **Version FK depends on grademap insert:** The in-memory `Grademap` struct must cache the last-inserted `id` from the `RETURNING id` clause of the INSERT so it can be stamped on outgoing fruit rows.
- **Write buffer does not depend on grademap versioning:** These are independent features and can be implemented in parallel or in any order.
- **Timing reconciliation approach depends on investigation first:** Implement grademap storage first (no dependency on reconciliation approach), then investigate fruit payload structure, then choose reconciliation path.

---

## MVP Definition

### Launch With (this milestone)

- [ ] Grademap append-only insert on MQTT receipt, with `received_at` and raw `payload` stored — enables audit log and all future reconciliation
- [ ] In-memory write buffer with configurable size limit and drop-oldest policy — closes the data-loss bug
- [ ] pgxpool ping-retry loop with automatic buffer flush on reconnect — makes the buffer useful
- [ ] Grademap version FK stamped on each fruit row (if grademap insert is implemented at the same time, this costs near-zero additional work)

### Add After Validation (v1.x)

- [ ] Timing reconciliation — investigate OEM fruit JSON for version field first; choose Path A or Path B based on findings; this is lower priority than preventing data loss
- [ ] Grademap diff on store — JSON diff between consecutive payloads stored alongside raw; useful for audit but not blocking

### Future Consideration (v2+)

- [ ] Persistent disk buffer — only if process-restart-during-outage is observed as a real operational problem
- [ ] Configurable drop-policy (drop-oldest vs drop-newest) — only if audit/compliance requirements emerge that change the recency preference

---

## Feature Prioritization Matrix

| Feature | Operational Value | Implementation Cost | Priority |
|---------|------------------|---------------------|----------|
| Grademap append-only insert | HIGH — foundational for audit and reconciliation | LOW | P1 |
| Write buffer (drop-oldest, bounded) | HIGH — closes data-loss bug | MEDIUM | P1 |
| Reconnect flush loop | HIGH — buffer is useless without it | MEDIUM | P1 |
| Grademap version FK on fruit row | HIGH — exact attribution | LOW (given grademap insert is done) | P1 |
| Timing reconciliation (Path A: version field) | HIGH — closes accuracy gap | LOW | P1 if version field found |
| Timing reconciliation (Path B: delay model) | MEDIUM — approximate improvement | MEDIUM | P2 if no version field |
| Grademap diff on store | MEDIUM — improves audit UX | LOW | P2 |
| Persistent disk buffer | LOW — handles restart-during-outage edge case | HIGH | P3 |

**Priority key:** P1 = must have this milestone, P2 = add after core is stable, P3 = future

---

## Implementation Variants Considered

### Grademap Storage — Variant: Dedicated Hypertable vs Regular Table

TimescaleDB hypertables partition by time and are optimised for append-heavy time-series with range queries. The grademap table is low-cardinality (a few rows per day at most) and not a time-series in the traditional sense — it is configuration history. A **regular PostgreSQL table with a B-tree index on `received_at`** is correct. Do not create a hypertable for it. Hypertable overhead is unnecessary and complicates schema management for a table that will have dozens of rows per year, not millions.

### Write Buffer — Variant: Channel vs Ring Buffer Slice

Go buffered channels are FIFO but do not support drop-oldest overflow. When a channel is full, the sender blocks (or drops with `select/default`, which drops newest). To get drop-oldest behaviour, a custom ring buffer slice is required. This is approximately 20 lines of idiomatic Go. The existing `deque` import (commented out in `processor.go`) indicates this was previously considered — a simple slice-based ring is sufficient and avoids re-adding the deque dependency.

### Reconnect — Variant: pgxpool.Reset() vs Ping Loop

`pgxpool.Reset()` closes all connections and forces a full reconnect. It is designed for scenarios where a network interruption corrupts all pool connections simultaneously. For a normal DB restart (which is the documented bug), individual connection health checks via `pool.Ping()` are sufficient and less disruptive. Use `Ping()` in the retry loop; reserve `Reset()` for cases where `Ping()` succeeds but `CopyFrom()` keeps failing.

---

## Sources

- [Event Sourcing pattern — Azure Architecture Center](https://learn.microsoft.com/en-us/azure/architecture/patterns/event-sourcing) (HIGH confidence — Microsoft official)
- [Event Sourcing vs Audit Log — Kurrent.io](https://www.kurrent.io/blog/event-sourcing-audit) (MEDIUM confidence — practitioner analysis)
- [How to query Slowly Changing Dimensions Type 2 — Maja Ferle](https://majaferle.com/how-to-query-slowly-changing-dimensions-type-2/) (HIGH confidence — canonical SCD-2 pattern)
- [Slowly Changing Dimensions in Postgres — Marc Linster, Medium](https://medium.com/@marclinster/slowly-changing-dimensions-in-postgres-7d0f4cac2191) (MEDIUM confidence — PostgreSQL-specific SCD-2 implementation)
- [Temporal Tables — SQL Server (Microsoft Learn)](https://learn.microsoft.com/en-us/sql/relational-databases/tables/temporal-tables) (HIGH confidence — SQL:2011 standard temporal validity patterns)
- [Go Channel Patterns - Drop — DEV Community](https://dev.to/b0r/go-channel-patterns-drop-4k19) (HIGH confidence — Go stdlib behaviour)
- [Edge Buffering — QuestDB Glossary](https://questdb.com/glossary/edge-buffering/) (MEDIUM confidence — IoT edge buffering design)
- [OpenTelemetry Collector Resiliency](https://opentelemetry.io/docs/collector/resiliency/) (MEDIUM confidence — industry reference for bounded queue drop behaviour)
- [pgxpool package documentation — pkg.go.dev](https://pkg.go.dev/github.com/jackc/pgx/v4/pgxpool) (HIGH confidence — official library docs)
- [MQTT Retained Messages — HiveMQ](https://www.hivemq.com/blog/mqtt-essentials-part-8-retained-messages/) (HIGH confidence — authoritative MQTT reference)
- [IoT Meets OLTP: Backfilling Challenges — Red Gate Simple Talk](https://www.red-gate.com/simple-talk/databases/sql-server/iot-meets-oltp-how-to-handle-backfilling-challenges-in-real-time-systems/) (MEDIUM confidence — IoT write-buffer design patterns)

---

*Feature research for: Breeze — Go MQTT/TimescaleDB fruit grading daemon*
*Researched: 2026-03-07*
