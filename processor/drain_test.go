package processor

import (
	"testing"
)

func TestDrainFruits(t *testing.T) {
	// Create a processor using a buffered next channel so batch signal does not block
	next := make(chan bool, 1)
	p := Create(next, 10)

	// Enqueue two Fruit values directly into p.queue and increment counter for each
	f1 := Fruit{CarrierId: "A001"}
	f2 := Fruit{CarrierId: "B002"}
	f3 := Fruit{CarrierId: "C003"}

	p.queue <- f1
	p.counter.Add(1)
	p.queue <- f2
	p.counter.Add(1)
	p.queue <- f3
	p.counter.Add(1)

	// Call DrainFruits
	result := DrainFruits(&p)

	// Assert: returned slice has the correct length
	if len(result) != 3 {
		t.Fatalf("expected 3 fruits, got %d", len(result))
	}

	// Assert: items are returned in FIFO order
	if result[0].CarrierId != "A001" {
		t.Errorf("expected first CarrierId A001, got %s", result[0].CarrierId)
	}
	if result[1].CarrierId != "B002" {
		t.Errorf("expected second CarrierId B002, got %s", result[1].CarrierId)
	}
	if result[2].CarrierId != "C003" {
		t.Errorf("expected third CarrierId C003, got %s", result[2].CarrierId)
	}

	// Assert: counter is zero after drain
	if p.counter.Load() != 0 {
		t.Errorf("expected counter 0 after drain, got %d", p.counter.Load())
	}

	// Assert: a second call to DrainFruits returns an empty (non-nil) slice
	second := DrainFruits(&p)
	if second == nil {
		t.Error("expected non-nil slice on second call, got nil")
	}
	if len(second) != 0 {
		t.Errorf("expected empty slice on second call, got %d items", len(second))
	}

}
