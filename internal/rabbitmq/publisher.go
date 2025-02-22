package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	maxRetries = 5
)

func PublishMessage(ctx context.Context, routingKey string, messageBody interface{}) error {
	body, err := json.Marshal(messageBody)
	if err != nil {
		return err
	}

	for i := 0; i < maxRetries; i++ {
		if env.Channel.IsClosed() {
			// last iteration
			if i+1 == maxRetries {
				log.Error().Msgf("channel was never opened, cannot send message for routing key %s", routingKey)
				return errors.New("channel was not opened to send message")
			}
			log.Debug().Msgf("Channel is closed, waiting for it to open. Retry %d out of %d", i+1, maxRetries)
			time.Sleep(time.Second * 1)

		} else {
			break
		}
	}

	// TODO: make exchange a constant
	return env.Channel.PublishWithContext(ctx, "mediapire-exch", routingKey, false, false, amqp091.Publishing{
		ContentType: "text/plain",
		Body:        body,
	})
}
