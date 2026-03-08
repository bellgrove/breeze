---
phase: 03-grademap-persistence
verified: 2026-03-08T08:00:00Z
status: passed
score: 10/10 must-haves verified
re_verification: false
human_verification:
  - test: "Live MQTT grademap message received and persisted"
    expected: "A row appears in breeze_grademap with received_at timestamp and raw JSONB payload; corresponding rows in breeze_grademap_changes show per-field deltas"
    why_human: "Requires live TimescaleDB instance and active MQTT broker with real grademap traffic"
  - test: "DB-unhealthy drop path"
    expected: "When the DB is down, grademap payloads are dropped with slog.Warn and the processor does not block"
    why_human: "Requires simulating a DB outage in a running environment"
---

# Phase 03: Grademap Persistence Verification Report

**Phase Goal:** Every grademap received from MQTT is stored in TimescaleDB with its receipt timestamp and a queryable change log, forming the foundation for timing reconciliation and audit queries.
**Verified:** 2026-03-08T08:00:00Z
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| #  | Truth | Status | Evidence |
|----|-------|--------|----------|
| 1  | `schema.sql` contains `CREATE TABLE breeze_grademap` with `id BIGSERIAL PRIMARY KEY`, `received_at TIMESTAMPTZ`, `payload JSONB` | VERIFIED | `schema.sql` lines 74–78: exact DDL present |
| 2  | `schema.sql` contains `CREATE TABLE breeze_grademap_changes` with FK to `breeze_grademap` | VERIFIED | `schema.sql` lines 82–90: table with `BIGINT NOT NULL REFERENCES breeze_grademap(id)` present |
| 3  | `run()` select loop handles `<-proc.GradeCh()`, inserts grademap row, computes diff, writes change rows | VERIFIED | `breeze.go` lines 340–351: case `payload := <-proc.GradeCh()` calls `writeGrademap()` which does INSERT RETURNING id (line 138), calls `processor.DiffGrademaps` (line 152), then `pool.CopyFrom` to `breeze_grademap_changes` (line 159) |
| 4  | `prevGrademapPayload` is only updated after a successful INSERT — diff self-corrects on DB failure | VERIFIED | `breeze.go` line 351: `prevGrademapPayload = next` only reached after `writeGrademap` returns nil error; `continue` on error at line 349 skips the assignment |
| 5  | `TestGrademapSchema` passes GREEN — `grademapDDL` and `grademapChangesDDL` constants exist | VERIFIED | `go test . -run TestGrademapSchema -count=1` exits 0 (ok) |
| 6  | `DiffGrademaps` returns one `GrademapChange` per changed leaf field with correct entity_type, entity_name, field, old_value, new_value | VERIFIED | Full implementation in `processor/grademap_diff.go` (129 lines); `TestDiffGrademaps` passes GREEN |
| 7  | `DiffGrademaps(nil, next)` treats all fields as new (OldValue is empty string) | VERIFIED | `grademap_diff.go` line 109: `oldStr := ""` when `prev == nil`; test case "nil prev — all fields in next are new" passes |
| 8  | `DiffGrademaps(prev, next)` where nothing changed returns empty/nil slice | VERIFIED | `grademap_diff.go` lines 114–125: skips when `oldStr == newStr`; test case "nothing changed — empty slice" passes |
| 9  | `OnMessage` sends a copy of the raw grademap payload on `gradeCh` (non-blocking, returns immediately) | VERIFIED | `processor.go` lines 33–39: `make([]byte, len)` + `copy` + `select/default`; `TestOnMessage_GrademapChannel` passes GREEN |
| 10 | `OnMessage` logs a warning and does not block when `gradeCh` is already full | VERIFIED | `processor.go` line 38: `default: slog.Warn("Dropped grademap update — channel full")`; second-send case in `TestOnMessage_GrademapChannel` passes GREEN |

