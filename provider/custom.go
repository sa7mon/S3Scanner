package provider

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"strings"
)

type CustomProvider struct {
	regions        []string
	clients        map[string]*s3.Client
	insecure       bool
	addressStyle   int
	endpointFormat string
}

func (cp CustomProvider) Insecure() bool {
	return cp.insecure
}

func (cp CustomProvider) AddressStyle() int {
	return cp.addressStyle
}

func (CustomProvider) Name() string {
	return "custom"
}

func (cp CustomProvider) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
	b.Provider = cp.Name()
	exists, region, err := bucketExists(cp.clients, b)
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

func (cp CustomProvider) Scan(b *bucket.Bucket, doDestructiveChecks bool) error {
	client := cp.getRegionClient(b.Region)
	return checkPermissions(client, b, doDestructiveChecks)
}

func (cp CustomProvider) Enumerate(b *bucket.Bucket) error {
	if b.Exists != bucket.BucketExists {
		return errors.New("bucket might not exist")
	}
	if b.PermAllUsersRead != bucket.PermissionAllowed {
		return nil
	}

	client := cp.getRegionClient(b.Region)
	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}

func (cp *CustomProvider) getRegionClient(region string) *s3.Client {
	c, ok := cp.clients[region]
	if ok {
		return c
	}
	return nil
}

/*
NewCustomProvider is a constructor which makes a new custom provider with the given options.
addressStyle should either be "path" or "vhost"
*/
func NewCustomProvider(addressStyle string, insecure bool, regions []string, endpointFormat string) (*CustomProvider, error) {
	cp := new(CustomProvider)
	cp.regions = regions
	cp.insecure = insecure
	cp.endpointFormat = endpointFormat
	if addressStyle == "path" {
		cp.addressStyle = PathStyle
	} else if addressStyle == "vhost" {
		cp.addressStyle = VirtualHostStyle
	} else {
		return cp, fmt.Errorf("unknown custom provider address style: %s. Expected 'path' or 'vhost'", addressStyle)
	}

	clients, err := cp.newClients()
	if err != nil {
		return nil, err
	}
	cp.clients = clients
	return cp, nil
}

func (cp *CustomProvider) newClients() (map[string]*s3.Client, error) {
	clients := make(map[string]*s3.Client, len(cp.regions))
	for _, r := range cp.regions {
		regionUrl := strings.Replace(cp.endpointFormat, "$REGION", r, -1)
		client, err := newNonAWSClient(cp, regionUrl)
		if err != nil {
			return nil, err
		}
		clients[r] = client
	}

	return clients, nil
}
