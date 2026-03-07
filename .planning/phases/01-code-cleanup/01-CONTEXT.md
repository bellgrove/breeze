# Phase 1: Code Cleanup - Context

**Gathered:** 2026-03-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Fix four pre-existing bugs in `breeze.go` and `processor/processor.go` that would corrupt data or block message delivery once Phase 2 write buffering is added. Also remove dead code and add log level configurability. No new features — only correctness fixes and cleanup.

</domain>

<decisions>
## Implementation Decisions

### Queue-full drop behavior (QUAL-01)
- Use non-blocking send (`select` with `default`) in `OnMessage` — drop the incoming fruit if queue is full
- Drop incoming fruit (not oldest) — simpler, consistent with Phase 2 ring buffer design
- Log every dropped fruit individually: `slog.Warn("Dropped fruit", "carrier", f.CarrierId)` — no rate limiting
- Queue size to become configurable (add to `Config` struct) — so it can be tuned alongside Phase 2's write buffer limit
- Default queue size: 50 (preserve current behaviour)

### Counter/channel race (QUAL-02)
- Fix `Values()` to read from channel before decrementing counter: `f := <-p.queue; p.counter.Add(-1)`

### Unmarshal error logging (QUAL-03)
- Log carrier ID + error on unmarshal failure: `slog.Error("Failed to parse fruit", "err", err)`
- Do NOT log raw payload bytes (too large, too noisy)
- Discard the malformed message — no zero-value fruit row written

### Shutdown defer order (QUAL-04)
- Move MQTT `client.Disconnect(250)` to register before `pool.Close()` so defers fire in correct order: disconnect MQTT first, then close pool

### Grademap.Update() error handling (bonus fix)
- `OnMessage` currently ignores the error returned by `p.grademap.Update()` — log it: `slog.Error("Failed to parse grademap", "err", err)`

### Code cleanup scope (QUAL-05)
- Remove all commented-out deque code (import, struct fields, `Create()` usage)
- Remove commented-out vision grade optimisation block in `grademap.Grade()`
- Style and dead code only — do not refactor working logic
- **Identify** structural/architectural/idiomatic issues for a future milestone but do NOT fix them in this phase; document findings in a code review note alongside the phase

### Debug log configuration
- Add `log_level` field to `Config` struct (YAML + env var override via envconfig, consistent with existing pattern)
- Wire into `slog.SetLogLoggerLevel()` at startup
- Default: `"info"` (preserves current behaviour)
- Valid values: `"debug"`, `"info"`, `"warn"`, `"error"`

</decisions>

<specifics>
## Specific Ideas

- User wants any structural/architecture/idiom issues *identified* and documented, not fixed, during this phase — keep a running note for next milestone backlog
- The configurable queue size should use the same config pattern as the rest of the struct (YAML tag + envconfig tag)

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Config` struct in `breeze.go`: already has YAML + envconfig tags — extend with `QueueSize int` and `LogLevel string` using the same pattern

### Established Patterns
- Config loading: YAML file first, then env var override via `envconfig.Process` — new fields follow same pattern
- Logging: `log/slog` throughout — use `slog.Error`/`slog.Warn` consistent with existing call sites
- Error handling: startup errors call `os.Exit(1)`; runtime errors use `slog.Error` and continue — unmarshal failures follow the runtime pattern

### Integration Points
- `OnMessage` in `processor/processor.go`: QUAL-01, QUAL-02, QUAL-03 all touch this function
- `run()` in `breeze.go`: QUAL-04 (defer order) is here
- `Processor.Values()` in `processor/processor.go`: QUAL-02 counter fix is here
- `Create()` in `processor/processor.go`: queue channel cap becomes `cfg.QueueSize`

</code_context>

<deferred>
## Deferred Ideas

- Structural/architectural/idiomatic issues found during review → document for future milestone backlog (do not fix in Phase 1)

</deferred>

---

*Phase: 01-code-cleanup*
*Context gathered: 2026-03-07*