**Score:** 10/10 truths verified

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `processor/grademap_diff.go` | `GrademapChange` struct + full `DiffGrademaps` implementation | VERIFIED | 129 lines; full implementation with nil-prev, scalar, nested-container cases; `containerEntityTypes` explicit map |
| `processor/grademap_diff_test.go` | Table-driven `TestDiffGrademaps` with 3+ cases | VERIFIED | 79 lines; 3 cases: nil-prev, field-changed, no-change |
| `processor/processor.go` | `gradeCh chan []byte` field, `GradeCh()` accessor, non-blocking send in `OnMessage` | VERIFIED | Lines 18, 25, 33–39; `gradeCh: make(chan []byte, 1)` in `Create()` at line 145 |
| `processor/processor_test.go` | `TestOnMessage_GrademapChannel` verifying channel delivery | VERIFIED | Test present, passes GREEN; `newTestProcessorSmallQueue` initialises `gradeCh` at line 37 |
| `schema.sql` | DDL for `breeze_grademap` and `breeze_grademap_changes` with indexes | VERIFIED | Lines 74–93; both tables + 3 indexes present |
| `breeze.go` | `grademapDDL` and `grademapChangesDDL` constants, `writeGrademap` helper, `gradeCh` case in `run()` | VERIFIED | Constants at lines 24–38; `writeGrademap` at lines 135–169; `run()` case at lines 340–351 |
| `breeze_test.go` | `TestGrademapSchema` checking DDL constants | VERIFIED | Lines 94–106; passes GREEN |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `breeze.go run() select loop` | `proc.GradeCh()` | `case payload := <-proc.GradeCh()` | WIRED | `breeze.go` line 340 |
| `breeze.go writeGrademap` | `pool.QueryRow RETURNING id` | `INSERT INTO breeze_grademap RETURNING id` | WIRED | `breeze.go` lines 137–141 |
| `breeze.go writeGrademap` | `pool.CopyFrom breeze_grademap_changes` | `pgx.CopyFromRows` | WIRED | `breeze.go` lines 158–163 |
| `breeze_test.go TestGrademapSchema` | `breeze.go grademapDDL` | `strings.Contains` constant check | WIRED | `breeze_test.go` line 95; constant at `breeze.go` line 24 |
| `processor/processor.go OnMessage` | `processor/processor.go gradeCh` | `select/default` non-blocking send | WIRED | `processor.go` lines 35–39; `select { case p.gradeCh <- rawPayload: default: ... }` |
| `processor/processor_test.go TestOnMessage_GrademapChannel` | `processor/processor.go GradeCh` | `p.GradeCh()` read-only channel | WIRED | Test passes GREEN; accessor at `processor.go` line 25 |
| `processor/grademap_diff_test.go` | `processor/grademap_diff.go` | `DiffGrademaps` function call | WIRED | `grademap_diff_test.go` line 48 calls `DiffGrademaps(tc.prev, tc.next)` |

---

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| GRAD-01 | 03-01, 03-03 | `breeze_grademap` table exists (append-only: `id`, `received_at TIMESTAMPTZ`, `payload JSONB`) | SATISFIED | `schema.sql` lines 74–78; `grademapDDL` constant in `breeze.go`; `TestGrademapSchema` GREEN |
| GRAD-02 | 03-01, 03-02, 03-03 | A record is inserted on every MQTT grademap update (sent via channel to `run()`, non-blocking in `OnMessage`) | SATISFIED | `processor.go` non-blocking send + `breeze.go` `case payload := <-proc.GradeCh()` → `writeGrademap()` INSERT; `TestOnMessage_GrademapChannel` GREEN |
| GRAD-03 | 03-01, 03-02, 03-03 | `breeze_grademap_changes` table with per-field change rows (`grademap_id FK`, `entity_type`, `entity_name`, `field`, `old_value`, `new_value`) | SATISFIED | `schema.sql` lines 82–90; `writeGrademap` calls `processor.DiffGrademaps` then `pool.CopyFrom` to `breeze_grademap_changes`; `TestDiffGrademaps` GREEN |

No orphaned requirements: GRAD-01, GRAD-02, GRAD-03 are all claimed by plans and implemented.

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `processor/fruit_test.go` | 134 | `reflect.Value.Equal` on `map[string]interface{}` panics — `TestFromJson/Class_2` always panics | Warning | Pre-existing bug documented in `deferred-items.md`; not introduced by Phase 03; does not affect any GRAD requirement |

No placeholder implementations, empty stubs, or TODO comments found in Phase 03 deliverables. The `DiffGrademaps` stub from Plan 01 was fully replaced in Plan 02.

---

### Human Verification Required

#### 1. Live end-to-end grademap persistence

**Test:** With a running TimescaleDB instance and an active MQTT broker, publish a grademap message on `tomra/211632/grademap` and inspect the database.
**Expected:** One row in `breeze_grademap` with `received_at` near the publish time and `payload` containing the raw JSONB; if a prior grademap exists, corresponding rows in `breeze_grademap_changes` showing per-field deltas.
**Why human:** Requires a live DB and MQTT broker; automated tests use unit-level isolation without I/O.

#### 2. DB-unhealthy drop behaviour

**Test:** Stop the database while breeze is running, then publish a grademap message.
**Expected:** `slog.Warn("Dropped grademap update — DB unhealthy")` appears in logs; process continues without blocking; no panic.
**Why human:** Requires simulating a DB outage in a running environment.

---

### Gaps Summary

No gaps. All 10 observable truths are verified, all 7 artifacts pass all three levels (exists, substantive, wired), all 7 key links are wired, and all 3 requirement IDs are satisfied. The pre-existing `TestFromJson` panic is documented in `deferred-items.md` and is out of scope for Phase 03.

---

_Verified: 2026-03-08T08:00:00Z_
_Verifier: Claude (gsd-verifier)_
