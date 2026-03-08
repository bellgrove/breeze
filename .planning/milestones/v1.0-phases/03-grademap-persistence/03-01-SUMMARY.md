---
phase: 03-grademap-persistence
plan: 01
subsystem: testing
tags: [go, tdd, grademap, postgres, ddl]

# Dependency graph
requires:
  - phase: 02-write-buffer-and-reconnect
    provides: DrainFruits, Processor queue infrastructure
provides:
  - GrademapChange struct and DiffGrademaps stub in processor package
  - TestDiffGrademaps (RED) covering nil prev, field change, and no-change cases
  - TestOnMessage_GrademapChannel (RED) verifying non-blocking grademap delivery via GradeCh()
  - TestGrademapSchema (RED) asserting DDL constants for breeze_grademap and breeze_grademap_changes tables
affects:
  - 03-02 (implements DiffGrademaps — makes TestDiffGrademaps GREEN)
  - 03-03 (defines grademapDDL constants — makes TestGrademapSchema GREEN)
  - 03-04 (adds GradeCh to Processor — makes TestOnMessage_GrademapChannel GREEN)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Nyquist wave-0 TDD: write failing stubs before implementation so every plan has verification from the start
    - Table-driven tests for diff logic with named cases covering nil-prev, mutation, and identity cases

key-files:
  created:
    - processor/grademap_diff.go
    - processor/grademap_diff_test.go
  modified:
    - processor/processor_test.go
    - breeze_test.go

key-decisions:
  - "DiffGrademaps takes map[string]any (not Grademap struct) to remain decoupled from the grademap model; implementation can evolve without changing the diff signature"
  - "TestOnMessage_GrademapChannel placed in processor_test.go (not a new file) to co-locate with existing OnMessage tests and share mockMsg/newTestProcessorSmallQueue helpers"

patterns-established:
  - "RED stubs: export the type/function first (compiles), then add tests referencing not-yet-existing methods (compile failure = RED)"

requirements-completed:
  - GRAD-01
  - GRAD-02
  - GRAD-03

# Metrics
duration: 3min
completed: 2026-03-08
---

# Phase 03 Plan 01: Grademap Persistence Test Stubs Summary

**Nyquist wave-0 failing test stubs for GrademapChange diff, non-blocking GradeCh channel, and grademap DDL schema constants — all RED, ready for implementation plans**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-03-07T23:07:58Z
- **Completed:** 2026-03-07T23:10:13Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Created `GrademapChange` struct and `DiffGrademaps` nil-stub in `processor/grademap_diff.go`
- Three-case table-driven `TestDiffGrademaps` (FAIL RED: nil stub returns no changes when changes expected)
- `TestOnMessage_GrademapChannel` appended to `processor/processor_test.go` (compile error: `GradeCh` undefined on `Processor`)
- `TestGrademapSchema` appended to `breeze_test.go` (compile error: `grademapDDL`/`grademapChangesDDL` undefined)

## Task Commits

1. **Task 1: Create grademap_diff.go stub and failing TestDiffGrademaps** - `312212c` (test)
2. **Task 2: Add TestOnMessage_GrademapChannel stub to processor_test.go** - `8d55613` (test)
3. **Task 3: Add TestGrademapSchema stub to breeze_test.go** - `de8de59` (test)

## Files Created/Modified

- `processor/grademap_diff.go` - GrademapChange struct + DiffGrademaps stub returning nil
- `processor/grademap_diff_test.go` - Table-driven TestDiffGrademaps with 3 cases
- `processor/processor_test.go` - TestOnMessage_GrademapChannel appended (references GradeCh — compile RED)
- `breeze_test.go` - TestGrademapSchema appended (references grademapDDL/grademapChangesDDL — compile RED)

## Decisions Made

- DiffGrademaps signature uses `map[string]any` (not Grademap struct) to remain decoupled from the grademap model; avoids creating a dependency between the diff logic and future schema changes.
- TestOnMessage_GrademapChannel placed in existing `processor_test.go` to share `mockMsg` and `newTestProcessorSmallQueue` helpers without an additional file.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All three RED stubs in place; Plans 02 and 03 can proceed in any order
- Plan 02 must add `GradeCh() chan []byte` method to `Processor` and initialize a buffered channel in `newTestProcessorSmallQueue` to make `TestOnMessage_GrademapChannel` GREEN
- Plan 03 must define `grademapDDL` and `grademapChangesDDL` string constants to make `TestGrademapSchema` GREEN
- Plan 02 must implement `DiffGrademaps` logic to make `TestDiffGrademaps` GREEN

---
*Phase: 03-grademap-persistence*
*Completed: 2026-03-08*
