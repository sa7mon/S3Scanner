package provider

import (
	"errors"
	"fmt"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
)

type providerScaleway struct {
	clients *clientmap.ClientMap
}

func NewProviderScaleway() (*providerScaleway, error) {
	sc := new(providerScaleway)

	clients, err := sc.newClients()
	if err != nil {
		return sc, err
	}
	sc.clients = clients
	return sc, nil
}

func (sc *providerScaleway) newClients() (*clientmap.ClientMap, error) {
	clients := clientmap.WithCapacity(len(ProviderRegions[sc.Name()]))
	for _, r := range ProviderRegions[sc.Name()] {
		client, err := newNonAWSClient(sc, fmt.Sprintf("https://s3.%s.scw.cloud", r))
		if err != nil {
			return nil, err
		}
		clients.Set(r, false, client)
	}

	return clients, nil
}

func (sc *providerScaleway) Scan(b *bucket.Bucket, doDestructiveChecks bool) error {
	client := sc.clients.Get(b.Region, false)
	return checkPermissions(client, b, doDestructiveChecks)
}

func (*providerScaleway) Insecure() bool {
	return false
}

func (*providerScaleway) Name() string {
	return "scaleway"
}

func (*providerScaleway) AddressStyle() int {
	return PathStyle
}

func (sc *providerScaleway) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	b.Provider = sc.Name()
	exists, region, err := bucketExists(sc.clients, b)
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

func (sc *providerScaleway) Enumerate(b *bucket.Bucket) error {
	if b.Exists != bucket.BucketExists {
		return errors.New("bucket might not exist")
	}

	client := sc.clients.Get(b.Region, false)
	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}
