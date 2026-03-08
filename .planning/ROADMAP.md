# Roadmap: Breeze

## Overview

Breeze has a working baseline; this milestone hardens it into a production-reliable daemon. The work proceeds in strict dependency order: fix existing bugs that would compound new features (Phase 1), then add write buffering and reconnection to eliminate data loss during DB outages (Phase 2), then persist grademap history as the audit foundation (Phase 3), then reconcile each fruit record to the exact historical grademap that produced its grade (Phase 4). Phase 4 is gated by a live-traffic investigation that must complete before implementation begins.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Code Cleanup** - Fix four existing bugs that become compounding problems once write buffering is added (completed 2026-03-07)
- [x] **Phase 2: Write Buffer and Reconnect** - Eliminate fruit data loss during PostgreSQL outages (completed 2026-03-07)
- [x] **Phase 3: Grademap Persistence** - Store every received grademap with timestamps for audit and reconciliation (completed 2026-03-07)
- [ ] **Phase 4: Timing Reconciliation** - Match each fruit record to the correct historical grademap

## Phase Details

### Phase 1: Code Cleanup
**Goal**: The existing codebase is free of bugs that would corrupt data or block message delivery when new features are added
**Depends on**: Nothing (first phase)
**Requirements**: QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05
**Success Criteria** (what must be TRUE):
  1. The paho MQTT dispatcher never stalls — `OnMessage` returns immediately whether the fruit queue is full or empty
  2. A malformed fruit JSON payload is logged with its raw content and discarded; no zero-value fruit row is written to the database
  3. On a clean shutdown, all fruit records queued at signal time are flushed to the database before the connection pool closes
  4. The fruit counter in `Processor` accurately reflects the number of records available to read (no race between counter and channel)
**Plans**: 3 plans

Plans:
- [ ] 01-01-PLAN.md — Write failing unit test stubs for QUAL-01, QUAL-02, QUAL-03 (Nyquist wave 0)
- [ ] 01-02-PLAN.md — Fix OnMessage blocking send, UnmarshalJSON error propagation, Values() counter race, shutdown defer order, and extend Config
- [ ] 01-03-PLAN.md — Remove dead code, run go mod tidy, write code review note

### Phase 2: Write Buffer and Reconnect
**Goal**: Fruit records are never lost during a PostgreSQL outage — they buffer in memory and flush automatically when the connection is restored
**Depends on**: Phase 1
**Requirements**: RECO-01, RECO-02, RECO-03, RECO-04
**Success Criteria** (what must be TRUE):
  1. When the database goes down, fruit records continue to accumulate in an in-memory buffer up to the configured limit; no records are silently dropped below that limit
  2. When the buffer reaches its configured maximum, the oldest records are dropped and a warning is logged per dropped record
  3. When the database comes back up, all buffered records are automatically flushed to TimescaleDB without operator intervention
  4. A batch write that fails mid-copy leaves no partial rows in the database — the batch either commits fully or is retried intact
**Plans**: 3 plans

Plans:
- [ ] 02-01-PLAN.md — Write failing test stubs for RECO-01 through RECO-04 and DrainFruits (Nyquist wave 0)
- [ ] 02-02-PLAN.md — Add processor.DrainFruits to expose []Fruit from internal queue
- [ ] 02-03-PLAN.md — Implement write buffer, reconnect probe, and transactional CopyFrom in breeze.go

### Phase 3: Grademap Persistence
**Goal**: Every grademap received from MQTT is stored in TimescaleDB with its receipt timestamp and a queryable change log, forming the foundation for timing reconciliation and audit queries
**Depends on**: Phase 1
**Requirements**: GRAD-01, GRAD-02, GRAD-03
**Success Criteria** (what must be TRUE):
  1. Every MQTT grademap message results in a row in `breeze_grademap` with the payload and the timestamp at which Breeze received it
  2. For each grademap update, `breeze_grademap_changes` contains one row per changed field showing the old and new values — enabling queries like "show all threshold changes this month"
  3. Grademap writes do not block MQTT message delivery — `OnMessage` returns immediately regardless of database latency
**Plans**: 3 plans

Plans:
- [ ] 03-01-PLAN.md — Write failing test stubs for GRAD-01, GRAD-02, GRAD-03 (Nyquist wave 0)
- [ ] 03-02-PLAN.md — Implement DiffGrademaps and add gradeCh channel to Processor
- [ ] 03-03-PLAN.md — Add schema DDL, define Go constants, wire gradeCh into run() with INSERT and CopyFrom

### Phase 4: Timing Reconciliation
**Goal**: Each fruit record is attributed to the exact historical grademap that the OEM used to grade it, accounting for the OEM's pipeline delay between publishing a new grademap and applying it to fruit in transit
**Depends on**: Phase 3
**Requirements**: TIME-01, TIME-02
**Success Criteria** (what must be TRUE):
  1. Each fruit row in TimescaleDB is linked to the `breeze_grademap` row that was active when the OEM graded it — not simply the most recently received grademap
  2. A configurable `grademap_propagation_delay` offset can be set in the YAML config to compensate for OEM pipeline lag; setting it to zero disables the offset with no behavioral change
**Plans**: 3 plans

Plans:
- [ ] 04-01-PLAN.md — Write failing test stubs for TIME-01 and TIME-02 (Nyquist wave 0)
- [ ] 04-02-PLAN.md — Implement resolveGrademapID, add GrademapPropagationDelayStr to Config, add GrademapID to Fruit struct
- [ ] 04-03-PLAN.md — Wire delay into writeBatch/flushBuffer/run(), add grademap_id DDL to schema.sql

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

**Note on Phase 4:** Investigation complete (2026-03-08) — Path B confirmed. No grademap version field in fruit payload. Implementation uses configurable delay offset for AS-OF timestamp lookups.

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Code Cleanup | 3/3 | Complete   | 2026-03-07 |
| 2. Write Buffer and Reconnect | 3/3 | Complete   | 2026-03-07 |
| 3. Grademap Persistence | 3/3 | Complete   | 2026-03-07 |
| 4. Timing Reconciliation | 2/3 | In Progress|  |
