package provider

import (
	"github.com/stretchr/testify/assert"
	"os"
	"s3scanner/bucket"
	"testing"
)

var providers = map[string]StorageProvider{}

func TestMain(m *testing.M) {
	var provider StorageProvider

	provider, err := NewProviderAWS()
	if err != nil {
		panic(err)
	}
	providers["aws"] = provider

	provider, err = NewCustomProvider(
		"path",
		false,
		[]string{"ewr1", "ams1"},
		"https://$REGION.vultrobjects.com")
	if err != nil {
		panic(err)
	}
	providers["custom"] = provider

	provider, err = NewProviderDO()
	if err != nil {
		panic(err)
	}
	providers["digitalocean"] = provider

	provider, err = NewProviderDreamhost()
	if err != nil {
		panic(err)
	}
	providers["dreamhost"] = provider

	provider, err = NewProviderGCP()
	if err != nil {
		panic(err)
	}
	providers["gcp"] = provider

	provider, err = NewProviderLinode()
	if err != nil {
		panic(err)
	}
	providers["linode"] = provider

	code := m.Run()
	os.Exit(code)
}

func failIfError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
func TestProvider_EnumerateListObjectsV2_short(t *testing.T) {
	t.Parallel()
	p, pErr := NewProviderAWS()
	failIfError(t, pErr)
	c, cErr := p.newClient("us-east-1")
	failIfError(t, cErr)

	// Bucket with "page" of objects (<1k keys)
	b := bucket.Bucket{Name: "s3scanner-bucketsize",
		Exists: bucket.BucketExists, Region: "us-east-1",
		PermAllUsersRead: bucket.PermissionAllowed}
	enumErr := enumerateListObjectsV2(c, &b)
	if enumErr != nil {
		t.Errorf("error enumerating s3scanner-bucketsize: %e", enumErr)
	}
	assert.True(t, b.ObjectsEnumerated)
	assert.Equal(t, 1, len(b.Objects))
	assert.Equal(t, uint64(43), b.BucketSize)
}

func Test_EnumerateListObjectsV2_long(t *testing.T) {
	t.Parallel()
	p, pErr := NewProviderAWS()
	failIfError(t, pErr)
	c, cErr := p.newClient("us-east-1")
	failIfError(t, cErr)

	// Bucket with more than 1k objects
	b2 := bucket.Bucket{Name: "s3scanner-long", Exists: bucket.BucketExists,
		Region: "us-east-1", PermAllUsersRead: bucket.PermissionAllowed}
	b2Err := enumerateListObjectsV2(c, &b2)
	if b2Err != nil {
		t.Errorf("error enumerating s3scanner-long: %e", b2Err)
	}
	assert.True(t, b2.ObjectsEnumerated)
	assert.Equal(t, 3501, len(b2.Objects))
	assert.Equal(t, uint64(4000), b2.BucketSize)
}

func Test_StorageProvider_Statics(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name         string
		provider     StorageProvider
		insecure     bool
		addressStyle int
	}{
		{name: "AWS", provider: providers["aws"], insecure: false, addressStyle: VirtualHostStyle},
		{name: "DO", provider: providers["digitalocean"], insecure: false, addressStyle: PathStyle},
		{name: "Dreamhost", provider: providers["dreamhost"], insecure: false, addressStyle: PathStyle},
		{name: "GCP", provider: providers["gcp"], insecure: false, addressStyle: PathStyle},
		{name: "Linode", provider: providers["linode"], insecure: false, addressStyle: VirtualHostStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			assert.Equal(t2, tt.insecure, tt.provider.Insecure())
			assert.Equal(t2, tt.addressStyle, tt.provider.AddressStyle())
		})
	}
}

