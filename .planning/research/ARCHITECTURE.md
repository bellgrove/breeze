# Architecture Research

**Domain:** Go MQTT-to-TimescaleDB fruit grading daemon
**Researched:** 2026-03-07
**Confidence:** HIGH — based on direct codebase analysis; pgxpool behaviour verified against official docs and upstream GitHub discussions

---

## Existing Architecture (Baseline)

The current pipeline is a straight line from MQTT callback to DB write. Every component is in the hot path:

```
MQTT broker
    │
    │ (paho library goroutine)
    ▼
Processor.OnMessage()
    │  Grademap.Update()  ← grademap topic
    │  Grademap.Grade()   ← fruit topic
    ▼
queue  chan Fruit  (cap 50)
    │
    │ (batching goroutine in processor.Create())
    ▼
next_batch  chan bool
    │
    │ (run() select loop — main goroutine)
    ▼
pool.CopyFrom() → TimescaleDB (breeze_fruit)
```

### Current component responsibilities

| Component | File | Responsibility |
|-----------|------|----------------|
| `run()` | `breeze.go` | Owns pgxpool, MQTT client, event loop |
| `Processor` | `processor/processor.go` | Bridges MQTT → DB; implements `pgx.CopyFromSource` |
| `Grademap` | `processor/grademap.go` | In-memory grade rule set; `Update()` / `Grade()` |
| `Fruit` | `processor/fruit.go` | Canonical row; `UnmarshalJSON` / `AsRow()` |
| batching goroutine | `processor/processor.go` | Coalesces timer signals; fires `next_batch` |

---

## Integration: Three Changes and Where They Slot In

### 1. Write Buffer (DB outage resilience)

**Problem with current design**

`pool.CopyFrom()` is called directly from `run()`'s select loop. When the DB is unavailable it returns an error and the batch is discarded — `Processor.Next()` / `Values()` have already drained `queue`. There is no retry; the fruit records are gone.

pgxpool does not auto-retry failed queries. It will reconnect idle connections on its `HealthCheckPeriod` (default 1 minute) and will establish a new connection on the next `Acquire()` call once the DB is back, but the application is responsible for buffering and replaying failed writes.

**Where the buffer sits**

The buffer belongs in `run()`, not in `Processor`. The reason: `Processor` currently implements `pgx.CopyFromSource` by draining `queue` on demand. If `CopyFrom` fails mid-stream, `Next()` / `Values()` have already consumed items off `queue`. A buffer added below `queue` (inside `Processor`) would fight with the existing `CopyFromSource` drain path. It is cleaner to have `run()` pull a snapshot of items before calling `CopyFrom`, then retain the snapshot for replay if the write fails.

Concretely, `run()` should:
1. On `next_batch` signal, drain `queue` into a local `[]Fruit` slice (up to the configured limit).
2. Attempt `CopyFrom` with that slice as the source.
3. On success: discard the slice.
4. On failure: prepend the slice to a `pendingBuf []Fruit` held in `run()`.
5. On every subsequent `next_batch`, try to flush `pendingBuf` first before accepting new items.

This keeps `Processor`'s `pgx.CopyFromSource` interface intact (or replaces it with a simpler `[]Fruit`-backed source) and puts retry ownership unambiguously in `run()`.

**Config addition needed**

`Config` needs a `WriteBuffer.MaxItems int` field. When `len(pendingBuf)` exceeds this limit, oldest items are dropped with a `slog.Warn`. This bounds memory during a long outage.

**Updated data flow (fruit path)**

```
MQTT callback
    ▼
Processor.OnMessage() — unchanged
    ▼
queue  chan Fruit  (cap 50) — unchanged
    ▼
batching goroutine — unchanged
    ▼
next_batch  chan bool — unchanged
    ▼
run() select loop
    ├── drain queue → []Fruit snapshot
    ├── prepend to pendingBuf (if prior failures exist)
    ├── trim pendingBuf to MaxItems (drop oldest, warn)
    └── attempt CopyFrom(pendingBuf)
            ├── success → clear pendingBuf
            └── failure → leave pendingBuf, log error, reconnect loop
```

