package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
	"net/http"
)

type Wasabi struct {
	clients      *clientmap.ClientMap
	existsClient *s3.Client
}

func (w *Wasabi) Insecure() bool {
	return false
}

func (w *Wasabi) AddressStyle() int {
	return PathStyle
}

func (w *Wasabi) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	b.Provider = w.Name()
	exists, region, err := bucketExists301(w.existsClient, "us-east-1", b)
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

func (w *Wasabi) Scan(bucket *bucket.Bucket, doDestructiveChecks bool) error {
	client := w.clients.Get(bucket.Region, false)
	return checkPermissions(client, bucket, doDestructiveChecks)
}

func (w *Wasabi) Enumerate(b *bucket.Bucket) error {
	if b.Exists != bucket.BucketExists {
		return errors.New("bucket might not exist")
	}
	client := w.getRegionClient(b.Region)
	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}

func (w *Wasabi) getRegionClient(region string) *s3.Client {
	return w.clients.Get(region, false)
}

func (w *Wasabi) newExistsClient() (*s3.Client, error) {
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { // don't follow redirects
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithHTTPClient(client),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	cfg.BaseEndpoint = aws.String("https://s3.wasabisys.com")
	return s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true }), nil
}

func NewProviderWasabi() (*Wasabi, error) {
	w := new(Wasabi)
	clients, err := w.newClients()
	if err != nil {
		return w, err
	}
	w.clients = clients

	c, cErr := w.newExistsClient()
	if cErr != nil {
		return w, cErr
	}
	w.existsClient = c
	return w, nil
}

func (w *Wasabi) newClients() (*clientmap.ClientMap, error) {
	clients := clientmap.WithCapacity(len(ProviderRegions[w.Name()]))
	for _, r := range ProviderRegions[w.Name()] {
		client, err := newNonAWSClient(w, fmt.Sprintf("https://s3.%s.wasabisys.com", r))
		if err != nil {
			return nil, err
		}
		clients.Set(r, false, client)
	}

	return clients, nil
}

func (w *Wasabi) Name() string { return "wasabi" }
