package pubsub

import (
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	DURABLE   SimpleQueueType = "durable"
	TRANSIENT SimpleQueueType = "transient"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // SimpleQueueType is an "enum" type I made to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// new channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not open a channel: %v", err)
	}

	amqTable := amqp.Table{
		"x-dead-letter-exchange": routing.ExchangePerilDLx,
	}
	// new queue
	queue, err := ch.QueueDeclare(
		queueName,
		queueType == DURABLE,   // durable
		queueType == TRANSIENT, // auto-delete
		queueType == TRANSIENT, // exclusive
		false,                  // no-wait
		amqTable,               // args
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	if err := ch.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	); err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, queue, nil
}
