package provider

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"s3scanner/bucket"
)

type providerDO struct {
	regions []string
	clients map[string]*s3.Client
}

func (pdo providerDO) Insecure() bool {
	return false
}

func (pdo providerDO) Name() string {
	return "digitalocean"
}

func (pdo providerDO) AddressStyle() int {
	return PathStyle
}

func (pdo providerDO) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	b.Provider = pdo.Name()
	exists, region, err := bucketExists(pdo.clients, b)
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

func (pdo providerDO) Scan(bucket *bucket.Bucket, doDestructiveChecks bool) error {
	client := pdo.getRegionClient(bucket.Region)
	return checkPermissions(client, bucket, doDestructiveChecks)
}

func (pdo providerDO) Enumerate(b *bucket.Bucket) error {
	if b.Exists != bucket.BucketExists {
		return errors.New("bucket might not exist")
	}

	client := pdo.getRegionClient(b.Region)
	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}

func (pdo *providerDO) Regions() []string {
	urls := make([]string, len(pdo.regions))
	for i, r := range pdo.regions {
		urls[i] = fmt.Sprintf("https://%s.digitaloceanspaces.com", r)
	}
	return urls
}

func (pdo *providerDO) newClients() (map[string]*s3.Client, error) {
	clients := make(map[string]*s3.Client, len(pdo.regions))
	for _, r := range pdo.Regions() {
		client, err := newNonAWSClient(pdo, r)
		if err != nil {
			return nil, err
		}
		clients[r] = client
	}

	return clients, nil
}

func (pdo *providerDO) getRegionClient(region string) *s3.Client {
	c, ok := pdo.clients[region]
	if ok {
		return c
	}
	return nil
}

func NewProviderDO() (*providerDO, error) {
	pdo := new(providerDO)
	pdo.regions = []string{"nyc3", "sfo2", "sfo3", "ams3", "sgp1", "fra1", "syd1"}

	clients, err := pdo.newClients()
	if err != nil {
		return pdo, err
	}
	pdo.clients = clients
	return pdo, nil
}
