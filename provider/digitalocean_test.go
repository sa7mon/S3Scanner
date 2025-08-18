package provider

import (
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanBucketPermissions_DO(t *testing.T) {
	t.Parallel()

	do, doErr := NewDigitalOcean()
	assert.Nil(t, doErr)

	// Bucket exists but isn't open
	c := bucket.NewBucket("admin")
	c2, cErr := do.BucketExists(&c)
	assert.Nil(t, cErr)
	cScanErr := do.Scan(c2, true)
	assert.Nil(t, cScanErr)
	assert.Equal(t, bucket.BucketExists, c2.Exists)
	assert.Equal(t, bucket.PermissionDenied, c2.PermAllUsersRead)

	// Bucket exists and has READ open
	o := bucket.NewBucket("stats")
	o2, oErr := do.BucketExists(&o)
	assert.Nil(t, oErr)
	oScanErr := do.Scan(o2, true)
	assert.Nil(t, oScanErr)
	assert.Equal(t, bucket.BucketExists, o2.Exists)
	assert.Equal(t, bucket.PermissionAllowed, o2.PermAllUsersRead)

	// Bucket with a dot that does not exist
	//dotNoBucket := bucket.NewBucket("s3.s3scanner.com")

	// Bucket with an invalid name (contains @ sign)
	//emailBucket := bucket.NewBucket("admin@example.com")

	// Bucket exists and has READ and READ_ACL open
	// TODO: Find a bucket for here
	//readAclOpenBucket := bucket.NewBucket("s3scanner-all-read-readacl")
	//err = ScanBucketPermissions(doClient, &readAclOpenBucket, false, doEndpoint)
	//if err != nil {
	//	t.Error(err)
	//}
	//assert.Equal(t, bucket.BucketExists, readAclOpenBucket.Exists)
	//assert.Equal(t, bucket.PermissionAllowed, readAclOpenBucket.PermAllUsersRead)
	//assert.Equal(t, bucket.PermissionAllowed, readAclOpenBucket.PermAllUsersReadACL)

	// Open bucket with a dot that exists
	// TODO: Find a bucket for here
	//dotBucket := bucket.NewBucket("flaws.cloud")
	//err = ScanBucketPermissions(doClient, &dotBucket, false, doEndpoint)
	//if err != nil {
	//	t.Error(err)
	//}
	//assert.Equal(t, bucket.BucketExists, dotBucket.Exists)
	//assert.Equal(t, bucket.PermissionAllowed, dotBucket.PermAllUsersRead)
}
