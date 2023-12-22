package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

func PublishMessage(ctx context.Context, routingKey string, messageBody interface{}) error {
	body, err := json.Marshal(messageBody)
	if err != nil {
		return err
	}

	// TODO: make exchange a constant
	return env.Channel.PublishWithContext(ctx, "mediapire-exch", routingKey, false, false, amqp091.Publishing{
		ContentType: "text/plain",
		Body:        body,
	})
}
