# Testing Patterns

**Analysis Date:** 2026-03-07

## Test Framework

**Runner:**
- Go standard library `testing` package
- No third-party test framework (no testify, gomock, etc.)
- Config: none (uses `go test ./...`)

**Assertion Library:**
- Standard `t.Errorf` / `t.Fatalf` — no external assertion library
- Custom field-by-field struct comparator `ValidateFruit()` in `processor/fruit_test.go`
- `reflect.DeepEqual` used for slice comparison within `ValidateFruit()`
- `reflect.Value.Equal()` used for scalar field comparison

**Run Commands:**
```bash
go test ./...              # Run all tests
go test ./processor/...    # Run processor package tests only
go test -v ./...           # Verbose output
go test -run TestGrademap  # Run specific test by name pattern
go test -cover ./...       # Coverage report
```

## Test File Organization

**Location:**
- Co-located with source in the same package directory: `processor/fruit_test.go`, `processor/grademap_test.go`
- Same package name (`package processor`) — white-box testing with full access to unexported identifiers

**Naming:**
- Test files: `{source_file}_test.go`
- Test functions: `Test{TypeOrFunction}_{Method}` — e.g., `TestGrademap_Grade`, `TestGrademap_Update`, `TestFromJson`

**Structure:**
```
processor/
├── fruit.go
├── fruit_test.go       # Tests for Fruit JSON parsing
├── grademap.go
├── grademap_test.go    # Tests for Grademap.Grade and Grademap.Update
└── processor.go
```

## Test Structure

**Suite Organization:**

All tests use the table-driven test pattern with a `tests` slice of anonymous structs:

```go
func TestGrademap_Update(t *testing.T) {
    type args struct {
        b []byte
    }
    tests := []struct {
        name string
        g    Grademap
        args args
    }{
        {"GrannySmith DPC 2024", Grademap{}, args{[]byte(gsm_gm_json)}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.g.Update(tt.args.b)
            if err != nil {
                t.Errorf("FromJson() error = %v", err)
                return
            }
            if tt.name != tt.g.Name {
                t.Errorf("Update() = %#v, want %#v", tt.g.Name, tt.name)
                return
            }
        })
    }
}
```

**Patterns:**
- `t.Run(tt.name, ...)` for sub-test scoping with named cases
- Early `return` after `t.Errorf` to stop further assertions in a failing case
- `args` struct nested inside test function for input grouping
- `want`/`wantErr` naming convention for expected values
- `got` for actual computed values

## Mocking

**Framework:** None — no mocking library used.

**Patterns:**
- No mocks or interfaces used in tests; tests operate directly on concrete types
- `Grademap` initialized as zero value `Grademap{}` and populated via `Update()` in test setup
- MQTT client and database pool not tested — only the pure `processor` package logic is unit tested

**What to Mock:**
- If testing `processor.go`'s `OnMessage`, MQTT client (`mqtt.Client`, `mqtt.Message`) would need mocking — currently not tested
- Database interactions in `breeze.go` not tested at all

**What NOT to Mock:**
- JSON parsing — tested directly against real payloads embedded as string constants

## Fixtures and Factories

**Test Data:**
- Large JSON payloads stored as package-level `const` or `var` string literals in test files
- Fruit structs pre-built as `var` package-level values for reuse across tests

```go
// Constant JSON payload (fruit_test.go)
const (
    c2_fruit_json = `{"carrierId": "0201020018A17DE8", ...}` // ~full production payload
)

// Pre-built expected Fruit struct (fruit_test.go)
// Initialized with field-by-field literal values
c2_fruit := Fruit{
    SizerTime:  c2_time,
    CarrierId:  "0201020018A17DE8",
    ...
}

// Package-level vars in grademap_test.go
var c1_time, _ = time.Parse(time.RFC3339, "2025-05-20T05:52:43.615Z")
var c1_avo = Fruit{ ... }         // Pre-built avocado fixture
var c1_avo_str = "{...}"          // JSON string form of same fixture
var gsm_gm_json = "..."           // Grademap JSON fixture (large)
var avo2_gm_json = "..."          // Second grademap JSON fixture
```

**Location:**
- All fixtures embedded inline in `_test.go` files within `processor/` — no separate fixtures directory

## Coverage

**Requirements:** None enforced — no coverage thresholds configured.

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Test Types

**Unit Tests:**
- All tests are unit tests targeting the `processor` package
- Test scope: JSON deserialization (`Fruit.UnmarshalJSON`), grademap parsing (`Grademap.Update`), and grade assignment (`Grademap.Grade`)

**Integration Tests:**
- None present

**E2E Tests:**
- Not used

## Custom Validation Helpers

**ValidateFruit (fruit_test.go):**

A custom field-by-field struct comparator is used instead of `reflect.DeepEqual` on the whole struct. This is because floating-point precision differences in JSON round-trips make direct equality checks unreliable for some fields.

```go
func ValidateFruit(x Fruit, y Fruit) (err error) {
    structType := reflect.TypeOf(x)
    structVal := reflect.ValueOf(x)
    structVal2 := reflect.ValueOf(y)
    fieldNum := structVal.NumField()

    for i := range fieldNum {
        field1 := structVal.Field(i)
        field2 := structVal2.Field(i)
        fieldName := structType.Field(i).Name

        if field1.Kind() == reflect.Slice {
            // []string or []byte branches
            ...
        } else {
            if !field1.Equal(field2) {
                err = fmt.Errorf("%v%s does not match; ", err, fieldName)
            }
        }
    }
    return err
}
```

Note: `ValidateFruit` is exported (capital V), meaning it is accessible outside the package if needed, but is only used in tests. The `map[string]any` fields (`MCenterOffsets`, etc.) are not validated by this helper — they are implicitly excluded because `reflect.Value.Equal` is not called for map kinds.

## Common Patterns

**Async/Channel Testing:**
- Not currently tested; `Processor`'s channel-based goroutine (`Create()`) is not covered by tests

**Error Testing:**
```go
// wantErr pattern
if (err != nil) != tt.wantErr {
    t.Errorf("FromJson() error = %v, wantErr %v", err, tt.wantErr)
    return
}
```

**Struct Assertion:**
```go
// Custom validator instead of reflect.DeepEqual
if err := ValidateFruit(got, tt.want); err != nil {
    t.Errorf("Unmarshal = %s", err)
}
```

---

*Testing analysis: 2026-03-07*
