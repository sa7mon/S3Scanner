package provider

import (
	"github.com/stretchr/testify/assert"
	"s3scanner/bucket"
	"testing"
)

func TestCustomProvider_BucketExists(t *testing.T) {
	t.Parallel()

	p := providers["custom"]
	var tests = []struct {
		name   string
		b      bucket.Bucket
		exists uint8
	}{
		{name: "exists, access denied", b: bucket.NewBucket("assets"), exists: bucket.BucketExists},
		{name: "exists, open", b: bucket.NewBucket("nurse-virtual-assistants"), exists: bucket.BucketExists},
		{name: "no such bucket", b: bucket.NewBucket("s3scanner-no-exist"), exists: bucket.BucketNotExist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			gb, err := p.BucketExists(&tt.b)
			assert.Nil(t2, err)
			assert.Equal(t2, tt.exists, gb.Exists)
		})
	}
}
