package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bellgrove/breeze/processor"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

const grademapDDL = `CREATE TABLE breeze_grademap (
    id          BIGSERIAL PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload     JSONB NOT NULL
);`

const grademapChangesDDL = `CREATE TABLE breeze_grademap_changes (
    id           BIGSERIAL PRIMARY KEY,
    grademap_id  BIGINT NOT NULL REFERENCES breeze_grademap(id),
    entity_type  TEXT NOT NULL,
    entity_name  TEXT NOT NULL,
    field        TEXT NOT NULL,
    old_value    TEXT,
    new_value    TEXT NOT NULL
);`

type Config struct {
	MQTT struct {
		URI  string `yaml:"uri"  envconfig:"SERVER_URI"`
		User string `yaml:"user" envconfig:"SERVER_USER"`
		Pass string `yaml:"pass" envconfig:"SERVER_PASS"`
	} `yaml:"mqtt"`
	Database struct {
		URL string `yaml:"url" envconfig:"DATABASE_URL"`
	} `yaml:"database"`
	LogLevel                    string `yaml:"log_level"                 envconfig:"LOG_LEVEL"`
	QueueSize                   int    `yaml:"queue_size"                envconfig:"QUEUE_SIZE"`
	WriteBufferSize             int    `yaml:"write_buffer_size"         envconfig:"WRITE_BUFFER_SIZE"`
	GrademapPropagationDelayStr string `yaml:"grademap_propagation_delay" envconfig:"GRADEMAP_PROPAGATION_DELAY"`
}

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	slog.Info("Connected")

	sub(client, "tomra/211632/fruit")
	sub(client, "tomra/211632/grademap")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	slog.Warn("Connection lost", "err", err)
}

// bufferFruits appends incoming fruits to buf, applying a drop-oldest policy
// when len(buf) >= maxSize. Logs a warning per dropped record.
func bufferFruits(buf []processor.Fruit, incoming []processor.Fruit, maxSize int) []processor.Fruit {
	for _, f := range incoming {
		if len(buf) >= maxSize {
			slog.Warn("Dropped buffered fruit", "carrier", buf[0].CarrierId)
			buf = buf[1:] // drop oldest; backing array reuse is acceptable at 10k cap
		}
		buf = append(buf, f)
	}
	return buf
}

// writeBatch writes a []processor.Fruit slice to breeze_fruit inside a transaction.
// On error the transaction is rolled back automatically by BeginTxFunc.
// delay is subtracted from the earliest SizerTime in the batch to find the grademap
// that was active when the OEM graded the fruit. A zero delay means no adjustment.
func writeBatch(ctx context.Context, pool *pgxpool.Pool, batch []processor.Fruit, delay time.Duration) error {
	var grademapID *int64
	if len(batch) > 0 {
		// Use the earliest SizerTime in the batch as the anchor (conservative).
		anchor := batch[0].SizerTime
		for _, f := range batch[1:] {
			if f.SizerTime.Before(anchor) {
				anchor = f.SizerTime
			}
		}
		id, ok, err := resolveGrademapID(ctx, pool, anchor, delay)
		if err != nil {
			slog.Warn("Failed to resolve grademap ID — writing fruit with NULL grademap_id", "err", err)
		} else if ok {
			grademapID = &id
		}
	}
	rows := make([][]any, len(batch))
	for i := range batch {
		batch[i].GrademapID = grademapID
		row, err := batch[i].AsRow()
		if err != nil {
			return fmt.Errorf("failed to convert fruit to row: %w", err)
		}
		rows[i] = row
	}
	return pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		n, err := tx.CopyFrom(ctx, pgx.Identifier{"breeze_fruit"}, processor.Columns(), pgx.CopyFromRows(rows))
		if err != nil {
			return err
		}
		slog.Debug("Wrote rows", "count", n)
		return nil
	})
}

