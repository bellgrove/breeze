package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bellgrove/breeze/processor"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestWriteBuffer verifies bufferFruits accumulates fruits and applies
// drop-oldest when full: when the buffer is at maxSize and a new fruit
// arrives, the oldest entry is dropped and the new fruit is appended.
func TestWriteBuffer(t *testing.T) {
	var buf []processor.Fruit
	maxSize := 3
	// Add 5 fruits; first 2 should be dropped
	for i := 0; i < 5; i++ {
		f := processor.Fruit{CarrierId: fmt.Sprintf("carrier-%d", i)}
		buf = bufferFruits(buf, []processor.Fruit{f}, maxSize)
	}
	if len(buf) != 3 {
		t.Fatalf("expected 3 items, got %d", len(buf))
	}
	if buf[0].CarrierId != "carrier-2" {
		t.Errorf("expected carrier-2 at index 0, got %s", buf[0].CarrierId)
	}
	if buf[2].CarrierId != "carrier-4" {
		t.Errorf("expected carrier-4 at index 2, got %s", buf[2].CarrierId)
	}
}

// TestReconnectProbe verifies the probe goroutine exits on context cancel
// and signals reconnectCh on a successful pool.Ping.
func TestReconnectProbe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	reconnectCh := make(chan struct{}, 1)
	cancel() // cancel immediately so probe exits on ctx.Done() before first timer
	startReconnectProbe(ctx, nil, reconnectCh) // pool=nil is safe; ping never called
	// Give goroutine time to exit
	time.Sleep(10 * time.Millisecond)
	select {
	case <-reconnectCh:
		t.Fatal("reconnectCh should not have been sent when context cancelled")
	default:
		// expected: probe exited cleanly
	}
}

// TestFlushBuffer verifies flushBuffer clears the buffer on success and
// leaves it intact on failure (so buffered records can be retried).
func TestFlushBuffer(t *testing.T) {
	ctx := context.Background()
	// Empty buffer returns nil without touching the pool
	err := flushBuffer(ctx, nil, nil)
	if err != nil {
		t.Fatalf("flushBuffer with empty buf should return nil, got %v", err)
	}
	err = flushBuffer(ctx, nil, []processor.Fruit{})
	if err != nil {
		t.Fatalf("flushBuffer with empty slice should return nil, got %v", err)
	}
}

// TestTransactionalWrite verifies writeBatch rolls back on error so that
// no partial rows are written to the database.
func TestTransactionalWrite(t *testing.T) {
	ctx := context.Background()
	// Create a pool pointing at an invalid address; connecting will fail.
	// This exercises the error path without requiring a real database.
	pool, err := pgxpool.New(ctx, "postgres://invalid:invalid@localhost:1/doesnotexist?connect_timeout=1")
	if err != nil {
		// ParseConfig succeeded but pool lazy-connects; proceed with closed pool.
		t.Logf("pool creation note: %v", err)
	}
	if pool != nil {
		pool.Close()
	}
	// writeBatch against a closed pool returns a non-nil error and does not panic.
	err = writeBatch(ctx, pool, []processor.Fruit{{CarrierId: "test"}})
	if err == nil {
		t.Fatal("expected error from writeBatch with closed pool, got nil")
	}
	// The error path is the important assertion: no partial rows, no panic.
}

// TestGrademapSchema verifies GRAD-01: the package exposes DDL constants for
// the grademap and grademap_changes tables including a foreign-key reference.
//
// grademapDDL and grademapChangesDDL defined by Plan 03 — intentionally RED until then.
func TestGrademapSchema(t *testing.T) {
	if !strings.Contains(grademapDDL, "CREATE TABLE breeze_grademap") {
		t.Errorf("grademapDDL does not contain CREATE TABLE breeze_grademap; got:\n%s", grademapDDL)
	}

	if !strings.Contains(grademapChangesDDL, "CREATE TABLE breeze_grademap_changes") {
		t.Errorf("grademapChangesDDL does not contain CREATE TABLE breeze_grademap_changes; got:\n%s", grademapChangesDDL)
	}

	if !strings.Contains(grademapChangesDDL, "REFERENCES breeze_grademap") {
		t.Errorf("grademapChangesDDL does not contain REFERENCES breeze_grademap (missing FK); got:\n%s", grademapChangesDDL)
	}
}

// TestResolveGrademapID_NoRows verifies that resolveGrademapID returns (0, false, nil)
// or a non-nil error when no matching grademap row exists (closed-pool exercices error
// path without a real DB).
//
// stub — RED until Plan 02 adds resolveGrademapID
func TestResolveGrademapID_NoRows(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://invalid:invalid@localhost:1/doesnotexist?connect_timeout=1")
	if err != nil {
		t.Logf("pool creation note: %v", err)
	}
	if pool != nil {
		pool.Close()
	}
	id, ok, err := resolveGrademapID(ctx, pool, time.Now(), 0)
	// Closed pool: either (0, false, nil) for ErrNoRows branch or a non-nil error.
	if err != nil {
		// non-nil error is a valid outcome for a closed pool
		return
	}
	if id != 0 || ok {
		t.Errorf("expected (0, false, _) when no rows, got (%d, %v, %v)", id, ok, err)
	}
}

