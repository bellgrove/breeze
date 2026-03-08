# Phase 2: Write Buffer and Reconnect - Research

**Researched:** 2026-03-08
**Domain:** Go in-memory buffering, pgx v5 transactions, exponential backoff reconnect
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **Write buffer capacity:** Default 10,000 records (~15 MB heap). Configurable via `write_buffer_size` field in `Config` struct, same YAML + envconfig pattern as `QueueSize` and `LogLevel`. Drop-oldest policy when buffer is full (RECO-01).
- **Write buffer location:** Buffer lives in `run()` as a local `[]Fruit` slice — not inside `Processor`. Only the write loop touches the buffer; no concurrency issues, no locking needed.

### Claude's Discretion

- Reconnect probe: background goroutine using `pool.Ping` with exponential backoff; parameters (initial delay, max cap, multiplier) can be hardcoded at reasonable defaults (e.g. 1s initial, 60s max, 2x multiplier) unless a compelling reason to make them configurable emerges during research
- Overflow logging: per-drop `slog.Warn` consistent with the QUAL-01 drop pattern (`slog.Warn("Dropped buffered fruit", "carrier", ...)`)
- Partial flush failure: if `CopyFrom` fails mid-recovery, re-buffer the unwritten records at the front of the buffer and continue probing — do not discard
- Transaction wrapping (RECO-04): wrap each `CopyFrom` in an explicit `BeginTx` / `Commit` / `Rollback` to prevent partial-batch duplicates on retry

### Deferred Ideas (OUT OF SCOPE)

- None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| RECO-01 | Fruit records are buffered in a bounded in-memory slice when `CopyFrom` fails (configurable max size, drop-oldest policy) | Slice-based buffer with `append` + length check is idiomatic Go; no library needed |
| RECO-02 | A background reconnect probe (`pool.Ping` with exponential backoff) runs when DB is unhealthy | `pgxpool.Pool.Ping(ctx)` is the correct probe; pure-stdlib backoff loop with `time.NewTimer` and `select` respects context cancellation |
| RECO-03 | Buffered records are automatically flushed when DB connection is restored | Flush-on-reconnect using `pgx.CopyFromRows` over buffered `[][]any` is the correct approach; re-buffer on failure |
| RECO-04 | All `CopyFrom` calls are wrapped in explicit transactions to prevent partial-batch duplicates on retry | `pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error { _, err = tx.CopyFrom(...) ; return err })` is the idiomatic pattern |
</phase_requirements>

---

## Summary

Phase 2 adds three cooperating mechanisms to `run()` in `breeze.go`: (1) a bounded `[]Fruit` write buffer that accumulates records when the DB is unavailable, (2) a DB-healthy boolean flag that gates whether new batches write directly or are buffered, and (3) a background reconnect goroutine that probes with `pool.Ping` using exponential backoff and signals `run()` when the DB recovers so buffered records can be flushed.

The buffer is a plain `[]Fruit` slice in `run()`. No library is needed. The drop-oldest policy is a one-liner: when the buffer is full before appending, drop `buffer[0]`, log a warning, and shift. All writes (normal and flush) are wrapped in `pgx.BeginTxFunc` with `tx.CopyFrom` so that a mid-batch failure leaves no partial rows and the batch is safe to retry intact.

The reconnect goroutine is a stdlib-only loop: `time.NewTimer` with doubled delay, capped at 60 s, reset on each failed ping, cancelled via `ctx`. No external backoff library is needed or warranted for this workload.

