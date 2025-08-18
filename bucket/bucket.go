package bucket

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/sa7mon/s3scanner/groups"
	log "github.com/sirupsen/logrus"
)

var BucketExists = uint8(1)
var BucketNotExist = uint8(0)
var BucketExistsUnknown = uint8(2)

var PermissionAllowed = uint8(1)
var PermissionDenied = uint8(0)
var PermissionUnknown = uint8(2)

var bucketRe = regexp.MustCompile(`[^.\-a-z0-9]`)

type Bucket struct {
	//gorm.Model
	ID                uint      `gorm:"primarykey" json:",omitempty"`
	Name              string    `json:"name" gorm:"name;size:64;index"`
	Region            string    `json:"region" gorm:"size:20"`
	Exists            uint8     `json:"exists"`
	DateScanned       time.Time `json:"date_scanned"`
	Objects           []Object  `json:"objects"`
	ObjectsEnumerated bool      `json:"objects_enumerated"`
	Provider          string    `json:"provider"`
	NumObjects        int32     `json:"num_objects"`

	// Total size of all bucket objects in bytes
	BucketSize       uint64 `json:"bucket_size"`
	OwnerID          string `json:"owner_id"`
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

type Object struct {
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

func (b *Bucket) String() string {
	if b.Exists == BucketNotExist {
		return fmt.Sprintf("%v | bucket_not_exist", b.Name)
	}

	var authUserPerms []string
	if b.PermAuthUsersRead == PermissionAllowed {
		authUserPerms = append(authUserPerms, "READ")
	}
	if b.PermAuthUsersWrite == PermissionAllowed {
		authUserPerms = append(authUserPerms, "WRITE")
	}
	if b.PermAuthUsersReadACL == PermissionAllowed {
		authUserPerms = append(authUserPerms, "READ_ACP")
	}
	if b.PermAuthUsersWriteACL == PermissionAllowed {
		authUserPerms = append(authUserPerms, "WRITE_ACP")
	}
	if b.PermAuthUsersFullControl == PermissionAllowed {
		authUserPerms = append(authUserPerms, "FULL_CONTROL")
	}

	var allUsersPerms []string
	if b.PermAllUsersRead == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "READ")
	}
	if b.PermAllUsersWrite == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "WRITE")
	}
	if b.PermAllUsersReadACL == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "READ_ACP")
	}
	if b.PermAllUsersWriteACL == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "WRITE_ACP")
	}
	if b.PermAllUsersFullControl == PermissionAllowed {
		allUsersPerms = append(allUsersPerms, "FULL_CONTROL")
	}

	return fmt.Sprintf("AuthUsers: [%v] | AllUsers: [%v]", strings.Join(authUserPerms, ", "), strings.Join(allUsersPerms, ", "))
}

func (b *Bucket) Permissions() map[*types.Grantee]map[string]uint8 {
	return map[*types.Grantee]map[string]uint8{
		groups.AllUsersv2: {
			"READ":         b.PermAllUsersRead,
			"WRITE":        b.PermAllUsersWrite,
			"READ_ACP":     b.PermAllUsersReadACL,
			"WRITE_ACP":    b.PermAllUsersWriteACL,
			"FULL_CONTROL": b.PermAllUsersFullControl,
		},
		groups.AuthenticatedUsersv2: {
			"READ":         b.PermAuthUsersRead,
			"WRITE":        b.PermAuthUsersWrite,
			"READ_ACP":     b.PermAuthUsersReadACL,
			"WRITE_ACP":    b.PermAuthUsersWriteACL,
			"FULL_CONTROL": b.PermAuthUsersFullControl,
		},
	}
}

func FromReader(r io.Reader, bucketChan chan Bucket) error {
	scanner := bufio.NewScanner(r)
	bucketsSeen := make(map[string]struct{})
	for scanner.Scan() {
		bucketName := strings.TrimSpace(scanner.Text())
		if !IsValidS3BucketName(bucketName) {
			log.Info(fmt.Sprintf("invalid   | %s", bucketName))
			continue
		}
		bucketName = strings.ToLower(bucketName)
		if _, seen := bucketsSeen[bucketName]; seen {
			continue
		}
		bucketsSeen[bucketName] = struct{}{}
		bucketChan <- NewBucket(bucketName)
	}

	if ferr := scanner.Err(); ferr != nil {
		return ferr
	}
	return nil
}

func ReadFromFile(bucketFile string, bucketChan chan Bucket) error {
	file, err := os.Open(bucketFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := FromReader(file, bucketChan); err != nil {
		return err
	}
	return nil
}

// ParseACLOutputV2 TODO: probably move this to providers.go
func (b *Bucket) ParseACLOutputV2(aclOutput *s3.GetBucketAclOutput) error {
	b.OwnerID = *aclOutput.Owner.ID
	if aclOutput.Owner.DisplayName != nil {
		b.OwnerDisplayName = *aclOutput.Owner.DisplayName
	}
	// Since we can read the permissions, there should be no unknowns. Set all to denied, then read each grant and
	// set the corresponding permission to allowed.
	b.DenyAll()

	for _, g := range aclOutput.Grants {
		if g.Grantee != nil && g.Grantee.Type == "Group" && *g.Grantee.URI == groups.AllUsersGroup {
			switch g.Permission {
			case types.PermissionRead:
				b.PermAllUsersRead = PermissionAllowed
			case types.PermissionWrite:
				b.PermAllUsersWrite = PermissionAllowed
			case types.PermissionReadAcp:
				b.PermAllUsersReadACL = PermissionAllowed
			case types.PermissionWriteAcp:
				b.PermAllUsersWriteACL = PermissionAllowed
			case types.PermissionFullControl:
				b.PermAllUsersFullControl = PermissionAllowed
			default:
				break
			}
		}
		if g.Grantee != nil && g.Grantee.Type == "Group" && *g.Grantee.URI == groups.AuthUsersGroup {
			switch g.Permission {
			case types.PermissionRead:
				b.PermAuthUsersRead = PermissionAllowed
			case types.PermissionWrite:
				b.PermAuthUsersWrite = PermissionAllowed
			case types.PermissionReadAcp:
				b.PermAuthUsersReadACL = PermissionAllowed
			case types.PermissionWriteAcp:
				b.PermAuthUsersWriteACL = PermissionAllowed
			case types.PermissionFullControl:
				b.PermAuthUsersFullControl = PermissionAllowed
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
	}
	return PermissionDenied
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

func (b *Bucket) DenyAll() {
	b.PermAllUsersRead = PermissionDenied
	b.PermAllUsersWrite = PermissionDenied
	b.PermAllUsersReadACL = PermissionDenied
	b.PermAllUsersWriteACL = PermissionDenied
	b.PermAllUsersFullControl = PermissionDenied
	b.PermAuthUsersRead = PermissionDenied
	b.PermAuthUsersWrite = PermissionDenied
	b.PermAuthUsersReadACL = PermissionDenied
	b.PermAuthUsersWriteACL = PermissionDenied
	b.PermAuthUsersFullControl = PermissionDenied
}
