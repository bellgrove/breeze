# Phase 1: Code Cleanup - Research

**Researched:** 2026-03-07
**Domain:** Go concurrency, MQTT dispatcher safety, structured logging, config patterns
**Confidence:** HIGH — all findings are grounded in direct code inspection of the actual source files

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Queue-full drop behavior (QUAL-01)**
- Use non-blocking send (`select` with `default`) in `OnMessage` — drop the incoming fruit if queue is full
- Drop incoming fruit (not oldest) — simpler, consistent with Phase 2 ring buffer design
- Log every dropped fruit individually: `slog.Warn("Dropped fruit", "carrier", f.CarrierId)` — no rate limiting
- Queue size to become configurable (add to `Config` struct) — so it can be tuned alongside Phase 2's write buffer limit
- Default queue size: 50 (preserve current behaviour)

**Counter/channel race (QUAL-02)**
- Fix `Values()` to read from channel before decrementing counter: `f := <-p.queue; p.counter.Add(-1)`

**Unmarshal error logging (QUAL-03)**
- Log carrier ID + error on unmarshal failure: `slog.Error("Failed to parse fruit", "err", err)`
- Do NOT log raw payload bytes (too large, too noisy)
- Discard the malformed message — no zero-value fruit row written

**Shutdown defer order (QUAL-04)**
- Move MQTT `client.Disconnect(250)` to register before `pool.Close()` so defers fire in correct order: disconnect MQTT first, then close pool

**Grademap.Update() error handling (bonus fix)**
- `OnMessage` currently ignores the error returned by `p.grademap.Update()` — log it: `slog.Error("Failed to parse grademap", "err", err)`

**Code cleanup scope (QUAL-05)**
- Remove all commented-out deque code (import, struct fields, `Create()` usage)
- Remove commented-out vision grade optimisation block in `grademap.Grade()`
- Style and dead code only — do not refactor working logic
- Identify structural/architectural/idiomatic issues for a future milestone but do NOT fix them in Phase 1; document findings in a code review note alongside the phase

**Debug log configuration**
- Add `log_level` field to `Config` struct (YAML + env var override via envconfig, consistent with existing pattern)
- Wire into `slog.SetLogLoggerLevel()` at startup
- Default: `"info"` (preserves current behaviour)
- Valid values: `"debug"`, `"info"`, `"warn"`, `"error"`

### Claude's Discretion

- The configurable queue size should use the same config pattern as the rest of the struct (YAML tag + envconfig tag)

### Deferred Ideas (OUT OF SCOPE)

- Structural/architectural/idiomatic issues found during review → document for future milestone backlog (do not fix in Phase 1)
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| QUAL-01 | `OnMessage` uses a non-blocking channel send with drop-and-warn so the paho MQTT dispatcher never stalls | Non-blocking select pattern documented below; exact blocking point identified in source |
| QUAL-02 | `Processor` counter decrements after channel read, not before (fix counter/channel race) | Exact race location pinpointed in `Values()`; correct ordering documented |
| QUAL-03 | `json.Unmarshal` errors in `OnMessage` are logged, not silently swallowed | Two unmarshal call sites identified — one in `processor.go`, one inside `UnmarshalJSON`; discard pattern documented |
| QUAL-04 | Shutdown defer order in `run()` is corrected — MQTT disconnect before pool close | Current defer registration order documented; correct defer-based pattern provided |
| QUAL-05 | General code review and cleanup (dead code, commented-out imports, style) | All dead code locations catalogued with line numbers |
</phase_requirements>

---

## Summary

Phase 1 is a targeted bug-fix and cleanup pass on two files: `breeze.go` and `processor/processor.go`. All five requirements map to specific, well-understood Go concurrency and correctness issues. No new libraries are needed. No architectural changes are in scope.

The four correctness bugs (QUAL-01 through QUAL-04) are tightly coupled: they all touch `OnMessage` or the shutdown sequence. QUAL-01 and QUAL-03 both modify the same `else if` branch in `OnMessage`, and QUAL-02 fixes the interacting `Values()` method. The bonus grademap error logging is also in `OnMessage`. A planner should treat these as a single coherent unit of work on that function.