### 2. Reconnection Loop

**How pgxpool behaves**

pgxpool does not block the caller waiting for the DB to come back. `CopyFrom` fails immediately when no healthy connection can be acquired. Retrying in a tight loop inside `run()`'s select case would starve `ctx.Done` and block MQTT message processing (because the select loop would never reach `next_batch` again while retrying synchronously).

**Recommended design: reconnect goroutine with exponential backoff**

Add a separate `reconnectLoop` goroutine, started in `run()`, that:
1. Watches a `dbHealthy bool` (or `atomic.Bool`) flag.
2. When `CopyFrom` fails in the main select loop, the loop sets `dbHealthy = false` and signals the reconnect goroutine via a channel (e.g., `needReconnect chan struct{}`).
3. The reconnect goroutine runs `pool.Ping(ctx)` on an exponential backoff schedule (start 1s, double each attempt, cap at 60s, with jitter). Standard library `time.Sleep` inside a goroutine is sufficient — no external library needed.
4. When `Ping` succeeds, it sets `dbHealthy = true` and drains `needReconnect` so the main loop knows it can attempt a flush.
5. The main select loop skips DB writes while `!dbHealthy`, preventing pointless `CopyFrom` calls that would immediately fail. Items continue to accumulate in `pendingBuf`.

This keeps the MQTT message path (the `OnMessage` → `queue` channel path) completely non-blocking throughout a DB outage.

```
run() goroutines:

  [main select loop]                    [reconnectLoop goroutine]
  ctx.Done ──────────────────────────── stop
  next_batch ──► if dbHealthy:          needReconnect ──► Ping with backoff
                   attempt flush                          on success: dbHealthy=true
                 else:
                   buffer only, signal
                   needReconnect
```

**Why not use pgxpool's BeforeAcquire hook**

`BeforeAcquire` fires per connection acquisition and is synchronous within the pool machinery. It cannot accumulate a pending buffer or signal a background goroutine cleanly. Application-level state management in `run()` is more transparent.

### 3. Grademap Versioning

**What the fruit JSON contains**

`Fruit.UnmarshalJSON` decodes `Sizer.Timestamp` (a `time.Time`) and `Sizer.Schema.Version` (an int). There is no dedicated grademap version or grademap ID field in the fruit payload. The fruit JSON references `Sizer.Grade.Id` / `Sizer.Grade.Name` (the output grade assigned by the OEM), not the grademap configuration version that produced it.

This means the "version field = clean fix" path from the pending decision in `PROJECT.md` is unavailable. The approach must be time-based: match each fruit to the grademap that was active at the fruit's `SizerTime`.

**Grademap DB write**

On every `grademap` MQTT message, write the full JSON payload (or a parsed canonical form) to a new `breeze_grademap` table with a `received_at TIMESTAMPTZ` column (server-side `NOW()`). This is an append — never update, never delete. Each row is one version of the grademap.

Schema outline:

```sql
CREATE TABLE breeze_grademap (
    received_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name         TEXT,
    raw          JSONB        -- full payload for auditability
);
```

This write should happen synchronously in `Processor.OnMessage()` before updating the in-memory `Grademap`. That ordering means `received_at` is always <= the time of the first fruit that sees the new grademap. Alternatively it can happen in `run()` via a separate `grademap` channel — see component placement below.

**Grademap timing lookup**

For reconciliation queries (or a future re-grade pass), the correct grademap for a fruit with `SizerTime = T` is:

```sql
SELECT raw
FROM breeze_grademap
WHERE received_at <= $1
ORDER BY received_at DESC
LIMIT 1;
```

This is a point-in-time lookup: "the grademap that was most recently published before this fruit was graded." It works correctly as long as `received_at` is recorded before the first fruit carrying that grademap's decisions arrives. Because the OEM applies the grademap with a pipeline delay (fruit already in transit are still graded on the old map), `received_at` will typically be earlier than the actual switch-over time — this is conservative and correct.

**Known gap (OEM pipeline delay)**

