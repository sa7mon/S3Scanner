package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/streadway/amqp"
	"log"
	"os"
	"strings"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%v - %v\n", msg, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("mqingest takes in a file of bucket names, one per line, and publishes them to a RabbitMQ queue")
	flag.PrintDefaults()
}

// Name of queue should indicate endpoint, no need to put endpoint in messgae

type BucketMessage struct {
	BucketName string `json:"bucket_name"`
}

func main() {
	var filename string
	var url string
	var queueName string

	flag.StringVar(&filename, "file", "", "File name of buckets to send to MQ")
	flag.StringVar(&url, "url", "amqp://guest:guest@localhost:5672/", "AMQP URI of RabbitMQ server")
	flag.StringVar(&queueName, "queue", "", "Name of message queue to publish buckets to")

	flag.Parse()

	if filename == "" || queueName == "" {
		fmt.Println("Flags 'file' and 'queue' are required")
		printUsage()
		os.Exit(1)
	}

	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare dead letter queue
	dlq, dlErr := ch.QueueDeclare(queueName+"_dead", true, false, false,
		false, nil)
	failOnError(dlErr, "Failed to declare dead letter queue")

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": dlq.Name,
		},
	)
	if err != nil {
		failOnError(err, "Failed to declare a queue")
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		failOnError(err, "Failed to set QoS on channel")
	}

	file, err := os.Open(filename)
	if err != nil {
		failOnError(err, "Failed to open file")
	}
	defer file.Close()

	msgsPublished := 0

	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		bucketName := strings.TrimSpace(fileScanner.Text())
		//bucketMsg := BucketMessage{BucketName: bucketName}
		bucketMsg := bucket.Bucket{Name: bucketName}
		bucketBytes, err := json.Marshal(bucketMsg)
		if err != nil {
			failOnError(err, "Failed to marshal bucket msg")
		}

		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{Body: bucketBytes, DeliveryMode: amqp.Persistent},
		)
		if err != nil {
			failOnError(err, "Failed to publish to channel")
		}
		msgsPublished++
	}
	if err := fileScanner.Err(); err != nil {
		failOnError(err, "fileScanner failed")
	}

	log.Printf("%v bucket names published to queue %v\n", msgsPublished, queueName)
}
