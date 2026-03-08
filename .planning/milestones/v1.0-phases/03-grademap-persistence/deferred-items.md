# Deferred Items — Phase 03

## Pre-existing Issues (Out of Scope)

### TestFromJson panic in fruit_test.go

**Found during:** Plan 03-02, Task 1 (GREEN phase)
**File:** processor/fruit_test.go — ValidateFruit()
**Issue:** `reflect.Value.Equal` panics when called on `map[string]interface{}` fields (MCenterOffsets, MClassifiedBlob, MColour, etc.). The test was not runnable before Plan 03-02 (package failed to compile due to missing GradeCh), so this was not previously observed.
**Impact:** TestFromJson always panics on the "Class 2" sub-test.
**Fix needed:** Replace `field1.Equal(field2)` with `reflect.DeepEqual(field1.Interface(), field2.Interface())` for non-scalar types in ValidateFruit.
**Deferred because:** Pre-existing bug, not introduced by Plan 03-02 changes.
