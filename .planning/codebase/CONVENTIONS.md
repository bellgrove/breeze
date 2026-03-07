# Coding Conventions

**Analysis Date:** 2026-03-07

## Naming Patterns

**Files:**
- `snake_case` for all `.go` files: `grademap.go`, `grademap_test.go`, `fruit.go`
- Test files co-located, suffixed with `_test`: `fruit_test.go`, `grademap_test.go`

**Types (structs):**
- `PascalCase`: `Processor`, `Fruit`, `Grademap`, `Grade`, `GmCh`, `GmPa`, `GmGr`, `GmGrCs`, `GmGrCsCr`
- Abbreviated type names used for internal/intermediate types: `GmCh` (GrademapCharacteristic), `GmPa` (GrademapPass), `GmGrCs` (GrademapGradeCriteriaSet)

**Functions and Methods:**
- `PascalCase` for exported functions and methods: `Columns()`, `Create()`, `Grade()`, `Update()`, `Check()`, `Next()`, `Values()`, `Err()`, `AsRow()`, `OnMessage()`
- `snake_case` for unexported helper functions: `cmp_crit()`, `crit_to_i()`
- Test helper functions exported with `PascalCase`: `ValidateFruit()`

**Variables:**
- `camelCase` for local variables: `batchId`, `grademap`, `next_batch` (some `snake_case` also used in `main`)
- `snake_case` used inconsistently in `breeze.go` for local vars: `next_batch`, `str_id`
- Loop variables use short names: `i1`, `i2`, `i3`, `k1`, `v1`, `v2`
- Test case table variables use short names: `tt`, `got`, `want`

**Constants:**
- `camelCase` for unexported constants: `exitCodeErr`, `exitCodeInterrupt`, `maxTimeout`, `maxItems`

**Struct Fields:**
- `PascalCase` for all exported struct fields: `SizerTime`, `CarrierId`, `SchemaVer`
- Internal JSON parsing structs use `snake_case` field names matching the JSON wire format: `Weight_g`, `Wai_g`, `Is_totalled`, `Is_sampled`

## Code Style

**Formatting:**
- Standard `gofmt` formatting (implied by Go convention; no custom formatter config present)
- No `.editorconfig` or Prettier config; rely on Go toolchain defaults

**Linting:**
- No `.golangci.yml` or lint configuration detected
- Standard Go toolchain (`go vet`) assumed

**Import Organization:**
- Standard library imports grouped first
- Third-party imports in a separate block
- Internal package imports (`github.com/bellgrove/breeze/processor`) in their own group
- Aliased imports used for naming clarity: `mqtt "github.com/eclipse/paho.mqtt.golang"`

Example from `breeze.go`:
```go
import (
    "context"
    "crypto/rand"
    "encoding/base64"
    ...

    "github.com/bellgrove/breeze/processor"
    mqtt "github.com/eclipse/paho.mqtt.golang"
    "github.com/jackc/pgx/v5"
    ...
)
```

## Error Handling

**Patterns:**
- Standard Go `if err != nil` pattern used throughout
- Errors returned as the last return value where applicable
- Some errors logged with `slog` and then ignored (no return): in `processor.go` `json.Unmarshal` return value silently discarded in `OnMessage`
- In `breeze.go`'s `run()` function, some errors call `os.Exit(1)` directly rather than returning: inconsistent with the function's `error` return type
- `fmt.Errorf` with `%w` used for error wrapping in `grademap.go`: `fmt.Errorf("unable to parse grademap: %w", err)`
- Errors accumulated as a single `error` value using string concatenation in `ValidateFruit()` — not idiomatic; should use `errors.Join` or similar

Example from `grademap.go`:
```go
if err := json.Unmarshal(b, &gm); err != nil {
    return fmt.Errorf("unable to parse grademap: %w", err)
}
```

## Logging

**Framework:** `log/slog` (Go standard library structured logger, Go 1.21+)

**Patterns:**
- `slog.Info("message", "key", value)` — structured key/value pairs as variadic args
- `slog.Debug(...)` for high-frequency/detailed events (fruit processing, batch incrementing)
- `slog.Warn(...)` for recoverable anomalies: connection loss, unexpected data
- `slog.Error(...)` for failures: write failures, comparison failures
- Log level `slog.LevelDebug` is commented out in `main()` — debug logging is disabled by default
- No log caller/source enrichment configured

## Comments

**When to Comment:**
- Doc comments on exported interface methods: `Next()`, `Values()`, `Err()` in `processor.go` have GoDoc-style comments explaining the `pgx.CopyFromSource` contract
- Inline comments used to explain non-obvious logic (goto usage, timer goroutine, grade index math)
- Commented-out code left in source: several blocks of commented alternatives and dead code throughout all files

**Example GoDoc comment:**
```go
// Next returns true if there is another row and makes the next row data
// available to Values(). When there are no more rows available or an error
// has occurred it returns false.
func (p *Processor) Next() bool {
```

## Function Design

**Size:** Functions are medium-sized; `UnmarshalJSON` in `fruit.go` is large (~160 lines) due to field-by-field mapping from wire format to domain struct.

**Parameters:** Methods use pointer receivers (`*Processor`, `*Fruit`, `*Grademap`, `*GmGrCsCr`) consistently for mutation. Value receivers not used.

**Return Values:**
- Errors returned as last value following Go convention
- Some functions return `([]any, error)` for the `pgx.CopyFromSource` interface

## Module Design

**Packages:**
- `main` package: `breeze.go` — entry point, config, MQTT wiring, DB write loop
- `processor` package: `processor/` — domain logic for fruit processing, grading, JSON parsing

**Exports:**
- Types and functions needed across packages exported with `PascalCase`
- Internal helpers unexported (`cmp_crit`, `crit_to_i`)

**Barrel Files:** Not used (Go does not use barrel/index files).

**Struct Initialization:**
- Positional struct initialization used in `Create()`:
  ```go
  return Processor{
      timer,
      make(chan Fruit, 50),
      atomic.Int32{},
      Grademap{},
  }
  ```
  This is fragile — adding fields silently reorders values.

---

*Convention analysis: 2026-03-07*
