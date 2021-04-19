from s3Bucket import s3Bucket, s3BucketObject, Permission

"""
Tests for s3Bucket class go here
"""


def test_invalid_bucket_name():
    try:
        b = s3Bucket(name="asdf,;0()")
    except ValueError as ve:
        if str(ve) != "Invalid bucket name":
            raise ve


def test_s3_bucket_object():
    o1 = s3BucketObject(key='index.html', size=8096, last_modified='2018-03-02T08:10:25.000Z')
    o2 = s3BucketObject(key='home.html', size=2, last_modified='2018-03-02T08:10:25.000Z')

    assert o1 != o2
    assert o2 < o1  # test __lt__ method which compares keys
    assert str(o1) == "Key: index.html, Size: 8096, LastModified: 2018-03-02T08:10:25.000Z"
    assert o1.getHumanReadableSize() == "7.9KB"


def test_check_bucket_name():
    b = s3Bucket(name="asdfasdf.s3.amazonaws.com")
    b2 = s3Bucket(name="asdf:us-west-1")


def test_get_human_readable_permissions():
    b = s3Bucket(name='asdf')
    b.AllUsersRead = Permission.ALLOWED
    b.AllUsersWrite = Permission.ALLOWED
    b.AllUsersReadACP = Permission.ALLOWED
    b.AllUsersWriteACP = Permission.ALLOWED
    b.AuthUsersRead = Permission.ALLOWED
    b.AuthUsersWrite = Permission.ALLOWED
    b.AuthUsersReadACP = Permission.ALLOWED
    b.AuthUsersWriteACP = Permission.ALLOWED

    b.getHumanReadablePermissions()

    b.AllUsersFullControl = Permission.ALLOWED
    b.AuthUsersFullControl = Permission.ALLOWED

    b.getHumanReadablePermissions()

