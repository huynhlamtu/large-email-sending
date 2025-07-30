package queue

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func Consume() (<-chan amqp.Delivery, error) {
	return mqChan.Consume(queueName, "", true, false, false, false, nil)
}
