# Permissions

S3Scanner will attempt to get all available information about a bucket, but it's up to you to interpret the results.

[Possible permissions](https://docs.aws.amazon.com/AmazonS3/latest/user-guide/set-bucket-permissions.html) for buckets:

* Read - List and view all files
* Write - Write files to bucket
* Read ACP - Read all Access Control Policies attached to bucket
* Write ACP - Write Access Control Policies to bucket
* Full Control - All above permissions

Any or all of these permissions can be set for the 2 main user groups:

* Authenticated Users
* Public Users (those without AWS credentials set)

The permissions can also be set on individual users, groups, and objects, but that is out of the scope of this tool and will require manual investigation.

**What this means**

There are many possible combinations of permissions that can be set on buckets. This tool will attempt to check as many as possible.