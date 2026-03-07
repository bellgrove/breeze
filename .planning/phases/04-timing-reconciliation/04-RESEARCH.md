# Phase 4: Timing Reconciliation - Research

**Researched:** 2026-03-08
**Domain:** AS-OF timestamp queries, PostgreSQL grademap linkage, Go config extension
**Confidence:** HIGH

---

## CRITICAL INVESTIGATION GATE: Path A vs Path B

> The roadmap and STATE.md both flag this as a blocker. This research resolves it.

**Finding: PATH B — No grademap version/sequence field exists in the fruit payload.**

### Evidence

The fruit MQTT payload has been inspected in full via `processor/fruit_test.go` (`c2_fruit_json`, a real production payload from 2025-05-20). The complete JSON structure is:

```
{
  "carrierId": "...",
  "sizer": {
    "schema": { "name": "sizer_fruit", "version": 12 },   // sizer schema version, NOT grademap reference
    "status", "lane", "frame", "carrier", "batch", "variety",
    "product", "outlet", "size", "grade", "timestamp": "2025-05-20T05:52:43.615Z"
    // no grademap_id, grademap_version, or grademap_sequence field
  },
  "spectrim": {
    "schema": { "name": "spectrim_fruit", "version": 2 }, // spectrim schema version, NOT grademap reference
    "timestamp": "2025-05-20T05:54:07Z"
    // no grademap_id, grademap_version, or grademap_sequence field
  }
}
```

The grademap MQTT payload (`Grademap.Update()` in `processor/grademap.go`) parses this structure:

```go
var gm struct {
    Characteristics       map[string]GmCh
    Defect_grading_passes map[string]GmPa
    Grades                map[string]GmGr
    Id                    int      // OEM internal grademap ID
    Is_primary            bool
    Name                  string
    Node_id               int      // grader node ID
}
```

The grademap has an `Id` and `Node_id`, but **neither field appears in the fruit payload**. The fruit payload carries no reference back to the grademap that was active when the OEM graded it.

**Conclusion:** Path A (use a version/sequence field directly) is not possible. **Path B applies: implement a configurable `grademap_propagation_delay` offset for the AS-OF timestamp lookup.**

### What "AS-OF" means in this context

The OEM (Tomra/Spectrim) grader:
1. Receives a new grademap via MQTT (Breeze records `received_at` in `breeze_grademap`)
2. Applies the new grademap to fruit — but only to fruit that enters the grader **after** the new grademap takes effect in the OEM's pipeline
3. Fruit already in the physical conveyor belt at the time of grademap update is graded using the **previous** grademap

This means: fruit arriving in Breeze at time T was graded using the grademap that was active at time `T - propagation_delay`, not the most recent grademap received before T.

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TIME-01 | Fruit records are matched to the correct historical grademap via AS-OF timestamp query (`WHERE received_at <= processing_time ORDER BY received_at DESC LIMIT 1`) instead of always using the current grademap | AS-OF query pattern against `breeze_grademap.received_at`; fruit carries `SizerTime` timestamp; lookup uses adjusted time `SizerTime - grademap_propagation_delay` |
| TIME-02 | A configurable `grademap_propagation_delay` offset is applied to the AS-OF lookup to compensate for OEM pipeline lag; setting it to zero disables the offset with no behavioral change | `Config` struct in `breeze.go` already has YAML+envconfig pattern; add `GrademapPropagationDelay time.Duration`; zero duration means no adjustment |
</phase_requirements>

---

## Summary

Phase 4 implements timing reconciliation for fruit-to-grademap attribution. The core operation is an AS-OF lookup: when writing a fruit record, query `breeze_grademap` for the row whose `received_at` is the most recent timestamp at or before `fruit.SizerTime - grademap_propagation_delay`. The result is a `grademap_id` foreign key that links each fruit row to the exact grademap in effect when the OEM graded it.

**Path B is confirmed.** No grademap version field exists in the fruit payload. The implementation is a configurable delay offset applied to timestamp-based AS-OF lookups.

The existing infrastructure is well-suited:
- `breeze_grademap` already has `received_at TIMESTAMPTZ` with a `DESC` index
- `breeze_fruit` (`time` column = `SizerTime`) already contains the timestamp needed for the AS-OF anchor
- `breeze.go`'s `Config` struct follows a consistent YAML+envconfig pattern that simply needs a new field
- `writeBatch()` in `breeze.go` is the correct place to resolve the grademap ID before writing

