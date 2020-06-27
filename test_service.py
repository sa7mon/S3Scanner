import os
import s3Bucket
from S3Service import S3Service
from s3Bucket import BucketExists, Permission
from TestUtils import TestBucketService
from exceptions import AccessDeniedException

s3scannerLocation = "./"
testingFolder = "./test/"
setupRan = False


"""
S3Service.py methods to test:

- init()
    - Test service.aws_creds_configured is false when forceNoCreds = False and vice-versa
- check_bucket_exists()
    - ✔️ Test against that exists
    - ✔️ Test against one that doesn't
    - Test previous scenarios with no creds   
- check_perm_read_acl()
    - ✔️ Test against bucket with AllUsers allowed
    - ✔️ Test against bucket with AuthUsers allowed
    - ✔️ Test against bucket with all denied 
- check_perm_read()
    - ✔️ Test against bucket with AuthUsers read permission
    - ✔️ Test against bucket with AllUsers read permission
    - ✔️ Test against bucket with no read permission
- check_perm_write()
    - ✔️ Test against bucket with no write permissions
    - ✔️ Test against bucket with AuthUsers write permission
    - ✔️ Test against bucket with AllUsers write permission
    - ✔️ Test against bucket with AllUsers and AuthUsers write permission
- check_perm_write_acl()
    - ✔️ Test against bucket with AllUsers allowed
    - ✔️ Test against bucket with AuthUsers allowed
    - Test against bucket with both AllUsers allowed
    - ✔️ Test against bucket with no groups allowed
- enumerate_bucket_objects()
    - ✔️ Test against empty bucket
    - ✔️ Test against not empty bucket with read permission
    - ✔️ Test against bucket without read permission
- parse_found_acl()
    - Test against JSON with FULL_CONTROL for AllUsers
    - Test against JSON with FULL_CONTROL for AuthUsers
    - Test against empty JSON
    - Test against JSON with ReadACP for AuthUsers and Write for AllUsers
"""


def test_setup_new():
    global setupRan
    if setupRan:    # We only need to run this once per test-run
        return

    # Create testingFolder if it doesn't exist
    if not os.path.exists(testingFolder) or not os.path.isdir(testingFolder):
        os.makedirs(testingFolder)
    setupRan = True


def test_bucket_exists():
    test_setup_new()

    s = S3Service()

    # Bucket that does exist
    b1 = s3Bucket.s3Bucket('s3scanner-private')
    s.check_bucket_exists(b1)
    assert b1.exists is BucketExists.YES

    # Bucket that doesn't exist (hopefully)
    b2 = s3Bucket.s3Bucket('asfasfasdfasdfasdf')
    s.check_bucket_exists(b2)
    assert b2.exists is BucketExists.NO


def test_check_perm_read():
    test_setup_new()

    s = S3Service()

    # Bucket that no one can list
    b1 = s3Bucket.s3Bucket('s3scanner-private')
    b1.exists = BucketExists.YES
    s.check_perm_read(b1)
    if s.aws_creds_configured:
        assert b1.AuthUsersRead == Permission.DENIED
    else:
        assert b1.AllUsersRead == Permission.DENIED

    # Bucket that only AuthenticatedUsers can list
    b2 = s3Bucket.s3Bucket('s3scanner-auth-read')
    b2.exists = BucketExists.YES
    s.check_perm_read(b2)
    if s.aws_creds_configured:
        assert b2.AuthUsersRead == Permission.ALLOWED
    else:
        assert b2.AllUsersRead == Permission.DENIED

    # Bucket that Everyone can list
    b3 = s3Bucket.s3Bucket('s3scanner-long')
    b3.exists = BucketExists.YES
    s.check_perm_read(b3)
    if s.aws_creds_configured:
        assert b3.AuthUsersRead == Permission.ALLOWED
    else:
        assert b3.AllUsersRead == Permission.ALLOWED


