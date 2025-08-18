package provider

import (
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/stretchr/testify/assert"
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

		{name: "exists, access denied", b: bucket.NewBucket("files"), exists: bucket.BucketExists},
		{name: "exists, open", b: bucket.NewBucket("logo"), exists: bucket.BucketExists},
		{name: "no such bucket", b: bucket.NewBucket("hammerbarn"), exists: bucket.BucketNotExist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			t2.Parallel()
			gb, err := p.BucketExists(&tt.b)
			assert.Nil(t2, err)
			assert.Equal(t2, tt.exists, gb.Exists)
		})
	}
}
