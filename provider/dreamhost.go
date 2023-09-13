package provider

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
)

// Dreamhost responds strangely if you attempt to access a bucket named 'auth'
var forbiddenBuckets = []string{"auth"}

type ProviderDreamhost struct {
	regions []string
	clients *clientmap.ClientMap
}

func (p ProviderDreamhost) Insecure() bool {
	return false
}

func (ProviderDreamhost) Name() string {
	return "dreamhost"
}

func (p ProviderDreamhost) AddressStyle() int {
	return PathStyle
}

func (p ProviderDreamhost) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	// Check for forbidden name
	for _, fb := range forbiddenBuckets {
		if strings.ToLower(b.Name) == fb {
			b.Exists = bucket.BucketNotExist
			return b, nil
		}
	}

	b.Provider = p.Name()
	exists, region, err := bucketExists(p.clients, b)
	if err != nil {
		return b, err
	}
	if exists {
		b.Exists = bucket.BucketExists
		b.Region = region
	} else {
		b.Exists = bucket.BucketNotExist
	}

	return b, nil
}

func (p ProviderDreamhost) Scan(bucket *bucket.Bucket, doDestructiveChecks bool) error {
	client := p.getRegionClient(bucket.Region)
	return checkPermissions(client, bucket, doDestructiveChecks)
}

func (p ProviderDreamhost) getRegionClient(region string) *s3.Client {
	return p.clients.Get(region)
}

func (p ProviderDreamhost) Enumerate(b *bucket.Bucket) error {
	if b.Exists != bucket.BucketExists {
		return errors.New("bucket might not exist")
	}

	client := p.getRegionClient(b.Region)
	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}

func (p ProviderDreamhost) Regions() []string {
	urls := make([]string, len(p.regions))
	for i, r := range p.regions {
		urls[i] = fmt.Sprintf("https://objects-%s.dream.io", r)
	}
	return urls
}

func (p *ProviderDreamhost) newClients() (*clientmap.ClientMap, error) {
	clients := clientmap.WithCapacity(len(p.regions))
	for _, r := range p.Regions() {
		client, err := newNonAWSClient(p, r)
		if err != nil {
			return nil, err
		}
		clients.Set(r, client)
	}

	return clients, nil
}

func NewProviderDreamhost() (*ProviderDreamhost, error) {
	pd := new(ProviderDreamhost)
	pd.regions = []string{"us-east-1"}

	clients, err := pd.newClients()
	if err != nil {
		return pd, err
	}
	pd.clients = clients
	return pd, nil
}
