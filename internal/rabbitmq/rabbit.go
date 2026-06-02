package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type connectionEnv struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
	mu         sync.Mutex
	closing    bool
}

func (ce *connectionEnv) observeClose(ctx context.Context, name string, closeCh <-chan *amqp091.Error) {
	err, ok := <-closeCh
	if !ok {
		return
	}

	if err != nil {
		log.Err(err).Msgf("RabbitMQ %s was closed", name)
	} else {
		log.Debug().Msgf("RabbitMQ %s was closed", name)
	}

	ce.reconnect(ctx)
}

func (ce *connectionEnv) reconnect(ctx context.Context) {
	for {
		ce.mu.Lock()
		if ce.closing {
			ce.mu.Unlock()
			return
		}

		err := ce.connectLocked(ctx)
		ce.mu.Unlock()

		if err != nil {
			log.Err(err).Msg("Failed to reconnect to rabbitmq")

			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		log.Info().Msg("Reconnected to rabbitmq")
		return
	}
}

func (ce *connectionEnv) connectLocked(ctx context.Context) error {
	if ce.Connection == nil || ce.Connection.IsClosed() {
		if ce.Channel != nil {
			ce.Channel.Close()
			ce.Channel = nil
		}

		rabbitCfg := app.GetApp().Rabbit
		conn, err := amqp091.DialConfig(
			fmt.Sprintf(connectionString, rabbitCfg.Username, rabbitCfg.Password, rabbitCfg.Address, rabbitCfg.Port),
			amqp091.Config{
				// Increase heartbeat timeout since some messages require I/O worker and could drop connections
				Heartbeat: 30 * time.Second,
			},
		)
		if err != nil {
			return err
		}

		ce.Connection = conn

		go ce.observeClose(ctx, "connection", ce.Connection.NotifyClose(make(chan *amqp091.Error, 1)))
	}

	if ce.Channel != nil && !ce.Channel.IsClosed() {
		return nil
	}

	ch, err := ce.Connection.Channel()
	if err != nil {
		return err
	}

	ce.Channel = ch

	go ce.observeClose(ctx, "channel", ce.Channel.NotifyClose(make(chan *amqp091.Error, 1)))

	return initializeConsumers(ctx, ce.Channel)
}

func (ce *connectionEnv) ConnectToChan(ctx context.Context) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	return ce.connectLocked(ctx)
}

const (
	defaultCreds     = "guest"
	connectionString = "amqp://%s:%s@%s:%d/"
)

var (
	env = &connectionEnv{}
)

func Setup(ctx context.Context) error {
	err := env.ConnectToChan(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to connect to rabbitmq")
		return err
	}

	return nil
}

func Cleanup() {
	env.mu.Lock()
	defer env.mu.Unlock()

	env.closing = true

	if env.Channel != nil {
		env.Channel.Close()
	}

	if env.Connection != nil {
		env.Connection.Close()
	}
}
