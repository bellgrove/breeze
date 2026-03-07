# Pitfalls Research

**Domain:** Go daemon — MQTT ingestion, reconnection buffering, TimescaleDB bulk write
**Researched:** 2026-03-07
**Confidence:** HIGH (most findings verified against pgx/paho source and official docs; grademap timing section is MEDIUM due to OEM-specific domain)

---

## Critical Pitfalls

### Pitfall 1: pgxpool Treats a Dead Pool as Healthy Until a Write Is Attempted

**What goes wrong:**
When PostgreSQL restarts (or drops TCP connections), pgxpool does not proactively detect the dead connections. The pool's periodic health check pings connections that have been idle for at least 1 second (default `ShouldPing` behaviour), but during a PostgreSQL outage the health check goroutine itself may be blocked waiting on a TCP timeout rather than receiving an immediate connection-refused. The connections appear to be "alive" to the pool's bookkeeping until an application-level `CopyFrom` or `Acquire` attempt actually exercises the socket. The first write after a restart therefore gets a socket error rather than a clean reconnect.

In the current codebase `run()` receives from `next_batch` and calls `pool.CopyFrom()` directly. If that call fails, the error is logged and the rows are discarded — no retry, no buffering, no reconnect probe. Any fruit that arrived during a DB outage is gone.

