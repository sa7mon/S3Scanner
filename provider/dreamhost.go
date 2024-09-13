package provider

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
	"strings"
)

// Dreamhost responds strangely if you attempt to access a bucket named 'auth'
var forbiddenBuckets = []string{"auth"}

type Dreamhost struct {
	clients *clientmap.ClientMap
}

func (p Dreamhost) Insecure() bool {
	return false
}

func (Dreamhost) Name() string {
	return "dreamhost"
}

func (p Dreamhost) AddressStyle() int {
	return PathStyle
}

func (p Dreamhost) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
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

func (p Dreamhost) Scan(bucket *bucket.Bucket, doDestructiveChecks bool) error {
	client := p.getRegionClient(bucket.Region)
	return checkPermissions(client, bucket, doDestructiveChecks)
}

func (p Dreamhost) getRegionClient(region string) *s3.Client {
	return p.clients.Get(region, false)
}

func (p Dreamhost) Enumerate(b *bucket.Bucket) error {
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

func (p *Dreamhost) newClients() (*clientmap.ClientMap, error) {
	clients := clientmap.WithCapacity(len(ProviderRegions[p.Name()]))
	for _, r := range ProviderRegions[p.Name()] {
		client, err := newNonAWSClient(p, fmt.Sprintf("https://objects-%s.dream.io", r))
		if err != nil {
			return nil, err
		}
		clients.Set(r, false, client)
	}

	return clients, nil
}

func NewProviderDreamhost() (*Dreamhost, error) {
	pd := new(Dreamhost)

	clients, err := pd.newClients()
	if err != nil {
		return pd, err
	}
	pd.clients = clients
	return pd, nil
}
