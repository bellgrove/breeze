# Deferred Items — Phase 02

## Pre-existing Issues (Out of Scope)

### TestFromJson/Class_2 Panic

- **Discovered during:** Plan 02-02, Task 1
- **Location:** `processor/fruit_test.go:134`
- **Error:** `panic: reflect.Value.Equal: values of type map[string]interface {} are not comparable`
- **Cause:** `ValidateFruit` uses `reflect.Value.Equal` on a `map[string]interface{}` field which is not comparable with that API.
- **Scope:** Pre-existing failure, not caused by DrainFruits changes. Confirmed by checking git stash state before our edits produced the same panic.
- **Action required:** Fix `ValidateFruit` in `fruit_test.go` to use `reflect.DeepEqual` or a map-safe comparison instead of `reflect.Value.Equal`.
