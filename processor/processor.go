package processor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Processor struct {
	timer   chan bool
	queue   chan Fruit
	counter atomic.Int32
	// dequeue  *deque.Deque[Fruit]
	// current  Fruit
	grademap Grademap
}

func (p *Processor) OnMessage(client mqtt.Client, msg mqtt.Message) {
	if strings.HasSuffix(msg.Topic(), "grademap") {
		slog.Info("New grademap")

		p.grademap.Update(msg.Payload())
	} else if strings.HasSuffix(msg.Topic(), "fruit") {
		var f Fruit
		json.Unmarshal(msg.Payload(), &f)
		p.grademap.Grade(&f)

		slog.Debug("New fruit", "carrier", f.CarrierId)

		p.queue <- f
		p.counter.Add(1)
		p.timer <- true
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
	if p.counter.Add(-1) >= 0 {
		f := <-p.queue
		return f.AsRow()
	} else {
		p.counter.Add(1)
		return nil, fmt.Errorf("queue is empty")
	}
}

// Err returns any error that has been encountered by the CopyFromSource. If
// this is not nil *Conn.CopyFrom will abort the copy.
func (p *Processor) Err() error {
	return nil
}

const (
	maxTimeout = time.Duration(1) * time.Second
	maxItems   = 30
)

func Create(next chan bool) Processor {
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

	// queue := new(deque.Deque[Fruit])
	// queue.SetBaseCap(100)

	return Processor{
		timer,
		make(chan Fruit, 50),
		atomic.Int32{},
		// Fruit{},
		Grademap{},
	}
}
