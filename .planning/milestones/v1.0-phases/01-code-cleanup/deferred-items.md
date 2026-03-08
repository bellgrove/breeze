# Deferred Items — Phase 01-code-cleanup

Items discovered during execution that are out-of-scope for the current plan.

## Pre-existing TestFromJson panic

**Discovered during:** 01-02 execution (Task 2 verification)
**File:** processor/fruit_test.go, line 138 — `ValidateFruit()` function
**Issue:** `reflect.Value.Equal` panics when called on `map[string]interface{}` values (maps are not comparable with reflect.Value.Equal)
**Impact:** `go test ./...` exits non-zero; only TestFromJson is affected; the three plan target tests (TestOnMessage_QueueFull, TestOnMessage_MalformedJSON, TestValues_CounterOrder) pass
**Confirmed pre-existing:** Tested against commit before plan 01-02 changes — same panic
**Suggested fix:** Replace `field1.Equal(field2)` with `reflect.DeepEqual(field1.Interface(), field2.Interface())` in the else branch of ValidateFruit(), or skip map fields in the comparison loop
**Priority:** Fix before Phase 2 — full `go test ./...` green is a quality gate
