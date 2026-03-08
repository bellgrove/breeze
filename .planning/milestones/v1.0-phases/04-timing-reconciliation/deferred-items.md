# Deferred Items — Phase 04 Timing Reconciliation

## Pre-existing Failures (Out of Scope)

### TestFromJson/Class_2 — processor package panic

**Discovered:** During 04-03 full test suite run
**Package:** `github.com/bellgrove/breeze/processor`
**Test:** `TestFromJson/Class_2` (and possibly other subtests)
**Error:** `panic: reflect.Value.Equal: values of type map[string]interface {} are not comparable`
**Location:** `processor/fruit_test.go:134` in `ValidateFruit`

**Nature:** Pre-existing failure confirmed — reproduced against HEAD before any 04-03 changes.
The test uses `reflect.Value.Equal` on `map[string]interface{}` which is not comparable.

**Impact:** Does not affect Phase 4 functionality. All Phase 4 tests in the main package pass.

**Suggested fix:** Replace `reflect.Value.Equal` with `reflect.DeepEqual` in `ValidateFruit` helper,
or restructure the assertion to compare marshalled JSON instead of raw map values.