**Primary recommendation:** Implement a DB-healthy state flag, a `[]Fruit` buffer, a `chan struct{}` reconnect signal, and a background probe goroutine — all inside `run()` — with zero new dependencies.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/jackc/pgx/v5` | v5.7.5 (already in go.mod) | `pgx.BeginTxFunc`, `pgx.Tx.CopyFrom`, `pgxpool.Pool.Ping` | Already imported; provides first-class transaction + bulk-copy on `Tx` |
| `github.com/jackc/pgx/v5/pgxpool` | same | Pool-level `BeginTx`, `Ping` | Already in use in `run()` |
| stdlib `time` | — | Timer-based backoff loop | No external dependency warranted |
| stdlib `sync/atomic` or plain bool | — | DB-healthy flag (single goroutine, no lock needed) | Buffer and flag only touched by `run()` goroutine; no concurrency |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib backoff loop | `cenkalti/backoff` | Library adds no value for a single probe loop; stdlib is clearer |
| `[]Fruit` slice buffer | ring-buffer library | Slice with manual drop-oldest is 10 lines; ring-buffer library adds a dependency for zero gain |
| `pgx.BeginTxFunc` helper | manual `BeginTx` / defer `Rollback` / `Commit` | Helper is more concise and less error-prone; either is correct |

**Installation:** No new packages. All required packages already in `go.mod`.

---

## Architecture Patterns

### Recommended Integration in `run()`

The existing `for { select { ... } }` loop gains three new concerns:

```
run()
├── dbHealthy bool          — false when last write or ping failed
├── writeBuffer []Fruit     — accumulates when !dbHealthy
├── reconnectCh chan struct{} — closed/sent by probe goroutine on recovery
└── probe goroutine         — started when dbHealthy flips to false; stopped on recovery
```

No new files required. All changes are in `breeze.go`.

### Pattern 1: Bounded Drop-Oldest Buffer

**What:** A `[]Fruit` slice bounded by `cfg.WriteBufferSize`. When full, drop index 0 (oldest), log a warn, then append.

**When to use:** Every time a `CopyFrom` fails and the batch needs to be buffered for later.

```go
// Source: standard Go slice idiom; no external reference needed
func bufferFruits(buf []Fruit, incoming []Fruit, maxSize int) []Fruit {
    for _, f := range incoming {
        if len(buf) >= maxSize {
            slog.Warn("Dropped buffered fruit", "carrier", buf[0].CarrierId)
            buf = buf[1:] // drop oldest; amortised O(1) for small N, fine for 10k limit
        }
        buf = append(buf, f)
    }
    return buf
}
```

Note: `buf[1:]` produces a new slice header sharing the backing array. For a 10,000-element cap this is acceptable; no heap churn from repeated drops because the backing array is reused. If drop frequency is expected to be very high (sustained outage), consider `copy(buf, buf[1:])` then `buf = buf[:len(buf)-1]` to avoid the slow memory drift — but for a 10k cap this is unlikely to matter.

### Pattern 2: CopyFrom Inside a Transaction

**What:** Wrap every `CopyFrom` in `pgx.BeginTxFunc` so partial batches are never committed.

**When to use:** Every write — both normal path and flush path.

```go
// Source: pgx v5 official docs — pkg.go.dev/github.com/jackc/pgx/v5
err = pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
    _, err := tx.CopyFrom(
        ctx,
        pgx.Identifier{"breeze_fruit"},
        processor.Columns(),
        rowSource, // pgx.CopyFromSource
    )
    return err
})
```

`pgx.BeginTxFunc` commits if the function returns nil, rolls back on any error. The `pgxpool.Tx` type has a `CopyFrom` method directly — no `conn.Raw()` hack is needed (that pattern is only required when using `database/sql`).

**Flush row source:** When flushing the buffer, convert `[]Fruit` to `[][]any` and use `pgx.CopyFromRows`:

```go
// Source: pgx v5 official docs — CopyFromRows
rows := make([][]any, len(writeBuffer))
for i, f := range writeBuffer {
    row, err := f.AsRow()
    if err != nil {
        return err
    }
    rows[i] = row
}
rowSrc := pgx.CopyFromRows(rows)
```

This avoids modifying `Processor` for the flush path. The existing `Processor` is still used for the normal (live) path because it implements `CopyFromSource` directly.

### Pattern 3: Exponential Backoff Reconnect Probe

**What:** A goroutine that pings the DB with increasing delays, signals `run()` on success, exits on context cancellation.

**When to use:** Started exactly once when `dbHealthy` flips from `true` to `false`.

```go
// Source: stdlib pattern; context-aware timer loop
func startReconnectProbe(ctx context.Context, pool *pgxpool.Pool, reconnectCh chan<- struct{}) {
    go func() {
        delay := time.Second // 1s initial
        const maxDelay = 60 * time.Second
        t := time.NewTimer(delay)
        defer t.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-t.C:
                if err := pool.Ping(ctx); err != nil {
                    slog.Warn("DB probe failed, retrying", "delay", delay, "err", err)
                    delay = min(delay*2, maxDelay)
                    t.Reset(delay)
                    continue
                }
                // DB is back
                slog.Info("DB reconnected")
                reconnectCh <- struct{}{}
                return
            }
        }
    }()
}
```

`min` is a builtin since Go 1.21. The project uses go 1.24.3 so this is safe.

### Pattern 4: DB-Healthy State Machine in `run()` Select Loop

**What:** The `run()` select gains a `reconnectCh` case. Normal batch writes set `dbHealthy = false` on error and buffer the records; the reconnect case flushes the buffer and restores `dbHealthy = true`.

```go
// Sketch only — exact code is the planner's job
var (
    writeBuffer []Fruit
    dbHealthy   = true
    reconnectCh = make(chan struct{}, 1)
)
if cfg.WriteBufferSize == 0 {
    cfg.WriteBufferSize = 10_000
}

