# Phase 2: Write Buffer and Reconnect - Context

**Gathered:** 2026-03-08
**Status:** Ready for planning

<domain>
## Phase Boundary

Add an in-memory write buffer so fruit records are never silently lost during a PostgreSQL outage. A background reconnect probe detects when the DB recovers and triggers an automatic flush of buffered records. No new features — only durability for the existing write path in `run()`.

</domain>

<decisions>
## Implementation Decisions

### Write buffer capacity
- Default size: 10,000 records (~15 MB heap)
- Configurable via `write_buffer_size` field in `Config` struct, same YAML + envconfig pattern as `QueueSize` and `LogLevel`
- Drop-oldest policy when buffer is full (RECO-01)

### Write buffer location
- Buffer lives in `run()` as a local `[]Fruit` slice — not inside `Processor`
- Only the write loop touches the buffer; no concurrency issues, no locking needed

### Claude's Discretion
- Reconnect probe: background goroutine using `pool.Ping` with exponential backoff; parameters (initial delay, max cap, multiplier) can be hardcoded at reasonable defaults (e.g. 1s initial, 60s max, 2× multiplier) unless a compelling reason to make them configurable emerges during research
- Overflow logging: per-drop `slog.Warn` consistent with the QUAL-01 drop pattern (`slog.Warn("Dropped buffered fruit", "carrier", ...)`)
- Partial flush failure: if `CopyFrom` fails mid-recovery, re-buffer the unwritten records at the front of the buffer and continue probing — do not discard
- Transaction wrapping (RECO-04): wrap each `CopyFrom` in an explicit `BeginTx` / `Commit` / `Rollback` to prevent partial-batch duplicates on retry

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Config` struct in `breeze.go`: already has `QueueSize int` and `LogLevel string` with YAML + envconfig tags — `WriteBufferSize int` follows the same pattern
- `pool.Ping(ctx)` from pgxpool: already imported, usable as the health probe
- `pool.CopyFrom(...)` in `run()`: the existing write call that needs transaction wrapping and retry/buffer integration

### Established Patterns
- Config loading: YAML first, env override via `envconfig.Process` — new `write_buffer_size` field follows same pattern
- Logging: `slog.Error` for failures, `slog.Warn` for drops, `slog.Debug` for normal ops
- Zero-value default pattern: `if cfg.QueueSize == 0 { cfg.QueueSize = 50 }` in `run()` — apply same for `WriteBufferSize`
- Error handling: runtime errors use `slog.Error` and continue (not `os.Exit`)

### Integration Points
- `run()` in `breeze.go`: the `for { select { case <-next_batch: ... } }` loop is where the write buffer, DB health state, and reconnect probe all integrate
- `processor.Processor.Values()`: returns `[]any` rows — the buffer stores `Fruit` structs, converting to rows at flush time (or buffer stores rows directly — researcher to determine best approach)

</code_context>

<specifics>
## Specific Ideas

- No specific references — standard Go backoff and buffering patterns apply

</specifics>

<deferred>
## Deferred Ideas

- None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-write-buffer-and-reconnect*
*Context gathered: 2026-03-08*