The dead code removal (QUAL-05) and log-level configurability are independent and can be planned as separate tasks. The configurable queue size is a Config extension that enables the QUAL-01 non-blocking send to use a runtime-tunable capacity instead of the hardcoded `50`.

**Primary recommendation:** Fix `OnMessage` first (QUAL-01 + QUAL-03 + grademap error), then fix `Values()` (QUAL-02), then fix shutdown defer order (QUAL-04), then clean dead code (QUAL-05), then add config extensions. Each task is a small, reviewable diff.

---

## Standard Stack

### Core (already in use — no new dependencies)

| Library | Version | Purpose | Role in Phase 1 |
|---------|---------|---------|-----------------|
| `log/slog` | stdlib (Go 1.21+) | Structured logging | `slog.Warn`, `slog.Error` calls for dropped/malformed fruit |
| `sync/atomic` | stdlib | Lock-free counter | `atomic.Int32` already used; ordering fix only |
| `github.com/kelseyhightower/envconfig` | v1.4.0 | Env var config override | Extend with `LOG_LEVEL`, `QUEUE_SIZE` fields |
| `gopkg.in/yaml.v2` | v2.4.0 | YAML config loading | Extend Config struct with new YAML tags |
| `github.com/eclipse/paho.mqtt.golang` | v1.5.0 | MQTT client | `OnMessage` callback constraint; must return promptly |

No new `go get` commands required. All libraries are in `go.mod` already.

### Alternatives Considered

| Instead of | Could Use | Why locked decision wins |
|------------|-----------|--------------------------|
| Drop incoming fruit on queue-full | Drop oldest (dequeue head) | Dequeue requires mutex or a second channel op; drop-incoming matches Phase 2 ring buffer approach |
| `slog.SetLogLoggerLevel()` | Custom `slog.Handler` | `SetLogLoggerLevel` is sufficient for the stated need; custom handler is out of scope |
| `envconfig` tag for log level | `os.Getenv` directly | Consistent with existing Config pattern |

---

## Architecture Patterns

### Exact Bug Locations and Fixes

#### QUAL-01: Blocking send in `OnMessage` — `processor/processor.go` lines 38-40

Current code (blocks if queue is full):
```go
p.queue <- f          // line 38 — blocks paho goroutine
p.counter.Add(1)      // line 39
p.timer <- true       // line 40
```

Correct pattern (non-blocking, drop-and-warn):
```go
select {
case p.queue <- f:
    p.counter.Add(1)
    p.timer <- true
default:
    slog.Warn("Dropped fruit", "carrier", f.CarrierId)
}
```

**Key constraint:** `p.counter.Add(1)` and `p.timer <- true` must only execute when the send succeeds. They are inside the `case` branch, not after the `select`. This preserves the invariant: counter == number of items in queue channel.

**Queue size config:** `Create()` currently hardcodes `make(chan Fruit, 50)`. Change to `make(chan Fruit, cfg.QueueSize)` once `Config` carries the value. `Create()` signature will need `cfg.QueueSize int` parameter (or receive the full `Config`). The `Processor` struct does not need to store it.

#### QUAL-02: Counter/channel race in `Values()` — `processor/processor.go` lines 55-60

Current code (decrements before read — race):
```go
func (p *Processor) Values() ([]any, error) {
    if p.counter.Add(-1) >= 0 {
        f := <-p.queue       // counter already decremented, queue read could block
        return f.AsRow()
    } else {
        p.counter.Add(1)     // compensating increment — racy
        return nil, fmt.Errorf("queue is empty")
    }
}
```

**The race:** `Next()` returns `true` when `counter > 0`. Between `Next()` returning and `Values()` running, another goroutine could consume the item (not currently the case but the counter/channel invariant is broken). More concretely, `counter.Add(-1)` fires before the queue item is removed, so a concurrent `Next()` call can observe `counter == 0` while an item sits in the queue.

