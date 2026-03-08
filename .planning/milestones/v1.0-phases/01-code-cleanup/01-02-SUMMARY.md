---
phase: 01-code-cleanup
plan: 02
subsystem: processor
tags: [go, mqtt, paho, channels, atomic, slog, yaml, envconfig]

# Dependency graph
requires:
  - phase: 01-code-cleanup
    plan: 01
    provides: "Failing tests for QUAL-01, QUAL-02, QUAL-03 bugs; grademap and fruit processor structure"
provides:
  - "Non-blocking OnMessage with select/drop pattern (QUAL-01)"
  - "UnmarshalJSON propagates inner error (QUAL-03 prerequisite)"
  - "Malformed JSON fruit payloads discarded and logged (QUAL-03)"
  - "Grademap update errors logged (bonus)"
  - "Values() reads queue before decrementing counter (QUAL-02)"
  - "Create() accepts configurable queueSize int parameter"
  - "Config carries LogLevel and QueueSize fields with yaml/envconfig tags"
  - "Log level wired to slog.SetLogLoggerLevel from config/env"
  - "QUAL-04: defer client.Disconnect fires before close(next_batch) before pool.Close()"
affects:
  - 02-write-buffering
  - 03-observability

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Non-blocking channel send with select/default — prevents MQTT dispatcher stall"
    - "Read-from-channel then decrement counter — prevents QUAL-02 window"
    - "LIFO defer ordering: Disconnect(250) -> close(next_batch) -> pool.Close()"
    - "QueueSize defaulting pattern: if cfg.QueueSize == 0 { cfg.QueueSize = 50 }"

key-files:
  created: []
  modified:
    - processor/processor.go
    - processor/fruit.go
    - breeze.go

key-decisions:
  - "Error from UnmarshalJSON propagated via fmt.Errorf wrapping — enables QUAL-03 check in OnMessage"
  - "QueueSize default 50 applied in run() before Create() call — preserves existing behavior when unconfigured"
  - "slog.SetLogLoggerLevel used for log level wiring — matches existing slog usage in codebase"

patterns-established:
  - "Non-blocking queue send: select { case p.queue <- f: ... ; default: slog.Warn(Dropped) }"
  - "Configurable Create() signature: Create(next chan bool, queueSize int)"

requirements-completed: [QUAL-01, QUAL-02, QUAL-03, QUAL-04]

# Metrics
duration: 12min
completed: 2026-03-08
---

# Phase 1 Plan 02: Correctness Bug Fixes Summary

**Four MQTT/queue correctness bugs fixed: non-blocking OnMessage, counter-after-read Values(), malformed JSON discard, and correct defer shutdown order — plus QueueSize and LogLevel config extension**

## Performance

- **Duration:** ~12 min
- **Started:** 2026-03-08T02:18:00Z
- **Completed:** 2026-03-08T02:30:00Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- QUAL-01: Replaced blocking `p.queue <- f` with non-blocking select/default; dropped fruits emit a Warn log
- QUAL-03: Added error check on json.Unmarshal in OnMessage; malformed payloads are discarded (not enqueued as zero-value fruit)
- QUAL-02: Values() now reads from channel first then decrements counter, eliminating the ordering race window
- QUAL-04: Registered `defer client.Disconnect(250)` immediately after MQTT connect; removed inline call from ctx.Done() case — LIFO order is now Disconnect -> close(next_batch) -> pool.Close()
- UnmarshalJSON in fruit.go now propagates inner error via fmt.Errorf wrapping (prerequisite for QUAL-03 to work correctly)
- Bonus: grademap.Update errors are now logged (were silently dropped)
- Config extended with LogLevel and QueueSize fields; log level wired to slog.SetLogLoggerLevel; QueueSize defaulted to 50 if unset

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix OnMessage blocking send and UnmarshalJSON error propagation (QUAL-01, QUAL-03)** - `6a970ab` (fix)
2. **Task 2: Fix Values() counter race and extend Config with QueueSize + LogLevel (QUAL-02, QUAL-04)** - `59d58e9` (fix)

## Files Created/Modified
- `processor/processor.go` - Non-blocking OnMessage select, error-checked json.Unmarshal, grademap error check, simplified Values(), Create(queueSize int)
- `processor/fruit.go` - UnmarshalJSON wraps inner error with fmt.Errorf; added "fmt" import
- `breeze.go` - Config extended with LogLevel/QueueSize, log level wiring, QueueSize default + Create() wiring, QUAL-04 defer fix

## Decisions Made
- Error propagation from UnmarshalJSON was required before QUAL-03 check could work — the two fixes were kept in the same task commit as the plan specified
- `p.timer <- true` kept strictly inside the `case p.queue <- f:` branch — prevents spurious batch write cycles on dropped fruit
- `fmt` import removed from processor.go after Values() simplification eliminated the only fmt usage

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated breeze.go Create() call site during Task 1 to pass queueSize**
- **Found during:** Task 1 (after changing Create() signature to accept queueSize)
- **Issue:** breeze.go called `processor.Create(next_batch)` with one argument; build failed with "not enough arguments"
- **Fix:** Temporarily updated call to `processor.Create(next_batch, 50)` so Task 1 build verification could pass; Task 2 then wired `cfg.QueueSize` properly
- **Files modified:** breeze.go
- **Verification:** `go build ./...` exits 0 after each task
- **Committed in:** 6a970ab (Task 1 commit), superseded by 59d58e9 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (Rule 3 — blocking compilation error)
**Impact on plan:** Necessary intermediate step; Task 2 completed the intended Config wiring. No scope creep.

## Issues Encountered
- `TestFromJson` panics with `reflect.Value.Equal: values of type map[string]interface{} are not comparable` — this is a pre-existing bug in `ValidateFruit()` unrelated to this plan's changes (confirmed by testing against commit before Task 1). Logged to deferred items. The plan's target tests (TestOnMessage_QueueFull, TestOnMessage_MalformedJSON, TestValues_CounterOrder) all pass GREEN.

## Next Phase Readiness
- All four correctness bugs eliminated; codebase is safe for Phase 2 write buffering
- QueueSize configurable via YAML/env, ready for production tuning
- Log level configurable via YAML/env, ready for operational use
- Pre-existing TestFromJson panic in ValidateFruit should be fixed before Phase 2 (deferred)

---
*Phase: 01-code-cleanup*
*Completed: 2026-03-08*
