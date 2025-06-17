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

	"github.com/bellgrove/breeze/processor"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	MQTT struct {
		URI  string `yaml:"uri" envconfig:"SERVER_URI"`
		User string `yaml:"user" envconfig:"SERVER_USER"`
		Pass string `yaml:"pass" envconfig:"SERVER_PASS"`
	} `yaml:"mqtt"`
	Database struct {
		URL string `yaml:"url" envconfig:"DATABASE_URL"`
	} `yaml:"database"`
}

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	slog.Info("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	slog.Warn("Connection lost", "err", err)
}

func main() {
	// slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	var cfg Config

	path := flag.String("config", "config.yml", "path to the config file")
	flag.Parse()

	readFile(path, &cfg)
	readEnv(&cfg)

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
	// conn, err := pgx.Connect(ctx, cfg.Database.URL)
	config, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse db config: %v\n", err)
		os.Exit(1)
	}

	// config.MaxConnIdleTime =
	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	next_batch := make(chan bool)
	defer close(next_batch)
	var proc = processor.Create(next_batch)

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

	sub(client, "tomra/211632/fruit")
	sub(client, "tomra/211632/grademap")

	slog.Info("Hello, World!")

	for {
		select {
		case <-ctx.Done():
			client.Disconnect(250)
			return nil
		case <-next_batch:
			// slog.Info("New write")
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
			// default:
			// 	// do a piece of work
			// 	time.Sleep(100 * time.Millisecond)
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