// flushBuffer writes all buffered fruits to breeze_fruit inside a transaction.
// On failure, the caller must leave the buffer intact and retry later.
// delay is subtracted from the earliest SizerTime in the buffer to resolve the
// grademap that was active at OEM grading time. A zero delay means no adjustment.
func flushBuffer(ctx context.Context, pool *pgxpool.Pool, buf []processor.Fruit, delay time.Duration) error {
	if len(buf) == 0 {
		return nil
	}
	var grademapID *int64
	// Use the earliest SizerTime in the buffer as the anchor (conservative).
	anchor := buf[0].SizerTime
	for _, f := range buf[1:] {
		if f.SizerTime.Before(anchor) {
			anchor = f.SizerTime
		}
	}
	id, ok, err := resolveGrademapID(ctx, pool, anchor, delay)
	if err != nil {
		slog.Warn("Failed to resolve grademap ID — flushing buffer with NULL grademap_id", "err", err)
	} else if ok {
		grademapID = &id
	}
	rows := make([][]any, len(buf))
	for i := range buf {
		buf[i].GrademapID = grademapID
		row, err := buf[i].AsRow()
		if err != nil {
			return fmt.Errorf("failed to convert buffered fruit to row: %w", err)
		}
		rows[i] = row
	}
	return pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		n, err := tx.CopyFrom(ctx, pgx.Identifier{"breeze_fruit"}, processor.Columns(), pgx.CopyFromRows(rows))
		if err != nil {
			return err
		}
		slog.Info("Flushed buffer", "count", n)
		return nil
	})
}