Correct pattern (read first, then decrement):
```go
func (p *Processor) Values() ([]any, error) {
    f := <-p.queue
    p.counter.Add(-1)
    return f.AsRow()
}
```

Note: The `else` branch that compensates with `counter.Add(1)` is eliminated. `Values()` is only called by `pgx.CopyFrom` after `Next()` returns true, and `Next()` gates on `counter > 0`. The channel read will not block in practice because the counter reflects actual queue depth. The simplified form is safe within the `pgxpool.CopyFrom` protocol.

#### QUAL-03: Silenced unmarshal error — `processor/processor.go` line 29 + `processor/fruit.go` line 275

There are two unmarshal call sites:

**Site 1** — `processor.go` line 29, `OnMessage` fruit branch:
```go
json.Unmarshal(msg.Payload(), &f)   // error ignored
```

This is the critical one. A malformed payload produces a zero-value `Fruit` (empty `CarrierId`, zero integers) which is then graded, possibly truncated, and sent to the queue — corrupting the database.

Fix (check error, discard on failure):
```go
if err := json.Unmarshal(msg.Payload(), &f); err != nil {
    slog.Error("Failed to parse fruit", "err", err)
    return
}
```

After this return, `p.grademap.Grade(&f)`, the PrimaryDefect truncation, the queue send, and the counter increment are all skipped. No zero-value row enters the queue.

**Site 2** — `fruit.go` line 275, inside `UnmarshalJSON`:
```go
json.Unmarshal(b, &breeze)   // error ignored
```

This is the custom `UnmarshalJSON` for the `Fruit` type. It is called by the Site 1 `json.Unmarshal`. If this inner call fails, Site 1's error check will catch it (because `UnmarshalJSON` returns `nil` unconditionally — line 359 `return nil`). To propagate the error correctly, `UnmarshalJSON` must return the inner unmarshal error:

```go
if err := json.Unmarshal(b, &breeze); err != nil {
    return fmt.Errorf("failed to parse fruit payload: %w", err)
}
```

This change makes Site 1's error check effective. Without it, `json.Unmarshal(msg.Payload(), &f)` will always return `nil` even on malformed input, because `UnmarshalJSON` swallows the error.

#### QUAL-04: Shutdown defer order — `breeze.go`

Current structure:
```go
pool, err := pgxpool.NewWithConfig(ctx, config)
// ...
defer pool.Close()            // registered first — fires LAST (LIFO)

next_batch := make(chan bool)
defer close(next_batch)       // registered second — fires second-to-last

// ...later in the for/select loop:
case <-ctx.Done():
    client.Disconnect(250)    // inline — runs immediately on ctx cancel
    return nil                // triggers defers: close(next_batch), then pool.Close()
```

**The problem:** When `ctx.Done()` fires, `client.Disconnect(250)` runs inline, stopping new MQTT messages. Then `return nil` fires the defers: `close(next_batch)` first (which the goroutine inside `Create()` uses to signal "stop"), then `pool.Close()`. Any fruit items sitting in the queue channel at shutdown time may not be written if `next_batch` closes before the main loop processes a final batch signal.

The user's decision is to register `client.Disconnect` as a defer before `pool.Close()`:

```go
pool, err := pgxpool.NewWithConfig(ctx, config)
// ...
defer pool.Close()                    // registered 1st — fires LAST

next_batch := make(chan bool)
defer close(next_batch)               // registered 2nd — fires 3rd

// ...after client is created:
defer client.Disconnect(250)          // registered 3rd — fires 2nd (before close(next_batch))

// Remove inline client.Disconnect from the ctx.Done() case:
case <-ctx.Done():
    return nil
```

With this order, shutdown sequence is: `client.Disconnect(250)` (stops MQTT), `close(next_batch)` (signals timer goroutine to stop), `pool.Close()` (closes DB connection). This is the correct order: no new fruit arrives after disconnect; the timer goroutine drains; pool closes last.

