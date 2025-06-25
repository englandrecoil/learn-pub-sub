package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	declaredCh, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("сouldn't open new channel: %v", err)
	}

	durable, autoDelete, exclusive := false, false, false
	if queueType == 0 {
		durable = true
	} else {
		autoDelete = true
		exclusive = true
	}

	table := amqp.Table{"x-dead-letter-exchange": "peril_dlx"}

	declaredQueue, err := declaredCh.QueueDeclare(queueName, durable, autoDelete, exclusive, false, table)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("сouldn't declare new queue: %v", err)
	}

	if err = declaredCh.QueueBind(declaredQueue.Name, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("сouldn't bind new queue to exchange: %v", err)
	}

	return declaredCh, declaredQueue, nil
}
