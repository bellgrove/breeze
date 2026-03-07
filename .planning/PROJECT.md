# Breeze

## What This Is

Breeze is a Go daemon that bridges a Tomra OEM fruit grader and a PostgreSQL/TimescaleDB database. It subscribes to MQTT topics for raw fruit measurements and grademap updates, re-applies the grademap locally to determine *why* each fruit was graded (primary defect and contributing defects), and bulk-writes the enriched records for analytics and auditing.

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

### Active

- [ ] Fix PostgreSQL reconnection: buffer writes when DB is down (up to a configurable limit), flush automatically when connection is restored
- [ ] Store grademap in TimescaleDB with timestamps — enables audit log of all grademap changes and what changed
- [ ] Fix grademap timing reconciliation: investigate whether fruit JSON contains a grademap version field; if not, model the OEM's delay between publishing a new grademap and applying it to fruit (so the correct historical grademap is used to re-grade each fruit)
- [ ] Code review and cleanup of existing codebase

### Out of Scope

- HTTP API or web interface — headless daemon only
- Alerting or notifications — analytics via TimescaleDB queries only
- Support for multiple grader machines — single Tomra unit (topic prefix `tomra/211632/`)

## Context

The OEM grader sends raw fruit measurements (size, colour spectroscopy, etc.) and a final grade but does not explain *why* a fruit was assigned that grade. Breeze re-applies the grademap rules locally to infer the cause. The grademap is a retained MQTT topic — it is not currently persisted to the database.

**Known bugs:**
1. **Grademap timing gap:** When the grademap changes, the OEM applies the change after some pipeline delay (fruit already in transit are graded on the old map). Breeze currently switches immediately on receipt, producing a window where reconciliation fails. Fruit JSON may contain a grademap version field — this must be investigated before choosing a fix approach.
2. **PostgreSQL reconnection failure:** If the database resets, the pgxpool connection is not recovered and writes are lost. The service must buffer unconsumed fruit records and replay them once the connection is restored.

**Analytics goal:** TimescaleDB stores the defect-attributed records, enabling queries like "insect damage rate per hour this week" — breaking down grade outcomes by cause rather than just grade category.

## Constraints

- **Tech stack**: Go 1.24.3, paho.mqtt.golang, pgx/v5, TimescaleDB — no changes to these
- **Schema**: Existing `breeze_fruit` TimescaleDB table; grademap storage will require new table(s)
- **Deployment**: Single binary, Linux, no HTTP layer

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Investigate fruit JSON for grademap version field | Determines fix approach for timing bug — version field = clean fix, no field = must model OEM delay | — Pending |
| Write buffer limit for DB reconnection | Bounds memory usage during outage; configurable value needed | — Pending |

---
*Last updated: 2026-03-07 after initialization*
