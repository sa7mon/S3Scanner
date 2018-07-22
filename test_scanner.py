import s3utils as s3
import sh
import os
import sys
import shutil
import time
import logging


pyVersion = sys.version_info
# pyVersion[0] can be 2 or 3


s3scannerLocation = "./"
testingFolder = "./test/"

setupRan = False


def test_setup():
    """ Setup code to run before we run all tests. """
    global setupRan

    if setupRan:    # We only need to run this once per test-run
        return

    # Check if AWS creds are configured
    s3.awsCredsConfigured = s3.checkAwsCreds()

    print("--> AWS credentials configured: " + str(s3.awsCredsConfigured))

    # Create testingFolder if it doesn't exist
    if not os.path.exists(testingFolder) or not os.path.isdir(testingFolder):
        os.makedirs(testingFolder)

    setupRan = True


def test_arguments():
    """
    Scenario mainargs.1: No args supplied
    Scenario mainargs.2: --out-file
    Scenario mainargs.3: --list
    Scenario mainargs.4: --dump
    """

    test_setup()

    # mainargs.1
    try:
        sh.python(s3scannerLocation + 's3scanner.py')
    except sh.ErrorReturnCode as e:
        assert e.stderr.decode('utf-8') == ""
        assert "usage: s3scanner [-h] [-o OUTFILE] [-c] [-d] [-l] buckets" in e.stdout.decode('utf-8')

    # mainargs.2

    # Put one bucket into a new file
    with open(testingFolder + "mainargs.2_input.txt", "w") as f:
        f.write('flaws.cloud\n')

    try:
        sh.python(s3scannerLocation + 's3scanner.py', '--out-file', testingFolder + 'mainargs.2_output.txt',
                  testingFolder + 'mainargs.2_input.txt')

        with open(testingFolder + "mainargs.2_output.txt") as f:
            line = f.readline().strip()

        assert line == 'flaws.cloud'

    finally:
        # No matter what happens with the test, clean up the test files at the end
        try:
            os.remove(testingFolder + 'mainargs.2_output.txt')
            os.remove(testingFolder + 'mainargs.2_input.txt')
        except OSError:
            pass

    # mainargs.3
    # mainargs.4


def test_checkAcl():
    """
    Scenario checkAcl.1 - ACL listing enabled
        Expected:
            found = True
            acls = {'allUsers': ['READ', 'READ_ACP'], 'authUsers': ['READ', 'READ_ACP']}
    Scenario checkAcl.2 - AccessDenied for ACL listing
        Expected:
            found = True
            acls = 'AccessDenied'
    Scenario checkAcl.3 - Bucket access is disabled
        Expected:
            found = True
            acls = "AllAccessDisabled"
    Scenario checkAcl.4 - Bucket doesn't exist
        Expected:
            found = False
            acls = {}
    """
    test_setup()

    if not s3.awsCredsConfigured:  # Don't run tests if AWS creds aren't configured
        return

    # checkAcl.1
    r1 = s3.checkAcl('aneta')
    assert r1["found"] is True
    assert r1["acls"] == {'allUsers': ['READ', 'READ_ACP'], 'authUsers': ['READ', 'READ_ACP']}

    # checkAcl.2
    result = s3.checkAcl('flaws.cloud')
    assert result["found"] is True
    assert result["acls"] == "AccessDenied"

    # checkAcl.3
    result = s3.checkAcl('amazon.com')
    assert result["found"] is True
    assert result["acls"] == "AllAccessDisabled"

    # checkAcl.4
    result = s3.checkAcl('hopethisdoesntexist1234asdf')
    assert result["found"] is False
    assert result["acls"] == {}


def test_checkAwsCreds():
    """
    Scenario checkAwsCreds.1 - Output of checkAwsCreds() matches a more intense check for creds
    """
    test_setup()

    # Check more thoroughly for creds being set.
    vars = os.environ

    keyid = vars.get("AWS_ACCESS_KEY_ID")
    key = vars.get("AWS_SECRET_ACCESS_KEY")
    credsFile = os.path.expanduser("~") + "/.aws/credentials"

    if keyid is not None and len(keyid) == 20:
        if key is not None and len(key) == 40:
            credsActuallyConfigured = True
        else:
            credsActuallyConfigured = False
    else:
        credsActuallyConfigured = False

    if os.path.exists(credsFile):
        print("credsFile path exists")
        if not credsActuallyConfigured:
            keyIdSet = None
            keySet = None

            # Check the ~/.aws/credentials file
            with open(credsFile, "r") as f:
                for line in f:
                    line = line.strip()
                    if line[0:17].lower() == 'aws_access_key_id':
                        if len(line) >= 38:  # key + value = length of at least 38 if no spaces around equals
                            keyIdSet = True
                        else:
                            keyIdSet = False

                    if line[0:21].lower() == 'aws_secret_access_key':
                        if len(line) >= 62:
                            keySet = True
                        else:
                            keySet = False

            if keyIdSet and keySet:
                credsActuallyConfigured = True

    # checkAwsCreds.1
    assert s3.checkAwsCreds() == credsActuallyConfigured

