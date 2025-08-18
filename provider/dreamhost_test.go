package provider

import (
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProviderDreamhost_BucketExists(t *testing.T) {
	t.Parallel()
	b := bucket.NewBucket("auth")
	dh, err := NewProviderDreamhost()
	if err != nil {
		t.Error(err)
	}
	beb, err := dh.BucketExists(&b)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, bucket.BucketNotExist, beb.Exists)
}

func TestProviderDreamhost_BucketExists_Cap(t *testing.T) {
	t.Parallel()
	b := bucket.NewBucket("aUtH")
	dh, err := NewProviderDreamhost()
	if err != nil {
		t.Error(err)
	}
	beb, err := dh.BucketExists(&b)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, bucket.BucketNotExist, beb.Exists)
}
