package worker

import (
	"encoding/json"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/mq"
	"github.com/sa7mon/s3scanner/provider"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

func publishBucket(ch *amqp.Channel, b bucket.Bucket) {
	bucketBytes, err := json.Marshal(b)
	if err != nil {
		FailOnError(err, "Failed to marshal bucket msg")
	}

	err = ch.Publish(
		"",
		"test",
		false,
		false,
		amqp.Publishing{Body: bucketBytes, DeliveryMode: amqp.Transient},
	)
	if err != nil {
		FailOnError(err, "Failed to publish to channel")
	}
}

func TestMqWork(t *testing.T) {
	_, testMQ := os.LookupEnv("TEST_MQ")
	if !testMQ {
		t.Skip("TEST_MQ not enabled")
	}

	aws, err := provider.NewProviderAWS()
	assert.Nil(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	assert.Nil(t, err)

	// Connect to queue and add a test bucket
	ch, err := mq.Connect(conn, "test", 1, 0)
	assert.Nil(t, err)
	publishBucket(ch, bucket.Bucket{Name: "mqtest"})

	WorkMQ(0, &wg, conn, aws, "test", 1,
		false, false)
}
