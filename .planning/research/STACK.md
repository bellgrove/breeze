# Stack Research

**Domain:** Go daemon — PostgreSQL reconnection buffering + TimescaleDB versioned records
**Researched:** 2026-03-07
**Confidence:** HIGH (pgx/v5 from official pkg.go.dev + GitHub source; TimescaleDB from official docs; Go buffering from standard library + well-documented patterns)

---

## Overview

This is a focused addition to an existing stack. The core technologies (pgx/v5, TimescaleDB, Go) are fixed. This document covers only the three additions needed:

1. pgxpool health checking and reconnection detection
2. In-memory write buffering during DB outages
3. TimescaleDB schema for versioned grademap history

---

## 1. pgxpool Reconnection and Health Checking

### What pgxpool Actually Does

pgxpool is designed to be created once at startup and live for the program's lifetime. It manages a pool of individual `*pgx.Conn` connections and handles their creation, destruction, and health checking automatically.

**Key behavior when the DB is down:**

- `pool.CopyFrom()` (and any other pool operation) acquires a connection before executing. If all existing connections are dead and new connections cannot be established, the call **blocks** until either the context is cancelled or a connection becomes available.
- pgxpool does NOT return a "DB is down" sentinel immediately — it will keep trying to acquire/establish a connection until `ctx.Done()` fires.
- The pool does **not** need to be recreated after a DB outage. It will automatically create new connections when the DB returns.

**Health check mechanics (HIGH confidence — from pkg.go.dev):**

| Mechanism | Default | What it does |
|-----------|---------|--------------|
| `HealthCheckPeriod` | 1 minute | Background goroutine pings idle connections; closes dead ones |
| `ShouldPing` callback | Ping if idle >= 1 second | Called after acquiring a connection; pings stale idle connections before use |
| `PrepareConn` callback | nil | Custom validation per acquire; returning `(false, nil)` destroys the connection and retries on a fresh one |
| `MaxConnIdleTime` | 30 minutes | Idle connections beyond this are closed |
| `MaxConnLifetime` | 1 hour | Connections beyond this age are closed and replaced |

**The key insight for reconnection detection:** The pool does not expose a "currently connected" boolean. The only reliable signals are:

- A call to `pool.Ping(ctx)` returns an error — DB is unreachable right now
- A call to `pool.CopyFrom()` (or any operation) returns an error — that operation failed

### Recommended API Usage

**Detecting DB is down:**

```go
// Use pool.Ping with a short timeout to probe without blocking the write path
pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()
if err := pool.Ping(pingCtx); err != nil {
    // DB is unreachable — activate buffer mode
}
```

**Detecting connection failure on a write:**

```go
rows, err := pool.CopyFrom(ctx, ...)
if err != nil {
    // Check if this is a connection-layer failure vs a PostgreSQL error
    var pgErr *pgconn.PgError
    var connectErr *pgconn.ConnectError
    switch {
    case errors.As(err, &connectErr):
        // Network-level failure: DB unreachable, safe to buffer and retry
    case errors.As(err, &pgErr):
        // PostgreSQL returned an error (constraint, syntax, etc.)
        // Check pgErr.Code — may or may not be retryable
    case pgconn.SafeToRetry(err):
        // Error occurred before data was sent — safe to retry
    }
}
```

**Import path for error types:**

```go
import "github.com/jackc/pgx/v5/pgconn"
```

Not `github.com/jackc/pgconn` — that is the standalone v4-era module. In pgx/v5, pgconn is embedded in the pgx module tree.

**`pool.Reset()` for forcing reconnection:**

```go
pool.Reset() // Closes all idle connections; acquired ones are closed when returned
```

Useful when resuming from a known outage — forces all connections to be re-established rather than reusing potentially stale ones.

### What NOT to Use

| Avoid | Why |
|-------|-----|
| `pool.Stat().TotalConns() == 0` to detect dead pool | Does not reliably indicate connectivity loss |
| `pool.Stat().IdleConns() == 0` | Same — pool may have zero idle conns but still be healthy |
| Recreating the pool after outage | Not needed; pool auto-recovers. Adds complexity and races |
| Context-free pool operations during outage | Without a timeout context, `pool.CopyFrom` will block indefinitely when DB is unreachable |