func Test_StorageProvider_BucketExists(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name       string
		provider   StorageProvider
		goodBucket bucket.Bucket
		badBucket  bucket.Bucket
	}{
		{name: "AWS", provider: providers["aws"], goodBucket: bucket.NewBucket("s3scanner-empty"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
		{name: "DO", provider: providers["digitalocean"], goodBucket: bucket.NewBucket("logo"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
		{name: "Dreamhost", provider: providers["dreamhost"], goodBucket: bucket.NewBucket("assets"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
		{name: "GCP", provider: providers["gcp"], goodBucket: bucket.NewBucket("books"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
		{name: "Linode", provider: providers["linode"], goodBucket: bucket.NewBucket("vantage"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			gb, err := tt.provider.BucketExists(&tt.goodBucket)
			assert.Nil(t2, err)
			assert.Equal(t2, bucket.BucketExists, gb.Exists)

			bb, err := tt.provider.BucketExists(&tt.badBucket)
			assert.Nil(t2, err)
			assert.Equal(t2, bucket.BucketNotExist, bb.Exists)

		})
	}
}

func Test_StorageProvider_Enum(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name       string
		provider   StorageProvider
		goodBucket bucket.Bucket
		numObjects int
	}{
		{name: "AWS", provider: providers["aws"], goodBucket: bucket.NewBucket("s3scanner-empty"), numObjects: 0},
		{name: "Custom public-read", provider: providers["custom"], goodBucket: bucket.NewBucket("alicante"), numObjects: 209},
		{name: "Custom no public-read", provider: providers["custom"], goodBucket: bucket.NewBucket("assets"), numObjects: 0},
		{name: "DO", provider: providers["digitalocean"], goodBucket: bucket.NewBucket("action"), numObjects: 2},
		{name: "Dreamhost", provider: providers["dreamhost"], goodBucket: bucket.NewBucket("bitrix24"), numObjects: 6},
		{name: "GCP", provider: providers["gcp"], goodBucket: bucket.NewBucket("assets"), numObjects: 3},
		{name: "Linode", provider: providers["linode"], goodBucket: bucket.NewBucket("vantage"), numObjects: 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			gb, err := tt.provider.BucketExists(&tt.goodBucket)
			assert.Nil(t2, err)
			err = tt.provider.Scan(&tt.goodBucket, false)
			assert.Nil(t2, err)
			scanErr := tt.provider.Enumerate(gb)
			assert.Nil(t2, scanErr)
			assert.Equal(t2, bucket.BucketExists, gb.Exists)
			assert.Equal(t2, int32(tt.numObjects), gb.NumObjects)
		})
	}
}

func Test_StorageProvider_Scan(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name        string
		provider    StorageProvider
		bucket      bucket.Bucket
		permissions string
	}{
		{name: "AWS", provider: providers["aws"], bucket: bucket.NewBucket("s3scanner-empty"), permissions: "AuthUsers: [] | AllUsers: [READ]"},
		{name: "Custom public-read-write", provider: providers["custom"], bucket: bucket.NewBucket("nurse-virtual-assistants"), permissions: "AuthUsers: [] | AllUsers: [READ, WRITE]"},
		{name: "Custom no public-read", provider: providers["custom"], bucket: bucket.NewBucket("assets"), permissions: "AuthUsers: [] | AllUsers: []"},
		{name: "DO", provider: providers["digitalocean"], bucket: bucket.NewBucket("logo"), permissions: "AuthUsers: [] | AllUsers: [READ]"},
		{name: "Dreamhost", provider: providers["dreamhost"], bucket: bucket.NewBucket("bitrix24"), permissions: "AuthUsers: [] | AllUsers: [READ]"},
		{name: "GCP", provider: providers["gcp"], bucket: bucket.NewBucket("hatrioua"), permissions: "AuthUsers: [] | AllUsers: []"},
		{name: "Linode", provider: providers["linode"], bucket: bucket.NewBucket("vantage"), permissions: "AuthUsers: [] | AllUsers: [READ]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			gb, err := tt.provider.BucketExists(&tt.bucket)
			scanErr := tt.provider.Scan(gb, true)
			assert.Nil(t2, err)
			assert.Nil(t2, scanErr)
			assert.Equal(t2, bucket.BucketExists, gb.Exists)
			assert.Equal(t2, tt.bucket.String(), tt.permissions)
		})
	}
}
