package provider

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/smithy-go"
)

// TODO: If user explicitly set profile name and we don't find creds, probably blow up
func HasCredentials(profile string) bool {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile(profile),
		config.WithEC2IMDSClientEnableState(imds.ClientDisabled), // Otherwise we wait 4 seconds to IMDSv2 to timeout
	)
	if err != nil {
		// TODO: log.debug(err)
		return false
	}

	_, credsErr := cfg.Credentials.Retrieve(context.TODO())
	if credsErr != nil {
		var oe *smithy.OperationError
		if errors.As(credsErr, &oe) {
			if !(oe.ServiceID == "ec2imds" && oe.OperationName == "GetMetadata") {
				// TODO: log.debug("something bad happened")
			}
			return false
		}
	}
	return true
}
