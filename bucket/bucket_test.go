package bucket

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/sa7mon/s3scanner/groups"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestIsValidS3BucketName_Good(t *testing.T) {
	t.Parallel()

	goodNames := []string{"my-bucket", "asd", "b-2", "b.2-a", "100", "a.bc", "asdfdshfkhasdfkjhasdjkhfgakjhsdfghkjalksjhflkajshdflkjahsdlfkj"}

	for _, name := range goodNames {
		assert.True(t, IsValidS3BucketName(name))
	}
}

func TestIsValidS3BucketName_Bad(t *testing.T) {
	t.Parallel()

	badNames := []string{"a", "aa", ".abc", "-abc", "mybucket-s3alias", "-s3alias",
		"abc.", "abc-", "xn--abc", "-000-", "asdfdshfkhasdfkjhasdjkhfgakjhsdfghkjalksjhflkajshdflkjahsdlfkjab"}

	for _, name := range badNames {
		assert.False(t, IsValidS3BucketName(name), name)
	}
}

func TestNewBucket(t *testing.T) {
	t.Parallel()

	b := NewBucket("mybucket")
	assert.Equal(t, PermissionUnknown, b.PermAuthUsersRead)
	assert.Equal(t, PermissionUnknown, b.PermAuthUsersWrite)
	assert.Equal(t, PermissionUnknown, b.PermAuthUsersReadACL)
	assert.Equal(t, PermissionUnknown, b.PermAuthUsersWriteACL)
	assert.Equal(t, PermissionUnknown, b.PermAuthUsersFullControl)
	assert.Equal(t, PermissionUnknown, b.PermAllUsersRead)
	assert.Equal(t, PermissionUnknown, b.PermAllUsersWrite)
	assert.Equal(t, PermissionUnknown, b.PermAllUsersReadACL)
	assert.Equal(t, PermissionUnknown, b.PermAllUsersWriteACL)
	assert.Equal(t, PermissionUnknown, b.PermAllUsersFullControl)
	assert.Equal(t, BucketExistsUnknown, b.Exists)
	assert.False(t, b.ObjectsEnumerated)
	assert.Equal(t, "mybucket", b.Name)
}

type testOwner struct {
	DisplayName string
	ID          string
}

func TestBucket_ParseAclOutputv2(t *testing.T) {
	t.Parallel()

	o := testOwner{
		DisplayName: "Test User",
		ID:          "1234",
	}

	cannedACLPrivate := s3.GetBucketAclOutput{
		Grants: []types.Grant{},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	cannedACLPublicRead := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AllUsersv2,
				Permission: "READ",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	cannedACLPublicReadWrite := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AllUsersv2,
				Permission: "READ",
			},
			{
				Grantee:    groups.AllUsersv2,
				Permission: "WRITE",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	publicReadACL := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AllUsersv2,
				Permission: "READ_ACP",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	publicWriteACL := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AllUsersv2,
				Permission: "WRITE_ACP",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	cannedACLPublicFullControl := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AllUsersv2,
				Permission: "FULL_CONTROL",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	cannedACLAuthRead := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AuthenticatedUsersv2,
				Permission: "READ",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	cannedACLAuthReadWrite := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AuthenticatedUsersv2,
				Permission: "READ",
			},
			{
				Grantee:    groups.AuthenticatedUsersv2,
				Permission: "WRITE",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	authReadACL := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AuthenticatedUsersv2,
				Permission: "READ_ACP",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	authWriteACL := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AuthenticatedUsersv2,
				Permission: "WRITE_ACP",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}
	cannedACLAuthFullControl := s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee:    groups.AuthenticatedUsersv2,
				Permission: "FULL_CONTROL",
			},
		},
		Owner: &types.Owner{
			DisplayName: &o.DisplayName,
			ID:          &o.ID,
		},
	}

	var tests = []struct {
		name            string
		acl             s3.GetBucketAclOutput
		expectedAllowed map[*types.Grantee][]string
		expectedDenied  map[*types.Grantee][]string
	}{
		{name: "private", acl: cannedACLPrivate},
		{name: "public read", acl: cannedACLPublicRead, expectedAllowed: map[*types.Grantee][]string{
			groups.AllUsersv2: {"READ"},
		}},
		{name: "public read-write", acl: cannedACLPublicReadWrite, expectedAllowed: map[*types.Grantee][]string{
			groups.AllUsersv2: {"READ", "WRITE"},
		}},
		{name: "public read acl", acl: publicReadACL, expectedAllowed: map[*types.Grantee][]string{
			groups.AllUsersv2: {"READ_ACP"},
		}},
		{name: "public write acl", acl: publicWriteACL, expectedAllowed: map[*types.Grantee][]string{
			groups.AllUsersv2: {"WRITE_ACP"},
		}},
		{name: "public full control", acl: cannedACLPublicFullControl, expectedAllowed: map[*types.Grantee][]string{
			groups.AllUsersv2: {"FULL_CONTROL"},
		}},
		{name: "auth read", acl: cannedACLAuthRead, expectedAllowed: map[*types.Grantee][]string{
			groups.AuthenticatedUsersv2: {"READ"},
		}},
		{name: "auth read-write", acl: cannedACLAuthReadWrite, expectedAllowed: map[*types.Grantee][]string{
			groups.AuthenticatedUsersv2: {"READ", "WRITE"},
		}},
		{name: "auth read acl", acl: authReadACL, expectedAllowed: map[*types.Grantee][]string{
			groups.AuthenticatedUsersv2: {"READ_ACP"},
		}},
		{name: "auth write acl", acl: authWriteACL, expectedAllowed: map[*types.Grantee][]string{
			groups.AuthenticatedUsersv2: {"WRITE_ACP"},
		}},
		{name: "auth full control", acl: cannedACLAuthFullControl, expectedAllowed: map[*types.Grantee][]string{
			groups.AuthenticatedUsersv2: {"FULL_CONTROL"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			t2.Parallel()
			b := NewBucket("mytestbucket")
			err := b.ParseACLOutputV2(&tt.acl)
			assert.Nil(t2, err)

			for grantee, perms := range tt.expectedAllowed {
				for _, perm := range perms {
					assert.Equal(t2, PermissionAllowed, b.Permissions()[grantee][perm])
				}
			}
			for grantee, perms := range tt.expectedDenied {
				for _, perm := range perms {
					assert.Equal(t2, PermissionDenied, b.Permissions()[grantee][perm])
				}
			}
		})
	}
}