**Primary recommendation:** Add `grademap_id BIGINT REFERENCES breeze_grademap(id)` to `breeze_fruit`, add `GrademapPropagationDelay time.Duration` to `Config`, perform a single AS-OF SELECT before each `CopyFrom`, and annotate each fruit row with the resolved ID.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `pgx/v5` | already in go.mod | AS-OF query via `pool.QueryRow` | Already used for all DB ops in `breeze.go` |
| `time` (stdlib) | Go stdlib | `time.Duration` for delay config | Used throughout; `time.Duration` is the idiomatic type for delay offsets |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `gopkg.in/yaml.v2` | already in go.mod | YAML config deserialization | Already used for all config fields |
| `github.com/kelseyhightower/envconfig` | already in go.mod | Env-var overrides | Already used for all config fields |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| AS-OF SELECT before CopyFrom | Subquery inside CopyFrom | Subquery approach adds complexity; SELECT-then-insert is simpler and testable |
| `time.Duration` in config | int64 seconds | `time.Duration` parses human strings ("30s") via envconfig; more ergonomic |

**Installation:** No new dependencies required.

---

## Architecture Patterns

### Existing Code Reference Points

**`breeze.go` `Config` struct** (lines 40-52) — add one field:
```go
GrademapPropagationDelay time.Duration `yaml:"grademap_propagation_delay" envconfig:"GRADEMAP_PROPAGATION_DELAY"`
```

**`schema.sql` `breeze_fruit` table** — add one column:
```sql
ALTER TABLE breeze_fruit ADD COLUMN grademap_id BIGINT REFERENCES breeze_grademap(id);
```
(nullable: fruit written before grademap history existed has no matching row)

**`processor/fruit.go` `Columns()` and `AsRow()`** — add `grademap_id` to both.

**`breeze.go` `writeBatch()`** — resolve grademap ID before CopyFrom.

### Pattern 1: AS-OF Timestamp Lookup

**What:** A single SELECT that finds the most recent grademap received at or before the adjusted fruit timestamp.

**When to use:** Once per batch, using the earliest `SizerTime` in the batch as the anchor (conservative: all fruit in a batch received in the same timer window are treated as using the same grademap).

**Alternative:** Per-fruit lookup (more accurate, more queries). For small batch sizes (default maxItems=30, maxTimeout=1s) per-fruit is feasible too. See Open Questions.

```go
// Source: existing pool.QueryRow pattern in breeze.go writeGrademap()
func resolveGrademapID(ctx context.Context, pool *pgxpool.Pool, fruitTime time.Time, delay time.Duration) (int64, error) {
    adjustedTime := fruitTime.Add(-delay)
    var id int64
    err := pool.QueryRow(ctx,
        `SELECT id FROM breeze_grademap
         WHERE received_at <= $1
         ORDER BY received_at DESC
         LIMIT 1`,
        adjustedTime,
    ).Scan(&id)
    if err != nil {
        return 0, err // pgx.ErrNoRows if no grademap received yet
    }
    return id, nil
}
```

### Pattern 2: Per-Fruit vs Per-Batch Grademap Resolution

**Per-batch** (simpler):
- Resolve once using `min(SizerTime)` across the batch
- One DB round-trip per batch
- Risk: if a grademap changes mid-batch, some fruit are misattributed
- Acceptable for batch sizes of 1 second / 30 items

**Per-fruit** (accurate):
- Resolve once per fruit using that fruit's `SizerTime`
- N DB round-trips per batch (can be reduced by caching last resolved ID)
- Correct if grademap changes arrive faster than one batch window

**Recommendation:** Per-batch with a cache check. If the resolved ID differs from the previous batch's ID, log a warning (grademap changed mid-session). This adds one query per batch, matching the existing `writeGrademap` pattern.

### Pattern 3: NULL handling for `grademap_id`

The column is nullable. Fruit written when no grademap has been received yet gets `grademap_id = NULL`. This is correct: Breeze may start receiving fruit before the OEM publishes a grademap. Callers of analytics queries must handle NULL.

### Pattern 4: `time.Duration` in YAML config

`gopkg.in/yaml.v2` does not natively parse duration strings for `time.Duration` fields. Two options:

