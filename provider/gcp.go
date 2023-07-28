package provider

import (
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"s3scanner-go/bucket"
)

// GCP like AWS, has a "universal" endpoint, but unlike AWS GCP does not require you to follow a redirect to the
// "proper" region. We can simply use storage.googleapis.com as the endpoint for all requests.
type GCP struct {
	client *s3.Client
}

func (g GCP) Insecure() bool {
	return false
}

func (GCP) Name() string {
	return "gcp"
}

// AddressStyle will return PathStyle, but GCP also supports VirtualHostStyle
func (g GCP) AddressStyle() int {
	return PathStyle
}

func (g GCP) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	b.Provider = g.Name()
	if !bucket.IsValidS3BucketName(b.Name) {
		return nil, errors.New("invalid bucket name")
	}
	exists, region, err := bucketExists(map[string]*s3.Client{"default": g.client}, b)
	if err != nil {
		return b, err
	}

	b.Region = region
	if exists {
		b.Exists = bucket.BucketExists
	} else {
		b.Exists = bucket.BucketNotExist
	}

	return b, nil
}

func (g GCP) Scan(bucket *bucket.Bucket, doDestructiveChecks bool) error {
	return checkPermissions(g.client, bucket, doDestructiveChecks)
}

func (g GCP) Enumerate(bucket *bucket.Bucket) error {
	return enumerateListObjectsV2(g.client, bucket)
}

func NewProviderGCP() (*GCP, error) {
	pg := new(GCP)
	c, err := newNonAWSClient(pg, "https://storage.googleapis.com")
	if err != nil {
		return pg, err
	}
	pg.client = c

	return pg, nil
}
