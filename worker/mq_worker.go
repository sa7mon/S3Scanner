package worker

import (
	"encoding/json"
	"fmt"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/db"
	"github.com/sa7mon/s3scanner/mq"
	"github.com/sa7mon/s3scanner/provider"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
	"sync"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func WorkMQ(threadID int, wg *sync.WaitGroup, conn *amqp.Connection, provider provider.StorageProvider, queue string,
	threads int, doEnumerate bool, writeToDB bool) {
	_, once := os.LookupEnv("TEST_MQ") // If we're being tested, exit after one bucket is scanned
	defer wg.Done()

	// Wrap the whole thing in a for (while) loop so if the mq server kills the channel, we start it up again
	for {
		ch, chErr := mq.Connect(conn, queue, threads, threadID)
		if chErr != nil {
			FailOnError(chErr, "couldn't connect to message queue")
		}

		msgs, consumeErr := ch.Consume(queue, fmt.Sprintf("%s_%v", queue, threadID), false, false, false, false, nil)
		if consumeErr != nil {
			log.Error(fmt.Errorf("failed to register a consumer: %w", consumeErr))
			return
		}

		for j := range msgs {
			bucketToScan := bucket.Bucket{}

			unmarshalErr := json.Unmarshal(j.Body, &bucketToScan)
			if unmarshalErr != nil {
				log.Error(unmarshalErr)
			}

			if !bucket.IsValidS3BucketName(bucketToScan.Name) {
				log.Info(fmt.Sprintf("invalid   | %s", bucketToScan.Name))
				FailOnError(j.Ack(false), "failed to ack")
				continue
			}

			b, existsErr := provider.BucketExists(&bucketToScan)
			if existsErr != nil {
				log.WithFields(log.Fields{"bucket": b.Name, "step": "checkExists"}).Error(existsErr)
				FailOnError(j.Reject(false), "failed to reject")
			}
			if b.Exists == bucket.BucketNotExist {
				// ack the message and skip to the next
				log.Infof("not_exist | %s", b.Name)
				FailOnError(j.Ack(false), "failed to ack")
				continue
			}

			scanErr := provider.Scan(b, false)
			if scanErr != nil {
				log.WithFields(log.Fields{"bucket": b}).Error(scanErr)
				FailOnError(j.Reject(false), "failed to reject")
				continue
			}

			if doEnumerate {
				if b.PermAllUsersRead != bucket.PermissionAllowed {
					PrintResult(&bucketToScan, false)
					FailOnError(j.Ack(false), "failed to ack")
					if writeToDB {
						dbErr := db.StoreBucket(&bucketToScan)
						if dbErr != nil {
							log.Error(dbErr)
						}
					}
					continue
				}

				log.WithFields(log.Fields{"method": "main.mqwork()",
					"bucket_name": b.Name, "region": b.Region}).Debugf("enumerating objects...")

				enumErr := provider.Enumerate(b)
				if enumErr != nil {
					log.Errorf("Error enumerating bucket '%s': %v\nEnumerated objects: %v", b.Name, enumErr, len(b.Objects))
					FailOnError(j.Reject(false), "failed to reject")
				}
			}

			PrintResult(&bucketToScan, false)
			ackErr := j.Ack(false)
			if ackErr != nil {
				// Acknowledge mq message. May fail if we've taken too long and the server has closed the channel
				// If it has, we break and start at the top of the outer for-loop again which re-establishes a new
				// channel
				log.WithFields(log.Fields{"bucket": b}).Error(ackErr)
				break
			}

			// Write to database
			if writeToDB {
				dbErr := db.StoreBucket(&bucketToScan)
				if dbErr != nil {
					log.Error(dbErr)
				}
			}
			if once {
				return
			}
		}
	}
}