**Why it happens:**
`pgxpool` is lazy by design: it reconnects on the next `Acquire`, not on connection loss. There are no application-level hooks that fire on connection failure, only on connection creation (`PrepareConn`/`BeforeAcquire`). The library also lacks a built-in back-off/retry mechanism for repeated connection failures (verified: [pgx discussion #1415](https://github.com/jackc/pgx/discussions/1415)).

**How to avoid:**
- Implement an application-level write buffer (the active PROJECT.md requirement). When `CopyFrom` returns an error, enqueue the batch into the buffer instead of discarding it.
- Use an exponential-backoff reconnect probe (a separate goroutine that calls `pool.Ping(ctx)` at increasing intervals). When the ping succeeds, flush the buffer.
- Set `config.MaxConnLifetime` and `config.MaxConnLifetimeJitter` so old connections cycle after a restart rather than accumulating in an error state. Example: `MaxConnLifetime = 30 * time.Minute`, `MaxConnLifetimeJitter = 5 * time.Minute`.
- Set `config.HealthCheckPeriod` and `config.PingTimeout` explicitly. The default `HealthCheckPeriod` is 1 minute; during that window dead connections remain in the pool. Setting `PingTimeout` (e.g. `5 * time.Second`) prevents the health check goroutine from hanging on a dead TCP socket.
- Do not use the deprecated `BeforeAcquire` hook; use `PrepareConn` instead (the v5 API changed).

**Warning signs:**
- Log line "Failed to write" followed immediately by silence (no retry, no reconnect attempt).
- `pool.Stat().TotalConns()` stays non-zero but `pool.Ping(ctx)` returns an error.
- MQTT messages continuing to arrive while DB write count stops incrementing.

**Phase to address:** Reconnection + buffering phase (the active milestone requirement).

---

### Pitfall 2: Write Buffer Drain Race at Shutdown — Pool Closes Before Final Flush

**What goes wrong:**
This bug already exists in the codebase (documented in CONCERNS.md) and will become more severe once a write buffer is added. The defer order in `run()` is:

```
defer pool.Close()       // line 99 — runs third
defer close(next_batch)  // line 102 — runs second
```

(Defers run LIFO.) When context is cancelled, `close(next_batch)` fires, signalling the batch goroutine to stop. But the batch goroutine may be mid-write. `pool.Close()` fires next, terminating any in-flight `CopyFrom`. With a write buffer added, the flush-on-reconnect path also needs to complete before the pool closes.

**Why it happens:**
The defer stack does not model the dependency: pool must stay open until all writes complete. Adding a buffer introduces a second flush path (reconnect-triggered flush) that is also not covered by the current shutdown sequence.

**How to avoid:**
- Replace implicit defer ordering with an explicit shutdown sequence: signal the batch goroutine to drain, wait for it to confirm drain complete (via a done channel), then close the pool.
- The buffer flush goroutine must also receive a context that is cancelled only after the last write completes.
- Use `sync.WaitGroup` or a done channel to gate `pool.Close()`.

**Warning signs:**
- Rows visible in the queue/buffer metric at shutdown that never appear in the database.
- Log shows "Gracefully shutting down" followed immediately by zero "Wrote rows" lines even when items were buffered.

**Phase to address:** Reconnection + buffering phase. Must be fixed as part of that work, not deferred.

---

### Pitfall 3: Paho MQTT Dispatcher Goroutine Blocks When the Fruit Channel Is Full

**What goes wrong:**
`OnMessage` runs on the paho MQTT dispatcher goroutine. The current code does a blocking send on `p.queue <- f` (processor.go line 38) and then a second blocking send on `p.timer <- true` (line 40). The `queue` channel has capacity 50. The `timer` channel has capacity 1.

If the DB is down and the write path is blocked, the batch goroutine stops consuming from `queue`. Within 50 messages the `p.queue <- f` send blocks the paho dispatcher. With `OrderMatters = true` (the default), all subsequent MQTT messages on all subscribed topics — including grademap updates — are stalled. The broker will eventually consider the client unresponsive.

Adding a write buffer does not fix this by itself. If the buffer is implemented as a `[]Fruit` slice that is flushed only when the DB reconnects, the queue will still fill and block paho during the outage.

**Why it happens:**
Paho v3 uses a single-goroutine dispatcher when `OrderMatters = true`. Any blocking operation in a message handler stalls the entire client. This is documented in the paho package and confirmed by [issue #474](https://github.com/eclipse-paho/paho.mqtt.golang/issues/474).

**How to avoid:**
- The `OnMessage` handler must never block. Use a non-blocking channel send with a drop-and-warn strategy when the queue is full:
  ```go
  select {
  case p.queue <- f:
      p.counter.Add(1)
      p.timer <- true // also non-blocking; timer channel already has cap 1
  default:
      slog.Warn("Queue full, dropping fruit", "carrier", f.CarrierId)
  }
  ```
- Alternatively set `opts.SetOrderMatters(false)` to allow concurrent handlers, but this changes message-ordering guarantees and requires re-evaluating grademap update concurrency safety.
- Make the buffer capacity configurable and sized relative to expected burst rates.

**Warning signs:**
- MQTT `connectLostHandler` fires during a DB outage (paho timeout, not DB issue).
- No new fruit messages logged after the queue fills, even though the broker is healthy.
- `pool.Stat().AcquiredConns()` stuck at 1 indefinitely.

**Phase to address:** Reconnection + buffering phase. The non-blocking send must be implemented at the same time as the buffer.

---

### Pitfall 4: Counter/Channel Race in `Processor` Makes `CopyFrom` Non-Deterministic Under Load

**What goes wrong:**
`Next()` reads `p.counter.Load() > 0` to decide whether more rows exist. `Values()` does `p.counter.Add(-1)` before reading from the channel. Between the decrement and the channel receive, another goroutine calling `Next()` can observe counter == 0 and stop iteration. In practice pgx's `CopyFrom` calls `Next()` and `Values()` from a single goroutine, so this race may not fire today. However, if reconnection logic or buffer flush involves concurrent `CopyFrom` calls, the race becomes live and rows will be silently dropped mid-batch.

**Why it happens:**
The counter and the channel are two independent state variables. Their combined invariant ("counter equals number of items in channel") is not maintained atomically. The correct pattern is to read from the channel first, then decrement (or avoid a separate counter entirely and derive length from the channel).

**How to avoid:**
- Fix the counter/channel ordering before adding reconnection: dequeue from `queue` first, then decrement counter.
- Better: remove the counter entirely. `len(p.queue)` is safe to read from one goroutine; pgx calls `Next/Values` sequentially, so `len(p.queue) > 0` as the `Next()` implementation is sufficient and removes the separate atomic.
- Add a concurrent `CopyFrom` test before any reconnection work touches the `Processor`.

**Warning signs:**
- DB write rows count is occasionally less than the batch trigger count (e.g. batch fires at 30 items but only 28 are written).
- Increased "queue is empty" errors in logs at higher message rates.

**Phase to address:** Code cleanup / pre-reconnection phase. Must be resolved before writing reconnection logic on top of it.

---

### Pitfall 5: `CopyFrom` Without a Transaction Leaves Partial Batches in the Database on Error

**What goes wrong:**
PostgreSQL's COPY protocol is transactional only when wrapped in an explicit transaction. The current `pool.CopyFrom(...)` call in `run()` is not inside a transaction. If `CopyFrom` errors partway through a batch (e.g. a constraint violation on one row, or a mid-copy DB disconnect), the rows already sent before the error are committed to the database while the remaining rows are discarded. After reconnection, the buffer replay will attempt to re-insert the full batch, producing duplicate rows for the portion that was already committed.

**Why it happens:**
The pgx documentation does not make this explicit, but PostgreSQL's COPY is auto-committed per statement when not inside a transaction. This is consistent with standard SQL; callers assume transactions are needed only for multi-statement sequences. The pgx v5 API surface does not prevent or warn about non-transactional COPY.

**How to avoid:**
- Wrap each `CopyFrom` in an explicit transaction:
  ```go
  tx, err := pool.Begin(ctx)
  // ... defer tx.Rollback(ctx)
  rows, err := tx.CopyFrom(...)
  // ... tx.Commit(ctx)
  ```
- On error, roll back the transaction. The batch is fully un-committed; retry the whole batch from the buffer.
- This also makes "at-least-once" delivery semantics clear: on rollback, retain the batch in the buffer and retry.

**Warning signs:**
- Duplicate rows in `breeze_fruit` with the same `carrier_id` after a reconnection event.
- Row count in DB after recovery is larger than expected batch size.

**Phase to address:** Reconnection + buffering phase. The transaction wrapper must be added at the same time as the retry buffer.

---

### Pitfall 6: Grademap Timing — Immediate Switch Misattributes Fruit in Transit

**What goes wrong:**
When the OEM publishes a new grademap, fruit already on the grader's physical belt has been scanned under the old grademap and will receive grades assigned by the old rules. The OEM applies the new grademap to grading after a pipeline delay (typically seconds to tens of seconds). Breeze currently calls `p.grademap.Update()` immediately on receiving the MQTT message, so fruit that arrives in the next few seconds is re-graded against the new rules even though the OEM used the old rules. This creates a window of misattribution.

The size of the window depends on the grader belt speed and pipeline depth — both OEM-specific and not currently known.

**Why it happens:**
The grademap is a retained MQTT topic; it does not include a version field or effective timestamp in the known payload structure. Without a synchronisation field in the fruit message, the only way to know which grademap was active when a fruit was graded is to model the delay or to investigate whether the fruit JSON includes a version or sequence number that correlates with the grademap.

**How to avoid:**
- **Investigate first:** Inspect a real fruit payload for any field that could serve as a grademap version or sequence number. This is explicitly called out as an open decision in PROJECT.md. Do this before implementing any timing fix.
- **If a version field exists:** Store each grademap with its version in TimescaleDB. At write time, join fruit record to grademap version. Clean and auditable.
- **If no version field exists:** Model the delay. Store each grademap with an ingestion timestamp plus a configurable `grademap_propagation_delay` offset. Fruit with `received_at` between `grademap_received_at` and `grademap_received_at + delay` is ambiguous — flag it rather than silently misattributing.
- Do not apply the new grademap to fruit already in the queue when the grademap update arrives (the queue may contain fruit scanned under the old rules).

**Warning signs:**
- Grade distribution shifts sharply at exact grademap update timestamps (no smooth transition).
- The same `carrier_id` range appears under two different primary defect codes.
- Fruit with `received_at` within 30 seconds of a grademap change has anomalous defect distribution.

**Phase to address:** Grademap persistence + timing reconciliation phase. The investigation step (inspect fruit JSON for version field) is a prerequisite gate for that phase.

---

### Pitfall 7: `json.Unmarshal` Errors Silently Enqueue Zero-Value Fruit

**What goes wrong:**
`OnMessage` in processor.go line 29 discards the error from `json.Unmarshal`. A zero-value `Fruit` (all empty strings, zero floats, nil maps) is graded (producing empty defect codes) and enqueued. The batch writes it to TimescaleDB. The database gains a row with no meaningful data. This is invisible at the application layer: no error, no counter, no log line.

With reconnection buffering added, zero-value fruit will also be stored in the in-memory buffer and replayed, compounding the corruption.

**Why it happens:**
Go's `json.Unmarshal` returns an error that must be explicitly checked. The current code does not check it.

**How to avoid:**
- Check the error immediately after `json.Unmarshal`. On error, log the raw payload (truncated) and return without enqueuing.
- Add a validity check on the resulting `Fruit` struct (e.g. `CarrierId` must be non-empty) as a second defence.
- This is a prerequisite fix before reconnection buffering, because a buffered replay of corrupted rows makes cleanup much harder.

**Warning signs:**
- Rows in `breeze_fruit` where `carrier_id` is empty or all numeric fields are 0.
- Sudden increase in zero-value rows correlated with MQTT payload format changes from OEM firmware update.

**Phase to address:** Code cleanup phase (before reconnection work).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Atomic counter separate from channel | Avoids `len(channel)` call, theoretically more flexible | TOCTOU race between counter decrement and channel read; silent row drops under concurrent `CopyFrom` | Never — channel length is the ground truth |
| `os.Exit()` inside `run()` | Fast failure path | Bypasses `defer pool.Close()` and buffer flush; connection leak on startup error | Never in a function that holds deferred resources |
| Non-transactional `CopyFrom` | Slightly simpler call site | Partial batch commits on error; duplicate rows after retry | Never when a retry buffer is in use |
| Grademap applied immediately on MQTT receipt | Simple, zero latency | Misattributes fruit in the OEM pipeline delay window | Acceptable only if grademap changes never happen during production runs |
| Fixed queue capacity (50) with no back-pressure | Simple to implement | Blocks paho dispatcher goroutine; stalls all MQTT processing | Never in a handler that blocks the MQTT dispatcher |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| pgxpool after DB restart | Assuming the pool reconnects automatically | Pool reconnects lazily on next acquire; dead connections persist until health check cycle or acquire attempt; implement application-level reconnect probe with back-off |
| pgxpool `BeforeAcquire` | Using deprecated hook | Use `PrepareConn` in pgx/v5; `BeforeAcquire` is deprecated and may be removed |
| paho MQTT `OnMessage` | Blocking inside the handler | The paho dispatcher is single-goroutine with `OrderMatters=true`; any blocking stalls all message processing; all blocking ops must use non-blocking sends or spawn a goroutine |
| paho MQTT reconnection | Assuming `SetAutoReconnect(true)` covers subscription re-subscription | `SetAutoReconnect` reconnects the transport; `OnConnect` handler must re-subscribe to topics on each connect event (already done correctly in `connectHandler`) |
| pgx `CopyFrom` | Calling without a transaction when retries are needed | COPY is not automatically atomic; wrap in `tx.Begin/Commit/Rollback` to get all-or-nothing semantics for safe retry |
| TimescaleDB + COPY | Expecting COPY to respect hypertable chunk boundaries | COPY works transparently across chunks; no special handling needed, but very old `received_at` timestamps may land in unexpected chunks if retention policies are active |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Fixed 50-item queue with blocking `OnMessage` | MQTT client appears to disconnect during DB outage; no new fruit after queue fills | Non-blocking send in `OnMessage` with drop-and-warn; configurable buffer size | At burst rate > 50 messages before a single 30-item or 1-second batch drains (approx. 1.6 batches in flight) |
| In-memory buffer during DB outage grows unbounded | OOM kill if DB is down for hours on a high-volume line | Enforce configurable `max_buffer` limit; drop or persist to disk when limit reached | Depends on message rate; at 30 fruit/sec and 1-hour outage: 108,000 buffered items |
| Regex recompiled on every `Grademap.Update()` | Negligible per-call, but unnecessary GC pressure | Move `regexp.MustCompile` to a package-level `var` | Grademap changes are rare; this is low priority but trivially fixable |
| Dual `[]byte`+`map[string]any` representation per fruit | Memory doubles for every fruit in the queue | Remove `[]byte` fields and marshal in `AsRow()` only | At 50 items in queue with large Spectrim blobs: measurable but not critical at current scale |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Buffer replay retries indefinitely | If a row consistently triggers a DB constraint, the buffer never drains and the daemon effectively DoSes itself retrying the same bad row | Track per-batch retry count; after N failures discard the batch and log the payload for offline inspection |
| Reconnect probe with no back-off | At DB-restart rate, rapid reconnect attempts can exhaust TCP connections on the DB host | Exponential back-off with jitter; cap at 60s; give up after configurable max attempts |
| Plaintext credentials in YAML config | Anyone with filesystem read access can extract DB and MQTT credentials | Document env-var-only config for production deployments; consider removing YAML support |

---

## "Looks Done But Isn't" Checklist

- [ ] **Reconnection:** Buffer drains completely before `pool.Close()` on shutdown — verify with a test that shuts down while items are in the buffer.
- [ ] **Reconnection:** Fruit that arrived during DB outage actually appears in the DB after reconnection — not just "no errors logged."
- [ ] **Buffering:** Buffer respects the configured size limit and drops/warns correctly rather than growing unbounded.
- [ ] **CopyFrom transaction:** On rollback, the full batch is retained in buffer (not partially re-inserted) — verify with a forced mid-copy disconnect.
- [ ] **OnMessage non-blocking:** MQTT client does not stall when the write buffer is full — verify by filling buffer while sending MQTT messages and confirming grademap updates still arrive.
- [ ] **Grademap timing:** The investigation for a version field in fruit JSON is complete and documented before any timing fix is implemented — not assumed absent.
- [ ] **Shutdown order:** Final in-memory batch is written to DB on SIGINT, not discarded — verify with a test that sends SIGINT while a partial batch is queued.
- [ ] **Zero-value fruit:** `json.Unmarshal` errors are checked; zero-value rows do not appear in TimescaleDB.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Rows lost due to missing reconnect buffer | HIGH | Query OEM for replayable raw data (unlikely); partial reconstruction from MQTT broker retained messages if broker persists history; accept data gap |
| Duplicate rows from non-transactional retry | MEDIUM | Identify duplicates via `carrier_id` + `received_at` uniqueness; de-duplicate with a one-time SQL query; add unique constraint to prevent recurrence |
| Zero-value fruit rows in DB | MEDIUM | Delete rows where `carrier_id IS NULL OR carrier_id = ''`; add a DB constraint or application check to prevent future insertion |
| Grademap misattribution window | HIGH | Cannot retroactively fix without knowing which grademap was active; mark rows in the timing window as `grade_confidence = 'uncertain'`; re-grade manually if OEM can supply the effective-time metadata |
| paho dispatcher stall (goroutine blocked) | MEDIUM | Restart daemon; implement non-blocking send to prevent recurrence |
| Shutdown data loss (pool closes before flush) | MEDIUM | Restart with reconnection buffer in place; fruit in transit at shutdown time is unrecoverable without buffer |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| pgxpool dead connection not detected | Reconnection + buffering | Integration test: kill postgres, wait, restart, confirm all buffered rows appear in DB |
| Shutdown before final flush | Reconnection + buffering (+ code cleanup) | Test: send SIGINT with 15 items in queue (below 30-item threshold); verify they are written |
| Paho dispatcher blocked by full queue | Reconnection + buffering | Test: fill buffer while sending MQTT; confirm grademap updates still arrive and are processed |
| Counter/channel race in `Processor` | Code cleanup (prerequisite) | Unit test: concurrent `Next`/`Values` calls at high rate; assert no rows dropped |
| Non-transactional `CopyFrom` | Reconnection + buffering | Test: force mid-copy failure; assert zero duplicate rows after retry |
| Grademap timing misattribution | Grademap persistence + timing phase | Inspection of fruit JSON for version field (gate); then integration test with known grademap switch time |
| Silent zero-value fruit from unmarshal error | Code cleanup (prerequisite) | Unit test: send malformed JSON payload; assert no row enqueued, error logged |
| Buffer grows unbounded | Reconnection + buffering | Test: sustain DB outage beyond configured limit; assert drop-and-warn behaviour, no OOM |

---

## Sources

- [pgxpool package docs — pgx/v5](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool)
- [Recovering from connection interruption in pgxpool — pgx discussion #1415](https://github.com/jackc/pgx/discussions/1415)
- [Connection pool deadlocks v5 regression — pgx issue #1354](https://github.com/jackc/pgx/issues/1354)
- [Pool not discarding bad connections — pgx issue #494](https://github.com/jackc/pgx/issues/494)
- [Deadlocks in paho when not wrapping callbacks in goroutine — paho issue #474](https://github.com/eclipse-paho/paho.mqtt.golang/issues/474)
- [Client can hang due to blocking push — paho issue #122](https://github.com/eclipse/paho.mqtt.golang/issues/122)
- [Goroutine Leaks: The Forgotten Sender — Ardan Labs](https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html)
- [Backpressure Patterns in Go — Medium Jan 2026](https://medium.com/@Realblank/backpressure-patterns-in-go-from-channels-to-queues-to-load-shedding-0841c9fe5607)
- [pgx transactions and COPY — danp.net](https://danp.net/posts/pgx-tx-copy/)
- Codebase audit: `.planning/codebase/CONCERNS.md` (2026-03-07)

---
*Pitfalls research for: Go MQTT/TimescaleDB daemon with reconnection and buffering*
*Researched: 2026-03-07*
