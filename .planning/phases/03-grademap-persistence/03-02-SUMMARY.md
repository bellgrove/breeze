---
phase: 03-grademap-persistence
plan: 02
subsystem: processor
tags: [go, mqtt, channels, diff, tdd]

# Dependency graph
requires:
  - phase: 03-01
    provides: GrademapChange/DiffGrademaps stub and TestDiffGrademaps/TestOnMessage_GrademapChannel failing tests
provides:
  - DiffGrademaps full implementation with nil-prev, scalar, and nested-map diff
  - Processor.gradeCh buffered channel (size 1) and GradeCh() read-only accessor
  - Non-blocking grademap payload send in OnMessage with drop-warning
affects:
  - 03-03 (run() wiring that reads from GradeCh())

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Payload copy before non-blocking channel send (paho reuses buffer after OnMessage returns)"
    - "select/default non-blocking send with slog.Warn on drop"
    - "fmt.Sprintf('%v', v) for uniform leaf value stringification in diff"
    - "Explicit container entity_type map (Characteristics→characteristic, Grades→grade, Defect_grading_passes→pass)"

key-files:
  created: []
  modified:
    - processor/grademap_diff.go
    - processor/processor.go
    - processor/processor_test.go

key-decisions:
  - "containerEntityTypes map is explicit (not heuristic) — unknown top-level map keys get entity_type 'unknown'"
  - "DiffGrademaps returns nil (not empty slice) when no changes — callers check len > 0"
  - "gradeCh initialised in both Create() and newTestProcessorSmallQueue — test helper bypasses Create()"

patterns-established:
  - "Payload copy pattern: make([]byte, len) + copy before channel send — prevents paho buffer aliasing"
  - "Channel buffer size 1 for low-frequency singleton updates (grademap updates are rare)"

requirements-completed: [GRAD-02, GRAD-03]

# Metrics
duration: 15min
completed: 2026-03-08
---

# Phase 3 Plan 2: DiffGrademaps and gradeCh Summary

**Pure diff function for grademap field-level change detection, plus Processor.gradeCh non-blocking channel for raw payload delivery using fmt.Sprintf stringification and explicit entity-type mapping**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-03-08T00:00:00Z
- **Completed:** 2026-03-08T00:15:00Z
- **Tasks:** 1 (TDD GREEN + implicit REFACTOR inline)
- **Files modified:** 3

## Accomplishments
- DiffGrademaps implemented: handles nil prev (all-new), scalar top-level changes, nested container changes (Characteristics/Grades/Defect_grading_passes), unchanged fields skipped, unknown map keys use entity_type "unknown"
- Processor.gradeCh chan []byte (size 1) added with GradeCh() read-only accessor
- OnMessage grademap branch copies payload before non-blocking channel send — prevents paho buffer aliasing
- All QUAL-01, QUAL-02, QUAL-03, drain, TestDiffGrademaps, TestOnMessage_GrademapChannel tests GREEN

## Task Commits

Each task was committed atomically:

1. **Task 1: GREEN — implement DiffGrademaps and gradeCh** - `a778a45` (feat)

**Plan metadata:** (in final docs commit)

_Note: RED was already committed by Plan 03-01. This plan started at GREEN._

## Files Created/Modified
- `processor/grademap_diff.go` - Full DiffGrademaps implementation replacing stub
- `processor/processor.go` - gradeCh field, GradeCh() method, non-blocking send in OnMessage
- `processor/processor_test.go` - newTestProcessorSmallQueue updated with gradeCh initialisation

## Decisions Made
- `containerEntityTypes` is an explicit map rather than a string-manipulation heuristic — future keys that don't match get "unknown", making gaps visible rather than silently wrong
- `DiffGrademaps` returns nil (not `[]GrademapChange{}`) on no changes — consistent with the stub and plan spec; callers use `len(changes) > 0`
- `gradeCh` initialised in `newTestProcessorSmallQueue` in addition to `Create()` because the test helper bypasses `Create()` to control queue size

## Deviations from Plan

None — plan executed exactly as written. The REFACTOR step from the plan (explicit entity_type derivation) was incorporated inline during GREEN implementation.

## Issues Encountered

**Pre-existing TestFromJson panic (deferred, out of scope):** After my changes made the processor package compile for the first time, `TestFromJson` panics because `ValidateFruit` calls `reflect.Value.Equal` on `map[string]interface{}` fields — a Go restriction. This panic predates Plan 03-02 (the package couldn't compile before, so the test couldn't run). Logged to `deferred-items.md`. Fix: replace `field1.Equal(field2)` with `reflect.DeepEqual(field1.Interface(), field2.Interface())` for non-scalar fields.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness
- `GradeCh()` is ready for Plan 03-03 `run()` wiring to read from and trigger DB persistence
- `DiffGrademaps` is ready for Plan 03-03 to call on each received grademap payload
- No blockers

---
*Phase: 03-grademap-persistence*
*Completed: 2026-03-08*