1. **Store as `string`, parse in `run()`** — more code but human-readable ("30s")
2. **Store as `int` (seconds or milliseconds), convert** — simpler parsing, less ergonomic

Existing pattern in this codebase uses `int` for numeric fields (`QueueSize`, `WriteBufferSize`). Consistent approach: use `int` seconds, convert to `time.Duration` in `run()`. Alternatively, use `string` with `time.ParseDuration`.

**Recommendation:** Use `string` with `time.ParseDuration` — it is idiomatic Go for human-facing config and the YAML value `"30s"` is self-documenting. Validate at startup, fail fast if unparseable.

```go
// In Config struct:
GrademapPropagationDelayStr string `yaml:"grademap_propagation_delay" envconfig:"GRADEMAP_PROPAGATION_DELAY"`

// In run():
var propagationDelay time.Duration
if cfg.GrademapPropagationDelayStr != "" {
    propagationDelay, err = time.ParseDuration(cfg.GrademapPropagationDelayStr)
    if err != nil {
        return fmt.Errorf("invalid grademap_propagation_delay: %w", err)
    }
}
// zero value (unset) means no adjustment — correct by default
```

### Recommended Project Structure

No new files are strictly required. Changes are localized to:

```
breeze.go              # Config field, resolveGrademapID(), writeBatch() updated
schema.sql             # ALTER TABLE breeze_fruit ADD COLUMN grademap_id
processor/fruit.go     # Columns() and AsRow() updated
breeze_test.go         # Tests for new functions
```

### Anti-Patterns to Avoid

- **Using the current in-memory `p.grademap`** for attribution: this is the grademap currently held in memory, not the one that was active at `SizerTime`. It would be wrong whenever a grademap update occurred between when the OEM graded the fruit and when Breeze wrote it.
- **Querying `breeze_grademap` without the DESC index**: the index `breeze_grademap_received_at_idx ON breeze_grademap (received_at DESC)` was created in Phase 3. The AS-OF query uses this index efficiently.
- **Making `grademap_id` NOT NULL**: fruit can arrive before any grademap is received. A NOT NULL constraint causes write failures.
- **Failing the fruit write if grademap lookup fails**: grademap attribution is an enhancement, not a correctness gate. If the lookup fails (DB error or no grademap rows), write the fruit with `grademap_id = NULL` and log a warning.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| AS-OF lookup | Custom in-memory grademap history | SQL `WHERE received_at <= $1 ORDER BY received_at DESC LIMIT 1` | `breeze_grademap` already stores the full history; SQL is the correct tool |
| Duration config parsing | Custom YAML type | `time.ParseDuration(string)` | stdlib handles "30s", "1m30s", "0s" correctly |
| Grademap version tracking | Sequence counter, sync.Mutex cache | SQL query result cached in a local variable | One SELECT per batch is negligible; no concurrency complications |

---

## Common Pitfalls

### Pitfall 1: `SizerTime` vs `time.Now()` as the anchor

**What goes wrong:** Using `time.Now()` (Breeze receive time) instead of `fruit.SizerTime` (OEM sizer timestamp) as the anchor for the AS-OF lookup.

**Why it happens:** `time.Now()` seems natural — "what was the grademap when we received this fruit?" But the OEM graded the fruit earlier; the relevant time is when the OEM processed it, not when Breeze received it over MQTT.

**How to avoid:** Always use `fruit.SizerTime` as the base time for the AS-OF lookup. The `propagation_delay` is subtracted from this time.

**Warning signs:** Fruit consistently attributed to a newer grademap than expected.

### Pitfall 2: `pgx.ErrNoRows` treated as a fatal error

**What goes wrong:** `pool.QueryRow().Scan()` returns `pgx.ErrNoRows` when `breeze_grademap` is empty (no grademap received yet). If this propagates to `writeBatch()` as a hard failure, all fruit writes fail until a grademap exists.

**Why it happens:** `pgx.ErrNoRows` looks like a DB error but is a valid state.

**How to avoid:** Check `errors.Is(err, pgx.ErrNoRows)` and treat it as "no grademap yet — use NULL".

**Warning signs:** All fruit writes failing at startup before the first grademap arrives.

### Pitfall 3: `time.Duration` zero value is correct "no delay"

**What goes wrong:** Treating zero duration as "not configured" and applying a fallback.

**Why it happens:** Checking `if cfg.Delay == 0` and setting a default.