for {
    select {
    case <-ctx.Done():
        return nil

    case <-next_batch:
        batch := drainProcessor(&proc) // collect []Fruit from proc queue
        if !dbHealthy {
            writeBuffer = bufferFruits(writeBuffer, batch, cfg.WriteBufferSize)
            continue
        }
        if err := writeBatch(ctx, pool, batch); err != nil {
            slog.Error("Failed to write", "err", err)
            dbHealthy = false
            writeBuffer = bufferFruits(writeBuffer, batch, cfg.WriteBufferSize)
            startReconnectProbe(ctx, pool, reconnectCh)
        }

    case <-reconnectCh:
        if err := flushBuffer(ctx, pool, writeBuffer); err != nil {
            slog.Error("Flush failed, re-buffering", "err", err)
            // Keep writeBuffer intact; probe will retry
            startReconnectProbe(ctx, pool, reconnectCh)
            continue
        }
        writeBuffer = writeBuffer[:0] // reset without deallocation
        dbHealthy = true
    }
}
```

### Key Integration Question: How to Drain Processor for Buffering

The existing `Processor` implements `CopyFromSource` (`Next`/`Values`/`Err`). When the DB is unhealthy, the batch arrived via `<-next_batch` but we need the `Fruit` structs for buffering — not `[][]any` rows. Two options:

**Option A (recommended):** Drain `proc` by calling `proc.Values()` N times (where N is `proc.counter.Load()`) to collect `[]Fruit` before deciding whether to write or buffer. Buffer stores `Fruit` structs. Flush converts to `[][]any` at flush time.

**Option B:** Buffer `[][]any` rows (already-converted). Simpler flush path. But loses the struct for the warn log's `carrier` field.

Option A is recommended because: (1) CONTEXT.md's overflow log example logs `"carrier", f.CarrierId` which requires the struct, and (2) it aligns with CONTEXT.md's explicit statement "the buffer stores `Fruit` structs."

Concretely, `drainProcessor` looks like:

```go
func drainProcessor(proc *processor.Processor) []Fruit {
    var out []Fruit
    for proc.Next() {
        row, _ := proc.Values() // Values drains the queue and returns []any
        // But Values() returns []any not Fruit — this approach doesn't reconstruct Fruit
    }
}
```

Wait — `Values()` returns `[]any`, not `Fruit`. The channel inside `Processor` holds `Fruit` and the counter tracks items. We cannot reconstruct a `Fruit` from `[]any`.

**Revised recommendation:** Expose the fruit structs from the Processor before conversion, or change `drainProcessor` to read from `proc.queue` directly. Since `proc.queue` is unexported, the cleanest approach is to add a package-level helper `DrainFruits(p *Processor) []Fruit` in the `processor` package that drains the queue channel into a slice. This is a small, contained change.

Alternatively: the buffer could store `[][]any` rows and accept that drop-warn logs only show `"count", 1` rather than `"carrier", f.CarrierId`. Review this with the planner — the choice affects the `Processor` package interface.

**Planner decision needed:** Either (a) add `processor.DrainFruits` to expose the queue, or (b) buffer `[][]any` and log drops without the carrier ID. Both are valid; the CONTEXT.md suggests carrier ID logging, pointing toward option (a).

### Anti-Patterns to Avoid

- **Starting multiple probe goroutines:** Guard with a `probeRunning bool` flag or nil-check the goroutine handle before calling `startReconnectProbe`. Without this, every failing write in an outage spawns a new goroutine.
- **Buffering after flush failure:** On flush failure, do NOT re-append to `writeBuffer` — the records are already there. Just restart the probe.
- **Calling `pool.Ping` without a deadline:** Pass a context with a short timeout for the ping to avoid a goroutine blocked indefinitely on a hung TCP connection. `context.WithTimeout(ctx, 5*time.Second)` per ping is appropriate.
- **Using `buf = buf[1:]` in a hot loop for huge buffers:** For the 10k cap this is fine. If the cap were much larger, a proper ring buffer would prevent slow memory drift. Document this as a known limitation.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Transactional bulk insert | Custom BEGIN/COPY/COMMIT with manual error paths | `pgx.BeginTxFunc` + `tx.CopyFrom` | `BeginTxFunc` handles rollback on error automatically; missing a `defer Rollback` is a classic bug |
| Converting slice to CopyFromSource | Custom iterator struct | `pgx.CopyFromRows(rows)` | Built into pgx; already tested |
| Exponential backoff | Custom recursive sleep or ticker | Inline `time.NewTimer` + `delay *= 2` with `min(delay, cap)` | Simple enough to inline; external library adds no value |
| Concurrency-safe buffer | Channel-based buffer or mutex-guarded slice | Plain `[]Fruit` in single goroutine | Buffer is only touched by `run()`'s single goroutine — no concurrency |

**Key insight:** This phase is a pure single-goroutine data flow change. The write buffer and state machine are internal to `run()`'s event loop, so no synchronization primitives are needed beyond the existing channels.

---

## Common Pitfalls

### Pitfall 1: Multiple Probe Goroutines

**What goes wrong:** Every failed `CopyFrom` in `<-next_batch` starts a new probe goroutine. After 10 failed writes you have 10 probes pinging simultaneously. On reconnect you get 10 signals in `reconnectCh` and 10 flush attempts.

**Why it happens:** No guard on `startReconnectProbe` call site.

**How to avoid:** Track a `probeRunning bool` flag. Set it `true` before launching the goroutine, `false` after receiving `reconnectCh`. Only call `startReconnectProbe` when `!probeRunning`.

**Warning signs:** Log shows multiple "DB probe failed" lines at identical timestamps with different delay values.

### Pitfall 2: Ping Context Not Time-Bounded

**What goes wrong:** `pool.Ping(ctx)` with the root context hangs for minutes on a TCP-level outage (no RST from server), blocking the probe goroutine indefinitely. The reconnect signal never fires even when the DB restarts.

**Why it happens:** `pgxpool.Ping` acquires a connection and sends an empty query. If the TCP connection is in a half-open state, it may not time out until the OS TCP keepalive fires (minutes).

**How to avoid:** Wrap ping in a short timeout: `pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second); defer cancel(); pool.Ping(pingCtx)`.

**Warning signs:** Probe goroutine is stuck and emitting no log output for minutes after the DB was restarted.

### Pitfall 3: Flush Failure Doubles Buffer

**What goes wrong:** On a failed flush, code re-appends the already-buffered records, doubling the buffer contents. On next successful flush, rows are written twice.

**Why it happens:** Treating flush failure the same as a write failure — feeding records back through `bufferFruits`.

**How to avoid:** On flush failure, leave `writeBuffer` unchanged (records are already in it). Simply restart the probe. Do not re-append.

**Warning signs:** Duplicate rows in `breeze_fruit` after an outage cycle.

### Pitfall 4: Holding Processor Items While Unhealthy

**What goes wrong:** On the normal write path, `CopyFrom` drains `proc` (via `Next`/`Values`). If the write is attempted and fails, the items are already consumed from the channel — they cannot be replayed from `proc`. If code tries to buffer "what was in proc" after the fact, it buffers nothing.

**Why it happens:** `CopyFrom` calls `Values()` which drains the internal queue channel. By the time the error is returned, the queue is empty.

**How to avoid:** Drain `proc` into a local `[]Fruit` (or `[][]any`) slice FIRST, then decide whether to write or buffer. Never pass `&proc` directly to `CopyFrom` when buffering is possible; use `pgx.CopyFromRows` over the pre-collected slice.

**Warning signs:** Buffer remains empty after an outage even though MQTT traffic continued.

### Pitfall 5: `buf = buf[1:]` Memory Leak for Very Long Outages

**What goes wrong:** Repeatedly doing `buf = buf[1:]` advances the slice header but does not release the backing array's head. Over a very long outage with constant drops at capacity, the backing array grows unbounded (the dropped elements are not GC'd).

**Why it happens:** Go slices share backing arrays; re-slicing does not shrink the allocation.

**How to avoid:** After a certain number of drops (or on buffer reset after flush), do `newBuf := make([]Fruit, len(buf)); copy(newBuf, buf); buf = newBuf`. For a 10,000-element cap this is negligible — call it out as a known trade-off in the code comment.

**Warning signs:** Heap grows monotonically during a sustained DB outage despite the buffer appearing bounded.

---

## Code Examples

Verified patterns from official sources:

### Transaction + CopyFrom (pgx v5 idiomatic)

```go
// Source: pkg.go.dev/github.com/jackc/pgx/v5 — BeginTxFunc
err = pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
    _, err := tx.CopyFrom(
        ctx,
        pgx.Identifier{"breeze_fruit"},
        processor.Columns(),
        pgx.CopyFromRows(rows), // rows is [][]any
    )
    return err
})
// BeginTxFunc commits on nil return, rolls back on error.
```

### Ping with Timeout

```go
// Source: pkg.go.dev/github.com/jackc/pgx/v5/pgxpool — Ping
pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
if err := pool.Ping(pingCtx); err != nil {
    // DB still unavailable
}
```

### Exponential Backoff Loop (stdlib only)

```go
// Source: stdlib time package — standard retry pattern
delay := time.Second
const maxDelay = 60 * time.Second
t := time.NewTimer(delay)
defer t.Stop()
for {
    select {
    case <-ctx.Done():
        return
    case <-t.C:
        if err := tryReconnect(); err != nil {
            delay = min(delay*2, maxDelay)
            t.Reset(delay)
            continue
        }
        return // success
    }
}
```

### CopyFromRows for Buffer Flush

```go
// Source: pkg.go.dev/github.com/jackc/pgx/v5 — CopyFromRows
rows := make([][]any, len(writeBuffer))
for i := range writeBuffer {
    row, err := writeBuffer[i].AsRow()
    if err != nil {
        return err
    }
    rows[i] = row
}
// Pass pgx.CopyFromRows(rows) to tx.CopyFrom
```

### Config Field Addition (zero-value default pattern)

```go
// Source: existing breeze.go pattern (QueueSize)
type Config struct {
    // ... existing fields ...
    WriteBufferSize int `yaml:"write_buffer_size" envconfig:"WRITE_BUFFER_SIZE"`
}

