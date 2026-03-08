# Phase 1 Code Review — Deferred Issues

**Date:** 2026-03-08
**Scope:** Issues observed during Phase 1 dead code removal pass. None of these were addressed in Phase 1. All are deferred to the next milestone backlog.

---

## Issue 1: Timer channel send can block under load

**Location:** `processor/processor.go`, line 45 (post-cleanup)
**Priority:** High

### Description

In `OnMessage`, after a fruit is placed on the queue via `case p.queue <- f:`, the timer is notified unconditionally:

```go
select {
case p.queue <- f:
    p.counter.Add(1)
    p.timer <- true   // <-- blocking send, no select guard
default:
    slog.Warn("Dropped fruit", ...)
}
```

The `timer` channel has a buffer of 1. If the batch goroutine has not yet consumed the previous `true` value (i.e., the first fruit has just been placed and the goroutine hasn't looped back to read timer yet), a second fruit arriving will block `OnMessage` on `p.timer <- true`. Since `OnMessage` is called from the MQTT client goroutine, this blocks the entire MQTT receive loop until the timer goroutine drains the channel.

The non-blocking `select` guard only protects the queue send — not the timer send.

### Impact

In high-throughput scenarios (rapid fruit arrivals), the MQTT client receive loop can stall momentarily while waiting to write to the timer channel. The stall resolves as soon as the batch goroutine cycles, but it creates back-pressure on the MQTT connection.

### Suggested Fix

Wrap the timer send in its own non-blocking select:

```go
select {
case p.timer <- true:
default:
    // timer already notified; batch goroutine will drain queue
}
```

This is safe because the batch goroutine polls on `timer` in a loop — if one notification is already pending, an additional one is not needed.

---

## Issue 2: TestFromJson panics on map comparison

**Location:** `processor/fruit_test.go`, `ValidateFruit` function (~line 108)
**Priority:** Medium

### Description

`ValidateFruit` uses `reflect.Value.Equal` to compare all struct fields. This panics at runtime when it encounters fields of type `map[string]interface{}` (e.g., `MCenterOffsets`, `MClassifiedBlob`, `MColour`, `MColourBlob`, `MDiameters`, `MFeatures`, `MFunction`, `MTimer`) because maps are not comparable via `reflect.Value.Equal`.

The test therefore always panics on the "Class 2" case and the processor package reports `FAIL` on every `go test` run.

This is a test bug, not a source bug — the `UnmarshalJSON` implementation itself is correct.

### Impact

`go test ./...` always fails for the processor package, masking any future real test failures. CI/CD cannot rely on the test suite passing.

### Suggested Fix

Option A — Use `reflect.DeepEqual` for map fields in `ValidateFruit`:

```go
if field1.Kind() == reflect.Map {
    if !reflect.DeepEqual(field1.Interface(), field2.Interface()) {
        err = fmt.Errorf("%v%s does not match; ", err, fieldName)
    }
    continue
}
```

Option B — Replace `ValidateFruit` entirely with a JSON round-trip comparison: marshal both `got` and `tt.want` to JSON and compare the resulting byte slices (or use `require.JSONEq` from testify).

Option C — Add a type-switch in `ValidateFruit` that handles map kinds before calling `field1.Equal`.

Any of the three options resolves the panic. Option A is the minimal change.

---

## Issue 3: Incomplete fruit JSON field — vision_value vs vision_grade_value

**Location:** `processor/fruit.go`, `UnmarshalJSON`, ~line 314
**Priority:** Low

### Description

The `VisionValue` field is populated from `breeze.Spectrim.Vision_grade_value`. However, the sizer JSON also contains a top-level `vision_value` field at `breeze.Sizer.Vision_value` (present in the test fixture as `"vision_value": 170`). The current code ignores the sizer-level `vision_value` and only stores the Spectrim value. Whether they differ in practice (e.g., one is rounded, one is raw) is not documented.

### Impact

Potential loss of data or silent discrepancy if the two values diverge for a given fruit. Low probability given both appear to be 170 in the test fixture.

### Suggested Fix

Inspect live payloads to determine whether `sizer.vision_value` and `spectrim.vision_grade_value` ever differ. If they match reliably, document the choice. If they differ, expose both fields or choose explicitly.

---

## Issue 4: snake_case field names in exported Go structs (idiomatic Go)

**Location:** `processor/grademap.go` — `GmCh`, `GmPa`, and related types
**Priority:** Low

### Description

Several exported struct fields use `snake_case` naming (e.g., `Defect_grading_pass_index`, `Display_order`, `Min_value`). Go convention is `CamelCase` for all exported identifiers. These names likely originate from JSON key matching, but since JSON tags are not used on these structs, the mapping relies on Go's case-insensitive JSON decoder matching.

### Impact

Code style inconsistency; not a correctness issue. Violates `golint`/`staticcheck` conventions. Makes the API surface look non-idiomatic.

### Suggested Fix

Add explicit `json:"..."` struct tags and rename fields to CamelCase. This is a two-step refactor that must be done carefully to avoid breaking the JSON parsing path.

---

## Issue 5: os.Exit called inside run() bypasses deferred cleanup

**Location:** `breeze.go`, `run()` function, ~lines 101-104 and 110-113
**Priority:** Medium

### Description

`run()` calls `os.Exit(1)` directly on database config parse errors and connection errors:

```go
if err != nil {
    fmt.Fprintf(os.Stderr, "Unable to parse db config: %v\n", err)
    os.Exit(1)
}
```

Calling `os.Exit` inside a function that has deferred cleanup in its caller (`defer pool.Close()`, `defer client.Disconnect(250)`) means those defers do not run. In addition, the `cancel()` defer in `main()` does not execute, leaving the context uncancelled.

### Impact

On startup failure, any resources already acquired are not cleanly released. For a startup failure this is usually inconsequential, but it sets a bad precedent and makes the function harder to test.

### Suggested Fix

Return an error from `run()` and let `main()` handle the exit:

```go
if err != nil {
    return fmt.Errorf("unable to parse db config: %w", err)
}
```

---

## Closing Note

All issues above are deferred to the next milestone backlog. They were identified during Phase 1 dead code removal and are **not** addressed by any Phase 1 plan. When planning Phase 2 or future phases, these items should be added to `RETROSPECTIVE.md` or a dedicated backlog file so they are not lost.

The two highest-priority items for the next milestone are:
1. **Issue 2** (TestFromJson panic) — blocks CI from reliably passing.
2. **Issue 1** (Timer channel blocking) — correctness risk under load.