**How to avoid:** Zero delay must be explicitly valid — `SizerTime.Add(0)` == `SizerTime`. Document in config YAML. No default fallback needed.

### Pitfall 4: Schema migration on existing `breeze_fruit` table

**What goes wrong:** Adding `grademap_id BIGINT NOT NULL` to an existing table with rows causes migration failure (existing rows violate NOT NULL).

**How to avoid:** Column must be nullable. If schema migration is needed for an existing deployment, `ALTER TABLE ... ADD COLUMN grademap_id BIGINT` (no NOT NULL, no DEFAULT) is safe.

### Pitfall 5: Grademap lookup inside CopyFrom

**What goes wrong:** Attempting a DB query inside the `CopyFrom` loop or callback — pgx does not support this.

**How to avoid:** Resolve the `grademap_id` before calling `CopyFrom`, then include it in each row's `[]any` slice.

---

## Code Examples

### AS-OF Query (verified pattern from existing codebase + pgx v5 docs)

```go
// Source: pool.QueryRow pattern from breeze.go writeGrademap()
func resolveGrademapID(ctx context.Context, pool *pgxpool.Pool, fruitTime time.Time, delay time.Duration) (int64, bool, error) {
    adjustedTime := fruitTime.Add(-delay)
    var id int64
    err := pool.QueryRow(ctx,
        `SELECT id FROM breeze_grademap
         WHERE received_at <= $1
         ORDER BY received_at DESC
         LIMIT 1`,
        adjustedTime,
    ).Scan(&id)
    if errors.Is(err, pgx.ErrNoRows) {
        return 0, false, nil // no grademap yet — caller uses NULL
    }
    if err != nil {
        return 0, false, fmt.Errorf("resolve grademap id: %w", err)
    }
    return id, true, nil
}
```

### Updated writeBatch signature (conceptual)

```go
func writeBatch(ctx context.Context, pool *pgxpool.Pool, batch []processor.Fruit, delay time.Duration) error {
    // Resolve grademap for the batch using the earliest SizerTime
    // (one query per batch, not per fruit)
    var grademapID *int64
    if len(batch) > 0 {
        anchor := batch[0].SizerTime
        for _, f := range batch[1:] {
            if f.SizerTime.Before(anchor) {
                anchor = f.SizerTime
            }
        }
        id, ok, err := resolveGrademapID(ctx, pool, anchor, delay)
        if err != nil {
            slog.Warn("Failed to resolve grademap ID — writing fruit with NULL", "err", err)
        } else if ok {
            grademapID = &id
        }
    }

    rows := make([][]any, len(batch))
    for i := range batch {
        row, err := batch[i].AsRow()
        if err != nil {
            return fmt.Errorf("failed to convert fruit to row: %w", err)
        }
        // grademap_id is the last element (appended by AsRow or here)
        rows[i] = append(row, grademapID) // nil pointer → SQL NULL
        rows[i] = row
    }
    // CopyFrom unchanged
}
```

### `processor/fruit.go` — adding `grademap_id`

`Columns()` adds `"grademap_id"` as the last column.
`AsRow()` appends `a.GrademapID` (type `*int64`, nil = NULL).

The `Fruit` struct gains one field:
```go
GrademapID *int64  // nil when no grademap row matched
```

### Schema migration

```sql
-- Safe for existing deployments (nullable, no DEFAULT)
ALTER TABLE breeze_fruit ADD COLUMN grademap_id BIGINT REFERENCES breeze_grademap(id);
-- Optional index for analytics joins
CREATE INDEX breeze_fruit_grademap_id_idx ON breeze_fruit (grademap_id);
```

---

## State of the Art

| Old Approach | Current Approach | Notes |
|--------------|------------------|-------|
| Always use current in-memory grademap | AS-OF SQL lookup at write time | Phase 3 built the foundation; Phase 4 uses it |
| Grademap not stored | `breeze_grademap` with `received_at` index | Completed Phase 3 |

---

## Open Questions

1. **Per-fruit vs per-batch grademap resolution**
   - What we know: batches are at most 30 items over 1 second (processor.go `maxItems=30`, `maxTimeout=1s`). Grademap changes in production are infrequent (daily or per-batch changes, not sub-second).
   - What's unclear: how often does a grademap change arrive mid-batch in live traffic?
   - Recommendation: implement per-batch (simpler, one query). If per-fruit accuracy is needed, it is a straightforward extension — same query, called N times with caching.

