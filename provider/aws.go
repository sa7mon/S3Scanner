package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/sa7mon/s3scanner/permission"
	"net/http"
	"time"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/provider/clientmap"
	log "github.com/sirupsen/logrus"
)

type AWS struct {
	existsClient   *s3.Client
	clients        *clientmap.ClientMap
	hasCredentials bool
}

func (a *AWS) BucketExists(b *bucket.Bucket) (*bucket.Bucket, error) {
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
	}
	// Error wasn't BucketNotFound or 403
	return b, err
}

func (a *AWS) Scan(b *bucket.Bucket, doDestructiveChecks bool) error {
	anonClient, anonErr := a.getRegionClient(b.Region, false)
	if anonErr != nil {
		return anonErr
	}

	if a.hasCredentials {
		authClient, authClientErr := a.getRegionClient(b.Region, true)
		if authClientErr != nil {
			return authClientErr
		}
		return checkPermissionsWithAuth(anonClient, authClient, b, doDestructiveChecks)
	}
	return checkPermissionsWithAuth(anonClient, nil, b, doDestructiveChecks)
}

func (a *AWS) Enumerate(b *bucket.Bucket) error {
	useCreds := false
	if b.PermAuthUsersRead == bucket.PermissionAllowed {
		useCreds = true
	}
	client, err := a.getRegionClient(b.Region, useCreds)
	if err != nil {
		return err
	}

	enumErr := enumerateListObjectsV2(client, b)
	if enumErr != nil {
		return enumErr
	}
	return nil
}

func NewProviderAWS() (*AWS, error) {
	pa := new(AWS)
	client, err := pa.newAnonClient("us-east-1")
	if err != nil {
		return nil, err
	}
	pa.existsClient = client

	// Seed the clients map with a common region
	usEastClient, usErr := pa.newClient("us-east-1")
	if usErr != nil {
		return nil, usErr
	}

	// check if the user has properly configured credentials for scanning
	clientHasCreds := ClientHasCredentials(usEastClient)
	pa.hasCredentials = clientHasCreds
	pa.clients = clientmap.New()
	pa.clients.Set("us-east-1", clientHasCreds, usEastClient)
	return pa, nil
}

func (a *AWS) AddressStyle() int {
	// AWS supports both styles
	return VirtualHostStyle
}

func (*AWS) Insecure() bool {
	return false
}

func (*AWS) Name() string {
	return "aws"
}

func (*AWS) newAnonClient(region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithHTTPClient(&http.Client{Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}}),
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

func (a *AWS) newClient(region string) (*s3.Client, error) {
	logFields := log.Fields{"method": "aws.newClient", "region": region}

	configOpts := []func(*config.LoadOptions) error{
		config.WithHTTPClient(&http.Client{Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}}),
		config.WithRegion(region),
		config.WithEC2IMDSClientEnableState(imds.ClientDisabled), // Otherwise we wait 4 seconds to IMDSv2 to timeout
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		configOpts...,
	)
	if err != nil {
		return nil, err
	}

	hasCreds, accessKeyID := HasCredentials(cfg)
	if hasCreds {
		log.WithFields(logFields).Debugf("using access key ID: %s", accessKeyID)
	} else {
		log.WithFields(logFields).Debugf("no credentials found")
	}

	return s3.NewFromConfig(cfg), nil
}

// TODO: This method is copied from Linode
func (a *AWS) getRegionClient(region string, useCreds bool) (*s3.Client, error) {
	c := a.clients.Get(region, useCreds)
	if c != nil {
		return c, nil
	}

	// No client for this region yet - create one
	var newClient *s3.Client
	var newClientErr error
	if useCreds {
		newClient, newClientErr = a.newClient(region)
	} else {
		newClient, newClientErr = a.newAnonClient(region)
	}
	if newClientErr != nil {
		return nil, newClientErr
	}
	a.clients.Set(region, useCreds, newClient)
	return newClient, nil
}

func checkPermissionsWithAuth(anonClient *s3.Client, authClient *s3.Client, b *bucket.Bucket, doDestructiveChecks bool) error {
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

	// Check for anon READ_ACP permission. If allowed, exit
	anonReadACL, err := permission.CheckPermReadACL(anonClient, b)
	if err != nil {
		return fmt.Errorf("error occurred while checking for anon ReadACL: %v", err.Error())
	}
	b.PermAllUsersReadACL = bucket.Permission(anonReadACL)
	if b.PermAllUsersReadACL == bucket.PermissionAllowed {
		return nil
	}

	// Check for auth READ_ACP permission. If allowed, exit
	if authClient != nil {
		authReadACL, authACLErr := permission.CheckPermReadACL(authClient, b)
		if authACLErr != nil {
			return fmt.Errorf("error occurred while checking for auth ReadACL: %v", authACLErr.Error())
		}
		b.PermAuthUsersReadACL = bucket.Permission(authReadACL)
		if b.PermAuthUsersReadACL == bucket.PermissionAllowed {
			return nil
		}
	}

	// Check for anon READ
	canRead, err := permission.CheckPermRead(anonClient, b)
	if err != nil {
		return fmt.Errorf("error occurred while checking for anon READ: %v", err.Error())
	}
	b.PermAllUsersRead = bucket.Permission(canRead)

	// Check for auth READ
	if authClient != nil {
		authCanRead, authReadErr := permission.CheckPermRead(authClient, b)
		if authReadErr != nil {
			return fmt.Errorf("error occurred while checking for auth READ: %v", authReadErr.Error())
		}
		b.PermAuthUsersRead = bucket.Permission(authCanRead)
	}

	if doDestructiveChecks {
		// Check for WRITE permission
		permWrite, writeErr := permission.CheckPermWrite(anonClient, b)
		if writeErr != nil {
			return fmt.Errorf("%v | error occurred while checking for WRITE: %v", b.Name, writeErr.Error())
		}
		b.PermAllUsersWrite = bucket.Permission(permWrite)

		// Check for WRITE_ACP permission
		permWriteACL, writeACLErr := permission.CheckPermWriteACL(anonClient, b)
		if writeACLErr != nil {
			return fmt.Errorf("error occurred while checking for WriteACL: %v", writeACLErr.Error())
		}
		b.PermAllUsersWriteACL = bucket.Permission(permWriteACL)
	}
	return nil
}
