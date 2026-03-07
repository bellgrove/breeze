# Phase 3: Grademap Persistence - Research

**Researched:** 2026-03-08
**Domain:** Go channels, PostgreSQL JSONB diff, pgx v5 single-row inserts
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| GRAD-01 | `breeze_grademap` table exists (append-only: `id`, `received_at TIMESTAMPTZ`, `payload JSONB`) | Plain `pgxpool.Pool.Exec` with a parameterised INSERT is sufficient; `RETURNING id` retrieves the row ID for the changes FK |
| GRAD-02 | A record is inserted on every MQTT grademap update (sent via channel to `run()`, non-blocking in `OnMessage`) | Follows the exact same channel pattern already used for fruit: a `chan []byte` with a buffered, non-blocking send; `run()` processes it in the select loop |
| GRAD-03 | `breeze_grademap_changes` table exists with queryable change rows — one row per changed field: `(grademap_id FK, entity_type, entity_name, field, old_value TEXT, new_value TEXT)` | Diff is computed in `run()` by comparing the new raw JSON map against a stored previous snapshot; `pgx.CopyFrom` (or a batched insert loop) writes the change rows |
</phase_requirements>

---

## Summary

Phase 3 adds grademap persistence alongside the existing fruit pipeline. Every raw MQTT grademap payload is stored verbatim as JSONB in `breeze_grademap` with a `received_at` timestamp. For each update, a field-level diff against the previous payload is computed and each changed field is written as a row in `breeze_grademap_changes`. Neither operation may block `OnMessage`.

The non-blocking requirement is met by the same channel pattern already established for fruit: `OnMessage` does a non-blocking send on a `chan []byte` carrying the raw grademap payload. The `run()` select loop receives it, inserts the grademap row, computes the diff, and writes the change rows — all inside the existing single-goroutine event loop, using the pool directly. No new goroutines or concurrency primitives are required.

The diff computation works by unmarshalling both the previous and new payloads into `map[string]any` and walking the key space. Nested objects (characteristics, grades, passes) each carry an entity type, entity name, and field name — these map directly to the `entity_type`, `entity_name`, `field` columns in `breeze_grademap_changes`. The diff only needs to compare scalar leaf values; nested maps are compared recursively to one level of depth (the structure of the grademap JSON is well-understood from the existing `grademap.go` parsing code).

