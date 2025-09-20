package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/okysetiawan/retainr/config"
	"log/slog"
)

type Client struct {
	client *nats.Conn
	logger *slog.Logger
}

func (cli *Client) Close() error {
	cli.client.Close()
	return nil
}

func (cli *Client) Subscribe(topic string, callback func(msg []byte)) error {
	_, err := cli.client.Subscribe(topic, func(m *nats.Msg) {
		callback(m.Data)
	})

	return err
}

func New(opts config.NatsConfig, logger *slog.Logger) (*Client, error) {
	options := &nats.Options{
		Servers:  opts.Servers,
		User:     opts.User,
		Password: opts.Password,
		Timeout:  opts.Timeout,
	}

	client, err := options.Connect()
	if err != nil {
		return nil, err
	}

	client.SetErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
		slog.With("error", err, "subject", subscription.Subject, "queue", subscription.Queue).Error("nats error encountered")
	})

	return &Client{
		client: client,
		logger: logger,
	}, nil
}
