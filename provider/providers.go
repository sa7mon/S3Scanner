package provider

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/permission"
	"github.com/sa7mon/s3scanner/provider/clientmap"
	log "github.com/sirupsen/logrus"
)

const (
	PathStyle        = 0
	VirtualHostStyle = 1
)

type StorageProvider interface {
	Insecure() bool
	AddressStyle() int
	BucketExists(*bucket.Bucket) (*bucket.Bucket, error)
	Scan(*bucket.Bucket, bool) error
	Enumerate(*bucket.Bucket) error
	Name() string
}

type bucketCheckResult struct {
	region string
	exists bool
}

var AllProviders = []string{
	"aws", "custom", "digitalocean", "dreamhost", "gcp", "linode",
}

var ProviderRegions = map[string][]string{
	"digitalocean": {"nyc3", "sfo2", "sfo3", "ams3", "sgp1", "fra1", "syd1"},
	"dreamhost":    {"us-east-1"},
	"linode":       {"us-east-1", "us-southeast-1", "eu-central-1", "ap-south-1"},
}

func NewProvider(name string) (StorageProvider, error) {
	var (
		provider StorageProvider
		err      error
	)
	switch name {
	case "aws":
		provider, err = NewProviderAWS()
	case "digitalocean":
		provider, err = NewProviderDO()
	case "dreamhost":
		provider, err = NewProviderDreamhost()
	case "gcp":
		provider, err = NewProviderGCP()
	case "linode":
		provider, err = NewProviderLinode()
	default:
		err = fmt.Errorf("unknown provider: %s", name)
	}
	return provider, err
}

func newNonAWSClient(sp StorageProvider, regionURL string) (*s3.Client, error) {
	var httpClient s3.HTTPClient

	if sp.Insecure() {
		httpClient = &http.Client{Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	} else {
		httpClient = &http.Client{Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}}
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: regionURL,
				}, nil
			})),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	addrStyleOption := func(o *s3.Options) { o.UsePathStyle = false }
	if sp.AddressStyle() == PathStyle {
		addrStyleOption = func(o *s3.Options) { o.UsePathStyle = true }
	}

	cfg.Credentials = nil // TODO: Remove and test
	return s3.NewFromConfig(cfg, addrStyleOption), nil
}

/*
enumerateListObjectsV2 will enumerate all objects stored in b using the ListObjectsV2 API endpoint. The endpoint will
be called until the IsTruncated field is false
*/
func enumerateListObjectsV2(client *s3.Client, b *bucket.Bucket) error {
	var continuationToken *string
	continuationToken = nil
	page := 0

	for {
		log.WithFields(log.Fields{
			"bucket_name": b.Name,
			"method":      "providers.enumerateListObjectsV2()",
		}).Debugf("requesting objects page %d", page)
		output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            &b.Name,
			ContinuationToken: continuationToken,
			EncodingType:      types.EncodingTypeUrl,
		},
		)
		if err != nil {
			return err
		}

		for _, obj := range output.Contents {
			b.Objects = append(b.Objects, bucket.BucketObject{Key: *obj.Key, Size: uint64(obj.Size)})
			b.BucketSize += uint64(obj.Size)
		}

		if !output.IsTruncated {
			b.ObjectsEnumerated = true
			break
		}
		continuationToken = output.NextContinuationToken
		page += 1
		if page >= 5000 { // TODO: Should this limit be lowered?
			return errors.New("more than 5000 pages of objects found. Skipping for now")
		}
	}
	b.NumObjects = int32(len(b.Objects))
	return nil
}

func checkPermissions(client *s3.Client, b *bucket.Bucket, doDestructiveChecks bool) error {
	/*
		// 1. Check if b exists
		// 2. Check for READ_ACP
		// 3. If FullControl is allowed for either AllUsers or AuthorizedUsers, skip the remainder of those tests
		// 4. Check for READ
		// 5. If doing destructive checks:
		// 5a. Check for Write
		// 5b. Check for WriteACP
	*/

	b.DateScanned = time.Now()

	// Check for READ_ACP permission
	canReadACL, err := permission.CheckPermReadACL(client, b)
	if err != nil {
		return fmt.Errorf("error occurred while checking for ReadACL: %v", err.Error())
	}
	b.PermAllUsersReadACL = bucket.Permission(canReadACL)
	// TODO: Can we skip the rest of the checks if READ_ACP is allowed?

	// We can skip the rest of the checks if FullControl is allowed
	if b.PermAuthUsersFullControl == bucket.PermissionAllowed {
		return nil
	}

	// Check for READ permission
	canRead, err := permission.CheckPermRead(client, b)
	if err != nil {
		return fmt.Errorf("%v | error occured while checking for READ: %v", b.Name, err.Error())
	}
	b.PermAllUsersRead = bucket.Permission(canRead)

	if doDestructiveChecks {
		// Check for WRITE permission
		permWrite, writeErr := permission.CheckPermWrite(client, b)
		if writeErr != nil {
			return fmt.Errorf("%v | error occured while checking for WRITE: %v", b.Name, writeErr.Error())
		}
		b.PermAllUsersWrite = bucket.Permission(permWrite)

		// Check for WRITE_ACP permission
		permWriteAcl, writeAclErr := permission.CheckPermWriteAcl(client, b)
		if writeAclErr != nil {
			return fmt.Errorf("error occured while checking for WriteACL: %v", writeAclErr.Error())
		}
		b.PermAllUsersWriteACL = bucket.Permission(permWriteAcl)
	}
	return nil
}

func bucketExists(clients *clientmap.ClientMap, b *bucket.Bucket) (bool, string, error) {
	// TODO: Should this return a client or a region name? If region name, we'll need GetClient(region)
	// TODO: Add region priority - order in which to check. maps are not ordered
	results := make(chan bucketCheckResult, clients.Len())
	e := make(chan error, 1)

	clients.Each(func(region string, client *s3.Client) {
		go func(bucketName string, client *s3.Client, region string) {
			logFields := log.Fields{
				"bucket_name": b.Name,
				"region":      region,
				"method":      "providers.bucketExists()",
			}
			_, regionErr := manager.GetBucketRegion(context.TODO(), client, bucketName)
			if regionErr == nil {
				log.WithFields(logFields).Debugf("no error - bucket exists")
				results <- bucketCheckResult{region: region, exists: true}
				return
			}

			var bnf manager.BucketNotFound
			var re2 *awshttp.ResponseError
			if errors.As(regionErr, &bnf) {
				log.WithFields(logFields).Debugf("BucketNotFound")
				results <- bucketCheckResult{region: region, exists: false}
			} else if errors.As(regionErr, &re2) && re2.HTTPStatusCode() == 403 {
				log.WithFields(logFields).Debugf("AccessDenied")
				results <- bucketCheckResult{region: region, exists: true}
			} else {
				// If regionErr is a ResponseError, only return the unwrapped error i.e. "Method Not Allowed"
				// Otherwise, return the whole error
				err := regionErr
				if errors.As(regionErr, &re2) {
					err = re2.Unwrap()
				}
				log.WithFields(logFields).Debug(fmt.Errorf("unhandled error: %w", regionErr))
				e <- err
			}
		}(b.Name, client, region)
	})

	for i := 0; i < clients.Len(); i++ {
		select {
		case err := <-e:
			return false, "", err
		case res := <-results:
			if res.exists {
				return true, res.region, nil
			}
		}
	}
	return false, "", nil
}
