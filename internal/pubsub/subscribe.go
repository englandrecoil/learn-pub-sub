package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T)) error {
	queueCh, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	ch, err := queueCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for item := range ch {
			var data T
			err := json.Unmarshal(item.Body, &data)
			if err != nil {
				log.Printf("couldn't unmarshall data: %v", err)
				continue
			}
			handler(data)
			if err = item.Ack(false); err != nil {
				log.Printf("couldn't send ACK to producer: %v", err)
			}
		}
	}()

	return nil
}
