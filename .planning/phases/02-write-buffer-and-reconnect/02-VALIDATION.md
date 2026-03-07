---
phase: 2
slug: write-buffer-and-reconnect
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 2 — Validation Strategy

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
| 2-01-01 | 01 | 0 | RECO-01,02,03,04 | unit stubs | `go test ./... -run "TestWriteBuffer\|TestReconnectProbe\|TestFlushBuffer\|TestTransactionalWrite" -count=1` | ❌ W0 | ⬜ pending |
| 2-02-01 | 02 | 1 | RECO-01 | unit | `go test ./... -run TestWriteBuffer -count=1` | ❌ W0 | ⬜ pending |
| 2-02-02 | 02 | 1 | RECO-02 | unit | `go test ./... -run TestReconnectProbe -count=1` | ❌ W0 | ⬜ pending |
| 2-02-03 | 02 | 1 | RECO-03 | unit | `go test ./... -run TestFlushBuffer -count=1` | ❌ W0 | ⬜ pending |
| 2-02-04 | 02 | 1 | RECO-04 | unit | `go test ./... -run TestTransactionalWrite -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `breeze_test.go` (or `buffer_test.go`) — stubs for RECO-01 (`TestWriteBuffer`), RECO-02 (`TestReconnectProbe`), RECO-03 (`TestFlushBuffer`), RECO-04 (`TestTransactionalWrite`)
- [ ] `processor/drain_test.go` — if `DrainFruits` helper is added to processor package

**Note:** All tests are pure unit tests using mocks — no live DB required. Existing `processor/processor_test.go` (Phase 1) must not be modified.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None | — | All behaviors have automated unit coverage | — |

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
