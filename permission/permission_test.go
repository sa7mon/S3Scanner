package permission

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var doDestructiveChecks = false

var east2AnonClient *s3.Client
var east1AnonClient *s3.Client
var euNorth1Client *s3.Client

func failIfError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func TestMain(m *testing.M) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		panic(err)
	}
	cfg.Credentials = nil
	east1AnonClient = s3.NewFromConfig(cfg)

	east2Cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-2"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		panic(err)
	}
	east2Cfg.Credentials = nil
	east2AnonClient = s3.NewFromConfig(east2Cfg)

	euCfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("eu-north-1"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		panic(err)
	}
	euNorth1Client = s3.NewFromConfig(euCfg)

	code := m.Run()
	os.Exit(code)
}

func TestCheckPermReadACL(t *testing.T) {
	t.Parallel()

	// Bucket with READ_ACP allowed for AllUsers
	permReadAllowed, err := CheckPermReadACL(east2AnonClient, &bucket.Bucket{
		Name:   "s3scanner-all-readacp",
		Region: "us-east-2",
	})
	failIfError(t, err)
	assert.True(t, permReadAllowed)

	// Bucket with READ_ACP allowed for AuthenticatedUsers
	//permReadAllowed, err = CheckPermReadACL(east2CredClient, &bucket.Bucket{
	//	Name:   "s3scanner-auth-read-acl",
	//	Region: "us-east-2",
	//})
	//failIfError(t, err)
	//assert.True(t, permReadAllowed)

	// Bucket without READ_ACP allowed
	permReadAllowed, err = CheckPermReadACL(east1AnonClient, &bucket.Bucket{
		Name:   "s3scanner-private",
		Region: "us-east-1",
	})
	failIfError(t, err)
	assert.False(t, permReadAllowed)

	// Bucket with READ_ACP allowed for AuthenticatedUsers, but we scan without creds
	permReadAllowed, err = CheckPermReadACL(east2AnonClient, &bucket.Bucket{
		Name:   "s3scanner-auth-read-acl",
		Region: "us-east-2",
	})
	failIfError(t, err)
	assert.False(t, permReadAllowed)
}

func TestCheckPermRead(t *testing.T) {
	t.Parallel()

	// Bucket with READ permission
	readAllowedBucket := bucket.Bucket{
		Name:   "s3scanner-bucketsize",
		Region: "us-east-1",
	}

	// Assert we can read the bucket without creds
	permReadAllowed, err := CheckPermRead(east1AnonClient, &readAllowedBucket)
	failIfError(t, err)
	assert.True(t, permReadAllowed)

	// Assert we can read the bucket with creds
	//permReadAllowed, err = CheckPermRead(east1CredClient, &readAllowedBucket)
	//failIfError(t, err)
	//assert.True(t, permReadAllowed)

	// Bucket without READ permission
	readNotAllowedBucket := bucket.Bucket{
		Name:   "s3scanner-private",
		Region: "us-east-1",
	}

	// Assert we can't read the bucket without creds
	permReadAllowed, err = CheckPermRead(east1AnonClient, &readNotAllowedBucket)
	failIfError(t, err)
	assert.False(t, permReadAllowed)

	// Assert we can't read the bucket even with creds
	//permReadAllowed, err = CheckPermRead(east2CredClient, &readNotAllowedBucket)
	//failIfError(t, err)
	//assert.False(t, permReadAllowed)
}

func TestCheckPermWrite(t *testing.T) {
	t.Parallel()
	if !doDestructiveChecks {
		t.Skip("skipped destructive check TestCheckPermWrite")
	}

	// Bucket with READ permission
	readAllowedBucket := bucket.Bucket{
		Name:   "s3scanner-bucketsize",
		Region: "us-east-1",
	}

	// Assert we can read the bucket without creds
	permWrite, err := CheckPermWrite(east1AnonClient, &readAllowedBucket)
	assert.Nil(t, err)
	assert.False(t, permWrite)
}

func TestCheckPermWriteACL(t *testing.T) {
	t.Parallel()
	if !doDestructiveChecks {
		t.Skip("skipped destructive check TestCheckPermWriteACL")
	}

	// Bucket with READ permission
	readAllowedBucket := bucket.Bucket{
		Name:                     "s3scanner-bucketsize",
		Region:                   "us-east-1",
		PermAllUsersRead:         bucket.PermissionAllowed,
		PermAllUsersWrite:        bucket.PermissionAllowed,
		PermAllUsersReadACL:      bucket.PermissionAllowed,
		PermAllUsersFullControl:  bucket.PermissionAllowed,
		PermAuthUsersRead:        bucket.PermissionAllowed,
		PermAuthUsersReadACL:     bucket.PermissionAllowed,
		PermAuthUsersWrite:       bucket.PermissionAllowed,
		PermAuthUsersFullControl: bucket.PermissionAllowed,
	}

	// Assert we can read the bucket without creds
	permWrite, err := CheckPermWriteACL(east1AnonClient, &readAllowedBucket)
	assert.Nil(t, err)
	assert.False(t, permWrite)
}
