# Codebase Concerns

**Analysis Date:** 2026-03-07

## Tech Debt

**Hardcoded MQTT device ID in connection handler:**
- Issue: The `connectHandler` in `breeze.go` (line 40-41) hardcodes both the device prefix `"tomra/211632/fruit"` and `"tomra/211632/grademap"`. The device ID `211632` is baked into a package-level `var`, not derived from config. Subscribing to different devices requires a code change.
- Files: `breeze.go` lines 37-42
- Impact: Cannot subscribe to multiple devices or change device at runtime. Config file has no device ID field.
- Fix approach: Add a `DeviceID` field to `Config.MQTT`, construct topic strings in `run()` after config is loaded.

**Unused dependency `github.com/gammazero/deque` still in go.mod:**
- Issue: `go.mod` lists `github.com/gammazero/deque v1.0.0` as a direct dependency. All usages are commented out in `processor/processor.go` (lines 17-19, 110-112). The import is absent from compiled code.
- Files: `go.mod`, `processor/processor.go`
- Impact: Unnecessary dependency in the module graph; `go mod tidy` would remove it and may cause CI friction if ever enforced.
- Fix approach: Run `go mod tidy` to remove the unused dependency.

**Dual representation of JSONB blob fields (byte slice + map):**
- Issue: `Fruit` struct carries both `[]byte` (e.g. `CenterOffsets`) and `map[string]any` (e.g. `MCenterOffsets`) for every Spectrim blob field — 8 pairs in total. The map is populated during `UnmarshalJSON` by re-marshaling from the parsed struct, then re-unmarshaling in the grademap test manually (`grademap_test.go` lines 34-38). This double-serialization path exists because `AsRow()` sends the `[]byte` to Postgres while `Grade()` uses the map.
- Files: `processor/fruit.go`, `processor/grademap.go`, `processor/grademap_test.go`
- Impact: Memory doubles for every fruit message. The map is not populated from the stored `[]byte` fields in the normal processing path — `Grade()` uses `f.MClassifiedBlob` etc., which are only populated during live `UnmarshalJSON`. If a `Fruit` is reconstructed from Postgres, grading will silently operate on nil maps.
- Fix approach: Remove the `[]byte` fields and marshal to `[]byte` only in `AsRow()`; or store only `[]byte` and unmarshal on demand in `Grade()`.

**`run()` calls `os.Exit()` directly instead of returning errors:**
- Issue: `breeze.go` lines 88-90 and 96-98 call `os.Exit(1)` inside `run()`. The `main()` function already handles errors returned from `run()` by writing to stderr and calling `os.Exit(exitCodeErr)`. Using `os.Exit` directly inside `run()` bypasses `defer pool.Close()` and the graceful shutdown logic.
- Files: `breeze.go` lines 84-148
- Impact: Database connection pool is not cleanly closed on startup errors. Deferred cleanup is skipped. The pattern is inconsistent with the rest of `run()`'s error handling (line 137 returns without exiting).
- Fix approach: Replace `os.Exit(1)` calls inside `run()` with `return fmt.Errorf(...)`.

**`Processor.Values()` has a racy counter decrement:**
- Issue: `processor/processor.go` `Values()` (lines 54-62) decrements the atomic counter before reading from the channel, then increments it back on the else branch. `Next()` reads the counter (line 50) to decide if more rows are available. Between the decrement in `Values()` and the channel read, another goroutine calling `Next()` could observe the counter as 0 and halt iteration prematurely. The channel read and counter decrement are not atomic together.
- Files: `processor/processor.go` lines 49-62
- Impact: Possible premature termination of a `CopyFrom` batch when throughput is high, silently dropping rows.
- Fix approach: Dequeue from channel first, then decrement; or restructure to avoid using a counter separate from channel length.

**Debug log level is commented out:**
- Issue: `breeze.go` line 49 has `// slog.SetLogLoggerLevel(slog.LevelDebug)` commented out. There is no runtime mechanism (flag, env var) to enable debug logging without code modification.
- Files: `breeze.go` line 49
- Impact: Troubleshooting production issues requires a recompile.
- Fix approach: Add a `--debug` flag or `LOG_LEVEL` env var that calls `slog.SetLogLoggerLevel` conditionally.

**Commented-out `next_batch` channel close ordering:**
- Issue: In `breeze.go` `run()`, `defer close(next_batch)` is declared on line 102. The goroutine in `Create()` closes its internal `timer` channel and exits when the outer `next_batch` channel is closed. However `pool.Close()` is deferred before `next_batch` is closed (line 99 vs 102), meaning the pool closes before the batch-flushing goroutine has a chance to write any in-flight batch.
- Files: `breeze.go` lines 99-102, `processor/processor.go` lines 78-108
- Impact: Rows that are buffered in the queue at shutdown time are silently dropped; they never get written to the database.
- Fix approach: Signal the processor to drain before closing the pool; or restructure defers so the batch goroutine can complete its final write before `pool.Close()`.

