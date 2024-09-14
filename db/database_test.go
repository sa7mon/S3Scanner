package db

import (
	"crypto/rand"
	"fmt"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"testing"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_/0123456789.")

func randPermission() uint8 {
	randInt, _ := rand.Int(rand.Reader, big.NewInt(2))
	return uint8(randInt.Int64())
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
		b[i] = letterRunes[randInt.Int64()]
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
		PermAuthUsersRead:        randPermission(),
		PermAuthUsersWrite:       randPermission(),
		PermAuthUsersReadACL:     randPermission(),
		PermAuthUsersWriteACL:    randPermission(),
		PermAuthUsersFullControl: randPermission(),
		PermAllUsersRead:         randPermission(),
		PermAllUsersWrite:        randPermission(),
		PermAllUsersReadACL:      randPermission(),
		PermAllUsersWriteACL:     randPermission(),
		PermAllUsersFullControl:  randPermission(),
	}
	bucketObjects := make([]bucket.Object, numObjects)
	for j := 0; j < numObjects; j++ {
		randSize, _ := rand.Int(rand.Reader, big.NewInt(250000000000))

		obj := bucket.Object{
			Key:  RandStringRunes(50),
			Size: randSize.Uint64(), // 25GB max
		}
		c.BucketSize += obj.Size
		bucketObjects[j] = obj
	}
	c.Objects = bucketObjects
	return c
}

func BenchmarkStoreBucket(b *testing.B) {
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