**Primary recommendation:** Add a buffered `chan []byte` to `Processor` for grademap payloads, mirror the fruit channel pattern for the non-blocking send, and process grademap inserts + diff writes directly in `run()` using `pgxpool.Pool.Exec` for the single grademap row and `pgx.CopyFromRows` for the bulk change rows.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/jackc/pgx/v5` | v5.7.5 (already in go.mod) | `pgxpool.Pool.Exec` for single-row INSERT RETURNING, `pgx.CopyFromRows` for change rows | Already imported; provides all needed write primitives |
| `github.com/jackc/pgx/v5/pgxpool` | same | Pool-level `Exec`, `BeginTxFunc` | Already in use in `run()` |
| stdlib `encoding/json` | — | Unmarshal raw payload to `map[string]any` for diff | Already imported; no external dependency |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stdlib `maps` | — | Iterate over map keys for diff | Already used in `grademap.go`; use for key enumeration in diff |
| stdlib `fmt` | — | Convert leaf values to TEXT for `old_value`/`new_value` columns | Already imported |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib `encoding/json` diff | `wI2L/jsondiff` library | Library is cleaner for complex JSON but adds a dependency; the grademap JSON structure is known and shallow enough that a manual map walk is ~30 lines |
| `pgx.CopyFromRows` for change rows | Individual `Exec` per change row | `CopyFrom` is faster for many rows but either works; for grademap changes (typically < 100 rows per update) `CopyFrom` is fine and consistent with the fruit pattern |
| Separate goroutine for DB writes | Inline in `run()` select loop | Separate goroutine adds concurrency risk; inline is consistent with the existing write pattern and keeps DB access single-goroutine |

**Installation:** No new packages. All required packages already in `go.mod`.

---

## Architecture Patterns

### How Phase 3 Integrates with the Existing Event Loop

The existing `run()` select loop already handles two concerns: `<-next_batch` (fruit writes) and `<-reconnectCh` (DB reconnect). Phase 3 adds a third case: `<-gradeCh` carrying a raw grademap payload as `[]byte`.

```
run() select loop
├── case <-ctx.Done()          — existing: graceful shutdown
├── case <-next_batch          — existing: fruit write / buffer
├── case <-reconnectCh         — existing: flush buffer on DB recovery
└── case payload := <-gradeCh  — NEW: insert grademap row, compute diff, write change rows
```

`gradeCh` is a `chan []byte` with buffer size 1 (consistent with the `timer` channel in `Processor`). The send in `OnMessage` is non-blocking (select/default).

### Recommended Project Structure

No new files are strictly required. Changes touch:

```
breeze.go                    — add gradeCh to run(); add writeGrademap() and diffGrademap() helpers
processor/processor.go       — add gradeCh chan []byte field; add non-blocking send in OnMessage for grademap topic
schema.sql                   — add CREATE TABLE for breeze_grademap and breeze_grademap_changes
```

Optionally, if the diff logic grows complex:

```
processor/grademap_diff.go   — package-level diffGrademap function (pure, testable)
```

The diff function is pure (no I/O) and the most complex new piece of logic, so putting it in a dedicated file inside `processor` (where `grademap.go` lives) and testing it in isolation is the recommended approach.

### Pattern 1: Non-Blocking Grademap Channel in OnMessage

**What:** Mirror the existing fruit-queue non-blocking send for grademap payloads.
**When to use:** Whenever `OnMessage` receives a message on the grademap topic.

```go
// Source: existing processor/processor.go pattern (select/default on p.queue)
// In OnMessage, grademap branch:
select {
case p.gradeCh <- msg.Payload():
    // payload sent; run() will persist it
default:
    slog.Warn("Dropped grademap update — channel full")
}
```

`p.gradeCh` must be buffered (size >= 1) so that `OnMessage` never blocks. A buffer of 1 is sufficient because `run()` processes each payload synchronously before the next `<-next_batch` can arrive.

**Important:** `msg.Payload()` returns a byte slice owned by the MQTT library that may be reused after `OnMessage` returns. Copy the payload before sending:

```go
payload := make([]byte, len(msg.Payload()))
copy(payload, msg.Payload())
select {
case p.gradeCh <- payload:
default:
    slog.Warn("Dropped grademap update — channel full")
}
```

This is a subtle but critical detail. The existing fruit path avoids this because `json.Unmarshal` creates new values; for the raw `[]byte` path we must copy.

### Pattern 2: Single-Row INSERT RETURNING for breeze_grademap

**What:** Insert the raw payload and receive the generated `id` back in one round-trip.
**When to use:** In `run()`, on every received grademap payload.

```go
// Source: pgx v5 docs — QueryRow with RETURNING
var grademapID int64
err := pool.QueryRow(ctx,
    `INSERT INTO breeze_grademap (received_at, payload) VALUES ($1, $2) RETURNING id`,
    time.Now(),
    payload, // raw []byte; pgx sends as JSONB if the column type is JSONB
).Scan(&grademapID)
if err != nil {
    slog.Error("Failed to insert grademap", "err", err)
    // Non-fatal: continue; loss of this grademap is acceptable vs blocking MQTT
}
```

pgx v5 accepts `[]byte` for a `JSONB` column directly — no explicit cast needed.

### Pattern 3: Field-Level Diff for breeze_grademap_changes

**What:** Compare the previous and new grademap payloads as `map[string]any` to produce a list of changed fields.
**When to use:** After every successful grademap insert.

The grademap JSON has a known structure (from `grademap.go`): top-level keys are `Characteristics`, `Defect_grading_passes`, `Grades`, `Id`, `Is_primary`, `Name`, `Node_id`. The nested objects contain the fields relevant to audit queries (e.g., threshold changes inside `Grades`).

The diff function signature:

```go
// In processor package (processor/grademap_diff.go)
type GrademapChange struct {
    EntityType string
    EntityName string
    Field      string
    OldValue   string
    NewValue   string
}

