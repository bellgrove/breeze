# Phase 5: Processor Stability Cleanup - Research

**Researched:** 2026-03-08
**Domain:** Go channel concurrency, reflect package, processor package test infrastructure
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| QUAL-01 | `OnMessage` uses a non-blocking channel send with drop-and-warn so the paho MQTT dispatcher never stalls | Fix 1: make the `p.timer <- true` send non-blocking with select/default — the queue send is already non-blocking but the subsequent timer send is not |
| QUAL-05 | General code review and cleanup (dead code, commented-out imports, style) | Fix 2: replace `reflect.Value.Equal` with `reflect.DeepEqual` in `ValidateFruit` so `go test ./...` exits clean |

</phase_requirements>

---

## Summary

Phase 5 closes two high-priority tech debt items identified during the v1.0 milestone audit. Both are isolated, surgical fixes with no dependency on new infrastructure. Each fix touches exactly one location in one file.

**Fix 1 — timer blocking send (processor/processor.go line 57):** Inside the `case p.queue <- f:` branch of `OnMessage`, after the fruit is enqueued the code does `p.timer <- true`. The timer channel has capacity 1. If the batch goroutine has not yet consumed the previous signal (which happens during rapid message bursts), a second `OnMessage` call blocks on this send. The queue send itself is already protected by a `select/default`, but the timer send is not. Fix: wrap the timer send in its own `select/default`. If the timer channel is already full, the batch goroutine already has a pending wake signal and will process the batch — skipping the second notification is safe.

**Fix 2 — ValidateFruit panic (processor/fruit_test.go line 134):** `ValidateFruit` uses `reflect.Value.Equal` for all non-slice fields. The `Fruit` struct contains eight `map[string]any` fields (`MCenterOffsets`, `MClassifiedBlob`, `MColour`, `MColourBlob`, `MDiameters`, `MFeatures`, `MFunction`, `MTimer`). `reflect.Value.Equal` panics on map kinds because maps are not directly comparable. This is a test-only bug; the production `UnmarshalJSON` is correct. Fix: add a `reflect.Map` kind check before calling `field1.Equal`, and use `reflect.DeepEqual(field1.Interface(), field2.Interface())` for those fields.

**Primary recommendation:** Two targeted edits — one `select/default` wrapper around `p.timer <- true`; one `reflect.Map` guard in `ValidateFruit`. Both changes are self-contained. No new tests need to be written (the existing `TestOnMessage_QueueFull` and `TestFromJson/Class_2` serve as the verification harness).

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go standard library `reflect` | stdlib | Field-by-field struct comparison in test | Already imported in fruit_test.go |
| Go channel mechanics | language | Non-blocking send via `select/default` | Already used in OnMessage for queue and gradeCh sends |

No new dependencies. Both fixes use language and stdlib primitives already present in the file.

**Installation:** none required.

---

## Architecture Patterns

### Pattern 1: Non-blocking channel send with select/default

**What:** A channel send wrapped in a `select` statement with a `default` branch so the goroutine never blocks.
**When to use:** Any time a fast-path goroutine (MQTT dispatcher) must signal a slower-path goroutine (batch processor) without waiting.

**Current (broken) code — processor.go lines 54–60:**
```go
select {
case p.queue <- f:
    p.counter.Add(1)
    p.timer <- true   // blocking — can stall MQTT dispatcher
default:
    slog.Warn("Dropped fruit", "carrier", f.CarrierId)
}
```

**Fixed code:**
```go
select {
case p.queue <- f:
    p.counter.Add(1)
    select {
    case p.timer <- true:
    default:
        // timer already notified; batch goroutine will drain queue
    }
default:
    slog.Warn("Dropped fruit", "carrier", f.CarrierId)
}
```

**Why this is safe:** The batch goroutine in `Create()` reads from `timer` in a loop. If one notification is already buffered (`chan bool, 1`), skipping a second notification does not skip processing — the goroutine will still wake and drain all queued fruits.

### Pattern 2: reflect.Map guard before reflect.Value.Equal

**What:** Check `field.Kind() == reflect.Map` before calling `field1.Equal(field2)`, and route map fields through `reflect.DeepEqual`.
**When to use:** Any `reflect`-based comparator that may encounter map-kinded fields.

**Current (broken) code — fruit_test.go line 134:**
```go
} else {
    if !field1.Equal(field2) {   // panics for map kinds
        ...
    }
}
```

**Fixed code (Option A from CODE-REVIEW.md — minimal change):**
```go
} else if field1.Kind() == reflect.Map {
    if !reflect.DeepEqual(field1.Interface(), field2.Interface()) {
        err = fmt.Errorf("%v%s does not match; ", err, fieldName)
    }
} else {
    if !field1.Equal(field2) {
        slog.Info("What", "f1", field1, "f2", field2)
        err = fmt.Errorf("%v%s does not match; ", err, fieldName)
    }
}
```

