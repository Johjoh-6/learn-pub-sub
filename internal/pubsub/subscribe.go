package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
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
			ack := handler(data)
			switch ack {
			case Ack:
				log.Println("ack ")
				msg.Ack(false)
			case NackRequeue:
				log.Println("nackrequeue: ")
				msg.Nack(false, true)
			case NackDiscard:
				log.Println("nackdiscard:")
				msg.Nack(false, false)
			}
		}
	}()

	return nil
}
