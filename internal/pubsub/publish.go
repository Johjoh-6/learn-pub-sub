package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	body, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	body, err := encodeGob(val)
	if err != nil {
		return err
	}
	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        body,
	})
	return nil
}

func encodeGob[T any](val T) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(val); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}
