package processor

import (
	"bytes"
	"testing"
	"time"
)

// mockMsg implements mqtt.Message for testing.
type mockMsg struct {
	topic   string
	payload []byte
}

func (m mockMsg) Topic() string     { return m.topic }
func (m mockMsg) Payload() []byte   { return m.payload }
func (m mockMsg) Qos() byte         { return 0 }
func (m mockMsg) Retained() bool    { return false }
func (m mockMsg) MessageID() uint16 { return 0 }
func (m mockMsg) Ack()              {}
func (m mockMsg) Duplicate() bool   { return false }

// newTestProcessorSmallQueue creates a Processor with a fruit queue of the
// given capacity. It bypasses Create() because Create() hard-codes queue
// size 50 and starts a goroutine that drives the next channel; for these
// unit tests we only care about the queue and counter behaviour.
//
// The timer channel is sized generously so that OnMessage's `p.timer <- true`
// send does not block during the test.
func newTestProcessorSmallQueue(queueSize int) Processor {
	// timer must be buffered so OnMessage's send to it does not block.
	timer := make(chan bool, queueSize+10)
	return Processor{
		timer:    timer,
		queue:    make(chan Fruit, queueSize),
		grademap: Grademap{},
		gradeCh:  make(chan []byte, 1),
	}
}

// newTestProcessorFull creates a Processor with a pre-filled queue of
// capacity 1, ready for the QueueFull test.
func fruitTopic() string { return "line1/fruit" }

// validFruitPayload returns an empty-object JSON.  UnmarshalJSON will
// produce a zero-value Fruit (unmarshal succeeds).
var validFruitPayload = []byte(`{}`)

// malformedFruitPayload cannot be decoded as JSON.
var malformedFruitPayload = []byte(`not valid json {{{`)

// TestOnMessage_QueueFull verifies QUAL-01: when the fruit queue is full,
// OnMessage must NOT block the MQTT dispatcher.
//
// Current behaviour (buggy): OnMessage does a blocking send `p.queue <- f`,
// which deadlocks when the queue is full.  The test detects this by running
// the second OnMessage call in a goroutine and timing out after 200 ms.  If
// the goroutine has not returned by then the test fails RED, documenting the
// expected non-blocking behaviour.
func TestOnMessage_QueueFull(t *testing.T) {
	p := newTestProcessorSmallQueue(1)

	// Fill the queue with the first fruit.
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		p.OnMessage(nil, mockMsg{topic: fruitTopic(), payload: validFruitPayload})
	}()
	select {
	case <-firstDone:
		// first call completed — queue now has 1 item
	case <-time.After(500 * time.Millisecond):
		t.Fatal("first OnMessage call blocked unexpectedly")
	}

	// Attempt a second OnMessage; with QUAL-01 unfixed this will block.
	secondDone := make(chan struct{})
	go func() {
		defer close(secondDone)
		p.OnMessage(nil, mockMsg{topic: fruitTopic(), payload: validFruitPayload})
	}()

	select {
	case <-secondDone:
		// Non-blocking: check queue length stayed at 1 (second fruit was dropped).
		if got := p.counter.Load(); got != 1 {
			t.Errorf("OnMessage_QueueFull: counter = %d, want 1 (second fruit must be dropped)", got)
		}
	case <-time.After(200 * time.Millisecond):
		// QUAL-01 is not fixed: OnMessage blocked on the second send.
		t.Error("OnMessage blocked when queue was full — QUAL-01 not yet fixed (blocking send detected)")
	}
}