// In run(), after cfg is loaded:
if cfg.WriteBufferSize == 0 {
    cfg.WriteBufferSize = 10_000
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `conn.CopyFrom` requiring `conn.Raw()` for tx | `tx.CopyFrom` directly on `pgx.Tx` / `pgxpool.Tx` | pgx v5 | No need for `database/sql` `Raw()` hack |
| Manual `BeginTx` / `defer Rollback` / `Commit` | `pgx.BeginTxFunc` helper | pgx v5 | Less boilerplate, rollback on error is automatic |
| `math.Min` for float64 delay cap | `min()` builtin | Go 1.21 | One less import; project uses Go 1.24.3 |

---

## Open Questions

1. **How to expose `[]Fruit` from Processor for buffering**
   - What we know: `proc.queue` is unexported; `Values()` returns `[]any`; `Processor` is the `CopyFromSource` passed to `CopyFrom`
   - What's unclear: Whether to add `processor.DrainFruits(p *Processor) []Fruit` or buffer `[][]any` rows instead of `Fruit` structs
   - Recommendation: Add a `DrainFruits` function to the `processor` package. It reads from the internal queue (accessing the unexported field from within the package) and returns `[]Fruit`. This is a small, contained change that preserves the carrier-ID warn log. The planner should create a task for this.

2. **Probe goroutine lifecycle on graceful shutdown**
   - What we know: Context cancellation propagates to the probe goroutine via `ctx.Done()`
   - What's unclear: Whether buffered records should be flushed on shutdown (before context cancels) or abandoned
   - Recommendation: Abandon buffered records on shutdown — attempting a flush on shutdown while the DB may still be down adds complexity for little benefit. Document this in a code comment. (The CONTEXT.md does not specify shutdown-flush behaviour.)

---

## Validation Architecture

nyquist_validation is enabled in .planning/config.json.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go standard testing (`testing` package) |
| Config file | none — `go test ./...` from repo root |
| Quick run command | `go test ./... -count=1 -timeout 30s` |
| Full suite command | `go test ./... -count=1 -race -timeout 60s` |

### Phase Requirements -> Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| RECO-01 | Buffer accumulates fruits on write failure; drops oldest when full; logs warn per drop | unit | `go test ./... -run TestWriteBuffer -count=1` | Wave 0 |
| RECO-02 | Probe goroutine starts when `dbHealthy` is false; uses exponential backoff; exits on ctx cancel | unit | `go test ./... -run TestReconnectProbe -count=1` | Wave 0 |
| RECO-03 | Buffered records are flushed when reconnectCh fires; buffer is cleared on success; probe restarted on flush failure | unit | `go test ./... -run TestFlushBuffer -count=1` | Wave 0 |
| RECO-04 | Each `CopyFrom` (normal and flush) is wrapped in a transaction; failure leaves no partial rows | unit (mock) | `go test ./... -run TestTransactionalWrite -count=1` | Wave 0 |

All tests are pure unit tests using mocks or table-driven logic — no live DB required. The existing test style (mock `mqtt.Message`, direct struct construction) is the model.

### Sampling Rate

- **Per task commit:** `go test ./... -count=1 -timeout 30s`
- **Per wave merge:** `go test ./... -count=1 -race -timeout 60s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `breeze_test.go` (or `buffer_test.go`) — covers RECO-01 through RECO-04; does not exist yet
- [ ] Consider `processor/drain_test.go` if `DrainFruits` is added to the processor package

*(Existing processor tests in `processor/processor_test.go` cover Phase 1 requirements and need not be changed.)*

---

## Sources

### Primary (HIGH confidence)

- `pkg.go.dev/github.com/jackc/pgx/v5` — `BeginTxFunc`, `CopyFromSource`, `CopyFromRows`, `TxOptions`
- `pkg.go.dev/github.com/jackc/pgx/v5/pgxpool` — `Pool.BeginTx`, `Pool.Ping`, `Tx.CopyFrom`, `Tx.Commit`, `Tx.Rollback`
- Go stdlib `time` package — `time.NewTimer`, timer reset pattern for backoff

### Secondary (MEDIUM confidence)

- `danp.net/posts/pgx-tx-copy/` — Transaction + CopyFrom pattern (confirms `database/sql` requires `conn.Raw()` but native pgx `Tx.CopyFrom` does not)
- `github.com/jackc/pgx` source (`pgxpool/tx.go`) — confirms `CopyFrom` method on `pgxpool.Tx`

### Tertiary (LOW confidence)

- Medium / dev.to backoff articles — confirmed that stdlib `time.NewTimer` + `select` + `ctx.Done()` is the idiomatic approach; no authoritative source links these directly to pgxpool probe pattern

---

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — pgx v5.7.5 already in go.mod; API signatures verified via pkg.go.dev
- Architecture: HIGH — single-goroutine buffer with no locking is confirmed by CONTEXT.md; transaction pattern verified against pgx docs
- Pitfalls: MEDIUM — multiple-probe and flush-doubling pitfalls are logic deductions from the design; confirmed by general Go goroutine hygiene literature
- Test map: HIGH — mirrors existing test file patterns in the repo

**Research date:** 2026-03-08
**Valid until:** 2026-06-08 (pgx v5 API is stable; Go stdlib is stable)
