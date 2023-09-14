package provider

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
)

type providerDO struct {
	clients *clientmap.ClientMap
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

func (pdo *providerDO) newClients() (*clientmap.ClientMap, error) {
	clients := clientmap.WithCapacity(len(ProviderRegions[pdo.Name()]))
	for _, r := range ProviderRegions[pdo.Name()] {
		client, err := newNonAWSClient(pdo, fmt.Sprintf("https://%s.digitaloceanspaces.com", r))
		if err != nil {
			return nil, err
		}
		clients.Set(r, client)
	}

	return clients, nil
}

func (pdo *providerDO) getRegionClient(region string) *s3.Client {
	return pdo.clients.Get(region)
}

func NewProviderDO() (*providerDO, error) {
	pdo := new(providerDO)

	clients, err := pdo.newClients()
	if err != nil {
		return pdo, err
	}
	pdo.clients = clients
	return pdo, nil
}
