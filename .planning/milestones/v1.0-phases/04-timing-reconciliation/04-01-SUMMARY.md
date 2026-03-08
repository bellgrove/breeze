---
phase: 04-timing-reconciliation
plan: "01"
subsystem: testing
tags: [tdd, wave0, nyquist, grademap, timing, go]

# Dependency graph
requires:
  - phase: 03-grademap-persistence
    provides: writeBatch, writeGrademap, grademapDDL constants that stubs reference
provides:
  - Wave 0 failing test stubs for TIME-01 (resolveGrademapID) and TIME-02 (configurable delay offset)
  - Nyquist compliance scaffold for Phase 4 implementation plans
affects:
  - 04-02 (implements resolveGrademapID and GrademapPropagationDelayStr — stubs turn GREEN)
  - 04-03 (updates writeBatch signature — TestWriteBatch_GrademapID turns GREEN)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Wave 0 TDD: write failing stubs before any implementation exists"
    - "Closed-pool pattern for exercising error paths without a real DB"

key-files:
  created: []
  modified:
    - breeze_test.go

key-decisions:
  - "TestResolveGrademapID_NoRows asserts (0, false, nil) OR non-nil error — covers both ErrNoRows and generic pool-error branches without locking in internal details"
  - "TestConfig_PropagationDelayZero handles empty string as zero without calling ParseDuration — matches intended caller logic where empty string means no offset"

patterns-established:
  - "Wave 0 stub pattern: test function references undefined symbol → compile fails RED → correct pre-implementation state"
  - "Closed-pool for error-path tests: pgxpool.New with invalid DSN then pool.Close() — consistent with existing TestTransactionalWrite"

requirements-completed: [TIME-01, TIME-02]

# Metrics
duration: 2min
completed: 2026-03-08
---

# Phase 4 Plan 01: Timing Reconciliation — Wave 0 Test Stubs Summary

**Seven failing Wave 0 test stubs for AS-OF grademap lookup (TIME-01) and configurable propagation delay (TIME-02) appended to breeze_test.go using the closed-pool pattern**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-08T00:01:56Z
- **Completed:** 2026-03-08T00:03:39Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Four TIME-01 stubs: `TestResolveGrademapID_NoRows`, `TestResolveGrademapID_ZeroDelay`, `TestResolveGrademapID`, `TestWriteBatch_GrademapID`
- Three TIME-02 stubs: `TestConfig_PropagationDelay`, `TestConfig_PropagationDelayZero`, `TestResolveGrademapID_WithDelay`
- Production code (`go build ./...`) compiles clean; test compilation is RED as required by Nyquist compliance

## Task Commits

Each task was committed atomically:

1. **Task 1: Wave 0 stubs for resolveGrademapID (TIME-01)** - `9c69f96` (test)
2. **Task 2: Wave 0 stubs for config delay and delay offset (TIME-02)** - `6fe7cdb` (test)

**Plan metadata:** TBD (docs: complete plan)

## Files Created/Modified
- `breeze_test.go` - Appended 7 new test functions (154 lines added across two commits)

## Decisions Made
- `TestResolveGrademapID_NoRows` accepts both `(0, false, nil)` and non-nil error outcomes — the closed-pool produces a generic connect error (not `pgx.ErrNoRows`) so locking in only the NoRows path would make the stub fragile and misleading before the implementation exists.
- `TestConfig_PropagationDelayZero` handles empty string without calling `time.ParseDuration` — mirrors the expected caller logic where empty string is treated as "no delay configured" and skips parsing entirely.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Wave 0 scaffold complete; Plan 02 can add `resolveGrademapID` and `Config.GrademapPropagationDelayStr` and these stubs will turn GREEN
- Plan 03 can update `writeBatch` signature to accept `delay time.Duration` and `TestWriteBatch_GrademapID` turns GREEN
- Nyquist compliance for Phase 4 fully satisfied

---
*Phase: 04-timing-reconciliation*
*Completed: 2026-03-08*
