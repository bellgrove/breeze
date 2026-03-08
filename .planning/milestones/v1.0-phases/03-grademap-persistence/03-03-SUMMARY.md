---
phase: 03-grademap-persistence
plan: "03"
subsystem: database
tags: [postgres, pgx, jsonb, grademap, schema, ddl]

# Dependency graph
requires:
  - phase: 03-grademap-persistence-01
    provides: TestGrademapSchema stub (RED), grademap DDL requirements
  - phase: 03-grademap-persistence-02
    provides: GradeCh() channel on Processor, DiffGrademaps() function, GrademapChange struct
provides:
  - grademapDDL and grademapChangesDDL constants in breeze.go
  - writeGrademap helper (INSERT RETURNING id + DiffGrademaps + CopyFrom changes)
  - gradeCh case in run() select loop (drop-on-unhealthy, self-correcting diff baseline)
  - schema.sql DDL for breeze_grademap and breeze_grademap_changes tables
affects:
  - Phase 04 (reads from breeze_grademap/breeze_grademap_changes for analysis)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "INSERT RETURNING id pattern for single-row inserts via pgxpool.QueryRow().Scan()"
    - "CopyFrom for batch change rows (mirrors fruit write pattern)"
    - "Self-correcting diff baseline: prevGrademapPayload only updated on successful INSERT"
    - "Drop-on-unhealthy: grademap payloads dropped (not buffered) when dbHealthy=false"
    - "CopyFrom failures for change rows logged but non-fatal (audit log, not data)"

key-files:
  created: []
  modified:
    - schema.sql
    - breeze.go

key-decisions:
  - "prevGrademapPayload not updated on INSERT failure — diff self-corrects on next successful write"
  - "CopyFrom error for breeze_grademap_changes is non-fatal — grademap row was written, diff self-corrects"
  - "Grademap payloads dropped (not second-buffered) when dbHealthy=false — consistent with design"

patterns-established:
  - "writeGrademap: INSERT-then-diff-then-CopyFrom without transaction wrapping (audit log pattern)"
  - "Prev-state tracking: package-local var updated only on confirmed success"

requirements-completed: [GRAD-01, GRAD-02, GRAD-03]

# Metrics
duration: 4min
completed: 2026-03-08
---

# Phase 3 Plan 3: Grademap Persistence Summary

**End-to-end grademap persistence: schema DDL, Go constants, writeGrademap helper with INSERT/diff/CopyFrom, and gradeCh case wired into run() select loop**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-03-08T07:17:15Z
- **Completed:** 2026-03-08T07:20:16Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Added `breeze_grademap` and `breeze_grademap_changes` DDL to schema.sql with appropriate indexes
- Defined `grademapDDL` and `grademapChangesDDL` package-level constants in breeze.go — TestGrademapSchema turns GREEN
- Implemented `writeGrademap` helper: single-row INSERT RETURNING id, DiffGrademaps comparison, CopyFrom change rows
- Added `case payload := <-proc.GradeCh()` to run() select loop with drop-on-unhealthy and self-correcting prev baseline

## Task Commits

Each task was committed atomically:

1. **Task 1: Add DDL to schema.sql and define Go constants** - `0f87bf9` (feat)
2. **Task 2: Add writeGrademap helper and gradeCh case to run()** - `cb48405` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `schema.sql` - Appended CREATE TABLE breeze_grademap and breeze_grademap_changes with indexes
- `breeze.go` - Added grademapDDL/grademapChangesDDL constants, encoding/json import, writeGrademap helper, prevGrademapPayload var, gradeCh case in run()

## Decisions Made
- `prevGrademapPayload` is not updated on INSERT failure — the diff self-corrects on the next successful write, ensuring the change log stays accurate even after transient DB errors
- CopyFrom errors for `breeze_grademap_changes` are logged but non-fatal — the grademap row was committed, so the change log is advisory (audit log, not data integrity)
- Grademap payloads are dropped (not buffered) when dbHealthy=false — consistent with the existing design decision; no second buffer for grademap updates

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

Pre-existing `TestFromJson` panic in `processor/fruit_test.go` (ValidateFruit uses reflect.Value.Equal on map[string]any fields, which panics). This was already failing before Plan 03-03 and is logged in deferred-items.md. Not caused by this plan's changes.

## Next Phase Readiness
- All three GRAD requirements (GRAD-01, GRAD-02, GRAD-03) are now complete
- Phase 03 grademap persistence is fully integrated: MQTT channel -> DB write -> change diff -> audit log
- Phase 4 can query breeze_grademap and breeze_grademap_changes for analysis
- Blocker remains: Phase 4 requires live OEM traffic inspection to determine grademap version field presence (Path A vs Path B)

---
*Phase: 03-grademap-persistence*
*Completed: 2026-03-08*
