# Technology Stack

**Analysis Date:** 2026-03-07

## Languages

**Primary:**
- Go 1.24.3 - All application code (`breeze.go`, `processor/`)

## Runtime

**Environment:**
- Go runtime 1.24.3

**Package Manager:**
- Go modules (`go mod`)
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- No web framework - this is a headless daemon/service, not an HTTP server

**Testing:**
- Go standard `testing` package - used in `processor/fruit_test.go`, `processor/grademap_test.go`

**Build/Dev:**
- VS Code with Go extension (`.vscode/launch.json`) for local debugging
- Standard `go build` / `go run` for building

## Key Dependencies

**Critical:**
- `github.com/eclipse/paho.mqtt.golang` v1.5.0 - MQTT client, the primary message ingestion protocol
- `github.com/jackc/pgx/v5` v5.7.5 - PostgreSQL driver with connection pooling (`pgxpool`), used for bulk `COPY FROM` writes
- `github.com/kelseyhightower/envconfig` v1.4.0 - Environment variable configuration binding
- `gopkg.in/yaml.v2` v2.4.0 - YAML config file parsing

**Infrastructure (indirect):**
- `github.com/jackc/puddle/v2` v2.2.2 - Connection pool implementation backing pgx
- `github.com/gorilla/websocket` v1.5.3 - WebSocket transport used internally by paho.mqtt.golang
- `golang.org/x/sync` v0.13.0 - Sync primitives
- `golang.org/x/crypto` v0.37.0 - Cryptography support for pgx TLS

**Previously Considered (now removed):**
- `github.com/gammazero/deque` v1.0.0 - Double-ended queue; imported in `go.mod` but commented out in `processor/processor.go`

## Configuration

**Environment:**
- Configured via YAML file (`config.yml` by default, path overridable with `-config` flag) AND environment variables
- `envconfig` processes env vars with prefix `""` (no prefix), overriding YAML values
- Key configs required:
  - `SERVER_URI` - MQTT broker URI
  - `SERVER_USER` - MQTT username
  - `SERVER_PASS` - MQTT password
  - `DATABASE_URL` - PostgreSQL connection string
- `config.yml` is gitignored (never committed); `.env` is also gitignored

**Build:**
- No build config files beyond `go.mod`/`go.sum`
- Compiled binary named `breeze` (gitignored)

## Platform Requirements

**Development:**
- Go 1.24.3+
- Access to an MQTT broker (Tomra/Spectrim fruit sizer system)
- Access to a PostgreSQL/TimescaleDB instance

**Production:**
- PostgreSQL with TimescaleDB extension (schema uses `tsdb.hypertable`, columnstore policies, and TimescaleDB-specific DDL)
- MQTT broker serving topics `tomra/211632/fruit` and `tomra/211632/grademap`
- Linux/macOS (standard Go deployment)

---

*Stack analysis: 2026-03-07*