def test_checkBucket():
    """
    checkBucket.1 - Bucket name
    checkBucket.2 - Domain name
    checkBucket.3 - Full s3 url
    checkBucket.4 - bucket:region
    """
    
    test_setup()
    
    testFile = './test/test_checkBucket.txt'
    
    # Create file logger
    flog = logging.getLogger('s3scanner-file')
    flog.setLevel(logging.DEBUG)              # Set log level for logger object
    
    # Create file handler which logs even debug messages
    fh = logging.FileHandler(testFile)
    fh.setLevel(logging.DEBUG)
    
    # Add the handler to logger
    flog.addHandler(fh)
    
    # Create secondary logger for logging to screen
    slog = logging.getLogger('s3scanner-screen')
    slog.setLevel(logging.CRITICAL)
    
    try:
        # checkBucket.1
        s3.checkBucket("flaws.cloud", slog, flog, False, False)
        
        # checkBucket.2
        s3.checkBucket("flaws.cloud.s3-us-west-2.amazonaws.com", slog, flog, False, False)
        
        # checkBucket.3
        s3.checkBucket("flaws.cloud:us-west-2", slog, flog, False, False)
        
        # Read in test loggin file and assert
        f = open(testFile, 'r')
        results = f.readlines()
        f.close()
        
        assert results[0].rstrip() == "flaws.cloud"
        assert results[1].rstrip() == "flaws.cloud"
        assert results[2].rstrip() == "flaws.cloud"
        
    finally:
        # Delete test file
        os.remove(testFile)

def test_checkBucketName():
    """
    Scenario checkBucketName.1 - Under length requirements
        Expected: False
    Scenario checkBucketName.2 - Over length requirements
        Expected: False
    Scenario checkBucketName.3 - Contains forbidden characters
        Expected: False
    Scenario checkBucketName.4 - Blank name
        Expected: False
    Scenario checkBucketName.5 - Good name
        Expected: True
    """
    test_setup()

    # checkBucketName.1
    result = s3.checkBucketName('ab')
    assert result is False

    # checkBucketName.2
    tooLong = "asdfasdf12834092834nMSdfnasjdfhu23y49u2y4jsdkfjbasdfbasdmn4asfasdf23423423423423"  # 80 characters
    result = s3.checkBucketName(tooLong)
    assert result is False

    # checkBucketName.3
    badBucket = "mycoolbucket:dev"
    assert s3.checkBucketName(badBucket) is False

    # checkBucketName.4
    assert s3.checkBucketName('') is False

    # checkBucketName.5
    assert s3.checkBucketName('arathergoodname') is True


def test_checkBucketWithoutCreds():
    """
    Scenario checkBucketwc.1 - Non-existent bucket
    Scenario checkBucketwc.2 - Good bucket
    Scenario checkBucketwc.3 - No public read perm
    """
    test_setup()

    if s3.awsCredsConfigured:
        return

    # checkBucketwc.1
    assert s3.checkBucketWithoutCreds('ireallyhopethisbucketdoesntexist') is False

    # checkBucketwc.2
    assert s3.checkBucketWithoutCreds('flaws.cloud') is True

    # checkBucketwc.3
    assert s3.checkBucketWithoutCreds('blog') is True


# def test_checkIncludeClosed():
#     """ Verify that the '--include-closed' argument is working correctly.
#         Expected:
#             The bucket name 'yahoo.com' is expected to exist, but be closed. The bucket name
#             and region should be included in the output buckets file in the format 'bucket:region'.
#     """
#     test_setup()
#
#     # Create a file called testing.txt and write 'yahoo.com' to it
#
#     inFile = testingFolder + 'test_checkIncludeClosed_in.txt'
#     outFile = testingFolder + 'test_checkIncludeClosed_out.txt'
#
#     f = open(inFile, 'w')
#     f.write('yahoo.com\n')  # python will convert \n to os.linesep
#     f.close()
#
#     sh.python(s3scannerLocation + "s3scanner.py", "--out-file", outFile, "--include-closed", inFile)
#
#     found = False
#     with open(outFile, 'r') as g:
#         for line in g:
#             if 'yahoo.com' in line:
#                 found = True
#
#     try:
#         assert found is True
#     finally:
#         # Cleanup testing files
#         os.remove(outFile)
#         os.remove(inFile)