**Note:** `paho.mqtt.golang`'s `Disconnect(quiesce uint)` blocks for up to `quiesce` ms (250ms here) waiting for in-flight messages. Registering it as a defer is safe — it is called on the main goroutine during `run()` return, not during a signal handler.

#### Bonus: `grademap.Update()` error ignored — `processor/processor.go` line 25

```go
p.grademap.Update(msg.Payload())   // error return ignored
```

Fix:
```go
if err := p.grademap.Update(msg.Payload()); err != nil {
    slog.Error("Failed to parse grademap", "err", err)
}
```

#### QUAL-05: Dead code inventory

All locations to remove:

**`processor/processor.go`:**
- Line 16: `// dequeue  *deque.Deque[Fruit]` — commented struct field
- Line 17: `// current  Fruit` — commented struct field
- Lines 110-111: `// queue := new(deque.Deque[Fruit])` and `// queue.SetBaseCap(100)` — commented code in `Create()`
- Lines 116-117: `// Fruit{},` — commented struct literal field in `Create()` return

**`processor/grademap.go`:**
- Lines 122-131: Commented vision grade optimisation block inside `Grade()`:
  ```go
  // A -> 3
  // B -> 2
  // len(grades) = 5
  // idx := -1
  // if len(f.VisionGrade) >= 1 {
  // 	idx = len(g.Grades) - int(f.VisionGrade[0]-'A') - 2
  // }
  // if idx < 0 || idx >= len(g.Grades)-1 {
  // 	slog.Error("invalid fruit grade", "vision_grade", f.VisionGrade)
  // 	return
  // }
  ```

**`breeze.go`:**
- Line 49: `// slog.SetLogLoggerLevel(slog.LevelDebug)` — to be replaced by the log_level config (not just deleted)
- Line 85: `// conn, err := pgx.Connect(ctx, cfg.Database.URL)` — commented old connection approach
- Line 92: `// config.MaxConnIdleTime =` — incomplete commented config
- Lines 131-132: `// slog.Info("New write")` — commented debug log
- Lines 143-145: `// default:`, `// 	// do a piece of work`, `// 	time.Sleep(100 * time.Millisecond)` — old polling approach

**`processor/fruit.go`:**
- Line 358: `// a.OtherDefects = []string{"ABCD"}` — debug test line

**`processor/grademap.go`:**
- Line 188: `// gm.Defect_grading_passes[k1] = v1` — commented mutation (not needed)

**Import cleanup:** Once deque struct fields are removed, if `github.com/gammazero/deque` has no remaining usages, remove it from the import block. Check: `go.mod` still references `github.com/gammazero/deque v1.0.0` — it must be removed from `go.mod`/`go.sum` as well with `go mod tidy`.

**`processor/fruit_test.go`:**
- Line 93: `// got, err := FromJson(tt.args.val)` — commented old API call
- Lines 101-103: `// if !reflect.DeepEqual(got, tt.want) {` block — commented alternative assertion
- Line 87: `// TODO: Add test cases.` — acceptable as-is (informational)

### Log Level Configuration Pattern

Extend the existing `Config` struct in `breeze.go`:

```go
type Config struct {
    MQTT struct {
        URI  string `yaml:"uri" envconfig:"SERVER_URI"`
        User string `yaml:"user" envconfig:"SERVER_USER"`
        Pass string `yaml:"pass" envconfig:"SERVER_PASS"`
    } `yaml:"mqtt"`
    Database struct {
        URL string `yaml:"url" envconfig:"DATABASE_URL"`
    } `yaml:"database"`
    LogLevel  string `yaml:"log_level"  envconfig:"LOG_LEVEL"`
    QueueSize int    `yaml:"queue_size" envconfig:"QUEUE_SIZE"`
}
```

Wire `LogLevel` in `main()` after `readEnv(&cfg)`:

