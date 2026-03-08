# Breeze

## What This Is

Breeze is a Go daemon that bridges a Tomra OEM fruit grader and a PostgreSQL/TimescaleDB database. It subscribes to MQTT topics for raw fruit measurements and grademap updates, re-applies the grademap locally to determine *why* each fruit was graded (primary defect and contributing defects), and bulk-writes the enriched records for analytics and auditing. It buffers records during database outages and attributes each fruit to the exact historical grademap that was active when the OEM graded it.

## Core Value

Every fruit record must be attributable to a specific defect cause — "10% class B" is useless; "4% insect damage, 3% sunburn, 2% punctures, 1% rot" is the goal.

## Requirements

### Validated

- ✓ Subscribe to `tomra/211632/fruit` MQTT topic and ingest raw fruit measurements — existing
- ✓ Subscribe to `tomra/211632/grademap` MQTT topic and maintain in-memory grademap — existing
- ✓ Re-apply grademap locally to derive `PrimaryDefect`, `OtherDefects`, `PrimaryReason`, `OtherReason` per fruit — existing
- ✓ Bulk-write enriched fruit records to TimescaleDB via `pgx` COPY protocol — existing
- ✓ Batch writes (30-item threshold or 1-second timer) to reduce DB round-trips — existing
- ✓ Config via YAML file with env var override — existing
- ✓ Graceful shutdown on OS signal — existing
- ✓ `OnMessage` uses a non-blocking channel send with drop-and-warn so the paho MQTT dispatcher never stalls — v1.0
- ✓ `Processor` counter decrements after channel read, not before (fix counter/channel race) — v1.0
- ✓ `json.Unmarshal` errors in `OnMessage` are logged, not silently swallowed — v1.0
- ✓ Shutdown defer order in `run()` corrected — MQTT disconnect before pool close — v1.0
- ✓ General code review and cleanup (dead code, commented-out imports, style) — v1.0
- ✓ Fruit records buffered in bounded in-memory slice when `CopyFrom` fails (configurable max, drop-oldest policy) — v1.0
- ✓ Background reconnect probe (`pool.Ping` with exponential backoff) runs when DB is unhealthy — v1.0
- ✓ Buffered records automatically flushed when DB connection is restored — v1.0
- ✓ All `CopyFrom` calls wrapped in explicit transactions to prevent partial-batch duplicates on retry — v1.0
- ✓ `breeze_grademap` table (append-only: `id`, `received_at TIMESTAMPTZ`, `payload JSONB`) — v1.0
- ✓ A record inserted on every MQTT grademap update — v1.0
- ✓ `breeze_grademap_changes` table with queryable per-field change rows — v1.0
- ✓ Fruit records matched to correct historical grademap via AS-OF timestamp query — v1.0
- ✓ Configurable `grademap_propagation_delay` offset applied to AS-OF lookup — v1.0

### Active

- [ ] Metrics endpoint exposing queue depth, buffer fill level, DB health state, and fruit throughput rate (OBS-01)
- [ ] Structured log fields for grademap reconciliation misses (OBS-02)
- [ ] Support for multiple Tomra grader units — configurable topic prefixes (MULTI-01)

### Out of Scope

- HTTP API or web interface — headless daemon only; analytics via TimescaleDB directly
- Alerting or notifications — analytics via TimescaleDB queries only
- Mobile or browser dashboard — not requested
- OAuth / external auth — no HTTP layer

## Context

**v1.0 shipped 2026-03-08.**

Codebase: 6,146 lines Go. Tech stack: Go 1.24.3, paho.mqtt.golang, pgx/v5, TimescaleDB. Single binary, Linux, no HTTP layer.

The daemon now correctly attributes every fruit to the specific grademap active when the OEM graded it, using a configurable propagation delay offset. Fruit records are buffered in memory during DB outages (configurable limit, drop-oldest policy) and auto-flushed on reconnect. Every grademap change is persisted to `breeze_grademap` and `breeze_grademap_changes` for audit and reconciliation queries.

Live infrastructure verification (E2E with live MQTT + TimescaleDB) remains pending — all unit and integration tests pass. Three VALIDATION.md files have stale frontmatter (nyquist_compliant: false) that does not reflect actual passing tests.

## Constraints

- **Tech stack**: Go 1.24.3, paho.mqtt.golang, pgx/v5, TimescaleDB — no changes to these
- **Schema**: `breeze_fruit`, `breeze_grademap`, `breeze_grademap_changes` TimescaleDB tables
- **Deployment**: Single binary, Linux, no HTTP layer

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Investigate fruit JSON for grademap version field | Determines fix approach for timing bug | ✓ Path B confirmed — no version field in payload; use AS-OF timestamp with propagation delay offset |
| Write buffer limit for DB reconnection | Bounds memory usage during outage | ✓ Configurable max size with drop-oldest eviction policy |
| Non-blocking `OnMessage` (inner select/default for timer) | Timer channel (cap 1) could stall dispatcher if queue branch triggered twice in quick succession | ✓ Implemented in Phase 5 — both queue and timer sends are now non-blocking |
| Decimal phase numbering for Phase 5 | Phase 5 inserted after milestone audit to close tech debt without disrupting planning | ✓ Works well — clear insertion semantics |
| `reflect.DeepEqual` over `reflect.Value.Equal` for map comparison in tests | `reflect.Value.Equal` panics on `map[string]interface{}`; `reflect.DeepEqual` is the correct Go idiom | ✓ Fixed TestFromJson/Class_2; `go test ./...` exits clean |

---
*Last updated: 2026-03-08 after v1.0 milestone*