// DiffGrademaps returns one GrademapChange per changed leaf field.
// prev may be nil on the first update (all fields are treated as new).
func DiffGrademaps(prev, next map[string]any) []GrademapChange
```

The diff walks the map recursively. For top-level scalars (`Name`, `Id`, etc.), `entity_type = "grademap"`, `entity_name = ""` (or the grademap name). For nested objects like `Characteristics["Russet"]`, `entity_type = "characteristic"`, `entity_name = "Russet"`, `field = changed_key`. For `Grades["1. Class A"]`, `entity_type = "grade"`, `entity_name = "Class A"`.

Stringify leaf values with `fmt.Sprintf("%v", v)` for TEXT storage.

### Pattern 4: Bulk Insert of Change Rows via pgx.CopyFrom

**What:** Write all change rows for a single grademap update in one CopyFrom call.
**When to use:** After computing the diff, if len(changes) > 0.

```go
// Source: existing breeze.go CopyFromRows pattern
changeRows := make([][]any, len(changes))
for i, c := range changes {
    changeRows[i] = []any{grademapID, c.EntityType, c.EntityName, c.Field, c.OldValue, c.NewValue}
}
_, err = pool.CopyFrom(ctx,
    pgx.Identifier{"breeze_grademap_changes"},
    []string{"grademap_id", "entity_type", "entity_name", "field", "old_value", "new_value"},
    pgx.CopyFromRows(changeRows),
)
if err != nil {
    slog.Error("Failed to write grademap changes", "err", err)
}
```

Note: `pool.CopyFrom` (not `tx.CopyFrom`) is used here because the grademap and changes writes are separate operations. If atomicity between the two inserts is required (either both succeed or neither), wrap both in `pgx.BeginTxFunc`. Given this is an audit log (not a transactional record), partial failure (grademap row written but changes not) is acceptable — log and continue.

### Pattern 5: Previous Payload Tracking in run()

**What:** `run()` must keep the previous raw JSON payload to compute diffs.
**When to use:** Maintained as local state in `run()`, updated after each successful grademap write.

```go
// In run(), alongside existing write buffer state:
var prevGrademapPayload map[string]any  // nil on first update
```

After each successful insert:

```go
var current map[string]any
if err := json.Unmarshal(payload, &current); err == nil {
    changes := processor.DiffGrademaps(prevGrademapPayload, current)
    // ... write changes ...
    prevGrademapPayload = current
}
```

This is consistent with the existing pattern of keeping state as local variables in `run()`.

### Pattern 6: Processor struct change — adding gradeCh

The `Processor` struct currently holds:

```go
type Processor struct {
    timer    chan bool
    queue    chan Fruit
    counter  atomic.Int32
    grademap Grademap
}
```

Add `gradeCh chan []byte`:

```go
type Processor struct {
    timer    chan bool
    queue    chan Fruit
    counter  atomic.Int32
    grademap Grademap
    gradeCh  chan []byte
}
```

`Create()` initialises it: `make(chan []byte, 1)`.

`run()` accesses it via a new exported method or by making `gradeCh` accessible. Since `run()` is in `package main` and `Processor` is in `package processor`, an exported getter is needed:

```go
// In processor/processor.go
func (p *Processor) GradeCh() <-chan []byte {
    return p.gradeCh
}
```

This follows the existing access pattern where `run()` calls `processor.DrainFruits(&proc)` rather than accessing fields directly.

### Anti-Patterns to Avoid

- **Calling `p.grademap.Update()` and then sending on `gradeCh`:** `OnMessage` currently calls `p.grademap.Update(msg.Payload())` for in-memory use. Continue doing that. Also send the raw payload on `gradeCh` for persistence. Both operations happen in the same `OnMessage` branch. This is not duplication — the in-memory grademap is used immediately for fruit grading; the channel payload is for persistence.
- **Not copying `msg.Payload()`:** The paho library may reuse the payload buffer after `OnMessage` returns. Always copy before sending on a channel.
- **Blocking on gradeCh send:** If `run()` is slow to drain `gradeCh`, a blocking send in `OnMessage` stalls the MQTT dispatcher. Always use `select/default`.
- **Storing `map[string]any` in `prevGrademapPayload` before the DB write succeeds:** If the INSERT fails, `prevGrademapPayload` should not be updated — the diff next time would skip the failed update's changes entirely. Only update `prevGrademapPayload` on successful DB write.
- **Using `json.RawMessage` or a custom type for the JSONB column:** pgx accepts `[]byte` directly for JSONB. No wrapper needed.

---

## Schema

### breeze_grademap

```sql
CREATE TABLE breeze_grademap (
    id          BIGSERIAL PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload     JSONB NOT NULL
);
```

This is a plain PostgreSQL table, not a TimescaleDB hypertable. Grademap updates are infrequent (a few per day at most) so time-partitioning provides no benefit. The table is append-only by convention — no UPDATE or DELETE operations.

An index on `received_at` supports the Phase 4 AS-OF lookup:

```sql
CREATE INDEX breeze_grademap_received_at_idx ON breeze_grademap (received_at DESC);
```

### breeze_grademap_changes

```sql
CREATE TABLE breeze_grademap_changes (
    id           BIGSERIAL PRIMARY KEY,
    grademap_id  BIGINT NOT NULL REFERENCES breeze_grademap(id),
    entity_type  TEXT NOT NULL,   -- 'grademap', 'characteristic', 'grade', 'pass'
    entity_name  TEXT NOT NULL,   -- e.g. 'Russet', 'Class A', ''
    field        TEXT NOT NULL,   -- e.g. 'Name', 'Min_value', 'Thresholds'
    old_value    TEXT,            -- NULL on first insert (no previous)
    new_value    TEXT NOT NULL
);
```

Indexes supporting the audit query ("show all threshold changes this month"):

```sql
CREATE INDEX breeze_grademap_changes_grademap_id_idx ON breeze_grademap_changes (grademap_id);
CREATE INDEX breeze_grademap_changes_field_idx ON breeze_grademap_changes (field);
```

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSONB insert with generated ID | Manual multi-step INSERT + SELECT | `QueryRow` with `RETURNING id` | Single round-trip; atomic; idiomatic pgx |
| Bulk change row insert | Loop of individual `Exec` per change | `pgx.CopyFromRows` + `pool.CopyFrom` | Consistent with fruit write pattern; single round-trip for N rows |
| JSON diff | Custom recursive comparator with maps | Plain `map[string]any` walk with `fmt.Sprintf` for stringification | Grademap structure is known and bounded; stdlib is sufficient and zero-dependency |
| Non-blocking channel send | Mutex + condition variable | Buffered channel + `select/default` | The proven pattern already used for fruit; no additional primitives |

**Key insight:** The grademap persistence path is structurally identical to the fruit persistence path — channel from `OnMessage` to `run()`, then DB write inside the select loop. The only new complexity is the diff computation, which is a pure function with no I/O.

---

## Common Pitfalls

### Pitfall 1: Not Copying msg.Payload()

**What goes wrong:** The raw bytes sent on `gradeCh` are modified or zeroed by the MQTT library after `OnMessage` returns, corrupting the payload before `run()` processes it.

**Why it happens:** `paho.mqtt.golang`'s `mqtt.Message.Payload()` returns a slice backed by an internal buffer that is reused per message.

**How to avoid:** Always copy: `payload := append([]byte(nil), msg.Payload()...)` or `make + copy` before the channel send.

**Warning signs:** Grademap payloads in the database contain zero bytes or garbled JSON.

### Pitfall 2: Updating prevGrademapPayload on DB Failure

**What goes wrong:** If the `breeze_grademap` INSERT fails, `prevGrademapPayload` is updated anyway. The next successful update diffs against the wrong baseline, missing all changes from the failed update.

**Why it happens:** Updating state before confirming success.

**How to avoid:** Gate the `prevGrademapPayload = current` assignment on `err == nil` from the INSERT.

**Warning signs:** `breeze_grademap_changes` is missing expected change rows after a period of DB errors.

### Pitfall 3: gradeCh Buffer Size Too Small

**What goes wrong:** If `run()` is briefly busy (e.g., writing a large fruit batch), `OnMessage` drops a grademap update because `gradeCh` is full (buffer = 1, already occupied by the previous update that hasn't been drained yet).

**Why it happens:** Grademap updates are rare but can arrive while `run()` is in the middle of a fruit write.

**How to avoid:** A buffer of 1 is borderline acceptable given grademap update frequency (< 1 per second). If drop warnings are observed in practice, increase to 4 or 8. Log the drop with `slog.Warn` (matching the fruit drop pattern) so the operator can detect it.

**Warning signs:** Repeated "Dropped grademap update" log lines.

### Pitfall 4: Blocking the MQTT Dispatcher on grademap.Update() Error

**What goes wrong:** If `p.grademap.Update()` is slow or panics, `OnMessage` stalls. This is pre-existing behavior, not new in Phase 3 — but Phase 3 adds a channel send immediately after, making timing issues more visible.

**Why it happens:** `Update()` does JSON parsing inline in `OnMessage`.

**How to avoid:** This is acceptable for Phase 3 — `Update()` is fast in practice (small payload). The non-blocking requirement is specifically about DB latency, which Phase 3 addresses. Document this as a known trade-off.

**Warning signs:** MQTT message rate drops when grademaps are received.

### Pitfall 5: JSONB vs JSON Column Type

**What goes wrong:** Creating the column as `JSON` instead of `JSONB` — queries like `payload->>'Name'` still work but cannot use GIN indexes and are slower for key extraction.

**Why it happens:** Conflating `JSON` (text storage) with `JSONB` (binary storage with indexing).

**How to avoid:** Always use `JSONB` for queryable JSON in PostgreSQL. The requirements specify JSONB. pgx accepts `[]byte` for both without any code change.

**Warning signs:** `payload->'key'` queries are slow; `CREATE INDEX ... USING GIN` fails.

---

## Code Examples

### QueryRow with RETURNING (pgx v5)

```go
// Source: pkg.go.dev/github.com/jackc/pgx/v5 — QueryRow / Scan
var grademapID int64
err := pool.QueryRow(ctx,
    `INSERT INTO breeze_grademap (received_at, payload) VALUES ($1, $2) RETURNING id`,
    time.Now(),
    payload,
).Scan(&grademapID)
```

### Non-Blocking Send with Payload Copy

```go
// Source: existing processor/processor.go OnMessage select/default pattern
rawPayload := make([]byte, len(msg.Payload()))
copy(rawPayload, msg.Payload())
select {
case p.gradeCh <- rawPayload:
default:
    slog.Warn("Dropped grademap update — channel full")
}
```

### Diff Function Signature (pure, testable)

```go
// In processor/grademap_diff.go
type GrademapChange struct {
    EntityType string // "grademap", "characteristic", "grade", "pass"
    EntityName string // e.g. "Russet", "Class A"
    Field      string // e.g. "Name", "Min_value"
    OldValue   string // empty string when prev is nil
    NewValue   string
}

