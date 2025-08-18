package provider

import (
	"errors"
	"fmt"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
)

type Scaleway struct {
	clients *clientmap.ClientMap
}

func NewProviderScaleway() (*Scaleway, error) {
	sc := new(Scaleway)

	clients, err := sc.newClients()
	if err != nil {
		return sc, err
	}
	sc.clients = clients
	return sc, nil
}

func (sc *Scaleway) newClients() (*clientmap.ClientMap, error) {
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

func (sc *Scaleway) Scan(b *bucket.Bucket, doDestructiveChecks bool) error {
	client := sc.clients.Get(b.Region, false)
	return checkPermissions(client, b, doDestructiveChecks)
}

func (*Scaleway) Insecure() bool {
	return false
}

func (*Scaleway) Name() string {
	return "scaleway"
}

func (*Scaleway) AddressStyle() int {
	return PathStyle
}

func (sc *Scaleway) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
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

func (sc *Scaleway) Enumerate(b *bucket.Bucket) error {
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