def test_enumerate_bucket_objects():
    test_setup_new()

    s = S3Service()

    # Empty bucket
    b1 = s3Bucket.s3Bucket('s3scanner-empty')
    b1.exists = BucketExists.YES
    s.check_perm_read(b1)
    if s.aws_creds_configured:
        assert b1.AuthUsersRead == Permission.ALLOWED
    else:
        assert b1.AllUsersRead == Permission.ALLOWED
    s.enumerate_bucket_objects(b1)
    assert b1.objects_enumerated is True
    assert b1.bucketSize == 0

    # Bucket with > 1000 items
    if s.aws_creds_configured:
        b2 = s3Bucket.s3Bucket('s3scanner-auth-read')
        b2.exists = BucketExists.YES
        s.check_perm_read(b2)
        assert b2.AuthUsersRead == Permission.ALLOWED
        s.enumerate_bucket_objects(b2)
        assert b2.objects_enumerated is True
        assert b2.bucketSize == 4143
        assert b2.getHumanReadableSize() == "4.0KB"
    else:
        print("[test_enumerate_bucket_objects] Skipping test due to no AWS creds")

    # Bucket without read permission
    b3 = s3Bucket.s3Bucket('s3scanner-private')
    b3.exists = BucketExists.YES
    s.check_perm_read(b3)
    if s.aws_creds_configured:
        assert b3.AuthUsersRead == Permission.DENIED
    else:
        assert b3.AllUsersRead == Permission.DENIED
    try:
        s.enumerate_bucket_objects(b3)
    except AccessDeniedException:
        pass


def test_check_perm_read_acl():
    test_setup_new()
    s = S3Service()

    # Bucket with no read ACL perms
    b1 = s3Bucket.s3Bucket('s3scanner-private')
    b1.exists = BucketExists.YES
    s.check_perm_read_acl(b1)
    if s.aws_creds_configured:
        assert b1.AuthUsersReadACP == Permission.DENIED
    else:
        assert b1.AllUsersReadACP == Permission.DENIED

    # Bucket that allows AuthenticatedUsers to read ACL
    if s.aws_creds_configured:
        b2 = s3Bucket.s3Bucket('s3scanner-auth-read-acl')
        b2.exists = BucketExists.YES
        s.check_perm_read_acl(b2)
        if s.aws_creds_configured:
            assert b2.AuthUsersReadACP == Permission.ALLOWED
        else:
            assert b2.AllUsersReadACP == Permission.DENIED

    # Bucket that allows AllUsers to read ACL
    b3 = s3Bucket.s3Bucket('s3scanner-all-readacp')
    b3.exists = BucketExists.YES
    s.check_perm_read_acl(b3)
    assert b3.AllUsersReadACP == Permission.ALLOWED
    assert b3.AllUsersWrite == Permission.DENIED
    assert b3.AllUsersWriteACP == Permission.DENIED
    assert b3.AuthUsersReadACP == Permission.DENIED
    assert b3.AuthUsersWriteACP == Permission.DENIED
    assert b3.AuthUsersWrite == Permission.DENIED


def test_check_perm_write(do_dangerous_test):
    test_setup_new()
    s = S3Service()
    sAnon = S3Service(forceNoCreds=True)

    # Bucket with no write perms
    b1 = s3Bucket.s3Bucket('flaws.cloud')
    b1.exists = BucketExists.YES
    s.check_perm_write(b1)

    if s.aws_creds_configured:
        assert b1.AuthUsersWrite == Permission.DENIED
    else:
        assert b1.AllUsersWrite == Permission.DENIED

    if do_dangerous_test:
        print("[test_check_perm_write] Doing dangerous test")
        ts = TestBucketService()

        danger_bucket_1 = ts.create_bucket(1)  # Bucket with AuthUser Write, WriteACP permissions
        try:
            b2 = s3Bucket.s3Bucket(danger_bucket_1)
            b2.exists = BucketExists.YES
            sAnon.check_perm_write(b2)
            s.check_perm_write(b2)
            assert b2.AuthUsersWrite == Permission.ALLOWED
            assert b2.AllUsersWrite == Permission.DENIED
        finally:
            ts.delete_bucket(danger_bucket_1)
        try:
            danger_bucket_2 = ts.create_bucket(2)  # Bucket with AllUser Write, WriteACP permissions
            b3 = s3Bucket.s3Bucket(danger_bucket_2)
            b3.exists = BucketExists.YES
            sAnon.check_perm_write(b3)
            s.check_perm_write(b3)
            assert b3.AllUsersWrite == Permission.ALLOWED
            assert b3.AuthUsersWrite == Permission.UNKNOWN
        finally:
            ts.delete_bucket(danger_bucket_2)

        # Bucket with AllUsers and AuthUser Write permissions
        danger_bucket_4 = ts.create_bucket(4)
        try:
            b4 = s3Bucket.s3Bucket(danger_bucket_4)
            b4.exists = BucketExists.YES
            sAnon.check_perm_write(b4)
            s.check_perm_write(b4)
            assert b4.AllUsersWrite == Permission.ALLOWED
            assert b4.AuthUsersWrite == Permission.UNKNOWN
        finally:
            ts.delete_bucket(danger_bucket_4)
    else:
        print("[test_check_perm_write] Skipping dangerous test")


