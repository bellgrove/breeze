---
phase: 1
slug: code-cleanup
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-07
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing stdlib (`testing` package), Go 1.26.1 |
| **Config file** | none — standard `go test` |
| **Quick run command** | `go test ./processor/... -timeout 30s` |
| **Full suite command** | `go test ./... -timeout 60s` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./processor/... -timeout 30s`
- **After every plan wave:** Run `go test ./... -timeout 60s`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 0 | QUAL-01,02,03 | unit stub | `go test ./processor/... -run TestOnMessage_QueueFull` | ❌ W0 | ⬜ pending |
| 1-02-01 | 01 | 1 | QUAL-01 | unit | `go test ./processor/... -run TestOnMessage_QueueFull` | ❌ W0 | ⬜ pending |
| 1-02-02 | 01 | 1 | QUAL-02 | unit | `go test ./processor/... -run TestValues_CounterOrder` | ❌ W0 | ⬜ pending |
| 1-02-03 | 01 | 1 | QUAL-03 | unit | `go test ./processor/... -run TestOnMessage_MalformedJSON` | ❌ W0 | ⬜ pending |
| 1-02-04 | 01 | 1 | QUAL-04 | manual | Code inspection + smoke test on SIGINT | N/A | ⬜ pending |
| 1-02-05 | 01 | 1 | QUAL-05 | lint | `go build ./... && go mod tidy` | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `processor/processor_test.go` — stubs for QUAL-01 (`TestOnMessage_QueueFull`), QUAL-02 (`TestValues_CounterOrder`), QUAL-03 (`TestOnMessage_MalformedJSON`)

**Note:** No framework install needed — `testing` is stdlib. New test file must NOT use the broken `ValidateFruit` helper in `fruit_test.go` (pre-existing panic on map fields).

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Shutdown defer order: MQTT disconnect before pool close | QUAL-04 | Integration-only — requires mocking MQTT client and pool to unit test | Start service, send `SIGINT`, observe logs show `disconnected` before `pool closed` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
