package worker

import (
	"bytes"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
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
			PrintResult(&tt.b, false)
			assert.Contains(t2, buf.String(), tt.expected)
		})
	}
}
