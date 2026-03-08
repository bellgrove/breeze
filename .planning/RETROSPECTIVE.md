# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0 — Production-reliable daemon

**Shipped:** 2026-03-08
**Phases:** 5 | **Plans:** 13 | **Commits:** 58

### What Was Built
- Non-blocking `OnMessage` with drop-and-warn — MQTT dispatcher cannot stall regardless of queue or timer state
- Bounded write buffer (configurable max, drop-oldest) + exponential backoff reconnect probe — zero fruit data loss during DB outages
- `breeze_grademap` + `breeze_grademap_changes` TimescaleDB tables — append-only grademap history with per-field change rows
- AS-OF timestamp grademap attribution with configurable `grademap_propagation_delay` offset — each fruit linked to the correct historical grademap
- Phase 5 tech debt closure: inner `select/default` for timer send, `reflect.DeepEqual` fix for clean `go test ./...`

### What Worked
- Strict dependency order (1 → 2 → 3 → 4 → 5) prevented integration surprises — each phase's outputs were clean inputs for the next
- Nyquist wave 0 test stubs written first meant implementation had a clear passing target
- Milestone audit before completion surfaced two high-priority tech debt items (timer blocking send, TestFromJson panic) that were immediately actionable as Phase 5 — good feedback loop
- Decimal phase numbering (Phase 5 inserted post-audit) was clean and unambiguous

### What Was Inefficient
- VALIDATION.md frontmatter was never updated after execution in Phases 1-4 — `nyquist_compliant` remained `false` in all four files despite passing tests, requiring re-investigation at audit time
- Live E2E verification (MQTT + TimescaleDB) remains pending — automated tests cover the logic but operational confidence requires live infrastructure
- Phase 2 investigation (`DrainFruits` appeared in working tree ahead of its plan) — indicates some code was written outside the plan cadence; not harmful but not tracked

### Patterns Established
- Inner `select/default` inside outer `select` for doubly non-blocking channel sends in `OnMessage`
- `reflect.DeepEqual` (not `reflect.Value.Equal`) for map comparison in Go tests
- Phase summaries with structured YAML frontmatter enable tooling to extract one-liners and requirements linkage
- Milestone audit before completion is worth running even when all plans are complete — it surfaces deferred items that should be closed before shipping

### Key Lessons
1. Update VALIDATION.md frontmatter immediately after each plan executes — stale `nyquist_compliant: false` causes confusion at audit time
2. Live E2E verification should be a planned task, not deferred — the operational confirmation of grademap persistence and timing reconciliation needs actual MQTT + DB infrastructure
3. Phase 5 as audit-driven tech debt closure worked well — if audit finds medium-high priority items, adding a cleanup phase before tagging is cleaner than leaving them as known debt

### Cost Observations
- Model mix: predominantly sonnet (balanced profile)
- Sessions: multiple across 2026-03-07 → 2026-03-08
- Notable: parallelized plan execution (yolo mode, coarse granularity) kept per-phase wall time low

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Phases | Plans | Key Change |
|-----------|--------|-------|------------|
| v1.0 | 5 | 13 | First milestone — established GSD workflow baseline |

### Cumulative Quality

| Milestone | Go LOC | `go test ./...` | Notes |
|-----------|--------|-----------------|-------|
| v1.0 | 6,146 | Clean | Unit tests pass; live E2E pending |

### Top Lessons (Verified Across Milestones)

1. Update validation frontmatter immediately after execution — stale status creates audit confusion
2. Milestone audit before tagging is worth the overhead — surfaces closable tech debt
