package rabbitmq

import (
	"context"
	"fmt"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/internal/utils"
	"github.com/rs/zerolog/log"

	"github.com/rabbitmq/amqp091-go"
)

type consumerHandler func(ctx context.Context, msg amqp091.Delivery) error

var consummerMapping = utils.NewConcurrentMap[string, consumerHandler]()

func RegisterConsumer(h consumerHandler, routingKey string) {
	consummerMapping.Add(routingKey, h)
}

func initializeConsumers(ctx context.Context, channel *amqp091.Channel) error {
	appInstance := app.GetApp()
	q, err := env.Channel.QueueDeclare(
		fmt.Sprintf("mediapire-mediahost-%s", appInstance.Name), // name
		true,  // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	for routingKey := range consummerMapping.Get() {
		log.Debug().Msgf("Setting up consumer for routing key %s", routingKey)

		err = env.Channel.QueueBind(
			q.Name,     // queue name
			routingKey, // routing key
			// TODO: make exchange a constant
			"mediapire-exch", // exchange
			false,
			nil)
		if err != nil {
			return err
		}
	}

	msgs, err := env.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			log.Debug().Msgf("Handling message for routing key %s", msg.RoutingKey)
			msg.Ack(false)

			if handler, ok := consummerMapping.GetKey(msg.RoutingKey); !ok {
				log.Debug().Msgf("No handler registered for routing key %s. Message acknowledge but no action taken", msg.RoutingKey)
			} else {
				go handler(context.Background(), msg)
			}
		}
	}()

	return nil
}