func DiffGrademaps(prev, next map[string]any) []GrademapChange
```

### CopyFrom for Change Rows

```go
// Source: existing breeze.go CopyFromRows + pool.CopyFrom pattern
changeRows := make([][]any, len(changes))
for i, c := range changes {
    changeRows[i] = []any{
        grademapID, c.EntityType, c.EntityName, c.Field, c.OldValue, c.NewValue,
    }
}
if len(changeRows) > 0 {
    _, err = pool.CopyFrom(ctx,
        pgx.Identifier{"breeze_grademap_changes"},
        []string{"grademap_id", "entity_type", "entity_name", "field", "old_value", "new_value"},
        pgx.CopyFromRows(changeRows),
    )
    if err != nil {
        slog.Error("Failed to write grademap changes", "err", err)
    }
}
```

### Exported GradeCh Accessor on Processor

```go
// In processor/processor.go
func (p *Processor) GradeCh() <-chan []byte {
    return p.gradeCh
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual `BEGIN` / `INSERT` / `COMMIT` | `QueryRow` + `RETURNING` | pgx v5 | Single round-trip; no manual transaction needed for one-row insert |
| `database/sql` requires `conn.Raw()` for COPY | `pool.CopyFrom` directly | pgx v5 | No unwrapping; works on pool directly |
| `json.RawMessage` for JSONB | `[]byte` accepted directly | pgx v5 | No wrapper type needed |

---

## Open Questions

1. **Should the grademap insert and its change rows be in one transaction?**
   - What we know: If the changes insert fails after the grademap insert succeeds, `breeze_grademap` has a row with no corresponding changes. The next update will compute the diff correctly because `prevGrademapPayload` was not updated on the failed changes write.
   - What's unclear: Whether the product requires "grademap row and its change rows are always atomic."
   - Recommendation: Do not wrap in a transaction for Phase 3. The diff will self-correct on the next successful update. Wrapping in a transaction adds complexity for a low-frequency operation where self-correction works. Revisit if Phase 4 reveals an atomicity requirement.

2. **Depth of the diff: how deep to walk nested JSON?**
   - What we know: The grademap JSON has at most 3 levels of nesting (top-level → entity → field → value). The `Characteristics` and `Grades` maps are 2 levels deep.
   - What's unclear: Whether threshold arrays (e.g., `Thresholds: [{Min: 0, Max: 50}]`) should be diffed element-by-element or stringified as a whole.
   - Recommendation: Stringify arrays and nested objects as JSON for `old_value`/`new_value`. This keeps the diff function simple and the change rows human-readable. If finer granularity is needed, it is an additive change in Phase 3 or later.

3. **What to do with the gradeCh on DB unhealthy state?**
   - What we know: Phase 2 introduced a `dbHealthy` flag. When the DB is down, fruit goes to the write buffer. Grademap payloads arriving while DB is unhealthy cannot be written.
   - What's unclear: Whether to buffer grademap payloads during DB outages (separate from fruit buffer) or drop and log.
   - Recommendation: Drop and log. Grademap updates are rare configuration changes, not data records. Missing a grademap update during a DB outage is acceptable — the OEM will resend the grademap when it reconnects. This is simpler than adding a second buffer. Note this in code comments.

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
| GRAD-01 | `breeze_grademap` and `breeze_grademap_changes` DDL is correct (column names, types, FK) | unit (schema string check or SQL parse) | `go test ./... -run TestGrademapSchema -count=1` | Wave 0 |
| GRAD-02 | Raw payload is sent non-blocking on `gradeCh`; `OnMessage` returns immediately regardless of DB latency | unit | `go test ./processor/... -run TestOnMessage_GrademapChannel -count=1` | Wave 0 |
| GRAD-03 | `DiffGrademaps` returns correct change rows: one per changed field, with correct entity_type/entity_name/field/old_value/new_value | unit | `go test ./processor/... -run TestDiffGrademaps -count=1` | Wave 0 |

All tests are pure unit tests — no live DB required. `TestGrademapSchema` can validate the SQL DDL strings as Go constants (the schema strings live in `schema.sql` or as `const` in a Go file). `TestOnMessage_GrademapChannel` uses the existing `mockMsg` pattern to verify the non-blocking send. `TestDiffGrademaps` is a table-driven test over the pure diff function.

### Sampling Rate

- **Per task commit:** `go test ./... -count=1 -timeout 30s`
- **Per wave merge:** `go test ./... -count=1 -race -timeout 60s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `processor/grademap_diff_test.go` — covers GRAD-03 (`TestDiffGrademaps` table-driven)
- [ ] `processor/grademap_diff.go` — the `DiffGrademaps` function and `GrademapChange` type
- [ ] `processor/processor_test.go` — add `TestOnMessage_GrademapChannel` (covers GRAD-02 non-blocking send)
- [ ] Schema DDL strings — either in `schema.sql` additions or as Go `const` for testability (GRAD-01)

---

## Sources

### Primary (HIGH confidence)

- `pkg.go.dev/github.com/jackc/pgx/v5` — `QueryRow`, `Scan`, `CopyFrom`, `CopyFromRows`, `Identifier`, JSONB `[]byte` acceptance
- `pkg.go.dev/github.com/jackc/pgx/v5/pgxpool` — `Pool.QueryRow`, `Pool.CopyFrom`
- Existing `breeze.go` — `writeBatch`, `flushBuffer` patterns for `CopyFromRows` and `pgx.Identifier`
- Existing `processor/processor.go` — non-blocking channel send pattern (`select/default` on `p.queue`)
- Existing `processor/grademap.go` — grademap JSON structure: top-level keys, nested `Characteristics`, `Grades`, `Defect_grading_passes`
- PostgreSQL docs — `JSONB` vs `JSON`, `BIGSERIAL`, `RETURNING`, index types

### Secondary (MEDIUM confidence)

- paho MQTT Go client source — `Payload()` returns library-managed buffer; copy before retaining is standard advice in MQTT client documentation
- PostgreSQL `JSONB` operator docs — `->`, `->>` operators confirmed to work on `JSONB` column; GIN index confirmed for `JSONB`

### Tertiary (LOW confidence)

- None — all findings are derived from already-imported libraries or the existing codebase.

---

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — all libraries already in `go.mod`; API signatures taken from existing code and pkg.go.dev
- Architecture: HIGH — channel pattern is identical to the existing fruit path; single-goroutine design is confirmed by existing `run()` structure
- Schema: HIGH — PostgreSQL `BIGSERIAL`, `JSONB`, `REFERENCES`, `TIMESTAMPTZ` are stable standard SQL; TimescaleDB is not used for these tables (infrequent data)
- Diff algorithm: HIGH — grademap JSON structure is fully known from `grademap.go`; map walk approach is well-established Go
- Pitfalls: HIGH — payload copy requirement is documented paho behavior; others are logic deductions from the existing architecture

**Research date:** 2026-03-08
**Valid until:** 2026-06-08 (pgx v5 API stable; Go stdlib stable; PostgreSQL DDL stable)
