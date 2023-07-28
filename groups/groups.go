package groups

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var AllUsersv2 = &types.Grantee{
	Type: types.TypeGroup,
	URI:  aws.String("http://acs.amazonaws.com/groups/global/AllUsers")}

var AuthenticatedUsersv2 = &types.Grantee{
	Type: types.TypeGroup,
	URI:  aws.String("http://acs.amazonaws.com/groups/global/AllUsers")}

const ALL_USERS_URI = "uri=http://acs.amazonaws.com/groups/global/AllUsers"
const AUTH_USERS_URI = "uri=http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