```go
// Parse log level (default: info)
level := slog.LevelInfo
switch strings.ToLower(cfg.LogLevel) {
case "debug":
    level = slog.LevelDebug
case "warn":
    level = slog.LevelWarn
case "error":
    level = slog.LevelError
}
slog.SetLogLoggerLevel(level)
```

Wire `QueueSize` with a default: if `cfg.QueueSize == 0` (zero value, not set), default to 50. Either set it in config struct directly or apply in `run()` before calling `processor.Create()`.

### `Create()` signature change

`Create()` currently takes `next chan bool`. It needs `queueSize int`:

```go
func Create(next chan bool, queueSize int) Processor {
    // ...
    return Processor{
        timer,
        make(chan Fruit, queueSize),
        atomic.Int32{},
        Grademap{},
    }
}
```

Caller in `breeze.go`:
```go
var proc = processor.Create(next_batch, cfg.QueueSize)
```

### Anti-Patterns to Avoid

- **Don't add a mutex around the counter+channel pair.** The fix is ordering (read then decrement), not synchronisation. Adding a mutex would change the concurrency model unnecessarily.
- **Don't log the raw payload bytes for QUAL-03.** The user explicitly decided against it (too large, too noisy). Log `"err"` and carrier ID only — but carrier ID is only available after successful unmarshal. If unmarshal fails, carrier ID is unavailable; log only `"err"`.
- **Don't refactor `OnMessage` beyond the specified fixes.** The grademap/fruit dispatch structure, the PrimaryDefect truncation, and the "unknown topic" warn are all working and out of scope.
- **Don't change `Values()` to handle the "empty queue" error path differently.** With the counter/channel invariant properly maintained by QUAL-01 and QUAL-02, `Values()` will never be called on an empty queue within the `pgx.CopyFrom` protocol.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Non-blocking channel send | Custom queue abstraction | `select` with `default` | stdlib idiom; one line |
| Structured log level parsing | String switch | `slog.Level.UnmarshalText` or plain switch | Switch is three lines; `UnmarshalText` works too but introduces indirect dependency on text marshalling |
| Env var config | Manual `os.Getenv` | `envconfig` tags (already in use) | Consistent with existing pattern |
| Config defaults | Constructor function | Zero-value check in `run()` | No new types needed |

---

## Common Pitfalls

### Pitfall 1: Sending to timer channel after failed queue send
**What goes wrong:** If `p.timer <- true` is placed outside the `select` block (after it), a dropped fruit still signals the batch timer, causing an empty batch write cycle.
**How to avoid:** Both `p.counter.Add(1)` and `p.timer <- true` must be inside the `case p.queue <- f:` branch, not after the `select`.

### Pitfall 2: `UnmarshalJSON` returning nil hides the error
**What goes wrong:** `processor.go`'s `json.Unmarshal(msg.Payload(), &f)` calls `f.UnmarshalJSON()` which currently returns `nil` unconditionally. Even after adding an error check at the call site, malformed payloads will still produce zero-value fruit until `UnmarshalJSON` is fixed to propagate the error.
**How to avoid:** Fix both sites together in the same task.

### Pitfall 3: `go mod tidy` not run after removing deque
**What goes wrong:** `github.com/gammazero/deque` remains in `go.mod` as a direct dependency even after all usages are removed, causing confusion for future readers and `go mod verify`.
**How to avoid:** Run `go mod tidy` as part of the dead code removal task. Verify with `go build ./...` before committing.

### Pitfall 4: defer registration order off-by-one
**What goes wrong:** If `defer client.Disconnect(250)` is registered after `defer close(next_batch)`, the timer goroutine will stop before MQTT disconnects, and in-flight messages during the 250ms quiesce window will not be processed.
**Correct order (registration):**
1. `defer pool.Close()` — first registered, fires last
2. `defer close(next_batch)` — second, fires third
3. `defer client.Disconnect(250)` — third registered, fires second (LIFO)
**Warning signs:** Disconnect fires after channel close — grademap or fruit `OnMessage` panics with "send on closed channel" if `p.timer` is closed before disconnect.

