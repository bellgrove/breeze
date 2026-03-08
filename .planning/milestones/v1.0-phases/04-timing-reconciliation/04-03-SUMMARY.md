---
phase: 04-timing-reconciliation
plan: "03"
subsystem: database
tags: [postgres, pgx, go, grademap, timing, fk]

# Dependency graph
requires:
  - phase: 04-01
    provides: "Test stubs and Fruit.GrademapID field, GrademapPropagationDelayStr config"
  - phase: 04-02
    provides: "resolveGrademapID function, GrademapID in Columns() and AsRow()"
provides:
  - "writeBatch(ctx, pool, batch, delay) resolves grademap FK before each CopyFrom batch"
  - "flushBuffer(ctx, pool, buf, delay) resolves grademap FK before flushing buffered records"
  - "run() parses GrademapPropagationDelayStr at startup and passes propagationDelay to all write call sites"
  - "schema.sql ALTER TABLE migration adding nullable grademap_id FK to breeze_fruit"
affects: [production deployment, analytics joins on grademap_id]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Batch-level grademap resolution: resolve once per batch using min(SizerTime) as anchor, not per-row"
    - "Non-fatal grademap resolution: log warning and write NULL on failure; fruit data is never lost"
    - "Zero-delay convention: empty string in config yields zero time.Duration — no anchor adjustment"

key-files:
  created: []
  modified:
    - breeze.go
    - schema.sql

key-decisions:
  - "writeBatch and flushBuffer resolve grademap ID once per call using min(SizerTime) as conservative anchor — avoids unsupported per-row queries inside pgx CopyFrom"
  - "grademap resolution failure is non-fatal — fruit written with NULL grademap_id rather than dropped"
  - "schema.sql column is NULLABLE with no DEFAULT — safe for existing deployments; rows before grademap history correctly get NULL"
  - "propagationDelay parsed in run() not in writeBatch — single parse at startup, zero value is valid (no adjustment)"

patterns-established:
  - "Batch-anchor pattern: derive a single timestamp anchor from the batch before the CopyFrom, use it for all DB lookups related to that batch"

requirements-completed: [TIME-01, TIME-02]

# Metrics
duration: 8min
completed: 2026-03-08
---

# Phase 4 Plan 03: Timing Reconciliation Wiring Summary

**writeBatch and flushBuffer updated to accept a propagation delay and stamp every fruit row with a grademap_id FK resolved via AS-OF timestamp lookup against breeze_grademap**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-03-08T08:10:00Z
- **Completed:** 2026-03-08T08:18:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- `writeBatch(ctx, pool, batch, delay)` resolves grademap ID from `min(SizerTime)` across the batch, sets `GrademapID` on each fruit before `CopyFrom` — non-fatal on lookup failure (writes NULL)
- `flushBuffer(ctx, pool, buf, delay)` applies the same resolution logic for buffered records
- `run()` parses `GrademapPropagationDelayStr` at startup; empty string yields zero duration (valid no-op); invalid string returns error
- `schema.sql` gains `ALTER TABLE breeze_fruit ADD COLUMN grademap_id BIGINT REFERENCES breeze_grademap(id)` with a nullable column (safe for existing deployments) and an analytics index
- All 12 Phase 4 tests in the main package pass with `-race`; `TestWriteBatch_GrademapID` is now GREEN (previously skipped stub)

## Task Commits

1. **RED test — TestWriteBatch_GrademapID** - `a419db5` (test)
2. **Task 1: wire propagation delay into writeBatch, flushBuffer, run()** - `dfe5822` (feat)
3. **Task 2: grademap_id DDL in schema.sql** - `c68db7c` (feat)

## Files Created/Modified

- `breeze.go` — Updated `writeBatch` and `flushBuffer` signatures to accept `delay time.Duration`; added grademap resolution logic; updated `run()` to parse `GrademapPropagationDelayStr` and pass `propagationDelay` to all call sites
- `schema.sql` — Added nullable `grademap_id BIGINT REFERENCES breeze_grademap(id)` column to `breeze_fruit` with an analytics index

## Decisions Made

- **Batch-level resolution** — `resolveGrademapID` called once per batch using `min(SizerTime)` as anchor. pgx does not support additional queries inside a `CopyFrom` loop; this is the only valid approach.
- **Non-fatal on lookup failure** — grademap resolution failure logs a warning and writes `NULL`, ensuring fruit data is never lost due to a grademap DB issue.
- **NULLABLE schema column** — No `NOT NULL` constraint, no `DEFAULT` value. Rows before grademap history correctly receive `NULL`. Safe to `ALTER TABLE` on existing deployments without backfilling.
- **Zero-delay via empty string** — Empty `GrademapPropagationDelayStr` bypasses `time.ParseDuration` and yields a zero `time.Duration`, which means no anchor adjustment. This matches the caller logic in `TestConfig_PropagationDelayZero`.

## Deviations from Plan

**1. [Rule 2 - Missing Critical] Updated TestTransactionalWrite and TestFlushBuffer call sites**
- **Found during:** Task 1 (writeBatch/flushBuffer signature update)
- **Issue:** Existing tests called `writeBatch` and `flushBuffer` with 3-arg signatures; updating the signatures broke compilation of those tests
- **Fix:** Updated call sites in tests to pass `0` as delay (zero = no adjustment, semantically correct for those unit tests)
- **Files modified:** `breeze_test.go`
- **Verification:** All tests compile and pass
- **Committed in:** `dfe5822` (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 2 — missing call site updates in test file required for compilation)
**Impact on plan:** Necessary for correctness; the plan specified updating call sites in `run()` but the test call sites also needed updating. No scope creep.

## Issues Encountered

**Pre-existing failure: TestFromJson/Class_2 in processor package**

`TestFromJson/Class_2` panics with `reflect.Value.Equal: values of type map[string]interface {} are not comparable`. Confirmed pre-existing before any 04-03 changes. All Phase 4 tests in the main `breeze` package pass with `-race`. Logged to `deferred-items.md` in this phase directory.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Phase 4 (timing reconciliation) is complete. Every fruit record written by `writeBatch` or `flushBuffer` now carries a `grademap_id` FK referencing the historical grademap row active at OEM grading time.

Operator deployment steps:
1. Apply the `ALTER TABLE` migration from `schema.sql`
2. Set `GRADEMAP_PROPAGATION_DELAY` env var (e.g., `30s`) or leave empty for no offset

Pre-existing `TestFromJson/Class_2` panic in the processor package should be addressed before the next development phase.

---
*Phase: 04-timing-reconciliation*
*Completed: 2026-03-08*
