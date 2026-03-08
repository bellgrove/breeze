---
phase: 02-write-buffer-and-reconnect
plan: "03"
subsystem: database
tags: [pgx, pgxpool, write-buffer, reconnect, exponential-backoff, state-machine]

# Dependency graph
requires:
  - phase: 02-01
    provides: RECO test stubs in breeze_test.go
  - phase: 02-02
    provides: DrainFruits package-level function in processor package

provides:
  - Write buffer state machine in run() with dbHealthy flag and probeRunning guard
  - bufferFruits helper with drop-oldest eviction policy
  - writeBatch wrapping CopyFrom in pgx.BeginTxFunc transaction
  - flushBuffer wrapping buffer flush in pgx.BeginTxFunc transaction
  - startReconnectProbe goroutine with exponential backoff (1s/2x/60s cap)
  - WriteBufferSize Config field with yaml/envconfig tags and 10,000 default
  - All four RECO tests passing (TestWriteBuffer, TestReconnectProbe, TestFlushBuffer, TestTransactionalWrite)

affects:
  - Phase 03 (any further DB resilience or observability work)
  - Phase 04 (grademap versioning - reads fruit records from DB)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "pgx.BeginTxFunc wraps all CopyFrom calls — automatic rollback on any error"
    - "Drop-oldest eviction via slice reslicing (buf = buf[1:]) at configured cap"
    - "Single-probe guard via probeRunning bool in select loop — prevents goroutine explosion"
    - "Exponential backoff: 1s initial, 2x multiplier, 60s cap, 5s ping timeout"
    - "Graceful shutdown abandons write buffer — DB may still be down"

key-files:
  created: []
  modified:
    - breeze.go
    - breeze_test.go

key-decisions:
  - "nil pool panics in pgx.BeginTxFunc — TestTransactionalWrite uses a created-then-closed pool instead of nil to exercise error path without panic"
  - "Graceful shutdown abandons buffered records — no flush attempt on ctx.Done(), acceptable per design"
  - "writeBuffer[:0] resets slice while reusing backing array — avoids allocation churn at 10k cap"

patterns-established:
  - "State machine pattern: dbHealthy + probeRunning booleans in select loop control DB resilience"
  - "BeginTxFunc pattern: all CopyFrom calls wrapped in transaction for atomicity"

requirements-completed: [RECO-01, RECO-02, RECO-03, RECO-04]

# Metrics
duration: 2min
completed: 2026-03-08
---

# Phase 2 Plan 03: Write Buffer and Reconnect State Machine Summary

**Write buffer state machine with drop-oldest eviction, transactional CopyFrom via pgx.BeginTxFunc, and exponential-backoff reconnect probe — all four RECO tests passing**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-07T20:37:37Z
- **Completed:** 2026-03-08T04:40:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Implemented write buffer state machine in run() with dbHealthy flag, probeRunning guard, and reconnectCh channel
- Added bufferFruits (drop-oldest), writeBatch (BeginTxFunc), flushBuffer (BeginTxFunc), and startReconnectProbe (exponential backoff) helpers
- Made all four RECO test stubs pass: TestWriteBuffer, TestReconnectProbe, TestFlushBuffer, TestTransactionalWrite
- Extended Config with WriteBufferSize field (yaml/envconfig tags, default 10,000)

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend Config, add helper functions, and implement state machine in breeze.go** - `32dbdab` (feat)
2. **Task 2: Implement the four breeze_test.go stubs** - `0470886` (test)

## Files Created/Modified

- `/home/ben/github.com/bellgrove/breeze/breeze.go` - WriteBufferSize Config field, four helper functions, rewritten run() select loop with state machine
- `/home/ben/github.com/bellgrove/breeze/breeze_test.go` - All four RECO test implementations replacing stubs

## Decisions Made

- `nil` pool panics inside `pgx.BeginTxFunc` (SIGSEGV on nil pool.Acquire), so TestTransactionalWrite uses a pool created against an invalid address then closed immediately — this exercises the BeginTxFunc error path without panic
- Graceful shutdown (ctx.Done()) abandons buffered records with no flush attempt — acceptable per design, DB may still be down
- `writeBuffer[:0]` on successful flush resets length while reusing backing array, avoiding GC pressure at 10k cap

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] TestTransactionalWrite: nil pool causes panic, not error return**

- **Found during:** Task 2 (TestTransactionalWrite implementation)
- **Issue:** Plan specified `writeBatch(ctx, nil, ...)` to verify error is returned. But `pgxpool.(*Pool).Acquire` dereferences the nil receiver immediately, causing SIGSEGV panic — not an error return.
- **Fix:** Created a `pgxpool.New()` against `postgres://invalid:invalid@localhost:1/doesnotexist` (invalid address), then called `pool.Close()`. The closed pool returns an error from BeginTxFunc without panicking.
- **Files modified:** `breeze_test.go`
- **Verification:** `go test -run TestTransactionalWrite -v` passes
- **Committed in:** `0470886` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - bug in test approach)
**Impact on plan:** Auto-fix was necessary to get a working test. The test intent (verify error is returned, no panic) is fully achieved by the corrected approach.

## Issues Encountered

- Pre-existing `TestFromJson/Class_2` panic in `processor/fruit_test.go` — `reflect.Value.Equal` called on `map[string]interface{}` (not comparable). This was present before any Plan 02-03 changes (confirmed via git stash). Logged in `deferred-items.md`. Does not affect RECO tests.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All four RECO requirements (RECO-01 through RECO-04) are implemented and tested
- Phase 02 write buffer and reconnect work is complete
- Pre-existing `TestFromJson` panic in processor package should be fixed before Phase 3 begins
- Binary builds cleanly (`go build ./...`), `go vet ./...` is clean

---
*Phase: 02-write-buffer-and-reconnect*
*Completed: 2026-03-08*