### Pitfall 5: `p.timer <- true` blocks if timer channel buffer is full
**What goes wrong:** `timer` is created as `make(chan bool, 1)` (capacity 1). If the timer goroutine hasn't consumed the previous signal, a second fruit arrival will block on `p.timer <- true`. This is a pre-existing latent issue.
**Scope:** This is NOT in scope for Phase 1. Document it in the code review note for the future milestone backlog. The QUAL-01 fix wraps the entire send block in a non-blocking select (for the queue send), but `p.timer <- true` inside the `case` branch still blocks if the timer buffer is full. The fix is either a second `select` for the timer send or changing timer to `make(chan bool, maxItems)`, but both are architectural changes deferred to a future phase.

### Pitfall 6: Test panic in `ValidateFruit` with map fields
**What goes wrong:** The existing `TestFromJson` test panics: `reflect.Value.Equal: values of type map[string]interface {} are not comparable`. This is a pre-existing test bug (not introduced by Phase 1).
**Scope:** Document as an existing test issue; Phase 1 does not need to fix it (QUAL-05 is dead code/style, not test repair). However, if any Phase 1 task adds new tests, they should use `reflect.DeepEqual` or JSON round-trip comparison for map fields, not `field1.Equal(field2)`.

---

## Code Examples

### Non-blocking send pattern (QUAL-01)
```go
// Source: Go specification — channel send statements
select {
case p.queue <- f:
    p.counter.Add(1)
    p.timer <- true
default:
    slog.Warn("Dropped fruit", "carrier", f.CarrierId)
}
```

### Unmarshal error propagation from custom UnmarshalJSON (QUAL-03)
```go
// In fruit.go UnmarshalJSON:
if err := json.Unmarshal(b, &breeze); err != nil {
    return fmt.Errorf("failed to parse fruit payload: %w", err)
}

// In processor.go OnMessage:
if err := json.Unmarshal(msg.Payload(), &f); err != nil {
    slog.Error("Failed to parse fruit", "err", err)
    return
}
```

### Defer registration for correct shutdown order (QUAL-04)
```go
// In run(), register defers in this order (fires in reverse):
defer pool.Close()            // fires 3rd (last)
// ... create next_batch channel ...
defer close(next_batch)       // fires 2nd
// ... create MQTT client ...
defer client.Disconnect(250)  // fires 1st

// Remove inline client.Disconnect from ctx.Done() case:
case <-ctx.Done():
    return nil
```

### slog level configuration
```go
// Source: log/slog stdlib documentation
import "log/slog"

level := slog.LevelInfo
switch strings.ToLower(cfg.LogLevel) {
case "debug":
    level = slog.LevelDebug
case "warn":
    level = slog.LevelWarn
case "error":
    level = slog.LevelError
}
slog.SetLogLoggerLevel(level)
```

---

## State of the Art

| Old Approach | Current Approach | Impact on Phase 1 |
|--------------|------------------|-------------------|
| `gammazero/deque` for queue | `chan Fruit` buffered channel | Deque import is dead — remove from go.mod |
| Hardcoded `slog.SetLogLoggerLevel` line (commented out) | `log_level` config field | New config field activates the commented pattern |
| `pgx.Connect` (commented out) | `pgxpool.NewWithConfig` | Old connect call is dead code to remove |

---

## Open Questions

1. **`p.timer <- true` blocking when timer buffer full**
   - What we know: `timer` is buffered at capacity 1. If the timer goroutine is slow, a second fruit arriving before the first timer signal is consumed will block inside the `case p.queue <- f:` branch of QUAL-01's select, even after the fix.
   - What's unclear: How frequently does this happen in production? Likely rare given 1-second batch timeout.
   - Recommendation: Document as a known latent issue in the code review note. Do not fix in Phase 1.