// TestValues_CounterOrder verifies QUAL-02: the fruit counter must accurately
// reflect the number of items in the queue at all observable points.
//
// Current behaviour (buggy): Values() decrements the counter BEFORE reading
// from the queue (`p.counter.Add(-1)` then `<-p.queue`).  This creates a
// window where counter is 0 but an item still sits in the queue — a
// concurrent caller of Next() would return false even though data exists.
//
// This test detects the ordering bug by:
//  1. Enqueueing one fruit (counter = 1).
//  2. Starting a goroutine that spins calling Next() and recording the first
//     false it sees, then immediately checking whether an item is actually
//     present in the queue.
//  3. Concurrently calling Values() to trigger the window.
//
// With the current buggy ordering, the monitor goroutine can observe
// counter == 0 while the queue still holds the fruit.  We treat any
// observation of Next() == false concurrent with a non-empty queue as a
// QUAL-02 failure.  If the fix is in place, Next() never returns false until
// the queue is genuinely empty.
func TestValues_CounterOrder(t *testing.T) {
	p := newTestProcessorSmallQueue(10)

	// Enqueue one fruit.
	enqDone := make(chan struct{})
	go func() {
		defer close(enqDone)
		p.OnMessage(nil, mockMsg{topic: fruitTopic(), payload: validFruitPayload})
	}()
	select {
	case <-enqDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("OnMessage blocked during setup")
	}
	if got := p.counter.Load(); got != 1 {
		t.Fatalf("setup: counter = %d, want 1", got)
	}

	// Monitor for the QUAL-02 ordering bug: Next() returns false while queue
	// is non-empty.
	bugDetected := make(chan struct{})
	monitorDone := make(chan struct{})
	go func() {
		defer close(monitorDone)
		deadline := time.Now().Add(500 * time.Millisecond)
		for time.Now().Before(deadline) {
			if !p.Next() && len(p.queue) > 0 {
				// Counter said "no items" but queue is non-empty — ordering bug.
				close(bugDetected)
				return
			}
		}
	}()

	// Trigger Values() to create the window.
	row, err := p.Values()
	<-monitorDone

	// Basic correctness checks on Values() itself.
	if err != nil {
		t.Errorf("Values() returned unexpected error: %v", err)
	}
	if row == nil {
		t.Error("Values() returned nil row, want non-nil")
	}
	if got := p.counter.Load(); got != 0 {
		t.Errorf("counter after Values() = %d, want 0", got)
	}

	// If the monitor detected the ordering bug, the test fails RED.
	select {
	case <-bugDetected:
		t.Error("QUAL-02 ordering bug detected: Next() returned false while queue was non-empty — counter is decremented before the queue read")
	default:
		// No bug observed — either the fix is in place or the race window was
		// not hit.  Accept this as a pass for now; Plan 02's fix ensures the
		// window cannot exist at all.
	}
}

// TestOnMessage_MalformedJSON verifies QUAL-03: a malformed JSON payload must
// be discarded; no zero-value fruit should be enqueued.
//
// Current behaviour (buggy): OnMessage ignores the error from json.Unmarshal
// and unconditionally enqueues the zero-value Fruit that results from a
// failed decode.  The test asserts counter == 0 after the call.
func TestOnMessage_MalformedJSON(t *testing.T) {
	p := newTestProcessorSmallQueue(10)

	done := make(chan struct{})
	go func() {
		defer close(done)
		p.OnMessage(nil, mockMsg{topic: fruitTopic(), payload: malformedFruitPayload})
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("OnMessage blocked with malformed payload")
	}

	// No fruit should have been enqueued.
	if got := p.counter.Load(); got != 0 {
		t.Errorf("OnMessage_MalformedJSON: counter = %d after malformed payload, want 0 — QUAL-03 not yet fixed (zero-value fruit was enqueued)", got)
	}
}

// TestOnMessage_GrademapChannel verifies GRAD-02: after a grademap-topic
// OnMessage call the processor exposes the raw payload via a GradeCh()
// channel, and the call returns immediately even when the channel is full.
//
// GradeCh() added by Plan 02 — this test is intentionally RED until then.
func TestOnMessage_GrademapChannel(t *testing.T) {
	p := newTestProcessorSmallQueue(5)

	payload := []byte(`{"Name":"GM1"}`)

	// OnMessage must not block; run it with a 200 ms timeout.
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.OnMessage(nil, mockMsg{topic: "tomra/211632/grademap", payload: payload})
	}()
	select {
	case <-done:
		// returned without blocking — good
	case <-time.After(200 * time.Millisecond):
		t.Fatal("OnMessage blocked on grademap topic — expected non-blocking call")
	}

	// GradeCh() must return a channel with at least one queued item.
	ch := p.GradeCh()
	if len(ch) < 1 {
		t.Fatalf("GradeCh(): expected at least 1 item, got %d", len(ch))
	}

	received := <-ch
	if !bytes.Equal(received, payload) {
		t.Errorf("GradeCh() item = %q, want %q", received, payload)
	}

	// Verify that OnMessage does not block when the gradeCh is already full
	// (buffer=1, occupied). Fill it first then send another grademap message.
	p2 := newTestProcessorSmallQueue(5)
	// Pre-fill: send one grademap to occupy the buffer.
	p2.OnMessage(nil, mockMsg{topic: "tomra/211632/grademap", payload: payload})

	// Second send — must not block even though channel is full.
	done2 := make(chan struct{})
	go func() {
		defer close(done2)
		p2.OnMessage(nil, mockMsg{topic: "tomra/211632/grademap", payload: []byte(`{"Name":"GM2"}`)})
	}()
	select {
	case <-done2:
		// non-blocking — good
	case <-time.After(200 * time.Millisecond):
		t.Fatal("OnMessage blocked on full gradeCh — GRAD-02 not yet implemented (Plan 02)")
	}
}
