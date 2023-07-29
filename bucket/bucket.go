package bucket

import (
	"bufio"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"s3scanner/groups"
	"strings"
	"time"
	"unicode"
)

var BucketExists = uint8(1)
var BucketNotExist = uint8(0)
var BucketExistsUnknown = uint8(2)

var PermissionAllowed = uint8(1)
var PermissionDenied = uint8(0)
var PermissionUnknown = uint8(2)

// var bucketReIP = regexp.MustCompile(`^[0-9]{1-3}\.[0-9]{1-3}\.[0-9]{1-3}\.[0-9]{1-3}$`)
var bucketRe = regexp.MustCompile(`[^.\-a-z0-9]`)

// Pattern from https://blogs.easydynamics.com/2016/10/24/aws-s3-bucket-name-validation-regex/
// Missing:
// No xn-- prefix
// No -s3alias suffix
// https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html

type Bucket struct {
	//gorm.Model
	ID                uint           `gorm:"primarykey" json:",omitempty"`
	Name              string         `json:"name" gorm:"name;size:64;index"`
	Region            string         `json:"region" gorm:"size:20"`
	Exists            uint8          `json:"exists"`
	DateScanned       time.Time      `json:"date_scanned"`
	Objects           []BucketObject `json:"objects"`
	ObjectsEnumerated bool           `json:"objects_enumerated"`
	Provider          string         `json:"provider"`
	NumObjects        int32          `json:"num_objects"`

	// Total size of all bucket objects in bytes
	BucketSize       uint64 `json:"bucket_size"`
	OwnerId          string `json:"owner_id"`
	OwnerDisplayName string `json:"owner_display_name"`

	PermAuthUsersRead        uint8 `json:"perm_auth_users_read"`
	PermAuthUsersWrite       uint8 `json:"perm_auth_users_write"`
	PermAuthUsersReadACL     uint8 `json:"perm_auth_users_read_acl"`
	PermAuthUsersWriteACL    uint8 `json:"perm_auth_users_write_acl"`
	PermAuthUsersFullControl uint8 `json:"perm_auth_users_full_control"`

	PermAllUsersRead        uint8 `json:"perm_all_users_read"`
	PermAllUsersWrite       uint8 `json:"perm_all_users_write"`
	PermAllUsersReadACL     uint8 `json:"perm_all_users_read_acl"`
	PermAllUsersWriteACL    uint8 `json:"perm_all_users_write_acl"`
	PermAllUsersFullControl uint8 `json:"perm_all_users_full_control"`
}

type BucketObject struct {
	//gorm.Model
	ID       uint   `gorm:"primarykey" json:",omitempty"`
	Key      string `json:"key" gorm:"type:string;size:1024"` // Keys can be up to 1,024 bytes long, UTF-8 encoded plus an additional byte just in case. https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-keys.html
	Size     uint64 `json:"size"`
	BucketID uint   `json:",omitempty"`
}

func NewBucket(name string) Bucket {
	return Bucket{
		Name:                     name,
		Exists:                   BucketExistsUnknown,
		ObjectsEnumerated:        false,
		PermAuthUsersRead:        PermissionUnknown,
		PermAuthUsersWrite:       PermissionUnknown,
		PermAuthUsersReadACL:     PermissionUnknown,
		PermAuthUsersWriteACL:    PermissionUnknown,
		PermAuthUsersFullControl: PermissionUnknown,
		PermAllUsersRead:         PermissionUnknown,
		PermAllUsersWrite:        PermissionUnknown,
		PermAllUsersReadACL:      PermissionUnknown,
		PermAllUsersWriteACL:     PermissionUnknown,
		PermAllUsersFullControl:  PermissionUnknown,
	}
}

func (bucket *Bucket) String() string {
	if bucket.Exists == BucketNotExist {
		return fmt.Sprintf("%v | bucket_not_exist", bucket.Name)
	}

	var authUserPerms []string
	if bucket.PermAuthUsersRead == PermissionAllowed {
		authUserPerms = append(authUserPerms, "READ")
	}
	if bucket.PermAuthUsersWrite == PermissionAllowed {
		authUserPerms = append(authUserPerms, "WRITE")
	}
	if bucket.PermAuthUsersReadACL == PermissionAllowed {
		authUserPerms = append(authUserPerms, "READ_ACP")
	}
	if bucket.PermAuthUsersWriteACL == PermissionAllowed {
		authUserPerms = append(authUserPerms, "WRITE_ACP")
	}
	if bucket.PermAuthUsersFullControl == PermissionAllowed {
		authUserPerms = append(authUserPerms, "FULL_CONTROL")
	}

	var allUsersPerms []string
	if bucket.PermAllUsersRead == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "READ")
	}
	if bucket.PermAllUsersWrite == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "WRITE")
	}
	if bucket.PermAllUsersReadACL == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "READ_ACP")
	}
	if bucket.PermAllUsersWriteACL == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "WRITE_ACP")
	}
	if bucket.PermAllUsersFullControl == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "FULL_CONTROL")
	}

	return fmt.Sprintf("AuthUsers: [%v] | AllUsers: [%v]", strings.Join(authUserPerms, ", "), strings.Join(allUsersPerms, ", "))
}

