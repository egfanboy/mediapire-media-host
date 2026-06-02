package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	var publishErr error
	for i := 0; i < maxRetries; i++ {
		env.mu.Lock()
		if env.closing {
			env.mu.Unlock()
			return errors.New("rabbitmq connection is closing")
		}

		if env.Channel == nil || env.Channel.IsClosed() {
			log.Debug().Msgf("RabbitMQ channel is closed, reconnecting before publishing. Retry %d out of %d", i+1, maxRetries)
			publishErr = env.connectLocked(ctx)
			if publishErr != nil {
				env.mu.Unlock()
				if err := waitForRetry(ctx); err != nil {
					return err
				}
				continue
			}
		}

		// TODO: make exchange a constant
		publishErr = env.Channel.PublishWithContext(ctx, "mediapire-exch", routingKey, false, false, amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
		env.mu.Unlock()

		if publishErr == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		log.Err(publishErr).Msgf("Failed to publish message for routing key %s. Retry %d out of %d", routingKey, i+1, maxRetries)
		if err := waitForRetry(ctx); err != nil {
			return err
		}
	}

	return fmt.Errorf("failed to publish message for routing key %s after %d attempts: %w", routingKey, maxRetries, publishErr)
}

func waitForRetry(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
		return nil
	}
}