## Known Bugs

**Unsafe type assertion in `UnmarshalJSON` for `ProcessingTime`:**
- Symptoms: Panic at runtime if the `"Total time for fruit processing"` key exists in the timer map but its value is not a `float64` (e.g. `null`, integer, or string from a schema change).
- Files: `processor/fruit.go` line 354
- Trigger: Receive a fruit message where the timer value is `null` or a non-float JSON value.
- Workaround: The nil guard on line 353 prevents the panic when the key is absent, but does not protect against non-`float64` typed values.

**`json.Unmarshal` errors ignored in `OnMessage` and `UnmarshalJSON`:**
- Symptoms: Silent data corruption. If a malformed JSON fruit message arrives, `json.Unmarshal` on line 29 of `processor/processor.go` returns an error that is discarded. The zero-value `Fruit` is then enqueued and written to Postgres.
- Files: `processor/processor.go` line 29, `processor/fruit.go` line 275
- Trigger: Any malformed or unexpected fruit payload on the MQTT topic.
- Workaround: None. Rows with all-zero values will appear in the database silently.

**`PrimaryDefect` truncation is a silent data mutation:**
- Symptoms: If `PrimaryDefect` from the grademap is longer than 3 characters, it is silently truncated in `OnMessage`. A warning is logged but the original value is lost and the truncated version is stored.
- Files: `processor/processor.go` lines 33-36
- Trigger: Grademap produces a defect code longer than 3 characters.
- Workaround: The `CHAR(3)` column in Postgres would reject it anyway, but the error would surface at write time rather than silently mutating data.

**`cmp_crit` returns 0 on parse error, causing non-deterministic sort:**
- Symptoms: If a criteria name does not match the expected `"X.N.name"` format, `crit_to_i` returns an error, `cmp_crit` logs it and returns 0 (equal). The sort order of criteria sets becomes undefined for malformed grademap entries, potentially grading fruit incorrectly.
- Files: `processor/grademap.go` lines 225-238
- Trigger: Grademap with a criteria name that does not contain at least 3 dot-separated segments.
- Workaround: None. Grading silently produces wrong ordering.

## Security Considerations

**MQTT credentials in YAML config file:**
- Risk: `config.yml` is excluded from git via `.gitignore`, but anyone with filesystem access to the deployment host can read plaintext credentials.
- Files: `breeze.go` lines 23-29, `Config` struct
- Current mitigation: `.gitignore` prevents accidental commit. `envconfig` supports env var overrides.
- Recommendations: Document that production deployments should use env vars (`SERVER_URI`, `SERVER_USER`, `SERVER_PASS`, `DATABASE_URL`) rather than the YAML file; consider removing YAML support entirely in favour of env-only config.

**No MQTT TLS verification config exposed:**
- Risk: `mqtt.NewClientOptions()` is used with default TLS settings. If the broker URI uses `ssl://`, certificate validation occurs, but there is no config option to specify a CA cert, client certificate, or disable verification for testing. In practice, a developer may use an `ssl://` URI with an untrusted cert and have no way to configure trust without modifying code.
- Files: `breeze.go` lines 105-118
- Current mitigation: Default Go TLS verification applies.
- Recommendations: Expose a `tls_insecure_skip_verify` or `tls_ca_cert` config option.

**Database URL contains credentials in plaintext:**
- Risk: The `DATABASE_URL` / `database.url` config value typically embeds username and password in the connection string (e.g. `postgres://user:pass@host/db`).
- Files: `breeze.go` lines 27-29
- Current mitigation: Config file is gitignored. Pool config is parsed by pgx, not logged.
- Recommendations: Prefer environment variable injection in production; avoid logging the URL.

## Performance Bottlenecks

**Queue channel capacity is fixed at 50 fruits:**
- Problem: `processor/processor.go` line 114 creates the fruit queue with a buffer of 50. The `maxItems` batch size is 30. If the MQTT broker delivers a burst larger than 50 fruits before the database write completes, `OnMessage` will block on `p.queue <- f` (line 38), which runs on the paho MQTT dispatcher goroutine. This stalls all MQTT message processing.
- Files: `processor/processor.go` lines 72-119
- Cause: Channel buffer size is hardcoded with no back-pressure signalling back to the broker.
- Improvement path: Increase buffer size, or use a non-blocking send with a drop/warn strategy.

