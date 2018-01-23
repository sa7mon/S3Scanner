import s3utils as s3
import sh
import os
import subprocess as sp

def test_getBucketSize():
    """
    Scenario 1: Bucket doesn't exist
        Expected: 255

    Scenario 2: Bucket exists, listing open to public
        Expected:
            Size: 9.1 KiB
        Note:
            Using flaws.cloud as example by permission of owner (@0xdabbad00)

    """

    # Scenario 1
    try:
        result = s3.getBucketSize('example-this-hopefully-wont-exist-123123123')
    except sh.ErrorReturnCode_255:
        assert True

    # Scenario 3
    assert s3.getBucketSize('flaws.cloud') == "9.1 KiB"


def test_checkBucket():
    """
    Scenario 1: Bucket name exists, region is wrong
        Expected:
            Code: 301
            Region: Region returned depends on the closest S3 region to the user. Since we don't know this,
                    just assert for 2 hyphens.
        Note:
            Amazon should always give us a 301 to redirect to the nearest s3 endpoint.
            Currently uses the ap-south-1 (Asia Pacific - Mumbai) region, so if you're running
            the test near there, change to a region far from
            you - https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region

    Scenario 2: Bucket exists, region correct
        Expected:
            Code: 200
            Message: Contains the domain name and region
        Note:
            Using flaws.cloud as example by permission of owner (@0xdabbad00)

    """
    # Scenario 1
    result = s3.checkBucket('amazon.com', 'ap-south-1')
    assert result[0] == 301
    assert result[1].count("-") == 2

    # Scenario 2
    result = s3.checkBucket('flaws.cloud', 'us-west-2')
    assert result[0] == 200
    assert 'flaws.cloud:us-west-2' in result[1]


def run_sh(args):
    # Run the s3finder file with sh, passing in args to it
    dir_path = os.path.dirname(os.path.realpath(__file__))
    script = dir_path + "/s3finder.py"

    try:
        run1 = sh.python(script)
    except sh.ErrorReturnCode as e:
        return e.stderr.decode('utf-8'), e.stdout.decode('utf-8')


def test_arguments():

