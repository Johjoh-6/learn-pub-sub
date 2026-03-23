package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	pubSubch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		log.Fatalf("could not declare and bind queue: %v", err)
	}

	ch, err := pubSubch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("could not consume from queue: %v", err)
	}
	go func() {
		defer pubSubch.Close()
		for msg := range ch {
			var data T
			err := json.Unmarshal(msg.Body, &data)
			if err != nil {
				log.Printf("could not unmarshal message: %v", err)
				continue
			}
			handler(data)
			msg.Ack(false)
		}
	}()

	return nil
}
