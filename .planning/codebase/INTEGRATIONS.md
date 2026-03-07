# External Integrations

**Analysis Date:** 2026-03-07

## APIs & External Services

**MQTT Broker (Tomra/Spectrim fruit sizer):**
- Purpose: Receives real-time fruit grading and grademap messages from industrial fruit sizing hardware
- SDK/Client: `github.com/eclipse/paho.mqtt.golang` v1.5.0
- Auth: Username/password credentials from config (`SERVER_USER`, `SERVER_PASS`)
- Connection: Broker URI from config (`SERVER_URI`)
- Client ID: Randomised per run (`go_breeze_<6-byte-base64>`) to avoid session conflicts
- Subscribed topics:
  - `tomra/211632/fruit` - Per-fruit measurement JSON payloads from Spectrim vision system
  - `tomra/211632/grademap` - Grade classification rule sets (updated dynamically)
- QoS: 0 (at most once delivery)
- Reconnect: `OnConnect` handler re-subscribes automatically on reconnect; `OnConnectionLost` logs a warning

## Data Storage

**Databases:**
- Type: PostgreSQL with TimescaleDB extension
- Connection env var: `DATABASE_URL` (full connection string) or YAML key `database.url`
- Client: `github.com/jackc/pgx/v5` with `pgxpool` connection pool
- Write strategy: Bulk `COPY FROM` via `pgxpool.CopyFrom` (not individual INSERTs) for throughput
- Schema file: `schema.sql`
- Primary table: `breeze_fruit` - wide time-series table with ~55 columns covering sizer, spectrim vision, and grading data
- TimescaleDB features used:
  - Hypertable on `time` column (partitioned by `time`)
  - Chunk interval: `1d`
  - Columnstore policy after `3d` (`add_columnstore_policy`)
  - Order by: `time DESC`
  - JSONB columns for nested spectrim data: `center_offsets`, `classified_blob`, `colour`, `colour_blob`, `diameters`, `features`, `function`, `timer`

**File Storage:**
- Not applicable

**Caching:**
- None

## Authentication & Identity

**Auth Provider:**
- None (no user authentication layer)
- MQTT broker auth via username/password in config
- PostgreSQL auth via connection string in config

## Monitoring & Observability

**Error Tracking:**
- None (no external error tracking service)

**Logs:**
- Go standard `log/slog` structured logging throughout (`breeze.go`, `processor/processor.go`, `processor/fruit.go`, `processor/grademap.go`)
- Log levels: Info (normal operations), Debug (per-fruit writes, batch details), Warn (connection loss, invalid data), Error (write failures)
- Debug logging is commented out at startup (`slog.SetLogLoggerLevel(slog.LevelDebug)` in `breeze.go` line 49)

## CI/CD & Deployment

**Hosting:**
- Not specified; designed for on-premises deployment alongside Tomra/Spectrim fruit sizer hardware

**CI Pipeline:**
- None detected

## Environment Configuration

**Required env vars (or YAML equivalents):**
- `SERVER_URI` - MQTT broker URI (e.g., `tcp://host:1883`)
- `SERVER_USER` - MQTT broker username
- `SERVER_PASS` - MQTT broker password
- `DATABASE_URL` - PostgreSQL DSN (e.g., `postgres://user:pass@host/db`)

**Secrets location:**
- `config.yml` (gitignored, not committed) - YAML config file read at startup
- `.env` (gitignored) - environment variables
- Config path overridable: `breeze -config /path/to/config.yml`

## Webhooks & Callbacks

**Incoming:**
- None (no HTTP server)

**Outgoing:**
- None

## Message Processing Pipeline

The integration data flow is:

1. MQTT broker pushes JSON messages to `tomra/211632/fruit` and `tomra/211632/grademap`
2. `processor.Processor.OnMessage` dispatches by topic suffix
3. Fruit messages are unmarshalled via custom `Fruit.UnmarshalJSON` (mapping nested Sizer + Spectrim JSON structures)
4. Grademap messages update in-memory `Grademap` state used to classify defects
5. Fruit records are queued in a channel (capacity 50); a goroutine batches up to 30 items or flushes after 1 second
6. Main loop receives batch-ready signals and calls `pgxpool.CopyFrom` to bulk-insert into `breeze_fruit`

---

*Integration audit: 2026-03-07*