**Why reflect.Value.Equal panics on maps:** The Go `reflect` package documents that `Value.Equal` panics if the value's Kind is `Map`, `Func`, or `Slice` (non-byte slices). The existing slice handling already works around this by using a type assertion + `reflect.DeepEqual`. The map case was missed.

### Anti-Patterns to Avoid

- **Adding a `GrademapID` nil guard to ValidateFruit:** The `GrademapID *int64` field is a pointer, not a map — `reflect.Value.Equal` handles pointer comparison correctly. No special case needed.
- **Increasing timer channel buffer:** Making `timer` larger (e.g., capacity 10) would reduce but not eliminate the blocking risk. Only a `select/default` eliminates it.
- **Replacing ValidateFruit entirely with JSON round-trip:** Option B from CODE-REVIEW.md is correct but is a larger change. The minimal Option A fix is sufficient and keeps the existing field-by-field diagnostic output.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Deep equality for maps | Custom recursive walker | `reflect.DeepEqual` | stdlib, handles nested maps, nil maps, type mismatches |
| Non-blocking send | Polling loop or goroutine spawn | `select { case ch <- v: default: }` | Language primitive — zero overhead, zero goroutine leak risk |

---

## Common Pitfalls

### Pitfall 1: Forgetting the GrademapID pointer field in ValidateFruit

**What goes wrong:** `GrademapID` is `*int64` (a pointer). When both values are nil, `reflect.Value.Equal` returns true. When one is nil and the other is not, it returns false. This is correct behaviour — no special handling needed. Adding an unnecessary nil-pointer branch would complicate the fix without benefit.

**How to avoid:** Only add a `reflect.Map` guard. The pointer kind is already handled correctly by `reflect.Value.Equal`.

### Pitfall 2: Misidentifying the fix scope for QUAL-01

**What goes wrong:** A developer reads QUAL-01 as "make OnMessage non-blocking" and concludes the fix is done (the queue send is already non-blocking via `select/default`). The timer send is a second, distinct blocking point inside the same `case` branch.

**How to avoid:** The fix is specifically `p.timer <- true` at processor.go line 57, not the outer `select`. The outer `select` protects the queue send. The inner send to `timer` has no protection.

**Warning signs:** The existing `TestOnMessage_QueueFull` test passes — it uses `newTestProcessorSmallQueue` which sizes the timer channel at `queueSize+10`, making the timer send non-blocking in tests but not in production (where `Create()` sizes timer at 1).

### Pitfall 3: Assuming newTestProcessorSmallQueue masks the timer bug

**What goes wrong:** The existing processor tests pass because `newTestProcessorSmallQueue` creates the timer channel with `queueSize+10` capacity (line 32 of processor_test.go). This means the timer blocking bug does not manifest in the test suite. There is no existing failing test for the timer fix — the verification is conceptual (code inspection + production scenario reasoning).

**How to avoid:** The plan must include either: (a) a new unit test that uses timer capacity 1 and sends two rapid fruits to verify non-blocking behaviour, or (b) explicit code inspection verification. Given the fix is trivially correct, code inspection + `go test ./...` green is sufficient.

### Pitfall 4: reflect.DeepEqual for []byte fields is already handled

**What goes wrong:** Assuming the byte-slice fields (`CenterOffsets`, `ClassifiedBlob`, etc.) also need a fix. They are already handled by the `field1.Kind() == reflect.Slice` branch that type-asserts to `[]byte` and calls `reflect.DeepEqual`. Only the `map[string]any` fields fall through to the broken `field1.Equal` call.

**How to avoid:** Read `ValidateFruit` carefully. The slice branch (lines 115–132) correctly handles all `[]byte` and `[]string` fields. The `else` at line 133 only runs for non-slice fields. The map fields (`MCenterOffsets` etc.) have Kind `reflect.Map`, not `reflect.Slice`, so they fall through to the else branch.

---

## Code Examples

### Exact current state — processor.go lines 54–60

```go
// Source: /processor/processor.go (as-read 2026-03-08)
select {
case p.queue <- f:
    p.counter.Add(1)
    p.timer <- true          // line 57 — blocking send, timer cap=1
default:
    slog.Warn("Dropped fruit", "carrier", f.CarrierId)
}
```

### Exact current state — fruit_test.go lines 133–138

```go
// Source: /processor/fruit_test.go (as-read 2026-03-08)
} else {
    if !field1.Equal(field2) {   // line 134 — panics for reflect.Map kind
        slog.Info("What", "f1", field1, "f2", field2)
        err = fmt.Errorf("%v%s does not match; ", err, fieldName)
    }
}
```

### Map fields that trigger the panic

The following `Fruit` struct fields have Kind `reflect.Map` and reach line 134:

| Field | Type |
|-------|------|
| `MCenterOffsets` | `map[string]any` |
| `MClassifiedBlob` | `map[string]any` |
| `MColour` | `map[string]any` |
| `MColourBlob` | `map[string]any` |
| `MDiameters` | `map[string]any` |
| `MFeatures` | `map[string]any` |
| `MFunction` | `map[string]any` |
| `MTimer` | `map[string]any` |