2. **`ValidateFruit` test panic on map fields**
   - What we know: `TestFromJson` panics on `MCenterOffsets` et al. (map fields not comparable with `reflect.Value.Equal`).
   - What's unclear: Whether the user wants this fixed as part of QUAL-05 cleanup.
   - Recommendation: Treat as out of scope for Phase 1 unless user confirms. Note the panic in the code review document. The `fruit_test.go` tests still run partially (the non-map field checks execute before the panic).

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing stdlib (`testing` package), Go 1.26.1 |
| Config file | None — standard `go test` |
| Quick run command | `/usr/local/go/bin/go test ./processor/...` |
| Full suite command | `/usr/local/go/bin/go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| QUAL-01 | `OnMessage` drops fruit without blocking when queue is full | unit | `go test ./processor/... -run TestOnMessage_QueueFull` | Wave 0 |
| QUAL-02 | `Values()` reads channel before decrementing counter | unit | `go test ./processor/... -run TestValues_CounterOrder` | Wave 0 |
| QUAL-03 | Malformed JSON is logged and discarded; no zero-value row queued | unit | `go test ./processor/... -run TestOnMessage_MalformedJSON` | Wave 0 |
| QUAL-04 | Shutdown defer order: disconnect before pool close | manual/smoke | Manual: observe logs during `SIGINT` | N/A — integration only |
| QUAL-05 | No commented-out dead code remains | lint/manual | `grep -rn "// dequeue\|// queue :=\|// conn, err" .` | N/A — grep check |

**Note on QUAL-04:** The shutdown order is not easily unit-testable without mocking the MQTT client and pool. It can be verified by code inspection during review and a manual smoke test. No automated test is required.

**Note on QUAL-05:** Dead code removal is verified by code review and successful `go build ./...` + `go mod tidy`. No dedicated test file needed.

### Sampling Rate

- **Per task commit:** `/usr/local/go/bin/go test ./processor/... -timeout 30s`
- **Per wave merge:** `/usr/local/go/bin/go test ./... -timeout 60s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `processor/processor_test.go` — covers QUAL-01 (`TestOnMessage_QueueFull`), QUAL-02 (`TestValues_CounterOrder`), QUAL-03 (`TestOnMessage_MalformedJSON`)
- [ ] No framework install needed — `testing` is stdlib

**Note on existing test panic:** `TestFromJson` in `fruit_test.go` currently panics on map fields (`MCenterOffsets` etc.) due to `reflect.Value.Equal` not supporting `map[string]interface{}`. This is a pre-existing bug. Wave 0 tests should be placed in a new file `processor_test.go` and do not depend on the broken `ValidateFruit` helper.

---

## Sources

### Primary (HIGH confidence)

- Direct inspection of `/home/ben/github.com/bellgrove/breeze/processor/processor.go` — bug locations at exact line numbers
- Direct inspection of `/home/ben/github.com/bellgrove/breeze/processor/fruit.go` — `UnmarshalJSON` returns nil unconditionally (line 359)
- Direct inspection of `/home/ben/github.com/bellgrove/breeze/breeze.go` — defer registration order, dead code locations
- Direct inspection of `/home/ben/github.com/bellgrove/breeze/go.mod` — Go 1.24.3 minimum (runtime is 1.26.1), `gammazero/deque` is a direct dependency despite no active usage
- Go specification (channel send statements) — non-blocking select/default is idiomatic stdlib pattern
- `log/slog` stdlib docs — `SetLogLoggerLevel` is the correct API for global level configuration

### Secondary (MEDIUM confidence)

- `paho.mqtt.golang` v1.5.0 README — `OnMessage` callback must not block; blocking the dispatcher stalls all subscriptions on the client

### Tertiary (LOW confidence)

- None

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new libraries; all existing dependencies verified via go.mod
- Architecture: HIGH — all patterns derived from direct code inspection and Go stdlib knowledge
- Bug locations: HIGH — line numbers confirmed by reading source files
- Pitfalls: HIGH — derived from Go language specification (channel semantics, defer LIFO order) and direct code analysis

**Research date:** 2026-03-07
**Valid until:** 2026-04-07 (stable codebase, no external API changes expected)
