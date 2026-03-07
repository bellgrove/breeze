package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/bellgrove/breeze/processor"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	MQTT struct {
		URI  string `yaml:"uri"  envconfig:"SERVER_URI"`
		User string `yaml:"user" envconfig:"SERVER_USER"`
		Pass string `yaml:"pass" envconfig:"SERVER_PASS"`
	} `yaml:"mqtt"`
	Database struct {
		URL string `yaml:"url" envconfig:"DATABASE_URL"`
	} `yaml:"database"`
	LogLevel  string `yaml:"log_level"  envconfig:"LOG_LEVEL"`
	QueueSize int    `yaml:"queue_size" envconfig:"QUEUE_SIZE"`
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
	if cfg.QueueSize == 0 {
		cfg.QueueSize = 50
	}
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

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-next_batch:
			rows, err := pool.CopyFrom(
				ctx,
				pgx.Identifier{"breeze_fruit"},
				processor.Columns(),
				&proc)
			if err != nil {
				slog.Error("Failed to write", "err", err, "count", rows)

			} else {
				slog.Debug("Wrote rows", "count", rows)
			}
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