def test_check_perm_write_acl(do_dangerous_test):
    test_setup_new()
    s = S3Service()
    sNoCreds = S3Service(forceNoCreds=True)

    # Bucket with no permissions
    b1 = s3Bucket.s3Bucket('s3scanner-private')
    b1.exists = BucketExists.YES
    s.check_perm_write_acl(b1)
    if s.aws_creds_configured:
        assert b1.AuthUsersWriteACP == Permission.DENIED
        assert b1.AllUsersWriteACP == Permission.UNKNOWN
    else:
        assert b1.AllUsersWriteACP == Permission.DENIED
        assert b1.AuthUsersWriteACP == Permission.UNKNOWN
    
    if do_dangerous_test:
        print("[test_check_perm_write_acl] Doing dangerous tests...")
        ts = TestBucketService()

        # Bucket with WRITE_ACP enabled for AuthUsers
        try:
            danger_bucket_3 = ts.create_bucket(3)
            b2 = s3Bucket.s3Bucket(danger_bucket_3)
            b2.exists = BucketExists.YES

            # Check for read/write permissions so when we check for write_acl we
            # send the same perms that it had originally
            sNoCreds.check_perm_read(b2)
            s.check_perm_read(b2)
            sNoCreds.check_perm_write(b2)
            s.check_perm_write(b2)

            # Check for WriteACP
            sNoCreds.check_perm_write_acl(b2)
            s.check_perm_write_acl(b2)

            # Grab permissions after our check so we can compare to original
            sNoCreds.check_perm_write(b2)
            s.check_perm_write(b2)
            sNoCreds.check_perm_read(b2)
            s.check_perm_read(b2)
            if s.aws_creds_configured:
                assert b2.AuthUsersWriteACP == Permission.ALLOWED

                # Make sure we didn't change the original permissions
                assert b2.AuthUsersWrite == Permission.ALLOWED
                assert b2.AllUsersWrite == Permission.DENIED
                assert b2.AllUsersRead == Permission.ALLOWED
                assert b2.AuthUsersRead == Permission.UNKNOWN
            else:
                assert b2.AllUsersRead == Permission.ALLOWED
                assert b2.AuthUsersWriteACP == Permission.UNKNOWN
        except Exception as e:
            raise e
        finally:
            ts.delete_bucket(danger_bucket_3)

        # Bucket with WRITE_ACP enabled for AllUsers
        try:
            danger_bucket_2 = ts.create_bucket(2)
            b3 = s3Bucket.s3Bucket(danger_bucket_2)
            b3.exists = BucketExists.YES
            sNoCreds.check_perm_read(b3)
            s.check_perm_read(b3)
            sNoCreds.check_perm_write(b3)
            s.check_perm_write(b3)
            sNoCreds.check_perm_write_acl(b3)
            s.check_perm_write_acl(b3)
            sNoCreds.check_perm_write(b3)
            s.check_perm_write(b3)
            sNoCreds.check_perm_read(b3)
            s.check_perm_read(b3)
            if s.aws_creds_configured:
                assert b3.AllUsersWriteACP == Permission.ALLOWED
                assert b3.AuthUsersWriteACP == Permission.UNKNOWN
                assert b3.AllUsersWrite == Permission.ALLOWED
            else:
                assert b3.AllUsersRead == Permission.ALLOWED
                assert b3.AuthUsersWriteACP == Permission.UNKNOWN
        except Exception as e:
            raise e
        finally:
            ts.delete_bucket(danger_bucket_2)
    else:
        print("[test_check_perm_write_acl] Skipping dangerous test...")


def test_parsing_found_acl():
    test_setup_new()
    s = S3Service()

    b1 = s3Bucket.s3Bucket('s3scanner-all-read-readacl')
    b1.exists = BucketExists.YES
    s.check_perm_read_acl(b1)
    s.parse_found_acl(b1)

    assert b1.foundACL is not None
    assert b1.AllUsersRead == Permission.ALLOWED
    assert b1.AllUsersReadACP == Permission.ALLOWED
    assert b1.AllUsersWrite == Permission.DENIED
    assert b1.AllUsersWriteACP == Permission.DENIED
    assert b1.AllUsersFullControl == Permission.DENIED

    assert b1.AuthUsersReadACP == Permission.DENIED
    assert b1.AuthUsersRead == Permission.DENIED
    assert b1.AuthUsersWrite == Permission.DENIED
    assert b1.AuthUsersWriteACP == Permission.DENIED
    assert b1.AuthUsersFullControl == Permission.DENIED