**Regex compiled on every `Grademap.Update()` call:**
- Problem: `processor/grademap.go` line 193 calls `regexp.MustCompile(...)` inside `Update()`. While grademap updates are infrequent, the regex is re-compiled each time.
- Files: `processor/grademap.go` line 193
- Cause: No package-level variable for the compiled regex.
- Improvement path: Move `regexp.MustCompile(`\[([^\]]*)\]`)` to a package-level `var`.

## Fragile Areas

**`Grademap.Update()` uses `v1.Index` as a slice index without bounds checking:**
- Files: `processor/grademap.go` lines 184-189
- Why fragile: `g.Passes` is allocated with `len(gm.Defect_grading_passes)`, then `g.Passes[v1.Index] = v1` is used to position each pass. If the index values in the JSON are not 0-based and contiguous, this panics with an index out of range.
- Safe modification: Add a bounds check or use a map rather than a slice for `Passes`.
- Test coverage: No test for `Defect_grading_passes` index handling.

**`GmGrCsCr.Check()` uses unchecked type assertion `val.(float64)`:**
- Files: `processor/grademap.go` line 108
- Why fragile: After the nil check, the code asserts `val.(float64)` without a comma-ok pattern. If the map value is a non-float type (e.g. integer, string), this panics. The maps are populated from `json.Unmarshal` into `map[string]any`, where JSON numbers unmarshal as `float64`, but this is an implicit contract not enforced by types.
- Safe modification: Use `val, ok := val.(float64)` and handle the false case.
- Test coverage: No test covers non-float map values.

**Single `Processor` instance is value-copied, not pointer-returned:**
- Files: `processor/processor.go` line 113, `breeze.go` line 103
- Why fragile: `Create()` returns a `Processor` value (not a pointer). It contains channels and an `atomic.Int32`. The caller in `breeze.go` stores it as `var proc = processor.Create(next_batch)` and passes `&proc` to `CopyFrom`. Copying a value containing channels is safe in Go (channels are reference types), but `atomic.Int32` must not be copied after first use per the `sync/atomic` docs. The copy happens at construction time before use, which is currently safe, but this is non-obvious and easy to break.
- Safe modification: Return `*Processor` from `Create()`.
- Test coverage: No concurrency tests.

## Test Coverage Gaps

**`processor.go` — `OnMessage`, `Next`, `Values`, `Err`, `Create` are untested:**
- What's not tested: All public methods of `Processor` including the batch timer goroutine logic, the `maxItems` flush path, and the `maxTimeout` flush path.
- Files: `processor/processor.go`
- Risk: The counter/channel race described above could go undetected. Timer-driven flush may stop working after refactors.
- Priority: High

**Error paths in `UnmarshalJSON` are untested:**
- What's not tested: Behaviour when JSON fields are absent, null, or wrong type. The `ProcessingTime` type assertion panic path is not covered.
- Files: `processor/fruit.go`, `processor/fruit_test.go`
- Risk: Silent data corruption or panics on schema changes from the upstream Spectrim system.
- Priority: High

**`breeze.go` main loop has no tests:**
- What's not tested: The `run()` function, MQTT connection setup, database write path, shutdown signal handling.
- Files: `breeze.go`
- Risk: Integration-level regressions are invisible without manual testing.
- Priority: Medium

**Grademap grading has only one test case (avocado class 1):**
- What's not tested: Pink Lady grading is commented out (`grademap_test.go` line 21). Only one fruit/grademap combination is exercised. Edge cases like empty grades list, unknown categories, or `Special` category (which falls through without setting `val`) are not covered.
- Files: `processor/grademap_test.go`, `processor/grademap.go`
- Risk: Grading logic for other fruit varieties may be silently broken.
- Priority: Medium

## Schema Concerns

**`schema.sql` contains orphaned template SQL at the top:**
- Issue: Lines 1-5 of `schema.sql` are Go template syntax (`{{ .table }}`, `{{ .columns }}`) that belong to a different (now absent) templating system. They are not valid SQL and cannot be applied directly.
- Files: `schema.sql` lines 1-5
- Impact: Running `schema.sql` against Postgres will fail on the first statement. Manual editing is required before applying.
- Fix approach: Remove the template lines or move them to a separate template file with clear documentation.

**No database migration tooling:**
- Issue: There is no migration framework (e.g. goose, golang-migrate). `schema.sql` is the only schema artifact. Schema evolution requires manual SQL application with no version tracking.
- Files: `schema.sql`
- Impact: Deployments to existing databases require manual intervention; there is no rollback mechanism.
- Fix approach: Introduce a migration tool and version the schema.

---

*Concerns audit: 2026-03-07*