The OEM delay means: after Breeze receives a new grademap at time T, the OEM may still be grading fruit on the old map for some seconds. Fruit arriving at T+5s with `SizerTime` in that window will be graded by Breeze on the new map, but the OEM graded them on the old map. Breeze's re-grading result will diverge from the OEM's result for this window.

This is a known, documented limitation. The correct long-term fix requires either:
- Empirical measurement of the OEM delay and storing it in config as `GrademapPipelineDelay duration`, then adjusting the lookup: `WHERE received_at <= ($1 - delay)`.
- Or investigation of live traffic to see if a grademap reference is embedded anywhere in the spectrim data (not currently mapped in `UnmarshalJSON`).

For the immediate milestone, recording `received_at` and using the point-in-time lookup is the right first step. The delay compensation can be added as a config value in a subsequent phase once the delay has been measured empirically.

**Component placement for grademap writes**

Two options:

Option A — write in `Processor.OnMessage()` directly:
- Requires passing the DB pool into `Processor` (currently `Processor` has no DB reference).
- Couples domain logic to infrastructure. Not recommended for the existing design.

Option B — expose a `grademap` channel from `Processor` and write in `run()`:
- `Processor` sends the raw grademap payload `[]byte` on a new `grademapUpdate chan []byte` channel when it receives a grademap message, in addition to updating the in-memory map.
- `run()` selects on `grademapUpdate` and writes to `breeze_grademap` using a simple `pool.Exec` or `pool.QueryRow`.
- Keeps `Processor` infrastructure-free. Consistent with the existing design where `run()` owns all DB interaction.

Option B is recommended. It adds one channel to `Processor` and one select case to `run()`.

---

## Revised Architecture (Post-Milestone)

```
MQTT broker
    │
    ▼ (paho goroutine)
Processor.OnMessage()
    ├─── grademap topic ──► Grademap.Update() (in-memory)
    │                   └─► grademapUpdate  chan []byte  ──► run() select
    │                                                            └─► pool.Exec INSERT breeze_grademap
    └─── fruit topic   ──► Grademap.Grade()
                       └─► queue  chan Fruit  (cap 50)
                               │
                               ▼ (batching goroutine)
                           next_batch  chan bool
                               │
                               ▼ (run() select loop)
                           drain queue → []Fruit snapshot
                               │
                               ▼
                           pendingBuf  []Fruit  (bounded by MaxItems)
                               │
                     ┌─────────┴──────────┐
                     │ dbHealthy           │ !dbHealthy
                     ▼                     ▼
               pool.CopyFrom()       buffer only
               breeze_fruit          signal needReconnect
                     │
                     ▼ (failure)
               pendingBuf retained
               signal needReconnect
                               │
                               ▼ (reconnectLoop goroutine)
                           pool.Ping() with exponential backoff
                               └─► on success: dbHealthy=true
```

---

## Component Responsibilities (Updated)

| Component | File | New Responsibility |
|-----------|------|--------------------|
| `run()` | `breeze.go` | Owns `pendingBuf`, `dbHealthy` flag, `reconnectLoop`, grademap DB writes |
| `Processor` | `processor/processor.go` | Add `grademapUpdate chan []byte`; expose channel accessor |
| `Grademap` | `processor/grademap.go` | No change |
| `Fruit` | `processor/fruit.go` | No change |
| `reconnectLoop` | `breeze.go` (new goroutine) | `pool.Ping` with backoff; signals `run()` on recovery |
| `breeze_grademap` table | `schema.sql` (new DDL) | Append-only grademap version log |

---

## Data Flow Changes Summary

| Flow | Before | After |
|------|--------|-------|
| Fruit → DB | `queue` → `CopyFrom`, errors discarded | `queue` → `pendingBuf` → `CopyFrom`, errors retained and replayed |
| Grademap → DB | Not written | `grademapUpdate` channel → `run()` → `INSERT breeze_grademap` |
| DB connection lost | `CopyFrom` error logged, data lost | error → `pendingBuf` retained, `reconnectLoop` goroutine retries |
| DB recovered | Manual restart required | `reconnectLoop` detects via `Ping`, sets `dbHealthy=true`, flush proceeds |

