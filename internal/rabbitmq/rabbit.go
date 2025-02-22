package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type connectionEnv struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
	ErrorsCh   chan *amqp091.Error
}

func observeChannelError(ctx context.Context, ce *connectionEnv) {
	err, ok := <-ce.ErrorsCh
	if !ok {
		return
	}

	log.Err(err).Msg("Channel was closed")

	ce.ConnectToChan(ctx)
}

func (ce *connectionEnv) ConnectToChan(ctx context.Context) error {
	if ce.Connection == nil {
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
	}

	ch, err := env.Connection.Channel()
	if err != nil {
		return err
	}

	ce.Channel = ch

	ce.ErrorsCh = ce.Channel.NotifyClose(make(chan *amqp091.Error))

	go observeChannelError(ctx, ce)

	return initializeConsumers(ctx, ce.Channel)
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
	if env.Channel != nil {
		env.Channel.Close()
	}

	if env.Connection != nil {
		env.Connection.Close()
	}
}
