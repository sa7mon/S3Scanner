package provider

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type providerAWS struct {
	existsClient *s3.Client
	clients      map[string]*s3.Client
}

func (a *providerAWS) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	b.Provider = a.Name()
	if !bucket.IsValidS3BucketName(b.Name) {
		return nil, errors.New("invalid bucket name")
	}
	region, err := manager.GetBucketRegion(context.TODO(), a.existsClient, b.Name)
	if err == nil {
		log.WithFields(log.Fields{"method": "aws.BucketExists()",
			"bucket_name": b.Name, "region": region}).Debugf("no error - bucket exists")
		b.Exists = bucket.BucketExists
		b.Region = region
		return b, nil
	}
	log.WithFields(log.Fields{"method": "aws.BucketExists()",
		"bucket_name": b.Name, "region": region}).Debug(err)

	var bnf manager.BucketNotFound
	var re2 *awshttp.ResponseError
	if errors.As(err, &bnf) {
		b.Exists = bucket.BucketNotExist
		return b, nil
	} else if errors.As(err, &re2) && re2.HTTPStatusCode() == 403 {
		// AccessDenied implies the bucket exists
		b.Exists = bucket.BucketExists
		b.Region = region
		return b, nil
	} else {
		// Error wasn't BucketNotFound or 403
		return b, err
	}
}

func (a *providerAWS) Scan(b *bucket.Bucket, doDestructiveChecks bool) error {
	client, err := a.getRegionClient(b.Region)
	if err != nil {
		return err
	}

	return checkPermissions(client, b, doDestructiveChecks)
}

func (a *providerAWS) Enumerate(b *bucket.Bucket) error {
	client, err := a.getRegionClient(b.Region)
	if err != nil {
		return err
	}

	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}

func NewProviderAWS() (*providerAWS, error) {
	pa := new(providerAWS)
	client, err := pa.newAnonClientNoRegion()
	if err != nil {
		return nil, err
	}
	pa.existsClient = client

	// Seed the clients map with a common region
	usEastClient, usErr := pa.newClient("us-east-1")
	if usErr != nil {
		return nil, usErr
	}
	pa.clients = map[string]*s3.Client{"us-east-1": usEastClient}
	return pa, nil
}

func (a *providerAWS) AddressStyle() int {
	// AWS supports both styles
	return VirtualHostStyle
}

func (*providerAWS) Insecure() bool {
	return false
}

func (*providerAWS) Name() string {
	return "aws"
}

func (*providerAWS) newAnonClientNoRegion() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithDefaultRegion("us-west-2"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithHTTPClient(&http.Client{Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}}),
	)
	if err != nil {
		return nil, err
	}

	cfg.Credentials = nil
	s3ClientNoRegion := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = false
	})

	return s3ClientNoRegion, nil
}

func (a *providerAWS) newClient(region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithHTTPClient(&http.Client{Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}}))

	if err != nil {
		return nil, err
	}

	cfg.Credentials = nil
	return s3.NewFromConfig(cfg), nil
}

// TODO: This method is copied from providerLinode
func (a *providerAWS) getRegionClient(region string) (*s3.Client, error) {
	c, ok := a.clients[region]
	if ok {
		return c, nil
	}

	// No client for this region yet - create one
	c, err := a.newClient(region)
	if err != nil {
		return nil, err
	}
	a.clients[region] = c // TODO: Make sure this is thread-safe
	return c, nil
}
