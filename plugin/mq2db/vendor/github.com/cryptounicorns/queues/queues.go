package queues

import (
	"strings"

	"github.com/corpix/loggers"

	"github.com/cryptounicorns/queues/errors"
	"github.com/cryptounicorns/queues/queue/channel"
	"github.com/cryptounicorns/queues/queue/kafka"
	"github.com/cryptounicorns/queues/queue/nsq"
	"github.com/cryptounicorns/queues/queue/websocket"
)

// Config is a configuration for Queue.
type Config struct {
	Type      string
	Kafka     kafka.Config
	Nsq       nsq.Config
	Channel   channel.Config
	Websocket websocket.Config
}

type GenericConfig struct {
	Format string
	Queue  Config
}

// New creates new Queue from Config.
func New(c Config, l loggers.Logger) (Queue, error) {
	switch strings.ToLower(c.Type) {
	case kafka.Name:
		return kafka.New(
			c.Kafka,
			prefixedLogger(kafka.Name, l),
		), nil
	case nsq.Name:
		return nsq.New(
			c.Nsq,
			prefixedLogger(nsq.Name, l),
		), nil
	case channel.Name:
		return channel.New(
			c.Channel,
			prefixedLogger(channel.Name, l),
		), nil
	case websocket.Name:
		return websocket.New(
			c.Websocket,
			prefixedLogger(websocket.Name, l),
		), nil
	default:
		return nil, errors.NewErrUnknownQueueType(c.Type)
	}
}
