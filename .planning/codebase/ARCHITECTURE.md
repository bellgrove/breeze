# Architecture

**Analysis Date:** 2026-03-07

## Pattern Overview

**Overall:** Event-driven pipeline with channel-based batching

**Key Characteristics:**
- Single binary Go application with no HTTP layer
- Inbound data arrives via MQTT subscription; outbound data writes to PostgreSQL/TimescaleDB
- Concurrency managed through Go channels and goroutines — no shared mutable state except an `atomic.Int32` counter
- Batching logic runs in a dedicated goroutine that coalesces MQTT messages before triggering bulk DB writes

## Layers

**Entry / Orchestration (`breeze.go`):**
- Purpose: Program startup, config loading, signal handling, MQTT client setup, and the main event loop
- Location: `breeze.go` (package `main`)
- Contains: `Config` struct, `main()`, `run()`, `sub()`, `readFile()`, `readEnv()`
- Depends on: `processor` package, `paho.mqtt.golang`, `pgx/v5/pgxpool`, `envconfig`, `yaml.v2`
- Used by: Nothing — this is the top of the dependency graph

**Message Processing (`processor/`):**
- Purpose: Decode MQTT payloads, apply grading logic, and expose rows for PostgreSQL bulk copy
- Location: `processor/processor.go`, `processor/fruit.go`, `processor/grademap.go`
- Contains: `Processor` struct, `Fruit` struct, `Grademap` struct, batching timer goroutine
- Depends on: standard library only (`encoding/json`, `sync/atomic`, `time`, etc.)
- Used by: `breeze.go` (`main` package)

## Data Flow

**Fruit message ingestion:**

1. MQTT broker publishes to `tomra/211632/fruit` or `tomra/211632/grademap`
2. `Processor.OnMessage()` in `processor/processor.go` receives the raw payload
3. If topic ends in `grademap`: `Grademap.Update()` parses and rebuilds the in-memory grade map
4. If topic ends in `fruit`: `json.Unmarshal` via `Fruit.UnmarshalJSON()` in `processor/fruit.go` flattens the nested Sizer/Spectrim JSON into a `Fruit` struct
5. `Grademap.Grade()` applies defect-grading criteria from the current grademap to the `Fruit`, setting `PrimaryDefect`, `OtherDefects`, `PrimaryReason`, `OtherReason`
6. The `Fruit` is pushed to a buffered channel (`queue`, capacity 50); a `true` is sent on the `timer` channel
7. The batching goroutine (started in `processor.Create()`) accumulates timer signals; it fires `next_batch <- true` when either 30 items have arrived or 1 second has elapsed since the last signal
8. The main event loop in `breeze.go:run()` receives from `next_batch` and calls `pool.CopyFrom()` using `Processor` as a `pgx.CopyFromSource`
9. `Processor.Next()` / `Processor.Values()` drain the `queue` channel, calling `Fruit.AsRow()` to produce ordered `[]any` slices for `pgx` bulk copy into `breeze_fruit`

**State Management:**
- Grademap state: held in `Processor.grademap` (a `Grademap` value); mutated only from the single MQTT callback goroutine
- Queue depth: tracked by `Processor.counter` (`atomic.Int32`), incremented on enqueue, decremented on dequeue
- Context cancellation: `context.WithCancel` propagates shutdown from OS signal through `run()` to MQTT disconnect and channel close

## Key Abstractions

**`Processor` (`processor/processor.go`):**
- Purpose: Bridges MQTT message receipt and PostgreSQL bulk copy; implements the `pgx.CopyFromSource` interface (`Next()`, `Values()`, `Err()`)
- Examples: `processor/processor.go`
- Pattern: Producer/consumer with a buffered channel queue; the timer goroutine acts as a debounce/coalesce trigger

**`Fruit` (`processor/fruit.go`):**
- Purpose: Canonical in-memory representation of one piece of fruit, including its graded defects
- Examples: `processor/fruit.go`
- Pattern: Custom `UnmarshalJSON` that flattens a deeply nested vendor JSON schema (`Sizer` + `Spectrim` sub-objects) into a flat struct; JSONB blob fields are stored both as `[]byte` (for DB write) and `map[string]any` (for grading logic)

**`Grademap` (`processor/grademap.go`):**
- Purpose: Holds the active grade rule set received from the MQTT broker; applies multi-level criteria checks against `Fruit` field values to determine defect codes
- Examples: `processor/grademap.go`
- Pattern: `Update()` parses a JSON grade map into sorted, indexed slices; `Grade()` walks grades in priority order and short-circuits on the first matching defect group

**`Config` (`breeze.go`):**
- Purpose: Typed configuration loaded from YAML file with env var override
- Examples: `breeze.go` lines 21–30
- Pattern: Struct tags drive both `yaml.Decoder` and `kelseyhightower/envconfig` — file values are set first, then env vars override

## Entry Points

**`main()` (`breeze.go`):**
- Location: `breeze.go`
- Triggers: `go run .` or compiled binary execution
- Responsibilities: Parse `-config` flag, load config (file then env), set up OS signal handling with graceful/forced shutdown, call `run()`

**`run()` (`breeze.go`):**
- Location: `breeze.go:84`
- Triggers: Called by `main()` after config is ready
- Responsibilities: Create pgxpool connection, create `Processor`, configure and connect MQTT client, block on the `ctx.Done` / `next_batch` select loop

**`Processor.OnMessage()` (`processor/processor.go`):**
- Location: `processor/processor.go:23`
- Triggers: Called by paho MQTT library on each incoming message (runs in MQTT library goroutine)
- Responsibilities: Route by topic suffix, decode payload, apply grading, enqueue fruit, signal timer

## Error Handling

**Strategy:** Log-and-continue for non-fatal errors; `os.Exit` for startup failures; context cancellation for graceful shutdown

**Patterns:**
- Startup errors (DB connect, MQTT connect) call `os.Exit(1)` immediately
- MQTT message decode errors are silently swallowed (`json.Unmarshal` return value ignored in `OnMessage`)
- DB write failures are logged with `slog.Error` but do not terminate the process
- Truncation guard applied for `PrimaryDefect` field exceeding 3 characters (logged as `slog.Warn`)
- Processing time parse failures logged as `slog.Warn` in `Fruit.UnmarshalJSON`

## Cross-Cutting Concerns

**Logging:** `log/slog` (structured, stdlib) throughout all packages; log level is `Info` by default; `slog.Debug` calls are present but the `SetLogLoggerLevel(slog.LevelDebug)` line is commented out in `main()`

**Validation:** Minimal — field-length truncation for `PrimaryDefect` only; no schema validation of MQTT payloads

**Authentication:** MQTT username/password from config; PostgreSQL DSN from config (pgx parses credentials from the URL)

---

*Architecture analysis: 2026-03-07*