**Always pass a context with a deadline to pool operations when operating in uncertain connectivity.** A context.Background() on a pool.CopyFrom() call will block forever if the DB is down.

---

## 2. In-Memory Write Buffering

### Problem

The current `queue` channel (capacity 50) in `Processor` is consumed by `pool.CopyFrom()` via the `pgx.CopyFromSource` interface (`Next()` / `Values()`). When the DB is down, `CopyFrom` fails, but the channel has already been partially or fully drained — fruit records are lost.

The fix requires decoupling ingestion from DB writing: buffer fruit records independently so they survive a failed write attempt and can be replayed.

### Recommended Approach: Bounded Slice with Mutex

**Use a `[]Fruit` slice protected by a `sync.Mutex`, not a channel.** This is the correct choice for this use case because:

1. The buffer must be **readable without consuming** — to pass records to `CopyFrom`, we need to know the full batch upfront before the copy starts, so we can replay on failure
2. The buffer must support **re-enqueue on failure** — if `CopyFrom` fails mid-batch, items must go back to the front
3. The buffer needs a **configurable hard cap** with drop-on-overflow or back-pressure behaviour
4. Channel semantics (`<-ch`) consume on read — once drained into a `CopyFromSource`, records are gone if the write fails

```go
type WriteBuffer struct {
    mu       sync.Mutex
    items    []Fruit
    capacity int
    dropped  int64 // atomic counter for monitoring
}

func (b *WriteBuffer) Enqueue(f Fruit) bool {
    b.mu.Lock()
    defer b.mu.Unlock()
    if len(b.items) >= b.capacity {
        // Drop oldest (index 0) to make room — preserves most recent data
        b.items = b.items[1:]
        atomic.AddInt64(&b.dropped, 1)
    }
    b.items = append(b.items, f)
    return true
}

func (b *WriteBuffer) Drain(n int) []Fruit {
    b.mu.Lock()
    defer b.mu.Unlock()
    if n > len(b.items) {
        n = len(b.items)
    }
    batch := make([]Fruit, n)
    copy(batch, b.items[:n])
    b.items = b.items[n:] // consume only on success
    return batch
}

func (b *WriteBuffer) Prepend(items []Fruit) {
    // Called on write failure to re-queue items at front
    b.mu.Lock()
    defer b.mu.Unlock()
    b.items = append(items, b.items...)
    // Trim to capacity if prepend caused overflow
    if len(b.items) > b.capacity {
        b.items = b.items[len(b.items)-b.capacity:]
    }
}

func (b *WriteBuffer) Len() int {
    b.mu.Lock()
    defer b.mu.Unlock()
    return len(b.items)
}
```

### Why Not a Channel

| Approach | Problem |
|----------|---------|
| `chan Fruit` (current) | `<-ch` is destructive — once read into CopyFromSource, records are gone on write failure |
| `chan []Fruit` (batch channel) | Batches can still be lost on write failure unless you re-enqueue |
| Ring buffer library | Adds a dependency; the pattern above is straightforward stdlib |

### Why Not `gammazero/deque`

`gammazero/deque` (already in go.mod) is not thread-safe and has no capacity bounding. It would require external locking and manual capacity enforcement — equivalent complexity to the slice approach, with an extra dependency.

### Capacity Sizing

| Scenario | Reasoning | Recommended default |
|----------|-----------|---------------------|
| Normal operation | DB writes every 1–30 items, latency <100ms | Buffer at 0 (no backlog) |
| Short outage (1–2 min) | ~30 items/sec × 120 sec = 3,600 fruits | 5,000 items |
| Extended outage | Memory: 5,000 × ~500 bytes/Fruit ≈ 2.5 MB | Still safe |
| Hard cap | Prevents unbounded growth on multi-hour outage | Configurable; default 5,000 |

