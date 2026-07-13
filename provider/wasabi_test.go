package provider

import (
	"testing"

	"github.com/sa7mon/s3scanner/bucket"
	"github.com/stretchr/testify/assert"
)

func TestWasabi_NewExistsClient(t *testing.T) {
	t.Parallel()
	w, wErr := NewProviderWasabi()
	assert.Nil(t, wErr)
	_, err := w.newExistsClient()
	assert.Nil(t, err)
}

func TestWasabi_BucketExists(t *testing.T) {
	t.Parallel()
	w, _ := NewProviderWasabi()
	exists, err := w.BucketExists(&bucket.Bucket{Name: "images"})
	assert.Nil(t, err)
	assert.Equal(t, bucket.BucketExists, exists.Exists)
	assert.Equal(t, "us-central-1", exists.Region)

	// exists in the default region - check returns a 200 instead of redirect
	exists, err = w.BucketExists(&bucket.Bucket{Name: "aedata"})
	assert.Nil(t, err)
	assert.Equal(t, bucket.BucketExists, exists.Exists)
	assert.Equal(t, "us-east-1", exists.Region)

	// no such bucket
	exists, err = w.BucketExists(&bucket.Bucket{Name: "helmet-rocket-trinket"})
	assert.Nil(t, err)
	assert.Equal(t, bucket.BucketNotExist, exists.Exists)
}
