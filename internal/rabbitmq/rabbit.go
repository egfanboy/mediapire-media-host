package rabbitmq

import (
	"context"
	"fmt"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/rabbitmq/amqp091-go"
)

type connectionEnv struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
}

const (
	defaultCreds     = "guest"
	connectionString = "amqp://%s:%s@%s:%d/"
)

var (
	env = connectionEnv{}
)

func Setup(ctx context.Context) error {
	rabbitCfg := app.GetApp().Rabbit
	var err error

	env.Connection, err = amqp091.Dial(fmt.Sprintf(connectionString, rabbitCfg.Username, rabbitCfg.Password, rabbitCfg.Address, rabbitCfg.Port))
	if err != nil {
		return err
	}

	env.Channel, err = env.Connection.Channel()
	if err != nil {
		return err

	}

	return initializeConsumers(ctx, env.Channel)
}

func Cleanup() {
	if env.Channel != nil {
		env.Channel.Close()
	}

	if env.Connection != nil {
		env.Connection.Close()
	}
}
