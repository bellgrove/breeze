---
phase: 01-code-cleanup
plan: "03"
subsystem: processor, breeze
tags: [cleanup, dead-code, dependencies, code-review]
dependency_graph:
  requires: [01-02]
  provides: [clean-source-files, go-mod-tidy, code-review-note]
  affects: [go.mod, go.sum, processor/processor.go, processor/grademap.go, processor/fruit.go, processor/fruit_test.go, breeze.go]
tech_stack:
  added: []
  patterns: [pure-deletion, go-mod-tidy]
key_files:
  created:
    - .planning/phases/01-code-cleanup/01-CODE-REVIEW.md
  modified:
    - processor/processor.go
    - processor/grademap.go
    - processor/fruit.go
    - processor/fruit_test.go
    - breeze.go
    - go.mod
    - go.sum
decisions: []
metrics:
  duration: "4 minutes"
  completed: "2026-03-08"
  tasks: 2
  files: 8
---

# Phase 1 Plan 03: Dead Code Removal and Code Review Note Summary

**One-liner:** Removed all commented dead code from five source files, purged gammazero/deque from go.mod via `go mod tidy`, and wrote a code review note documenting five deferred structural issues.

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Remove dead code from all source files and run go mod tidy | 0ced037 | processor/processor.go, processor/grademap.go, processor/fruit.go, processor/fruit_test.go, breeze.go, go.mod, go.sum |
| 2 | Write code review note documenting deferred structural issues | df1f649 | .planning/phases/01-code-cleanup/01-CODE-REVIEW.md |

## What Was Done

### Task 1: Dead Code Removal

Performed a pure deletion pass across all five files listed in the plan:

- **processor/processor.go:** Removed two commented struct fields (`dequeue`, `current`), two commented lines in `Create()` body (`queue := new(deque.Deque[Fruit])`, `queue.SetBaseCap(100)`), and one commented struct literal field (`Fruit{}`). The `gammazero/deque` import was not present in the import block (it was already absent from the live import list — only the struct fields referenced it by comment).
- **processor/grademap.go:** Removed the 10-line commented vision grade optimisation block inside `Grade()` (the `A -> 3`, `B -> 2`, `idx` logic), and the single commented mutation line `gm.Defect_grading_passes[k1] = v1`.
- **processor/fruit.go:** Removed the debug test line `// a.OtherDefects = []string{"ABCD"}`.
- **processor/fruit_test.go:** Removed the commented old API call `// got, err := FromJson(tt.args.val)` and the three-line commented assertion block `// if !reflect.DeepEqual(got, tt.want)`. Left `// TODO: Add test cases.` in place per plan.
- **breeze.go:** Removed five dead comment locations — `// slog.SetLogLoggerLevel(slog.LevelDebug)` in `main()`, `// conn, err := pgx.Connect(...)` in `run()`, `// config.MaxConnIdleTime =` in `run()`, `// slog.Info("New write")` in the batch write case, and the three-line `// default:` polling block.

After all deletions, `go mod tidy` was run. This removed `github.com/gammazero/deque v1.0.0` from `go.mod` and its hash from `go.sum`.

### Task 2: Code Review Note

Created `.planning/phases/01-code-cleanup/01-CODE-REVIEW.md` documenting five deferred issues:

1. **Timer channel blocking** (high) — `p.timer <- true` is an unguarded blocking send that can stall the MQTT receive goroutine under load.
2. **TestFromJson panic** (medium) — `ValidateFruit` uses `reflect.Value.Equal` on `map[string]interface{}` fields, which panics; causes `go test ./...` to always fail for the processor package.
3. **Incomplete vision_value mapping** (low) — sizer-level `vision_value` ignored in favour of spectrim `vision_grade_value`; values may diverge.
4. **snake_case exported struct fields** (low) — `GmCh.Defect_grading_pass_index` etc. violate Go naming conventions.
5. **os.Exit inside run()** (medium) — bypasses deferred cleanup in caller and cannot be tested.

## Verification Results

- `go build ./...` — PASS
- `go test ./...` — pre-existing TestFromJson panic (out of scope, documented in 01-CODE-REVIEW.md)
- `grep gammazero go.mod` — no matches (PASS)
- Dead comment grep — no matches (PASS)
- `01-CODE-REVIEW.md` — exists, contains all required issues

## Deviations from Plan

None — plan executed exactly as written. The pre-existing TestFromJson failure was anticipated in the plan and is documented in the code review note.

## Self-Check: PASSED

- `processor/processor.go` — verified (clean)
- `processor/grademap.go` — verified (clean)
- `processor/fruit.go` — verified (clean)
- `processor/fruit_test.go` — verified (clean)
- `breeze.go` — verified (clean)
- `go.mod` — gammazero absent
- `.planning/phases/01-code-cleanup/01-CODE-REVIEW.md` — exists
- Task 1 commit `0ced037` — verified
- Task 2 commit `df1f649` — verified