Note: these fields are populated by `UnmarshalJSON` but the `c2_fruit` expected value in `TestFromJson` leaves them as `nil` (zero value). When both `got.MCenterOffsets` and `want.MCenterOffsets` are nil, `reflect.Value.Equal` still panics because the Kind is Map. `reflect.DeepEqual(nil, nil)` returns true — safe.

---

## State of the Art

| Old Approach | Current Approach | Impact |
|--------------|------------------|--------|
| Blocking timer send inside `case p.queue <- f:` | Non-blocking `select/default` on timer send | Eliminates MQTT dispatcher stall risk under load |
| `reflect.Value.Equal` for all non-slice fields | `reflect.DeepEqual` for map kinds, `reflect.Value.Equal` for others | `TestFromJson/Class_2` passes; `go test ./...` exits 0 |

---

## Open Questions

1. **Should a new test be added to catch the timer blocking bug directly?**
   - What we know: existing tests pass because `newTestProcessorSmallQueue` over-sizes the timer channel. There is no test that exercises timer channel capacity 1 with two rapid sends.
   - What's unclear: the phase description says "1 plan" covering both fixes. Adding a new test for the timer fix would be correct Nyquist practice but is not explicitly required.
   - Recommendation: add a short test `TestOnMessage_TimerNonBlocking` that creates a Processor via `Create()` (timer cap 1) and fires two rapid fruit messages, asserting no blocking. This closes QUAL-01 fully at the test level. The plan should decide whether to include this or treat code inspection as sufficient.

2. **Do the eight M-prefixed map fields in Fruit appear in the test's expected value?**
   - What we know: in `TestFromJson`, `c2_fruit` does not set any `M*` fields — they are nil. After `json.Unmarshal`, `got.MCenterOffsets` is also nil (the JSON is unmarshalled into `[]byte` fields, not the map fields). So both sides are nil maps.
   - What's unclear: does `reflect.Value.Equal` panic on two nil maps? Answer: yes, because Kind is still `Map` even for nil map values. `reflect.DeepEqual(nil_map, nil_map)` returns true and does not panic.
   - Recommendation: the fix is correct and complete even for the nil-map case.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none — `go test ./...` |
| Quick run command | `go test ./processor/...` |
| Full suite command | `go test ./...` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| QUAL-01 (full closure) | Timer send in `OnMessage` is non-blocking | unit | `go test ./processor/... -run TestOnMessage` | Partial — existing `TestOnMessage_QueueFull` does not exercise timer capacity=1; new test needed if planner decides to add one |
| QUAL-05 (full closure) | `go test ./...` exits 0; `TestFromJson/Class_2` does not panic | unit | `go test ./processor/... -run TestFromJson` | Yes — `processor/fruit_test.go` exists, currently panics |

### Sampling Rate

- **Per task commit:** `go test ./processor/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** `go test ./...` exits 0 before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] Optional: `processor/processor_test.go` — add `TestOnMessage_TimerNonBlocking` using a `Create()`-produced processor (timer cap 1) to directly verify the timer fix. If planner omits this, the fix is still verifiable by code inspection + `TestFromJson` green run.

*(The `TestFromJson/Class_2` test already exists and is the primary gate for Fix 2. No new test file is required — the fix makes an existing test stop panicking.)*

---

## Sources

### Primary (HIGH confidence)

- Direct code read: `/processor/processor.go` — confirmed blocking send at line 57, timer channel created with `make(chan bool, 1)` in `Create()`
- Direct code read: `/processor/fruit_test.go` — confirmed `reflect.Value.Equal` at line 134, confirmed `Fruit` map fields at lines 64–71 of `fruit.go`
- Direct code read: `/processor/processor_test.go` — confirmed `newTestProcessorSmallQueue` uses `queueSize+10` timer buffer (line 32), explaining why existing tests pass
- `.planning/phases/01-code-cleanup/01-CODE-REVIEW.md` — authoritative description of both issues, suggested fix patterns
- `.planning/v1.0-MILESTONE-AUDIT.md` — confirms both items are open tech debt, confirms severity ratings

### Secondary (MEDIUM confidence)

- Go language specification: `select` statement with `default` is the canonical non-blocking send idiom — this is stable language behaviour, not a library API
- Go `reflect` package documentation (stdlib): `Value.Equal` panics for Map, Func, and non-byte Slice kinds — confirmed by CODE-REVIEW.md description and code inspection

### Tertiary (LOW confidence)

None.

---

## Metadata

**Confidence breakdown:**
- Fix 1 (timer send): HIGH — code read confirms exact location, behaviour, and safe fix
- Fix 2 (reflect panic): HIGH — code read confirms exact location, all eight map fields identified, DeepEqual fix is standard stdlib pattern
- Test infrastructure: HIGH — all test files read; existing tests confirmed; wave 0 gap is optional extension only

**Research date:** 2026-03-08
**Valid until:** 2026-04-08 (stable — no external dependencies, all stdlib)
