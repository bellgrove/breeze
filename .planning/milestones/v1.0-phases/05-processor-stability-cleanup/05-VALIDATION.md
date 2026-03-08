---
phase: 5
slug: processor-stability-cleanup
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-08
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none — `go test ./...` from repo root |
| **Quick run command** | `go test ./processor/... -count=1` |
| **Full suite command** | `go test ./... -count=1 -race` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./processor/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1 -race`
- **Before `/gsd:verify-work`:** Full suite must exit 0 (no panics, no failures)
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | QUAL-05 | unit | `go test ./processor/... -run TestFromJson -count=1` | ✅ | ⬜ pending |
| 05-01-02 | 01 | 1 | QUAL-01 | unit | `go test ./processor/... -run TestOnMessage -count=1` | ✅ | ⬜ pending |
| 05-01-03 | 01 | 1 | QUAL-01, QUAL-05 | integration | `go test ./... -count=1 -race` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. No new test files required.

- `TestFromJson/Class_2` (in `processor/fruit_test.go`) already exists — currently panics, will pass after Fix 2
- `TestOnMessage_QueueFull` (in `processor/processor_test.go`) already exists — will continue to pass after Fix 1

**Optional wave 0 extension (planner decision):** Add `TestOnMessage_TimerNonBlocking` using `Create()` with timer cap 1 to directly exercise the timer fix. The fix is also verifiable by code inspection + clean test run.

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
