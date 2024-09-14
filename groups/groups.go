package groups

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const AuthUsersGroup = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
const AllUsersGroup = "http://acs.amazonaws.com/groups/global/AllUsers"

var AllUsersv2 = &types.Grantee{
	Type: types.TypeGroup,
	URI:  aws.String(AllUsersGroup)}

var AuthenticatedUsersv2 = &types.Grantee{
	Type: types.TypeGroup,
	URI:  aws.String(AuthUsersGroup)}

const AllUsersURI = "uri=" + AllUsersGroup
const AuthUsersURI = "uri=" + AuthUsersGroup
