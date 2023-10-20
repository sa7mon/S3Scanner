package provider

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
)

type providerOVH struct {
	clients *clientmap.ClientMap
}

func NewProviderOVH() (*providerOVH, error) {
	po := new(providerOVH)

	clients, err := po.newClients()
	if err != nil {
		return po, err
	}
	po.clients = clients
	return po, nil
}

func (po *providerOVH) getRegionClient(region string) *s3.Client {
	return po.clients.Get(region)
}

func (po *providerOVH) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	b.Provider = po.Name()
	exists, region, err := bucketExists(po.clients, b)
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

func (po *providerOVH) Enumerate(b *bucket.Bucket) error {
	if b.Exists != bucket.BucketExists {
		return errors.New("bucket might not exist")
	}

	client := po.getRegionClient(b.Region)
	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}

func (po *providerOVH) newClients() (*clientmap.ClientMap, error) {
	clients := clientmap.WithCapacity(len(ProviderRegions[po.Name()]))
	for _, r := range ProviderRegions[po.Name()] {
		client, err := newNonAWSClient(po, fmt.Sprintf("https://s3.%s.io.cloud.ovh.us", r))
		if err != nil {
			return nil, err
		}
		clients.Set(r, client)
	}

	return clients, nil
}

func (po *providerOVH) Scan(b *bucket.Bucket, doDestructiveChecks bool) error {
	client := po.getRegionClient(b.Region)
	return checkPermissions(client, b, doDestructiveChecks)
}

func (*providerOVH) Insecure() bool {
	return false
}

func (*providerOVH) Name() string {
	return "ovh"
}

func (*providerOVH) AddressStyle() int {
	return VirtualHostStyle
}