---

## Build Order

These three changes have dependencies that dictate sequence:

**Phase 1: Grademap persistence (independent)**
- Add `breeze_grademap` table to `schema.sql`
- Add `grademapUpdate chan []byte` to `Processor`; emit on grademap receipt
- Add `grademapUpdate` select case in `run()` with `pool.Exec` insert
- No dependency on write buffer; can ship alone

**Phase 2: Write buffer + reconnect loop (depends on Phase 1 being stable)**
- Add `pendingBuf []Fruit` to `run()`; add `MaxItems` config field
- Add `dbHealthy atomic.Bool` and `needReconnect chan struct{}`
- Add `reconnectLoop` goroutine with exponential backoff
- Modify `run()` select to conditionally write vs buffer
- Note: this requires changing how `run()` drains `queue` — currently `Processor` implements `pgx.CopyFromSource` with direct channel drain; with buffering, `run()` needs to drain `queue` itself before calling `CopyFrom`. Consider introducing a helper `drainQueue(p *Processor, max int) []Fruit` that reads from the queue channel directly, or refactoring `Processor` to expose a `Drain(n int) []Fruit` method.

**Phase 3: Grademap timing compensation (depends on Phase 1)**
- Once `breeze_grademap` is populated with live data, measure the actual OEM pipeline delay empirically
- Add `GrademapPipelineDelay time.Duration` config field
- Queries and re-grade reconciliation can use `received_at - delay` as the effective switch-over time

---

## Anti-Patterns to Avoid

### Putting the buffer inside Processor

**What people do:** Add a `[]Fruit` overflow slice inside `Processor`, checked in `Next()`.

**Why it is wrong:** `Processor` implements `pgx.CopyFromSource` — its `Next()` / `Values()` are called by pgx during an active `CopyFrom` transaction. There is no hook for "write failed, give me back the items". Once `CopyFrom` starts consuming, failed items cannot be un-consumed from inside the interface.

**Do this instead:** Buffer in `run()`, before calling `CopyFrom`.

### Retrying CopyFrom synchronously in the select case

**What people do:** Wrap `pool.CopyFrom()` in a retry loop inside the `case <-next_batch:` branch.

**Why it is wrong:** The `run()` select loop processes MQTT batches. A blocking retry loop inside a select case prevents `ctx.Done` from being handled. If the DB is down for minutes, graceful shutdown becomes impossible and the MQTT message path (via `queue`) backs up.

**Do this instead:** Return immediately from the select case with an error, retain the buffer, and let a separate `reconnectLoop` goroutine signal when the DB is healthy again.

### Polling the DB from OnMessage

**What people do:** Write the grademap to the DB directly inside `Processor.OnMessage()`.

**Why it is wrong:** `OnMessage` runs in the paho library goroutine. Blocking it on a DB call holds up all MQTT message delivery. During a DB outage, every incoming message would block.

**Do this instead:** Send the payload on a channel and handle the DB write in `run()` where blocking behaviour is explicit and context-cancellable.

---

## Scaling Considerations

This is a single-machine embedded daemon. The single Tomra unit produces bounded throughput (fruit per second is physically constrained). Scaling is not a concern.

The only capacity parameter that matters is `WriteBuffer.MaxItems` — it should be sized to cover the maximum tolerable outage duration multiplied by the peak fruit throughput (approximately 2–5 fruit/second for a single-lane grader). A default of 10,000 items covers ~30 minutes at peak, using roughly 10–15 MB of heap.

---

## Sources

- Codebase direct analysis: `breeze.go`, `processor/processor.go`, `processor/fruit.go`, `processor/grademap.go`
- [pgxpool package documentation](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) — `HealthCheckPeriod`, `BeforeAcquire`, connection lifecycle
- [pgx GitHub Discussion #1415: Recovering from a connection interruption when using pgxpool](https://github.com/jackc/pgx/discussions/1415) — confirms pool does not auto-retry queries; circuit-breaker recommendation

---

*Architecture research for: Breeze MQTT/TimescaleDB fruit grading daemon — milestone integration*
*Researched: 2026-03-07*