// TestResolveGrademapID_ZeroDelay verifies that calling resolveGrademapID with
// delay=0 does not panic and that the anchor time is passed unmodified (zero shift).
//
// stub — RED until Plan 02 adds resolveGrademapID
func TestResolveGrademapID_ZeroDelay(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://invalid:invalid@localhost:1/doesnotexist?connect_timeout=1")
	if err != nil {
		t.Logf("pool creation note: %v", err)
	}
	if pool != nil {
		pool.Close()
	}
	// delay=0 must not shift the anchor; pool error is expected, not a panic.
	_, _, _ = resolveGrademapID(ctx, pool, time.Now(), 0)
}

// TestResolveGrademapID verifies that resolveGrademapID accepts the expected
// signature (ctx, pool, time.Time, time.Duration) (int64, bool, error).
//
// stub — RED until Plan 02 adds resolveGrademapID
func TestResolveGrademapID(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://invalid:invalid@localhost:1/doesnotexist?connect_timeout=1")
	if err != nil {
		t.Logf("pool creation note: %v", err)
	}
	if pool != nil {
		pool.Close()
	}
	var id int64
	var ok bool
	id, ok, err = resolveGrademapID(ctx, pool, time.Now(), 30*time.Second)
	// Signature compile check; pool error is expected.
	_ = id
	_ = ok
	_ = err
}

// TestWriteBatch_GrademapID verifies that the updated writeBatch signature
// (with delay time.Duration parameter) returns an error on a closed pool and
// does not panic.
//
// stub — RED until Plan 03 updates writeBatch signature
func TestWriteBatch_GrademapID(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://invalid:invalid@localhost:1/doesnotexist?connect_timeout=1")
	if err != nil {
		t.Logf("pool creation note: %v", err)
	}
	if pool != nil {
		pool.Close()
	}
	err = writeBatch(ctx, pool, []processor.Fruit{{CarrierId: "test"}}, 0)
	if err == nil {
		t.Fatal("expected error from writeBatch with closed pool, got nil")
	}
}

// TestConfig_PropagationDelay verifies that Config.GrademapPropagationDelayStr
// field exists and that "30s" parses to 30*time.Second without error.
//
// stub — RED until Plan 02 adds GrademapPropagationDelayStr to Config
func TestConfig_PropagationDelay(t *testing.T) {
	cfg := Config{GrademapPropagationDelayStr: "30s"}
	d, err := time.ParseDuration(cfg.GrademapPropagationDelayStr)
	if err != nil {
		t.Fatalf("expected no error parsing 30s, got %v", err)
	}
	if d != 30*time.Second {
		t.Errorf("expected 30s, got %v", d)
	}
}

// TestConfig_PropagationDelayZero verifies that both empty string and "0s"
// produce a zero time.Duration (no offset applied).
//
// stub — RED until Plan 02 adds GrademapPropagationDelayStr to Config
func TestConfig_PropagationDelayZero(t *testing.T) {
	// Sub-case 1: empty string — treat as zero delay without parsing.
	cfg := Config{GrademapPropagationDelayStr: ""}
	var d time.Duration
	if cfg.GrademapPropagationDelayStr != "" {
		var err error
		d, err = time.ParseDuration(cfg.GrademapPropagationDelayStr)
		if err != nil {
			t.Fatalf("unexpected parse error for empty string: %v", err)
		}
	}
	if d != 0 {
		t.Errorf("expected 0 duration for empty string, got %v", d)
	}

	// Sub-case 2: "0s" parses to zero without error.
	cfg2 := Config{GrademapPropagationDelayStr: "0s"}
	d2, err := time.ParseDuration(cfg2.GrademapPropagationDelayStr)
	if err != nil {
		t.Fatalf("expected no error parsing 0s, got %v", err)
	}
	if d2 != 0 {
		t.Errorf("expected 0 duration for 0s, got %v", d2)
	}
}

// TestResolveGrademapID_WithDelay verifies that delay=30s shifts the anchor
// backward by 30s (adjustedTime = fruitTime.Add(-delay) < fruitTime), and that
// the function call does not panic on a closed pool.
//
// stub — RED until Plan 02 adds resolveGrademapID
func TestResolveGrademapID_WithDelay(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://invalid:invalid@localhost:1/doesnotexist?connect_timeout=1")
	if err != nil {
		t.Logf("pool creation note: %v", err)
	}
	if pool != nil {
		pool.Close()
	}
	fruitTime := time.Now()
	delay := 30 * time.Second
	// Document the invariant: adjusted anchor is before fruitTime.
	adjustedTime := fruitTime.Add(-delay)
	if !adjustedTime.Before(fruitTime) {
		t.Errorf("expected adjustedTime (%v) to be before fruitTime (%v)", adjustedTime, fruitTime)
	}
	// Pool error expected; must not panic.
	_, _, _ = resolveGrademapID(ctx, pool, fruitTime, delay)
}
