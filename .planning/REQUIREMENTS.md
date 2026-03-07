# Requirements: Breeze

**Defined:** 2026-03-07
**Core Value:** Every fruit record must be attributable to a specific defect cause — "10% class B" is useless; "4% insect damage, 3% sunburn, 2% punctures, 1% rot" is the goal.

## v1 Requirements

### Code Quality

- [x] **QUAL-01**: `OnMessage` uses a non-blocking channel send with drop-and-warn so the paho MQTT dispatcher never stalls
- [x] **QUAL-02**: `Processor` counter decrements after channel read, not before (fix counter/channel race)
- [x] **QUAL-03**: `json.Unmarshal` errors in `OnMessage` are logged, not silently swallowed
- [x] **QUAL-04**: Shutdown defer order in `run()` is corrected — MQTT disconnect before pool close
- [ ] **QUAL-05**: General code review and cleanup (dead code, commented-out imports, style)

### Database Reconnection

- [ ] **RECO-01**: Fruit records are buffered in a bounded in-memory slice when `CopyFrom` fails (configurable max size, drop-oldest policy)
- [ ] **RECO-02**: A background reconnect probe (`pool.Ping` with exponential backoff) runs when DB is unhealthy
- [ ] **RECO-03**: Buffered records are automatically flushed when DB connection is restored
- [ ] **RECO-04**: All `CopyFrom` calls are wrapped in explicit transactions to prevent partial-batch duplicates on retry

### Grademap History

- [ ] **GRAD-01**: `breeze_grademap` table exists (append-only: `id`, `received_at TIMESTAMPTZ`, `payload JSONB`)
- [ ] **GRAD-02**: A record is inserted on every MQTT grademap update (sent via channel to `run()`, non-blocking in `OnMessage`)
- [ ] **GRAD-03**: `breeze_grademap_changes` table exists with queryable change rows — one row per changed field: `(grademap_id FK, entity_type, entity_name, field, old_value TEXT, new_value TEXT)` — enabling queries like "show all threshold changes this month"

### Timing Reconciliation

- [ ] **TIME-01**: Fruit records are matched to the correct historical grademap via AS-OF timestamp query (`WHERE received_at <= processing_time ORDER BY received_at DESC LIMIT 1`) instead of always using the current grademap
- [ ] **TIME-02**: A configurable `grademap_propagation_delay` offset is applied to the AS-OF lookup to compensate for OEM pipeline lag (fruit already in transit when grademap changes)

## v2 Requirements

### Observability

- **OBS-01**: Metrics endpoint exposing queue depth, buffer fill level, DB health state, and fruit throughput rate
- **OBS-02**: Structured log fields for grademap reconciliation misses (fruit processed but no matching grademap found)

### Multi-machine

- **MULTI-01**: Support for multiple Tomra grader units (configurable topic prefixes)

## Out of Scope

| Feature | Reason |
|---------|--------|
| HTTP API or web interface | Headless daemon only; analytics via TimescaleDB directly |
| Real-time alerting/notifications | Out of scope for v1; analytics queries serve this need |
| Mobile or browser dashboard | Not requested |
| OAuth / external auth | No HTTP layer |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| QUAL-01 | Phase 1 | Complete |
| QUAL-02 | Phase 1 | Complete |
| QUAL-03 | Phase 1 | Complete |
| QUAL-04 | Phase 1 | Complete |
| QUAL-05 | Phase 1 | Pending |
| RECO-01 | Phase 2 | Pending |
| RECO-02 | Phase 2 | Pending |
| RECO-03 | Phase 2 | Pending |
| RECO-04 | Phase 2 | Pending |
| GRAD-01 | Phase 3 | Pending |
| GRAD-02 | Phase 3 | Pending |
| GRAD-03 | Phase 3 | Pending |
| TIME-01 | Phase 4 | Pending |
| TIME-02 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-07*
*Last updated: 2026-03-07 after initial definition*
