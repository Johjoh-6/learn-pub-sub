package pubsub

import (
	"bytes"
	"encoding/gob"
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
	return subscribe[T](conn, exchange, queueName, key, queueType, handler, func(data []byte) (T, error) {
		var t T
		return t, json.Unmarshal(data, &t)
	})
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	return subscribe[T](conn, exchange, queueName, key, queueType, handler, decodeGob)
}
func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) Acktype,
	unmarshaller func([]byte) (T, error),
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
			data, err := unmarshaller(msg.Body)
			if err != nil {
				log.Printf("could not unmarshal message: %v", err)
				continue
			}
			ack := handler(data)
			switch ack {
			case Ack:
				msg.Ack(false)
			case NackRequeue:
				msg.Nack(false, true)
			case NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}

func decodeGob[T any](data []byte) (T, error) {
	var val T
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&val)
	return val, err
}
