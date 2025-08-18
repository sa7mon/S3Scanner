package permission

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/groups"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func CheckPermReadACL(s3Client *s3.Client, bucket *bucket.Bucket) (bool, error) {
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
			}
			return false, err
		}
		return false, err
	}
	err = bucket.ParseACLOutputV2(aclOutput)
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckPermWriteACL(svc *s3.Client, b *bucket.Bucket) (bool, error) {
	// TODO: Ensure bucket exists
	// TODO: Make sure this works with a bucket that allows PutACL. 400's returned always right now, is that because no creds?

	grants := map[string][]string{}
	if b.PermAuthUsersFullControl == bucket.PermissionAllowed {
		grants["FULL_CONTROL"] = append(grants["FULL_CONTROL"], groups.AuthUsersURI)
	}
	if b.PermAuthUsersWriteACL == bucket.PermissionAllowed {
		grants["WRITE_ACP"] = append(grants["WRITE_ACP"], groups.AuthUsersURI)
	}
	if b.PermAuthUsersWrite == bucket.PermissionAllowed {
		grants["WRITE"] = append(grants["WRITE"], groups.AuthUsersURI)
	}
	if b.PermAuthUsersReadACL == bucket.PermissionAllowed {
		grants["READ_ACP"] = append(grants["READ_ACP"], groups.AuthUsersURI)
	}
	if b.PermAuthUsersRead == bucket.PermissionAllowed {
		grants["READ"] = append(grants["READ"], groups.AuthUsersURI)
	}

	if b.PermAllUsersFullControl == bucket.PermissionAllowed {
		grants["FULL_CONTROL"] = append(grants["FULL_CONTROL"], groups.AllUsersURI)
	}
	if b.PermAllUsersWriteACL == bucket.PermissionAllowed {
		grants["WRITE_ACP"] = append(grants["WRITE_ACP"], groups.AllUsersURI)
	}
	if b.PermAllUsersWrite == bucket.PermissionAllowed {
		grants["WRITE"] = append(grants["WRITE"], groups.AllUsersURI)
	}
	if b.PermAllUsersReadACL == bucket.PermissionAllowed {
		grants["READ_ACP"] = append(grants["READ_ACP"], groups.AllUsersURI)
	}
	if b.PermAllUsersRead == bucket.PermissionAllowed {
		grants["READ"] = append(grants["READ"], groups.AllUsersURI)
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
			}
			return false, err
		}
		return false, err
	}

	return true, nil
}

func CheckPermWrite(svc *s3.Client, bucket *bucket.Bucket) (bool, error) {
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
			}
			return false, err
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

func CheckPermRead(svc *s3.Client, bucket *bucket.Bucket) (bool, error) {
	_, err := svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{Bucket: &bucket.Name})
	if err != nil {
		log.Debugf("[%v][CheckPermRead] err: %v", bucket.Name, err)
		var re *awshttp.ResponseError
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 403 { // No permission
				return false, nil
			}
			return false, fmt.Errorf("[CheckPermRead] %s : %s : %w", bucket.Name, bucket.Region, err)
		}
		return false, fmt.Errorf("[CheckPermRead] %s : %s : %w", bucket.Name, bucket.Region, err)
	}
	return true, nil
}
