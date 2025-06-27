package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	queueCh, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	if err = queueCh.Qos(10, 10, true); err != nil {
		return err
	}
	ch, err := queueCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range ch {
			var data T
			err := json.Unmarshal(msg.Body, &data)
			if err != nil {
				log.Printf("couldn't unmarshall data: %v", err)
				continue
			}

			ackType := handler(data)
			switch ackType {
			case Ack:
				msg.Ack(true)
			case NackRequeue:
				msg.Nack(false, true)
			case NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()
	return err
}