// writeGrademap inserts a raw grademap payload into breeze_grademap and writes
// per-field change rows to breeze_grademap_changes. prev may be nil on the first
// call (all fields treated as new). prevGrademapPayload must only be updated by
// the caller on nil error return (not on failure) so the diff baseline stays accurate.
//
// Returns the parsed map for use as the next prev, and any error from the INSERT.
// A CopyFrom error for changes is logged but does not propagate (audit log, not data).
func writeGrademap(ctx context.Context, pool *pgxpool.Pool, payload []byte, prev map[string]any) (map[string]any, error) {
	var grademapID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO breeze_grademap (received_at, payload) VALUES ($1, $2) RETURNING id`,
		time.Now(),
		payload,
	).Scan(&grademapID)
	if err != nil {
		return nil, fmt.Errorf("insert breeze_grademap: %w", err)
	}

	var current map[string]any
	if jsonErr := json.Unmarshal(payload, &current); jsonErr != nil {
		slog.Error("Failed to unmarshal grademap for diff", "err", jsonErr)
		return current, nil // grademap row written; skip diff
	}

	changes := processor.DiffGrademaps(prev, current)
	if len(changes) > 0 {
		changeRows := make([][]any, len(changes))
		for i, c := range changes {
			changeRows[i] = []any{grademapID, c.EntityType, c.EntityName, c.Field, c.OldValue, c.NewValue}
		}
		if _, copyErr := pool.CopyFrom(ctx,
			pgx.Identifier{"breeze_grademap_changes"},
			[]string{"grademap_id", "entity_type", "entity_name", "field", "old_value", "new_value"},
			pgx.CopyFromRows(changeRows),
		); copyErr != nil {
			slog.Error("Failed to write grademap changes", "err", copyErr)
			// Non-fatal: grademap row was written; diff self-corrects on next update
		}
	}

	return current, nil
}

// resolveGrademapID returns the id of the most recent breeze_grademap row whose
// received_at is at or before fruitTime-delay. Returns (0, false, nil) when no
// matching row exists (pgx.ErrNoRows is not treated as an error). Returns
// (0, false, err) for any other query failure.
func resolveGrademapID(ctx context.Context, pool *pgxpool.Pool, fruitTime time.Time, delay time.Duration) (int64, bool, error) {
	adjustedTime := fruitTime.Add(-delay)
	var id int64
	err := pool.QueryRow(ctx,
		`SELECT id FROM breeze_grademap WHERE received_at <= $1 ORDER BY received_at DESC LIMIT 1`,
		adjustedTime,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("resolve grademap id: %w", err)
	}
	return id, true, nil
}

// startReconnectProbe launches a background goroutine that pings the DB with
// exponential backoff (1s initial, 2x multiplier, 60s cap, 5s ping timeout).
// On successful ping it sends to reconnectCh and exits. On ctx cancellation it exits.
func startReconnectProbe(ctx context.Context, pool *pgxpool.Pool, reconnectCh chan<- struct{}) {
	go func() {
		delay := time.Second
		const maxDelay = 60 * time.Second
		t := time.NewTimer(delay)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err := pool.Ping(pingCtx)
				cancel()
				if err != nil {
					slog.Warn("DB probe failed, retrying", "delay", delay, "err", err)
					delay = min(delay*2, maxDelay)
					t.Reset(delay)
					continue
				}
				slog.Info("DB reconnected")
				reconnectCh <- struct{}{}
				return
			}
		}
	}()
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	var cfg Config

	path := flag.String("config", "config.yml", "path to the config file")
	flag.Parse()

	readFile(path, &cfg)
	readEnv(&cfg)

	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	slog.SetLogLoggerLevel(level)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			slog.Info("Gracefully shutting down")
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		slog.Warn("Forcing shutdown")
		os.Exit(exitCodeInterrupt)
	}()
	if err := run(ctx, os.Args, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitCodeErr)
	}
}

func run(ctx context.Context, _ []string, cfg *Config) error {
	config, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse db config: %v\n", err)
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	next_batch := make(chan bool)
	defer close(next_batch)

	// Apply defaults
	if cfg.QueueSize == 0 {
		cfg.QueueSize = 50
	}
	if cfg.WriteBufferSize == 0 {
		cfg.WriteBufferSize = 10_000
	}

	var propagationDelay time.Duration
	if cfg.GrademapPropagationDelayStr != "" {
		propagationDelay, err = time.ParseDuration(cfg.GrademapPropagationDelayStr)
		if err != nil {
			return fmt.Errorf("invalid grademap_propagation_delay %q: %w", cfg.GrademapPropagationDelayStr, err)
		}
	}
	// zero value means no adjustment — correct

	var proc = processor.Create(next_batch, cfg.QueueSize)

	opts := mqtt.NewClientOptions()
	slog.Info(cfg.MQTT.URI)
	opts.AddBroker(cfg.MQTT.URI)

	token := make([]byte, 6)
	rand.Read(token)
	str_id := base64.StdEncoding.EncodeToString(token)
	opts.SetClientID("go_breeze_" + str_id)
	opts.SetUsername(cfg.MQTT.User)
	opts.SetPassword(cfg.MQTT.Pass)
	opts.SetDefaultPublishHandler(proc.OnMessage)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(250)

	slog.Info("Hello, World!")

	// Write buffer state
	var (
		writeBuffer         []processor.Fruit
		dbHealthy           = true
		probeRunning        = false
		reconnectCh         = make(chan struct{}, 1)
		prevGrademapPayload map[string]any // nil on first grademap update
	)

	for {
		select {
		case <-ctx.Done():
			// Abandon buffered records on shutdown — DB may still be down.
			// Records in writeBuffer are lost; this is acceptable per design.
			return nil

		case <-next_batch:
			batch := processor.DrainFruits(&proc)
			if !dbHealthy {
				writeBuffer = bufferFruits(writeBuffer, batch, cfg.WriteBufferSize)
				continue
			}
			if err := writeBatch(ctx, pool, batch, propagationDelay); err != nil {
				slog.Error("Failed to write batch", "err", err)
				dbHealthy = false
				writeBuffer = bufferFruits(writeBuffer, batch, cfg.WriteBufferSize)
				if !probeRunning {
					probeRunning = true
					startReconnectProbe(ctx, pool, reconnectCh)
				}
			}

		case <-reconnectCh:
			probeRunning = false
			if err := flushBuffer(ctx, pool, writeBuffer, propagationDelay); err != nil {
				slog.Error("Flush failed, re-buffering", "err", err)
				// writeBuffer is left intact — do not re-append
				probeRunning = true
				startReconnectProbe(ctx, pool, reconnectCh)
				continue
			}
			writeBuffer = writeBuffer[:0] // reset slice, reuse backing array
			dbHealthy = true

		case payload := <-proc.GradeCh():
			if !dbHealthy {
				slog.Warn("Dropped grademap update — DB unhealthy")
				continue
			}
			next, err := writeGrademap(ctx, pool, payload, prevGrademapPayload)
			if err != nil {
				slog.Error("Failed to persist grademap", "err", err)
				// Do not update prevGrademapPayload — diff self-corrects on next success
				continue
			}
			prevGrademapPayload = next
		}
	}
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 0, nil)
	token.Wait()
	slog.Info("Subscribed to topic", "topic", topic)
}

func readFile(path *string, cfg *Config) error {
	f, err := os.Open(*path)
	if err != nil {
		slog.Error("unable to open config", "file", *path, "err", err)
		return err
	}
	defer f.Close()
	slog.Info("loading config file", "file", *path)

	decoder := yaml.NewDecoder(f)
	return decoder.Decode(cfg)
}

func readEnv(cfg *Config) error {
	return envconfig.Process("", cfg)
}
