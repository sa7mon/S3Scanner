package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"s3scanner-go/bucket"
	"testing"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_/0123456789.")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func makeRandomBucket(numObjects int) bucket.Bucket {
	c := bucket.Bucket{
		Name:                     fmt.Sprintf("s3scanner-test_%s", RandStringRunes(8)),
		Region:                   "us-east-1",
		Exists:                   1,
		DateScanned:              time.Now(),
		ObjectsEnumerated:        true,
		Provider:                 "aws",
		BucketSize:               0,
		PermAuthUsersRead:        uint8(rand.Intn(2)),
		PermAuthUsersWrite:       uint8(rand.Intn(2)),
		PermAuthUsersReadACL:     uint8(rand.Intn(2)),
		PermAuthUsersWriteACL:    uint8(rand.Intn(2)),
		PermAuthUsersFullControl: uint8(rand.Intn(2)),
		PermAllUsersRead:         uint8(rand.Intn(2)),
		PermAllUsersWrite:        uint8(rand.Intn(2)),
		PermAllUsersReadACL:      uint8(rand.Intn(2)),
		PermAllUsersWriteACL:     uint8(rand.Intn(2)),
		PermAllUsersFullControl:  uint8(rand.Intn(2)),
	}
	bucketObjects := make([]bucket.BucketObject, numObjects)
	for j := 0; j < numObjects; j++ {
		obj := bucket.BucketObject{
			Key:  RandStringRunes(50),
			Size: uint64(rand.Intn(250000000000)), // 25GB max
		}
		c.BucketSize += obj.Size
		bucketObjects[j] = obj
	}
	c.Objects = bucketObjects
	return c
}

func BenchmarkStoreBucket(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	// Connect to database
	err := Connect("host=localhost user=postgres password=example dbname=postgres port=5432 sslmode=disable TimeZone=America/Chicago", false)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		c := makeRandomBucket(50)
		sErr := StoreBucket(&c)
		if sErr != nil {
			b.Error(sErr)
		}
	}
}

func TestStoreBucket(t *testing.T) {
	_, testDB := os.LookupEnv("TEST_DB")
	if !testDB {
		t.Skip("TEST_DB not enabled")
	}
	err := Connect("host=localhost user=postgres password=example dbname=postgres port=5432 sslmode=disable", true)
	assert.Nil(t, err)

	b := makeRandomBucket(100)
	sErr := StoreBucket(&b)
	assert.Nil(t, sErr)
}