func TestFromReader(t *testing.T) {
	t.Parallel()

	reader := strings.NewReader(`test
bar
bucket
bar
test
foo
bucket
foo
bar`)

	testChan := make(chan Bucket)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	g, _ := errgroup.WithContext(ctx)
	defer cancel()

	g.Go(func() error {
		err := FromReader(reader, testChan)
		close(testChan)
		return err
	})

	i := 0
	for range testChan {
		i++
	}
	assert.Equal(t, 4, i)

	if err := g.Wait(); err != nil {
		t.Error(err)
	}
}

func TestReadFromFile(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	testFile := fmt.Sprintf("%s/_test_/buckets.txt", filepath.Dir(filename))

	testChan := make(chan Bucket)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	g, _ := errgroup.WithContext(ctx)
	defer cancel()

	g.Go(func() error {
		err := ReadFromFile(testFile, testChan)
		close(testChan)
		return err
	})

	var i = 0
	for b := range testChan {
		assert.Equal(t, fmt.Sprintf("mybucket%v", i), b.Name)
		i++
	}
	assert.Equal(t, 5, i)

	if err := g.Wait(); err != nil {
		t.Error(err)
	}
}

func TestBucket_String(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name   string
		bucket Bucket
		string string
	}{
		{name: "public read", bucket: Bucket{
			Exists:           BucketExists,
			PermAllUsersRead: PermissionAllowed,
		}, string: "AuthUsers: [] | AllUsers: [READ]"},
		{name: "public read-write", bucket: Bucket{
			Exists:            BucketExists,
			PermAllUsersRead:  PermissionAllowed,
			PermAllUsersWrite: PermissionAllowed,
		}, string: "AuthUsers: [] | AllUsers: [READ, WRITE]"},
		{name: "public read acl", bucket: Bucket{
			Exists:              BucketExists,
			PermAllUsersReadACL: PermissionAllowed,
		}, string: "AuthUsers: [] | AllUsers: [READ_ACP]"},
		{name: "public write acl", bucket: Bucket{
			Exists:               BucketExists,
			PermAllUsersWriteACL: PermissionAllowed,
		}, string: "AuthUsers: [] | AllUsers: [WRITE_ACP]"},
		{name: "public full control", bucket: Bucket{
			Exists:                  BucketExists,
			PermAllUsersFullControl: PermissionAllowed,
		}, string: "AuthUsers: [] | AllUsers: [FULL_CONTROL]"},
		{name: "auth read", bucket: Bucket{
			Exists:            BucketExists,
			PermAuthUsersRead: PermissionAllowed,
		}, string: "AuthUsers: [READ] | AllUsers: []"},
		{name: "auth read-write", bucket: Bucket{
			Exists:             BucketExists,
			PermAuthUsersRead:  PermissionAllowed,
			PermAuthUsersWrite: PermissionAllowed,
		}, string: "AuthUsers: [READ, WRITE] | AllUsers: []"},
		{name: "auth read acl", bucket: Bucket{
			Exists:               BucketExists,
			PermAuthUsersReadACL: PermissionAllowed,
		}, string: "AuthUsers: [READ_ACP] | AllUsers: []"},
		{name: "auth write acl", bucket: Bucket{
			Exists:                BucketExists,
			PermAuthUsersWriteACL: PermissionAllowed,
		}, string: "AuthUsers: [WRITE_ACP] | AllUsers: []"},
		{name: "auth full control", bucket: Bucket{
			Exists:                   BucketExists,
			PermAuthUsersFullControl: PermissionAllowed,
		}, string: "AuthUsers: [FULL_CONTROL] | AllUsers: []"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			t2.Parallel()
			assert.Equal(t2, tt.string, tt.bucket.String())
		})
	}
}
