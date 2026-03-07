# Codebase Structure

**Analysis Date:** 2026-03-07

## Directory Layout

```
breeze/                     # Module root: github.com/bellgrove/breeze
├── breeze.go               # package main — entry point, config, MQTT client, DB write loop
├── go.mod                  # Module definition and direct dependencies
├── go.sum                  # Dependency checksums
├── schema.sql              # TimescaleDB DDL for breeze_fruit table
├── .gitignore              # Excludes binaries, config.yml, .env, go.work
├── .vscode/
│   └── launch.json         # VS Code debug launch configuration
├── .planning/
│   └── codebase/           # GSD mapping documents (this directory)
└── processor/              # package processor — message decoding, grading, DB row production
    ├── processor.go        # Processor struct, OnMessage, Next/Values/Err, Create, batching goroutine
    ├── fruit.go            # Fruit struct, UnmarshalJSON, Columns(), AsRow()
    ├── grademap.go         # Grademap struct, Update(), Grade(), criteria check logic
    ├── fruit_test.go       # Tests for Fruit JSON unmarshalling and row mapping
    └── grademap_test.go    # Tests for Grademap parsing and grading logic
```

## Directory Purposes

**Root (`/`):**
- Purpose: `package main` entry point and project-level config files
- Contains: Single Go source file (`breeze.go`), module files, database schema
- Key files: `breeze.go`, `go.mod`, `schema.sql`

**`processor/`:**
- Purpose: All domain logic isolated from infrastructure wiring
- Contains: Data model (`Fruit`), grade rule engine (`Grademap`), pipeline coordinator (`Processor`), and their tests
- Key files: `processor/processor.go`, `processor/fruit.go`, `processor/grademap.go`

**`.vscode/`:**
- Purpose: Editor-specific debug configuration
- Contains: `launch.json` only
- Not committed to CI, used for local development

**`.planning/codebase/`:**
- Purpose: GSD codebase mapping documents consumed by `/gsd:plan-phase` and `/gsd:execute-phase`
- Generated: Yes (by GSD tooling)
- Committed: Yes

## Key File Locations

**Entry Points:**
- `breeze.go`: `main()` and `run()` — program startup, MQTT client, DB write loop

**Configuration:**
- `go.mod`: Module path `github.com/bellgrove/breeze`, Go version `1.24.3`, dependency declarations
- `schema.sql`: TimescaleDB DDL — `CREATE TABLE breeze_fruit`, hypertable config, columnstore policy
- `config.yml` (gitignored): Runtime YAML config for MQTT URI/credentials and database URL

**Core Logic:**
- `processor/fruit.go`: `Fruit` struct with custom JSON unmarshaller; `Columns()` and `AsRow()` for DB copy
- `processor/grademap.go`: Grade rule engine — `Grademap.Update()` and `Grademap.Grade()`
- `processor/processor.go`: `Processor` (implements `pgx.CopyFromSource`); batching goroutine in `Create()`

**Testing:**
- `processor/fruit_test.go`: Unit tests for `Fruit` parsing (23 KB)
- `processor/grademap_test.go`: Unit tests for grademap parsing and grading (125 KB, contains large fixture data)

## Naming Conventions

**Files:**
- Lowercase, no separators: `fruit.go`, `grademap.go`, `processor.go`
- Test files follow Go convention: `<name>_test.go`

**Directories:**
- Lowercase, single word: `processor/`

**Types:**
- PascalCase: `Fruit`, `Grademap`, `Processor`, `Config`, `Grade`, `GmCh`, `GmPa`, `GmGr`, `GmGrCs`, `GmGrCsCr`
- Abbreviated grademap sub-types use `Gm` prefix: `GmCh` (characteristic), `GmPa` (pass), `GmGr` (grade), `GmGrCs` (criteria set), `GmGrCsCr` (criteria)

**Functions / Methods:**
- Exported: PascalCase — `OnMessage`, `Create`, `Columns`, `AsRow`, `Grade`, `Update`
- Unexported: camelCase or snake_case — `readFile`, `readEnv`, `sub`, `cmp_crit`, `crit_to_i`

**Variables:**
- Local: camelCase (`pool`, `proc`, `next_batch`, `str_id`)
- Package-level handlers: camelCase (`connectHandler`, `connectLostHandler`)

**Struct fields from external JSON:**
- Vendor API fields use snake_case to match JSON keys: `Weight_g`, `Is_totalled`, `Defect_grading_passes`

## Where to Add New Code

**New MQTT topic handler:**
- Add topic subscription in `connectHandler` in `breeze.go`
- Add routing branch in `Processor.OnMessage()` in `processor/processor.go`

**New fruit field (from vendor payload):**
- Add field to `Fruit` struct in `processor/fruit.go`
- Add mapping in `Fruit.UnmarshalJSON()` in `processor/fruit.go`
- Add column name to `Columns()` slice in `processor/fruit.go`
- Add value to `AsRow()` return slice in `processor/fruit.go` (order must match `Columns()`)
- Add column to `CREATE TABLE breeze_fruit` in `schema.sql`

**New grading characteristic category:**
- Add `case` branch in `GmGrCsCr.Check()` in `processor/grademap.go`

**New utility / helper:**
- If specific to a domain type, add as a method on that type in the relevant file in `processor/`
- If shared infrastructure, add to `breeze.go` (package `main`) or a new file in `processor/`

**New test:**
- Place in `processor/` alongside the file under test, following `<name>_test.go` naming

## Special Directories

**`.git/`:**
- Purpose: Git repository metadata
- Generated: Yes
- Committed: N/A

**`.planning/`:**
- Purpose: GSD planning and codebase analysis documents
- Generated: Yes (by GSD commands)
- Committed: Yes

---

*Structure analysis: 2026-03-07*
