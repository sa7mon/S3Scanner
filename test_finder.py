import s3utils as s3
import sh


def test_getBucketSize():
    """
    Scenario 1: Bucket doesn't exist
        Expected: 255

    Scenario 2: Bucket exists, region correct, listing denied

    Scenario 3: Bucket exists, region correct, listing open to public

    """

    # Scenario 1
    try:
        s3.getBucketSize('example-this-hopefully-wont-exist-123123123')
    except sh.ErrorReturnCode_255:
        assert True


def test_checkBucket():
    """
    Scenario 1: Bucket name exists, region is wrong
        Expected:
            Code: 301
            Region: Region returned depends on the closest S3 region to the user. Since we don't know this,
                    just assert for 2 hyphens.

        Note:
            Amazon should always give us a 301 to redirect to the nearest s3 endpoint.
            Currently uses the ap-south-1 (Asia Pacific - Mumbai) region, so if you're running the test
            near there, change to a region far from you like us-west-1.

    Scenario 2: Bucket exists, region correct

    """
    # Scenario 1
    result = s3.checkBucket('amazon.com', 'ap-south-1')
    assert result[0] == 301
    assert result[1].count("-") == 2

    # Scenario 2
