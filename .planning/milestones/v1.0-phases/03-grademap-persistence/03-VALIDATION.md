---
phase: 3
slug: grademap-persistence
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 3 — Validation Strategy

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
| 03-01-01 | 01 | 1 | GRAD-03 | unit | `go test ./processor/... -run TestDiffGrademaps -count=1` | Wave 0 | ⬜ pending |
| 03-01-02 | 01 | 1 | GRAD-02 | unit | `go test ./processor/... -run TestOnMessage_GrademapChannel -count=1` | Wave 0 | ⬜ pending |
| 03-01-03 | 01 | 1 | GRAD-01 | unit | `go test ./... -run TestGrademapSchema -count=1` | Wave 0 | ⬜ pending |
| 03-02-01 | 02 | 2 | GRAD-03 | unit | `go test ./processor/... -run TestDiffGrademaps -count=1` | ✅ | ⬜ pending |
| 03-02-02 | 02 | 2 | GRAD-02 | unit | `go test ./processor/... -run TestOnMessage_GrademapChannel -count=1` | ✅ | ⬜ pending |
| 03-03-01 | 03 | 2 | GRAD-01 | unit | `go test ./... -run TestGrademapSchema -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `processor/grademap_diff.go` — `GrademapChange` type and `DiffGrademaps` function stub
- [ ] `processor/grademap_diff_test.go` — `TestDiffGrademaps` table-driven test stubs for GRAD-03
- [ ] `processor/processor_test.go` — add `TestOnMessage_GrademapChannel` stub for GRAD-02 non-blocking send
- [ ] Schema DDL as testable Go constants or `schema.sql` additions — `TestGrademapSchema` for GRAD-01

*All tests are pure unit tests — no live DB required.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None | — | — | — |

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