def test_dumpBucket():
    """
    Scenario dumpBucket.1 - Public read permission enabled
        Expected: Dumping the bucket "flaws.cloud" should result in 6 files being downloaded into the buckets folder.
                    The expected file sizes of each file are listed in the 'expectedFiles' dictionary.
    Scenario dumpBucket.2 - Public read objects disabled
        Expected: The function returns false and the bucket directory doesn't exist
    Scenario dumpBucket.3 - Authenticated users read enabled, public users read disabled
        Expected: The function returns true and the bucket directory exists. Opposite for if no aws creds are set
    """
    test_setup()

    # dumpBucket.1

    s3.dumpBucket("flaws.cloud")

    dumpDir = './buckets/flaws.cloud/'  # Folder to look for the files in

    # Expected sizes of each file
    expectedFiles = {'hint1.html': 2575, 'hint2.html': 1707, 'hint3.html': 1101, 'index.html': 3082,
                     'robots.txt': 46, 'secret-dd02c7c.html': 1051, 'logo.png': 15979}

    try:
        # Assert number of files in the folder
        assert len(os.listdir(dumpDir)) == len(expectedFiles)

        # For each file, assert the size
        for file, size in expectedFiles.items():
            assert os.path.getsize(dumpDir + file) == size
    finally:
        # No matter what happens with the asserts, cleanup after the test by deleting the flaws.cloud directory
        shutil.rmtree(dumpDir)

    # dumpBucket.2
    assert s3.dumpBucket('app-dev') is False
    assert os.path.exists('./buckets/app-dev') is False

    # dumpBucket.3
    assert s3.dumpBucket('1904') is s3.awsCredsConfigured  # Asserts should both follow whether or not creds are set
    assert os.path.exists('./buckets/1904') is s3.awsCredsConfigured


def test_getBucketSize():
    """
    Scenario getBucketSize.2 - Public read enabled
        Expected: The flaws.cloud bucket returns size: 9.1KiB
    Scenario getBucketSize.3 - Public read disabled
        Expected: app-dev bucket has public read permissions disabled
    Scenario getBucketSize.4 - Bucket doesn't exist
        Expected: We should get back "NoSuchBucket"
    """
    test_setup()

    # getBucketSize.2
    assert s3.getBucketSize('flaws.cloud') == "24.9 KiB"

    # getBucketSize.3
    assert s3.getBucketSize('app-dev') == "AccessDenied"

    # getBucketSize.4
    assert s3.getBucketSize('thiswillprobablynotexistihope') == "NoSuchBucket"


def test_getBucketSizeTimeout():
    """
    Scenario getBucketSize.1 - Too many files to list so it times out
        Expected: The function returns a timeout error after the specified wait time
        Note: Use e27.co to test with. Verify that getBucketSize returns an unknown size and doesn't take longer
        than sizeCheckTimeout set in s3utils
    """
    test_setup()

    s3.awsCredsConfigured = False

    startTime = time.time()

    output = s3.getBucketSize("e27.co")
    duration = time.time() - startTime

    # Assert that getting the bucket size took less than or equal to the alloted time plus 1 second to account
    # for processing time.
    assert duration <= s3.sizeCheckTimeout + 1
    assert output == "Unknown Size - timeout"


def test_listBucket():
    """
    Scenario listBucket.1 - Public read enabled
        Expected: Listing bucket flaws.cloud will create the directory, create flaws.cloud.txt, and write the listing to file
    Scenario listBucket.2 - Public read disabled

    """
    test_setup()

    # listBucket.1

    listFile = './list-buckets/flaws.cloud.txt'

    s3.listBucket('flaws.cloud')

    assert os.path.exists(listFile)              # Assert file was created in the correct location

    lines = []
    with open(listFile, 'r') as g:
        for line in g:
            lines.append(line)

    assert lines[0][26:41] == '2575 hint1.html'  # Assert the first line is correct
    assert len(lines) == 7                       # Assert number of lines in the file is correct

    # listBucket.2
    assert s3.listBucket('app-dev') == "AccessDenied"