func (bucket *Bucket) Permissions() map[*types.Grantee]map[string]uint8 {
	return map[*types.Grantee]map[string]uint8{
		groups.AllUsersv2: {
			"READ":         bucket.PermAllUsersRead,
			"WRITE":        bucket.PermAllUsersWrite,
			"READ_ACP":     bucket.PermAllUsersReadACL,
			"WRITE_ACP":    bucket.PermAllUsersWriteACL,
			"FULL_CONTROL": bucket.PermAllUsersFullControl,
		},
		groups.AuthenticatedUsersv2: {
			"READ":         bucket.PermAuthUsersRead,
			"WRITE":        bucket.PermAuthUsersWrite,
			"READ_ACP":     bucket.PermAuthUsersReadACL,
			"WRITE_ACP":    bucket.PermAuthUsersWriteACL,
			"FULL_CONTROL": bucket.PermAuthUsersFullControl,
		},
	}
}

func ReadFromFile(bucketFile string, bucketChan chan Bucket) error {
	file, err := os.Open(bucketFile)
	if err != nil {
		return err
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		bucketName := strings.TrimSpace(fileScanner.Text())
		if !IsValidS3BucketName(bucketName) {
			log.Info(fmt.Sprintf("invalid   | %s", bucketName))
		} else {
			bucketChan <- NewBucket(strings.ToLower(bucketName))
		}
	}

	if ferr := fileScanner.Err(); ferr != nil {
		return ferr
	}

	return err
}

// ParseAclOutputv2 TODO: probably move this to providers.go
func (bucket *Bucket) ParseAclOutputv2(aclOutput *s3.GetBucketAclOutput) error {
	bucket.OwnerId = *aclOutput.Owner.ID
	if aclOutput.Owner.DisplayName != nil {
		bucket.OwnerDisplayName = *aclOutput.Owner.DisplayName
	}

	for _, b := range aclOutput.Grants {
		if b.Grantee == groups.AllUsersv2 {
			switch b.Permission {
			case types.PermissionRead:
				bucket.PermAllUsersRead = PermissionAllowed
			case types.PermissionWrite:
				bucket.PermAllUsersWrite = PermissionAllowed
			case types.PermissionReadAcp:
				bucket.PermAllUsersReadACL = PermissionAllowed
			case types.PermissionWriteAcp:
				bucket.PermAllUsersWriteACL = PermissionAllowed
			case types.PermissionFullControl:
				bucket.PermAllUsersFullControl = PermissionAllowed
			default:
				break
			}
		}
		if b.Grantee == groups.AuthenticatedUsersv2 {
			switch b.Permission {
			case types.PermissionRead:
				bucket.PermAuthUsersRead = PermissionAllowed
			case types.PermissionWrite:
				bucket.PermAuthUsersWrite = PermissionAllowed
			case types.PermissionReadAcp:
				bucket.PermAuthUsersReadACL = PermissionAllowed
			case types.PermissionWriteAcp:
				bucket.PermAuthUsersWriteACL = PermissionAllowed
			case types.PermissionFullControl:
				bucket.PermAuthUsersFullControl = PermissionAllowed
			default:
				break
			}
		}
	}
	return nil
}

// Permission is a convenience method to convert a boolean into either a PermissionAllowed or PermissionDenied
func Permission(canDo bool) uint8 {
	if canDo {
		return PermissionAllowed
	} else {
		return PermissionDenied
	}
}

func IsValidS3BucketName(bucketName string) bool {
	// TODO: Optimize the heck out of this
	/*
		Bucket names must not be formatted as an IP address (for example, 192.168.5.4).
	*/

	// Bucket names can consist only of lowercase letters, numbers, dots (.), and hyphens (-).
	if bucketRe.MatchString(bucketName) {
		return false
	}

	// Bucket names must be between 3 (min) and 63 (max) characters long.
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return false
	}

	// Bucket names must begin and end with a letter or number.
	firstChar := []rune(bucketName[0:1])[0]
	lastChar := []rune(bucketName[len(bucketName)-1:])[0]
	if !unicode.IsLetter(firstChar) && !unicode.IsNumber(firstChar) {
		return false
	}
	if !unicode.IsLetter(lastChar) && !unicode.IsNumber(lastChar) {
		return false
	}

	// Bucket names must not start with the prefix 'xn--'
	if strings.HasPrefix(bucketName, "xn--") {
		return false
	}

	// Bucket names must not end with the suffix "-s3alias"
	if strings.HasSuffix(bucketName, "-s3alias") {
		return false
	}

	return true
}