2. **`time.Duration` YAML config representation**
   - What we know: `gopkg.in/yaml.v2` does not natively unmarshal `time.Duration`; `envconfig` supports it via `time.ParseDuration`.
   - What's unclear: whether the operator prefers seconds (integer) or duration strings ("30s").
   - Recommendation: use `string` field, parse with `time.ParseDuration` in `run()` for human ergonomics. Document `"0s"` as the zero-delay value.

3. **Whether `breeze_fruit` already has rows in production**
   - What we know: schema includes `breeze_fruit` with no `grademap_id`.
   - What's unclear: whether a migration script or `IF NOT EXISTS` pattern is needed in `schema.sql`.
   - Recommendation: DDL plan should ADD COLUMN (not recreate table), and planner should note this as an operator migration step.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none — `go test ./...` |
| Quick run command | `go test ./... -run TestResolveGrademap -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TIME-01 | AS-OF lookup returns correct grademap ID for a given fruit time | unit | `go test . -run TestResolveGrademapID -v` | Wave 0 |
| TIME-01 | Zero-delay case returns same result as exact timestamp match | unit | `go test . -run TestResolveGrademapID_ZeroDelay -v` | Wave 0 |
| TIME-01 | No grademap rows → returns false/nil, no error | unit | `go test . -run TestResolveGrademapID_NoRows -v` | Wave 0 |
| TIME-01 | `writeBatch` passes `grademap_id` as last column value | unit | `go test . -run TestWriteBatch_GrademapID -v` | Wave 0 |
| TIME-02 | Non-zero delay shifts anchor time correctly | unit | `go test . -run TestResolveGrademapID_WithDelay -v` | Wave 0 |
| TIME-02 | Config parses `"30s"` propagation delay without error | unit | `go test . -run TestConfig_PropagationDelay -v` | Wave 0 |
| TIME-02 | Zero delay ("0s" or unset) produces no offset | unit | `go test . -run TestConfig_PropagationDelayZero -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `breeze_test.go` — add `TestResolveGrademapID`, `TestResolveGrademapID_ZeroDelay`, `TestResolveGrademapID_NoRows`, `TestWriteBatch_GrademapID`, `TestResolveGrademapID_WithDelay`, `TestConfig_PropagationDelay`, `TestConfig_PropagationDelayZero`

Note: `resolveGrademapID` requires a real DB to test meaningfully. Unit tests should use an invalid pool (same pattern as `TestTransactionalWrite`) to verify error handling, and integration-style tests (if desired) would require `testcontainers` or a test DB — which is out of scope for v1. Unit tests should cover: the `pgx.ErrNoRows` handling branch, correct argument passing, and zero-delay behavior. DB interaction tests are manual-only or skipped.

---

## Sources

### Primary (HIGH confidence)
- Codebase inspection — `processor/fruit.go`, `processor/grademap.go`, `processor/processor.go`, `breeze.go`, `schema.sql`
- `processor/fruit_test.go` — contains a complete real production fruit JSON payload (2025-05-20)
- `processor/grademap_test.go` — confirms grademap JSON structure via `Grademap.Update()` struct
- `breeze.go` `writeGrademap()` — existing `pool.QueryRow` pattern used verbatim as AS-OF template
- `schema.sql` — confirms `breeze_grademap_received_at_idx ON breeze_grademap (received_at DESC)` exists

### Secondary (MEDIUM confidence)
- pgx v5 documentation — `pgx.ErrNoRows` sentinel, `pool.QueryRow().Scan()` behavior

---

## Metadata

**Confidence breakdown:**
- Path A/B determination: HIGH — based on direct inspection of real production payload in test file; no grademap version field present
- Standard stack: HIGH — all libraries already in use; no new dependencies
- Architecture (AS-OF query): HIGH — existing `pool.QueryRow` pattern is identical; index already exists
- `time.Duration` YAML config: MEDIUM — `yaml.v2` behavior with Duration confirmed by known limitation; string workaround is established Go idiom
- Per-batch vs per-fruit recommendation: MEDIUM — based on observed batch parameters; production grademap change frequency not directly measured

**Research date:** 2026-03-08
**Valid until:** 2026-06-08 (stable domain — pgx v5 and grademap schema are not changing)
