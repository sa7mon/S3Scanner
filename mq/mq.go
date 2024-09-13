package mq

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func Connect(conn *amqp.Connection, queue string, threads int, threadID int) (*amqp.Channel, error) {
	log.Debugf("{thread%v} Opening channel...", threadID)
	ch, channelErr := conn.Channel()
	if channelErr != nil {
		return nil, fmt.Errorf("[Connect()] failed to open a channel: %w", channelErr)
	}
	//defer ch.Close()

	// Declare dead letter queue
	dlq, dlErr := ch.QueueDeclare(queue+"_dead", true, false, false, false, nil)
	if dlErr != nil {
		return nil, fmt.Errorf("[Connect()] failed to declare dead letter queue: %w", dlErr)
	}

	// Declare queue to consume messages from
	_, queueErr := ch.QueueDeclare(queue, true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": dlq.Name,
		},
	)
	if queueErr != nil {
		return nil, fmt.Errorf("[Connect()] failed to declare a queue: %w", queueErr)
	}

	qosErr := ch.Qos(threads, 0, false)
	if qosErr != nil {
		return nil, fmt.Errorf("[Connect()] failed to set QoS: %w", qosErr)
	}

	return ch, nil
}
