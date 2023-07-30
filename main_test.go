package main

import (
	"bytes"
	"encoding/json"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/mq"
	"github.com/sa7mon/s3scanner/provider"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

func publishBucket(ch *amqp.Channel, b bucket.Bucket) {
	bucketBytes, err := json.Marshal(b)
	if err != nil {
		failOnError(err, "Failed to marshal bucket msg")
	}

	err = ch.Publish(
		"",
		"test",
		false,
		false,
		amqp.Publishing{Body: bucketBytes, DeliveryMode: amqp.Transient},
	)
	if err != nil {
		failOnError(err, "Failed to publish to channel")
	}
}

func TestArgCollection_Validate(t *testing.T) {
	goodInputs := []argCollection{
		{
			bucketName: "asdf",
			bucketFile: "",
			useMq:      false,
		},
		{
			bucketName: "",
			bucketFile: "buckets.txt",
			useMq:      false,
		},
		{
			bucketName: "",
			bucketFile: "",
			useMq:      true,
		},
	}
	tooManyInputs := []argCollection{
		{
			bucketName: "asdf",
			bucketFile: "asdf",
			useMq:      false,
		},
		{
			bucketName: "adsf",
			bucketFile: "",
			useMq:      true,
		},
		{
			bucketName: "",
			bucketFile: "asdf.txt",
			useMq:      true,
		},
	}

	for _, v := range goodInputs {
		err := v.Validate()
		if err != nil {
			t.Errorf("%v: %e", v, err)
		}
	}
	for _, v := range tooManyInputs {
		err := v.Validate()
		if err == nil {
			t.Errorf("expected error but did not find one: %v", v)
		}
	}
}

func TestWork(t *testing.T) {
	b := bucket.NewBucket("s3scanner-bucketsize")
	aws, err := provider.NewProviderAWS()
	assert.Nil(t, err)
	b2, exErr := aws.BucketExists(&b)
	assert.Nil(t, exErr)

	wg := sync.WaitGroup{}
	wg.Add(1)
	c := make(chan bucket.Bucket, 1)
	c <- *b2
	close(c)
	work(&wg, c, aws, true, false)
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

	mqwork(0, &wg, conn, aws, "test", 1,
		false, false)
}

func TestLogs(t *testing.T) {
	var buf bytes.Buffer
	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: &buf,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
			log.InfoLevel,
		},
	})

	tests := []struct {
		name     string
		b        bucket.Bucket
		enum     bool
		expected string
	}{
		{name: "enumerated, public-read, empty", b: bucket.Bucket{
			Name:              "test-logging",
			Exists:            bucket.BucketExists,
			ObjectsEnumerated: true,
			NumObjects:        0,
			BucketSize:        0,
			PermAllUsersRead:  bucket.PermissionAllowed,
		}, enum: true, expected: "exists    | test-logging |  | AuthUsers: [] | AllUsers: [READ] | 0 objects (0 B)"},
		{name: "enumerated, closed", b: bucket.Bucket{
			Name:              "enumerated-closed",
			Exists:            bucket.BucketExists,
			ObjectsEnumerated: true,
			NumObjects:        0,
			BucketSize:        0,
			PermAllUsersRead:  bucket.PermissionDenied,
		}, enum: true, expected: "exists    | enumerated-closed |  | AuthUsers: [] | AllUsers: [] | 0 objects (0 B)"},
		{name: "closed", b: bucket.Bucket{
			Name:              "no-enumerate-closed",
			Exists:            bucket.BucketExists,
			ObjectsEnumerated: false,
			PermAllUsersRead:  bucket.PermissionDenied,
		}, enum: true, expected: "exists    | no-enumerate-closed |  | AuthUsers: [] | AllUsers: []"},
		{name: "no-enum-not-exist", b: bucket.Bucket{
			Name:   "no-enum-not-exist",
			Exists: bucket.BucketNotExist,
		}, enum: false, expected: "not_exist | no-enum-not-exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			printResult(&tt.b)
			assert.Contains(t2, buf.String(), tt.expected)
		})
	}

}
