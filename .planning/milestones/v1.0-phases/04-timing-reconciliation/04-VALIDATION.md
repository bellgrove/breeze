---
phase: 4
slug: timing-reconciliation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard testing (`testing` package) |
| **Config file** | none — `go test ./...` from repo root |
| **Quick run command** | `go test ./... -count=1 -timeout 30s` |
| **Full suite command** | `go test ./... -count=1 -race -timeout 60s` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1 -timeout 30s`
- **After every plan wave:** Run `go test ./... -count=1 -race -timeout 60s`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | TIME-01 | unit | `go test . -run "TestResolveGrademapID\|TestWriteBatch_GrademapID" -count=1` | Wave 0 | ⬜ pending |
| 04-01-02 | 01 | 1 | TIME-02 | unit | `go test . -run "TestConfig_PropagationDelay\|TestResolveGrademapID_WithDelay" -count=1` | Wave 0 | ⬜ pending |
| 04-02-01 | 02 | 2 | TIME-01 | unit | `go test . -run TestResolveGrademapID -count=1` | ✅ | ⬜ pending |
| 04-02-02 | 02 | 2 | TIME-02 | unit | `go test . -run "TestConfig_PropagationDelay\|TestResolveGrademapID_WithDelay" -count=1` | ✅ | ⬜ pending |
| 04-03-01 | 03 | 3 | TIME-01 | unit | `go test . -run TestWriteBatch_GrademapID -count=1` | ✅ | ⬜ pending |
| 04-03-02 | 03 | 3 | TIME-01, TIME-02 | unit | `go test ./... -count=1 -timeout 30s` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `breeze_test.go` — add stubs: `TestResolveGrademapID`, `TestResolveGrademapID_ZeroDelay`, `TestResolveGrademapID_NoRows`, `TestWriteBatch_GrademapID`, `TestResolveGrademapID_WithDelay`, `TestConfig_PropagationDelay`, `TestConfig_PropagationDelayZero`

*Note: `resolveGrademapID` requires a DB connection. Unit stubs should exercise error-handling branches (pgx.ErrNoRows, nil pool) using the same closed-pool pattern as `TestTransactionalWrite`. Live DB integration is manual-only.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Fruit row in DB has correct `grademap_id` FK after live grademap change | TIME-01 | Requires live OEM + running DB | Change grademap via MQTT; confirm fruit rows after change reference new grademap_id |
| Propagation delay shifts attribution correctly in live traffic | TIME-02 | Requires production timing data | Set `grademap_propagation_delay: 30s`; verify fruit timestamped within delay window attribute to previous grademap |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
