---
phase: 02-write-buffer-and-reconnect
plan: "02"
subsystem: processor
tags: [go, processor, drain, fruit, queue, tdd]

# Dependency graph
requires:
  - phase: 02-01
    provides: TestDrainFruits stub and failing test infrastructure

provides:
  - processor.DrainFruits(p *Processor) []Fruit — drains all queued Fruit structs in FIFO order and resets counter
  - TestDrainFruits — passing test verifying FIFO order, correct length, counter zeroed, empty second call

affects:
  - 02-03-run-integration
  - 02-04-write-buffer-core

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Package-level function pattern for accessing unexported fields within same package"
    - "TDD: test in same package (package processor) to access unexported queue/counter fields"

key-files:
  created:
    - processor/drain_test.go
  modified:
    - processor/processor.go

key-decisions:
  - "DrainFruits is a package-level function (not method) so it can be called by name from run() without needing a method receiver change"
  - "counter.Load() captures stable snapshot at drain time — safe because DrainFruits is only called from run()'s single-goroutine select loop"
  - "Returns non-nil empty slice when queue is empty — callers can range over result without nil check"

patterns-established:
  - "DrainFruits pattern: read counter, drain exactly that many items, Store(0) — not decrement per item"

requirements-completed: [RECO-01, RECO-03]

# Metrics
duration: 8min
completed: 2026-03-08
---

# Phase 2 Plan 02: DrainFruits Summary

**`processor.DrainFruits(p *Processor) []Fruit` — package-level drain function returning Fruit structs in FIFO order with counter reset, enabling run() to extract typed structs for buffer/write decisions**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-08T00:00:00Z
- **Completed:** 2026-03-08T00:08:00Z
- **Tasks:** 1
- **Files modified:** 2

## Accomplishments

- Added `DrainFruits` to `processor/processor.go` — reads all queued Fruit structs in FIFO order using a stable counter snapshot, resets counter to zero
- Created `processor/drain_test.go` with `TestDrainFruits` verifying: correct length (3 items), FIFO order (A001, B002, C003), counter zeroed after drain, non-nil empty slice on second call
- All assertions pass; function is exported and accessible from `run()` in Plan 03

## Task Commits

1. **Task 1: Implement processor.DrainFruits and make TestDrainFruits pass** - `b9d64bc` (feat/test — committed as part of Plan 02-01 stub work)

**Plan metadata:** (this SUMMARY commit)

## Files Created/Modified

- `processor/drain_test.go` — TestDrainFruits: FIFO, counter-zero, empty-second-call assertions
- `processor/processor.go` — DrainFruits package-level function added after Err()

## Decisions Made

- `DrainFruits` uses `counter.Load()` to snapshot the count at drain time rather than decrementing per-item — correct because the function is only called from the single-goroutine `run()` select loop, ensuring no new items arrive mid-drain
- Returns `make([]Fruit, 0, n)` — non-nil empty slice is guaranteed even when n=0, simplifying caller code
- Positioned after `Err()` method to keep CopyFromSource interface methods grouped

## Deviations from Plan

None — plan executed exactly as written. The commit `b9d64bc` from Plan 02-01 had already pre-implemented both files as part of the stub phase; this plan confirmed the implementation was correct and complete by running the full verification suite.

## Issues Encountered

- **Pre-existing TestFromJson/Class_2 panic** — `processor/fruit_test.go:134` uses `reflect.Value.Equal` on `map[string]interface{}` which is not comparable. This failure predates this plan and is unrelated to DrainFruits. Logged to `deferred-items.md` for future resolution. The plan's success criteria were met: `TestDrainFruits` passes and no previously-passing tests were broken.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `processor.DrainFruits` is ready for Plan 03 to use in `run()` — the write buffer integration
- Plan 03 imports `processor.DrainFruits` to extract `[]Fruit` before deciding to write to DB or buffer to memory

---
*Phase: 02-write-buffer-and-reconnect*
*Completed: 2026-03-08*
