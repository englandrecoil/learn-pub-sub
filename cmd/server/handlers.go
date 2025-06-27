package main

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, queueType pubsub.SimpleQueueType, handler func(T) pubsub.AckType) error {
	queueCh, _, err := pubsub.DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	ch, err := queueCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range ch {
			reader := bytes.NewReader(msg.Body)
			dec := gob.NewDecoder(reader)
			var data T
			if err := dec.Decode(&data); err != nil {
				log.Printf("couldn't decode data:  %v", err)
			}

			AckType := handler(data)
			switch AckType {
			case pubsub.Ack:
				msg.Ack(true)
			case pubsub.NackRequeue:
				msg.Nack(false, true)
			case pubsub.NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}
