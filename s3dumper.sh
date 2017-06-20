#!/bin/bash

#############################################################################
#
# testS3.sh - Test if S3 buckets are open to everyone and dump if they are
#
# Author: Dan Salmon (twitter.com/bltjetpack, github.com/sa7mon)
# Created: 6/20/2017
#
# Requirements: aws-cli [pip install awscli]
# 
#############################################################################

# Actually obey ctrl+c while inside while loop
trap trapint 2
function trapint {
    exit 0
}

# Read file with lines in format:
# domain:s3region
#
# i.e.  google.com:us-west-2
while read line
do
    # Do what you want to $domain
    domain=`echo $line | awk -F ':' '{print $1}'`
    region=`echo $line | awk -F ':' '{print $2}'`

    echo "[$domain] Checking..."
    awsls=`aws s3 ls s3://$domain --no-sign-request --region $region 2>&1`

    # Check if we get an "AccessDenied" error while attempting to 
    # list the bucket's contents

    if [[ $awsls == *"AccessDenied"* ]] 
    then
  		echo "[$domain]    Nope. Access denied."
  	else
  		echo "[$domain]    Seems good!"
  		# Create directory if it doesn't already exist
  		if [ ! -d "./buckets/$domain" ]; then
		  mkdir -p "./buckets/$domain";
		fi

		# Download bucket into directory
		aws s3 sync s3://$domain ./buckets/$domain/ --no-sign-request --region $region
	fi
done < "domains.test"