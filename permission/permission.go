package permission

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	. "github.com/sa7mon/s3scanner/bucket"
	. "github.com/sa7mon/s3scanner/groups"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func CheckPermReadACL(s3Client *s3.Client, bucket *Bucket) (bool, error) {
	aclOutput, err := s3Client.GetBucketAcl(context.TODO(),
		&s3.GetBucketAclInput{Bucket: &bucket.Name})
	if err != nil {
		log.WithFields(log.Fields{
			"bucket_name": bucket.Name,
			"method":      "permission.CheckPermReadACL()",
		}).Debugf("error getting ACL: %v", err)
		var re *awshttp.ResponseError
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 403 {
				//fmt.Println("Access Denied!")
				return false, nil
			} else {
				return false, err
			}
		}
		return false, err
	}
	err = bucket.ParseAclOutputv2(aclOutput)
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckPermWriteAcl(svc *s3.Client, b *Bucket) (bool, error) {
	// TODO: Ensure bucket exists
	// TODO: Make sure this works with a bucket that allows PutACL. 400's returned always right now, is that because no creds?

	grants := map[string][]string{}
	if b.PermAuthUsersFullControl == PermissionAllowed {
		grants["FULL_CONTROL"] = append(grants["FULL_CONTROL"], AuthUsersUri)
	}
	if b.PermAuthUsersWriteACL == PermissionAllowed {
		grants["WRITE_ACP"] = append(grants["WRITE_ACP"], AuthUsersUri)
	}
	if b.PermAuthUsersWrite == PermissionAllowed {
		grants["WRITE"] = append(grants["WRITE"], AuthUsersUri)
	}
	if b.PermAuthUsersReadACL == PermissionAllowed {
		grants["READ_ACP"] = append(grants["READ_ACP"], AuthUsersUri)
	}
	if b.PermAuthUsersRead == PermissionAllowed {
		grants["READ"] = append(grants["READ"], AuthUsersUri)
	}

	if b.PermAllUsersFullControl == PermissionAllowed {
		grants["FULL_CONTROL"] = append(grants["FULL_CONTROL"], AllUsersUri)
	}
	if b.PermAllUsersWriteACL == PermissionAllowed {
		grants["WRITE_ACP"] = append(grants["WRITE_ACP"], AllUsersUri)
	}
	if b.PermAllUsersWrite == PermissionAllowed {
		grants["WRITE"] = append(grants["WRITE"], AllUsersUri)
	}
	if b.PermAllUsersReadACL == PermissionAllowed {
		grants["READ_ACP"] = append(grants["READ_ACP"], AllUsersUri)
	}
	if b.PermAllUsersRead == PermissionAllowed {
		grants["READ"] = append(grants["READ"], AllUsersUri)
	}

	_, err := svc.PutBucketAcl(context.TODO(), &s3.PutBucketAclInput{
		Bucket:           &b.Name,
		GrantFullControl: aws.String(strings.Join(grants["FULL_CONTROL"], ",")),
		GrantWriteACP:    aws.String(strings.Join(grants["WRITE_ACP"], ",")),
		GrantWrite:       aws.String(strings.Join(grants["WRITE"], ",")),
		GrantReadACP:     aws.String(strings.Join(grants["READ_ACP"], ",")),
		GrantRead:        aws.String(strings.Join(grants["READ"], ",")),
	})
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 400 || re.HTTPStatusCode() == 403 {
				//fmt.Println("Access Denied!")
				return false, nil
			} else {
				return false, err
			}
		}
		return false, err
	}

	return true, nil
}

func CheckPermWrite(svc *s3.Client, bucket *Bucket) (bool, error) {
	// TODO: Ensure bucket exists
	// TODO: What happens if we fail to clean up temp file

	// Try to put an object with a unique name and no body
	timestampFile := fmt.Sprintf("%v_%v.txt", time.Now().Unix(), bucket.Name)
	_, err := svc.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket.Name),
		Key:    &timestampFile,
		Body:   nil,
	})
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 403 { // No permission
				return false, nil
			} else {
				return false, err
			}
		}
	}

	// Clean up temporary file if it was successful
	_, err = svc.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket.Name),
		Key:    &timestampFile,
	})
	if err != nil {
		return true, err
	}

	return true, nil
}

func CheckPermRead(svc *s3.Client, bucket *Bucket) (bool, error) {
	_, err := svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{Bucket: &bucket.Name})
	if err != nil {
		log.Debugf("[%v][CheckPermRead] err: %v", bucket.Name, err)
		var re *awshttp.ResponseError
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 403 { // No permission
				return false, nil
			} else {
				return false, fmt.Errorf("[CheckPermRead] %s : %s : %w", bucket.Name, bucket.Region, err)
			}
		}
		return false, fmt.Errorf("[CheckPermRead] %s : %s : %w", bucket.Name, bucket.Region, err)
	}
	return true, nil
}
