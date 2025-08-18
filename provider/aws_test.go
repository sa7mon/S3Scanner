package provider

import (
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/stretchr/testify/assert"
	"testing"
)

type bucketPermissionTestCase struct {
	b                                bucket.Bucket
	ExpectedPermAuthUsersRead        uint8
	ExpectedPermAuthUsersWrite       uint8
	ExpectedPermAuthUsersReadACL     uint8
	ExpectedPermAuthUsersWriteACL    uint8
	ExpectedPermAuthUsersFullControl uint8
	ExpectedPermAllUsersRead         uint8
	ExpectedPermAllUsersWrite        uint8
	ExpectedPermAllUsersReadACL      uint8
	ExpectedPermAllUsersWriteACL     uint8
	ExpectedPermAllUsersFullControl  uint8
}

//	// Bucket exists and has READ_ACL open for AuthenticatedUsers
//	authReadAclOpenBucket := bucket.NewBucket("s3scanner-auth-read-acl")
//	err = ScanBucketPermissions(awsClientNoRegion, &authReadAclOpenBucket, false, awsEndpoint, false)
//	failIfError(t, err)
//	err = ScanBucketPermissions(awsClientNoRegion, &authReadAclOpenBucket, false, awsEndpoint, true)
//	failIfError(t, err)
//	assert.Equal(t, bucket.BucketExists, authReadAclOpenBucket.Exists)
//	assert.Equal(t, bucket.PermissionDenied, authReadAclOpenBucket.PermAllUsersRead)
//	assert.Equal(t, bucket.PermissionDenied, authReadAclOpenBucket.PermAllUsersReadACL)
//	assert.Equal(t, bucket.PermissionDenied, authReadAclOpenBucket.PermAuthUsersRead)
//	assert.Equal(t, bucket.PermissionAllowed, authReadAclOpenBucket.PermAuthUsersReadACL)
//
//	// Bucket exists and has READ open for AuthenticatedUsers
//	authReadOpenBucket := bucket.NewBucket("s3scanner-auth")
//	err = ScanBucketPermissions(awsClientNoRegion, &authReadOpenBucket, false, awsEndpoint, false)
//	failIfError(t, err)
//	err = ScanBucketPermissions(awsClientNoRegion, &authReadOpenBucket, false, awsEndpoint, true)
//	failIfError(t, err)
//	assert.Equal(t, bucket.BucketExists, authReadOpenBucket.Exists)
//	assert.Equal(t, bucket.PermissionDenied, authReadOpenBucket.PermAllUsersRead)
//	assert.Equal(t, bucket.PermissionDenied, authReadOpenBucket.PermAllUsersReadACL)
//	assert.Equal(t, bucket.PermissionAllowed, authReadOpenBucket.PermAuthUsersRead)
//	assert.Equal(t, bucket.PermissionDenied, authReadOpenBucket.PermAuthUsersReadACL)

func TestProviderAWS_BucketExists(t *testing.T) {
	t.Parallel()
	p := providers["aws"]

	testCases := []struct {
		b           bucket.Bucket
		shouldExist uint8
	}{
		{bucket.NewBucket("s3scanner-private"), bucket.BucketExists},           // Bucket that exists
		{bucket.NewBucket("asdfasdfdoesnotexist"), bucket.BucketNotExist},      // Bucket that doesn't exist
		{bucket.NewBucket("flaws.cloud"), bucket.BucketExists},                 // Bucket with dot that exists
		{bucket.NewBucket("asdfasdf.danthesalmon.com"), bucket.BucketNotExist}, // Bucket with dot that doesn't exist
	}

	for _, testCase := range testCases {
		e, err := p.BucketExists(&testCase.b)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, testCase.shouldExist, e.Exists, testCase.b.Name)
	}

	// Bucket with invalid name
	b := bucket.NewBucket("asdf@test.com")
	_, err := p.BucketExists(&b)
	if err == nil {
		t.Error("expected error but didn't find one")
	}
	assert.Equal(t, "invalid bucket name", err.Error())
}

func TestProviderAWS_Scan(t *testing.T) {
	t.Parallel()
	p := providers["aws"]

	testCases := []bucketPermissionTestCase{
		{
			b:                            bucket.NewBucket("test"), // Bucket exists but isn't open
			ExpectedPermAllUsersRead:     bucket.PermissionDenied,
			ExpectedPermAllUsersReadACL:  bucket.PermissionDenied,
			ExpectedPermAuthUsersRead:    bucket.PermissionDenied,
			ExpectedPermAuthUsersReadACL: bucket.PermissionDenied,
		},
		{
			b:                            bucket.NewBucket("s3scanner-bucketsize"), // Bucket exists and has READ open for auth and all
			ExpectedPermAllUsersRead:     bucket.PermissionAllowed,
			ExpectedPermAllUsersReadACL:  bucket.PermissionDenied,
			ExpectedPermAuthUsersRead:    bucket.PermissionAllowed,
			ExpectedPermAuthUsersReadACL: bucket.PermissionDenied,
		},
		{
			b:                            bucket.NewBucket("s3scanner-all-read-readacl"), // Bucket exists and has READ and READ_ACL open for auth and all
			ExpectedPermAllUsersRead:     bucket.PermissionAllowed,
			ExpectedPermAllUsersReadACL:  bucket.PermissionAllowed,
			ExpectedPermAuthUsersRead:    bucket.PermissionDenied,
			ExpectedPermAuthUsersReadACL: bucket.PermissionDenied,
		},
		{
			b:                            bucket.NewBucket("s3scanner-auth-read"), // AuthUsers: [READ] | AllUsers: []
			ExpectedPermAllUsersRead:     bucket.PermissionDenied,
			ExpectedPermAllUsersReadACL:  bucket.PermissionDenied,
			ExpectedPermAuthUsersRead:    bucket.PermissionAllowed,
			ExpectedPermAuthUsersReadACL: bucket.PermissionDenied,
		},
		{
			b:                            bucket.NewBucket("s3scanner-auth-read-acl"), // AuthUsers: [READ_ACP] | AllUsers: []
			ExpectedPermAllUsersRead:     bucket.PermissionDenied,
			ExpectedPermAllUsersReadACL:  bucket.PermissionDenied,
			ExpectedPermAuthUsersRead:    bucket.PermissionDenied,
			ExpectedPermAuthUsersReadACL: bucket.PermissionAllowed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.b.Name, func(t2 *testing.T) {
			t2.Parallel()
			b, err := p.BucketExists(&testCase.b)
			if err != nil {
				t2.Error(err)
			}
			scanErr := p.Scan(b, false)
			if scanErr != nil {
				t2.Error(scanErr)
			}
			assert.Equal(t2, testCase.ExpectedPermAllUsersRead, b.PermAllUsersRead)
			assert.Equal(t2, testCase.ExpectedPermAllUsersReadACL, b.PermAllUsersReadACL)
			assert.Equal(t2, testCase.ExpectedPermAuthUsersRead, b.PermAuthUsersRead)
			assert.Equal(t2, testCase.ExpectedPermAuthUsersReadACL, b.PermAuthUsersReadACL)
		})
	}
}

func TestProviderAWS_Enumerate(t *testing.T) {
	t.Parallel()
	p := providers["aws"]

	b := bucket.NewBucket("s3scanner-bucketsize")
	b2, err := p.BucketExists(&b)
	assert.Nil(t, err)
	assert.Equal(t, bucket.BucketExists, b2.Exists)

	err = p.Enumerate(b2)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, b2.NumObjects)
	assert.EqualValues(t, 1, len(b2.Objects))
}
