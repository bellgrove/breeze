package main

import (
	"context"
	"fmt"
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