Add a `BufferSize` field to the existing `Config` struct with a default of 5,000. Log a `slog.Warn` each time a record is dropped due to overflow.

### Reconnection Loop Pattern

The write path in `breeze.go:run()` needs a reconnection loop that:

1. On write failure, logs the error and activates "buffer mode"
2. In buffer mode, periodically calls `pool.Ping(ctx)` with a short timeout (e.g., 5-second interval)
3. On successful ping, calls `pool.Reset()` then attempts to flush the buffer in batches
4. On flush success, exits buffer mode
5. On flush failure, re-prepends the batch to the buffer and continues pinging

This loop runs within the existing `select` event loop — no new goroutines needed.

---

## 3. TimescaleDB Schema for Grademap Versioning

### Goal

Store every grademap received from MQTT with a timestamp, enabling:
- Audit log of all grademap changes
- Temporal queries: "which grademap was active when this fruit was processed?"
- Diff queries: "what changed between version N and version N+1?"

### Recommended Schema

**Option A: Hypertable (recommended)**

```sql
CREATE TABLE breeze_grademap (
    received_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name          TEXT        NOT NULL,
    payload       JSONB       NOT NULL
);

SELECT create_hypertable('breeze_grademap', by_range('received_at'));

-- Index for lookup by name (e.g., "what was the active map at time T?")
CREATE INDEX ON breeze_grademap (name, received_at DESC);
```

**Query: active grademap at a given fruit timestamp:**

```sql
SELECT payload
FROM breeze_grademap
WHERE name = $1
  AND received_at <= $2   -- fruit's processing_time
ORDER BY received_at DESC
LIMIT 1;
```

This is an AS-OF query — "the most recent grademap version that existed at or before this fruit's timestamp."

**Option B: Plain table (simpler, still correct)**

```sql
CREATE TABLE breeze_grademap (
    id          BIGSERIAL   PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name        TEXT        NOT NULL,
    payload     JSONB       NOT NULL
);

CREATE INDEX ON breeze_grademap (name, received_at DESC);
```

Use Option B if grademap changes are rare (tens per day at most). TimescaleDB hypertable overhead is not worth it for low-volume configuration tables. Use Option A if you want compression, retention policies, or continuous aggregates over grademap history.

### Why JSONB for payload

The grademap JSON structure is complex and vendor-defined. Storing the raw payload as JSONB:
- Preserves the full original document for audit purposes
- Allows `->` / `->>` JSON path queries if specific fields need indexing later
- Requires no schema migration if the OEM changes the grademap format
- Supports `jsonb_diff` style comparisons between versions

### What NOT to Do

| Avoid | Why |
|-------|-----|
| Storing only the current grademap in a single-row table | Loses history; cannot answer "what map was active at T?" |
| Normalising grademap fields into columns | OEM format may change; normalization locks schema to current format |
| Using `UPDATE ... WHERE name = $1` | Overwrites history — use `INSERT` only, never update |
| Partitioning grademap table on `name` | Low cardinality (one grademap name); not useful |

### Timing Reconciliation (the bug)

From PROJECT.md: there is a known bug where the OEM applies a new grademap after a pipeline delay. The schema above supports the reconciliation fix either way:

- **If fruit JSON contains a `grademap_version` field:** Query `breeze_grademap` by that field added as an indexed column or JSONB path
- **If no version field exists:** Use `received_at <= fruit.processing_time` AS-OF query with a configurable backoff duration (e.g., subtract 5 seconds from `fruit.processing_time` to account for pipeline delay)

The schema does not need to change between these two approaches — only the query changes.

---

## Recommended Core Technologies (Additions)

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| `pgxpool` (existing) | v5.7.5 | Connection pool | No new dependency; use `Ping`, `Reset`, error type checks |
| `pgconn` (existing, sub-pkg) | v5.7.5 | Error type discrimination | `ConnectError` vs `PgError` vs `SafeToRetry` |
| `sync.Mutex` + `[]Fruit` | stdlib | Write buffer | No extra dependency; replay-safe; bounded capacity |
| TimescaleDB hypertable | existing | Grademap history | Enables AS-OF temporal queries; compression if needed |

