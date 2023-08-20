package worker

import (
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

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
	Work(&wg, c, aws, true, false, false)
}
