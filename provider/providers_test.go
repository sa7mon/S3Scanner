package provider

import (
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/stretchr/testify/assert"
	"os"
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
		[]string{"nyc3", "sfo2", "sfo3", "ams3", "sgp1", "fra1", "syd1"},
		"https://$REGION.digitaloceanspaces.com")
	if err != nil {
		panic(err)
	}
	providers["custom"] = provider

	provider, err = NewDigitalOcean()
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

	provider, err = NewProviderScaleway()
	if err != nil {
		panic(err)
	}
	providers["scaleway"] = provider

	provider, err = NewProviderWasabi()
	if err != nil {
		panic(err)
	}
	providers["wasabi"] = provider

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
		{name: "Scaleway", provider: providers["scaleway"], insecure: false, addressStyle: PathStyle},
		{name: "Wasabi", provider: providers["wasabi"], insecure: false, addressStyle: PathStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			t2.Parallel()
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
		{name: "Dreamhost", provider: providers["dreamhost"], goodBucket: bucket.NewBucket("images"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
		{name: "GCP", provider: providers["gcp"], goodBucket: bucket.NewBucket("books"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
		{name: "Linode", provider: providers["linode"], goodBucket: bucket.NewBucket("vantage"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
		{name: "Scaleway", provider: providers["scaleway"], goodBucket: bucket.NewBucket("2017"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
		{name: "Wasabi", provider: providers["wasabi"], goodBucket: bucket.NewBucket("acp"), badBucket: bucket.NewBucket("s3scanner-no-exist")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			t2.Parallel()
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
		{name: "DO", provider: providers["digitalocean"], goodBucket: bucket.NewBucket("action"), numObjects: 4},
		{name: "Dreamhost", provider: providers["dreamhost"], goodBucket: bucket.NewBucket("acc"), numObjects: 310},
		{name: "GCP", provider: providers["gcp"], goodBucket: bucket.NewBucket("assets"), numObjects: 3},
		{name: "Linode", provider: providers["linode"], goodBucket: bucket.NewBucket("vantage"), numObjects: 50},
		{name: "Scaleway", provider: providers["scaleway"], goodBucket: bucket.NewBucket("3d-builder"), numObjects: 1},
		{name: "Wasabi", provider: providers["wasabi"], goodBucket: bucket.NewBucket("animals"), numObjects: 102},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			t2.Parallel()
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
		{name: "AWS", provider: providers["aws"], bucket: bucket.NewBucket("s3scanner-bucketsize"), permissions: "AuthUsers: [READ] | AllUsers: [READ]"},
		{name: "Custom public-read-write", provider: providers["custom"], bucket: bucket.NewBucket("nurse-virtual-assistants"), permissions: "AuthUsers: [] | AllUsers: []"},
		{name: "Custom no public-read", provider: providers["custom"], bucket: bucket.NewBucket("assets"), permissions: "AuthUsers: [] | AllUsers: []"},
		{name: "DO", provider: providers["digitalocean"], bucket: bucket.NewBucket("logo"), permissions: "AuthUsers: [] | AllUsers: [READ]"},
		{name: "Dreamhost", provider: providers["dreamhost"], bucket: bucket.NewBucket("acc"), permissions: "AuthUsers: [] | AllUsers: [READ]"},
		{name: "GCP", provider: providers["gcp"], bucket: bucket.NewBucket("hatrioua"), permissions: "AuthUsers: [] | AllUsers: []"},
		{name: "Linode", provider: providers["linode"], bucket: bucket.NewBucket("vantage"), permissions: "AuthUsers: [] | AllUsers: [READ]"},
		{name: "Scaleway", provider: providers["scaleway"], bucket: bucket.NewBucket("3d-builder"), permissions: "AuthUsers: [] | AllUsers: [READ]"},
		{name: "Wasabi", provider: providers["wasabi"], bucket: bucket.NewBucket("acceptance"), permissions: "AuthUsers: [] | AllUsers: [READ, READ_ACP]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			t2.Parallel()
			gb, err := tt.provider.BucketExists(&tt.bucket)
			scanErr := tt.provider.Scan(gb, false)
			assert.Nil(t2, err)
			assert.Nil(t2, scanErr)
			assert.Equal(t2, bucket.BucketExists, gb.Exists)
			assert.Equal(t2, tt.permissions, tt.bucket.String())
		})
	}
}