---

## Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `gammazero/deque` | v1.0.0 | Already in go.mod; NOT recommended | Do not use — not thread-safe, unbounded; prefer slice+mutex |
| `context.WithTimeout` | stdlib | Bound pool operations | Always wrap pool.CopyFrom / pool.Ping when DB may be unreachable |

---

## Installation

No new dependencies required. All needed APIs are in the existing `github.com/jackc/pgx/v5` module.

```bash
# Verify current version
go list -m github.com/jackc/pgx/v5
# Should show: github.com/jackc/pgx/v5 v5.7.5

# Import paths to add in new code:
# "github.com/jackc/pgx/v5/pgconn"   — error type discrimination
# "github.com/jackc/pgx/v5/pgxpool"  — already imported in breeze.go
```

---

## Alternatives Considered

| Recommended | Alternative | When Alternative Makes Sense |
|-------------|-------------|------------------------------|
| `sync.Mutex` + `[]Fruit` slice | `chan Fruit` (current) | Only when read is always destructive (current non-replay design) |
| `sync.Mutex` + `[]Fruit` slice | `gammazero/deque` | If you needed O(1) front-prepend at extremely high throughput (not the case here) |
| Plain `INSERT` into `breeze_grademap` | `UPDATE` upsert | Never — would destroy history |
| AS-OF query for reconciliation | In-memory map[version] | Only if grademap version field exists in fruit JSON and is exact |
| `pool.Ping` for connectivity probe | `pool.Stat()` checks | `Stat()` does not test live connectivity |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Recreating `pgxpool.Pool` after outage | Pool auto-recovers; recreation adds race conditions | `pool.Reset()` + retry |
| `context.Background()` on pool ops during uncertain state | Blocks forever if DB unreachable | `context.WithTimeout(ctx, N*time.Second)` |
| `gammazero/deque` as write buffer | Not thread-safe, not bounded | `sync.Mutex` + `[]Fruit` |
| Importing `github.com/jackc/pgconn` (standalone) | That is the v4-era module | `github.com/jackc/pgx/v5/pgconn` |
| `pool.Stat()` to detect connectivity | Stat reflects pool metadata, not live DB reachability | `pool.Ping(ctx)` |
| `UPDATE` on grademap table | Destroys audit trail | `INSERT`-only; query with `ORDER BY received_at DESC LIMIT 1` |

---

## Version Compatibility

| Package | Version | Notes |
|---------|---------|-------|
| `github.com/jackc/pgx/v5` | v5.7.5 | pgconn is sub-package, not standalone — import path `pgx/v5/pgconn` |
| `github.com/jackc/puddle/v2` | v2.2.2 | Backing pool implementation — do not import directly |
| TimescaleDB | ≥ 2.0 | `create_hypertable` with `by_range()` syntax requires ≥ 2.13 |

---

## Sources

- [pgxpool pkg.go.dev — Pool, Config, Ping, Reset, Stat, ShouldPing, PrepareConn](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) — HIGH confidence
- [pgconn pkg.go.dev — ConnectError, PgError, SafeToRetry, Timeout](https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn) — HIGH confidence
- [pgx Discussion #1415 — Recovering from connection interruption](https://github.com/jackc/pgx/discussions/1415) — MEDIUM confidence (community discussion, confirmed by pool.go source)
- [pgx Issue #891 — Detecting dead pgxpool](https://github.com/jackc/pgx/issues/891) — MEDIUM confidence (maintainer confirmed pool auto-recovery behaviour)
- [TimescaleDB metadata table best practices](https://www.tigerdata.com/learn/best-practices-for-time-series-metadata-tables) — MEDIUM confidence (official Timescale blog)
- [gammazero/deque pkg.go.dev — thread-safety and capacity behaviour](https://pkg.go.dev/github.com/gammazero/deque) — HIGH confidence

---

*Stack research for: Breeze — reconnection buffering + grademap versioning additions*
*Researched: 2026-03-07*
