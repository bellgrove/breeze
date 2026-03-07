package processor

import (
	"encoding/json"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Processor struct {
	timer    chan bool
	queue    chan Fruit
	counter  atomic.Int32
	grademap Grademap
	gradeCh  chan []byte
}

// GradeCh returns a read-only channel that receives a copy of the raw
// grademap payload each time a grademap MQTT message is processed.
// The channel is buffered (size 1); when it is already full the payload
// is dropped and a warning is logged.
func (p *Processor) GradeCh() <-chan []byte { return p.gradeCh }

func (p *Processor) OnMessage(client mqtt.Client, msg mqtt.Message) {
	if strings.HasSuffix(msg.Topic(), "grademap") {
		if err := p.grademap.Update(msg.Payload()); err != nil {
			slog.Error("Failed to parse grademap", "err", err)
		}
		slog.Info("New grademap", "name", p.grademap.Name)
		rawPayload := make([]byte, len(msg.Payload()))
		copy(rawPayload, msg.Payload())
		select {
		case p.gradeCh <- rawPayload:
		default:
			slog.Warn("Dropped grademap update — channel full")
		}
	} else if strings.HasSuffix(msg.Topic(), "fruit") {
		var f Fruit
		if err := json.Unmarshal(msg.Payload(), &f); err != nil {
			slog.Error("Failed to parse fruit", "err", err)
			return
		}
		p.grademap.Grade(&f)

		slog.Debug("New fruit", "carrier", f.CarrierId)
		if len(f.PrimaryDefect) > 3 {
			slog.Warn("PrimaryDefect too long", "pd", f.PrimaryDefect, "fruit", f)
			f.PrimaryDefect = f.PrimaryDefect[:3]
		}

		select {
		case p.queue <- f:
			p.counter.Add(1)
			p.timer <- true
		default:
			slog.Warn("Dropped fruit", "carrier", f.CarrierId)
		}
	} else {
		slog.Warn("Unknown message", "topic", msg.Topic(), "msg", msg.Payload())
	}
}

// Next returns true if there is another row and makes the next row data
// available to Values(). When there are no more rows available or an error
// has occurred it returns false.
func (p *Processor) Next() bool {
	return p.counter.Load() > 0
}

// Values returns the values for the current row.
func (p *Processor) Values() ([]any, error) {
	f := <-p.queue
	p.counter.Add(-1)
	return f.AsRow()
}

// Err returns any error that has been encountered by the CopyFromSource. If
// this is not nil *Conn.CopyFrom will abort the copy.
func (p *Processor) Err() error {
	return nil
}

// DrainFruits reads all available items from the Processor queue and returns
// them as a []Fruit slice, resetting the counter to zero. It is the caller's
// responsibility not to call DrainFruits concurrently with Next/Values.
// Used by run() to collect fruits before deciding to write or buffer.
func DrainFruits(p *Processor) []Fruit {
	n := int(p.counter.Load())
	out := make([]Fruit, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, <-p.queue)
	}
	p.counter.Store(0)
	return out
}

const (
	maxTimeout = time.Duration(1) * time.Second
	maxItems   = 30
)

func Create(next chan bool, queueSize int) Processor {
	timer := make(chan bool, 1)

	go func() {
		defer close(timer)

		for keepGoing := true; keepGoing; {
			batch := 0
			expire := time.After(maxTimeout)
			for {
				select {
				case val, ok := <-timer:
					slog.Debug("Inc batch", "val", val, "ok", ok)
					if !ok || !val {
						keepGoing = false
						goto done
					}

					batch = batch + 1
					if batch == maxItems {
						goto done
					}

				case <-expire:
					goto done
				}
			}

		done:
			if batch > 0 {
				next <- true
			}
		}
	}()

	return Processor{
		timer:    timer,
		queue:    make(chan Fruit, queueSize),
		counter:  atomic.Int32{},
		grademap: Grademap{},
		gradeCh:  make(chan []byte, 1),
	}
}
