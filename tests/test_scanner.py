import sys
import subprocess
import os
import time
import shutil

from S3Scanner.S3Service import S3Service


def test_arguments():
    s = S3Service()

    a = subprocess.run([sys.executable, '-m', 'S3Scanner', '--version'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert a.stdout.decode('utf-8').strip() == '2.0.2'

    b = subprocess.run([sys.executable, '-m', 'S3Scanner', 'scan', '--bucket', 'flaws.cloud'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert_scanner_output(s, 'flaws.cloud | bucket_exists | AuthUsers: [], AllUsers: [Read]', b.stdout.decode('utf-8').strip())

    c = subprocess.run([sys.executable, '-m', 'S3Scanner', 'scan', '--bucket', 'asdfasdf---,'], stdout=subprocess.PIPE,
                       stderr=subprocess.PIPE)
    assert_scanner_output(s, 'asdfasdf---, | bucket_invalid_name', c.stdout.decode('utf-8').strip())

    d = subprocess.run([sys.executable, '-m', 'S3Scanner', 'scan', '--bucket', 'isurehopethisbucketdoesntexistasdfasdf'],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert_scanner_output(s, 'isurehopethisbucketdoesntexistasdfasdf | bucket_not_exist', d.stdout.decode('utf-8').strip())

    e = subprocess.run([sys.executable, '-m', 'S3Scanner', 'scan', '--bucket', 'flaws.cloud', '--dangerous'],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert_scanner_output(s, f"INFO: Including dangerous checks. WARNING: This may change bucket ACL destructively{os.linesep}flaws.cloud | bucket_exists | AuthUsers: [], AllUsers: [Read]", e.stdout.decode('utf-8').strip())

    f = subprocess.run([sys.executable, '-m', 'S3Scanner', 'dump', '--bucket', 'flaws.cloud', '--dump-dir', './asfasdf'],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert_scanner_output(s, "Error: Given --dump-dir does not exist or is not a directory", f.stdout.decode('utf-8').strip())

    # Create temp folder to dump into
    test_folder = os.path.join(os.getcwd(), 'testing_' + str(time.time())[0:10], '')
    os.mkdir(test_folder)

    try:
        f = subprocess.run([sys.executable, '-m', 'S3Scanner', 'dump', '--bucket', 'flaws.cloud', '--dump-dir', test_folder],
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        assert_scanner_output(s, f"flaws.cloud | Enumerating bucket objects...{os.linesep}flaws.cloud | Total Objects: 7, Total Size: 25.0KB{os.linesep}flaws.cloud | Dumping contents using 4 threads...{os.linesep}flaws.cloud | Dumping completed", f.stdout.decode('utf-8').strip())

        g = subprocess.run([sys.executable, '-m', 'S3Scanner', 'dump', '--bucket', 'asdfasdf,asdfasd,', '--dump-dir', test_folder],
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        assert_scanner_output(s, "asdfasdf,asdfasd, | bucket_name_invalid", g.stdout.decode('utf-8').strip())

        h = subprocess.run([sys.executable, '-m', 'S3Scanner', 'dump', '--bucket', 'isurehopethisbucketdoesntexistasdfasdf', '--dump-dir', test_folder],
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        assert_scanner_output(s, 'isurehopethisbucketdoesntexistasdfasdf | bucket_not_exist', h.stdout.decode('utf-8').strip())
    finally:
        shutil.rmtree(test_folder)  # Cleanup the testing folder


def test_endpoints():
    """
    Test the handling of non-AWS endpoints
    :return:
    """
    s = S3Service()
    b = subprocess.run([sys.executable, '-m', 'S3Scanner', '--endpoint-url', 'https://sfo2.digitaloceanspaces.com',
                        'scan', '--bucket', 's3scanner'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert_scanner_output(s, 's3scanner | bucket_not_exist',
                          b.stdout.decode('utf-8').strip())

    c = subprocess.run([sys.executable, '-m', 'S3Scanner', '--endpoint-url', 'http://example.com', 'scan', '--bucket',
                        's3scanner'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert c.stdout.decode('utf-8').strip() == "Error: Endpoint 'http://example.com' does not appear to be S3-compliant"


def assert_scanner_output(service, expected_output, found_output):
    """
    If the tests are run without AWS creds configured, all the output from scanner.py will have a warning banner.
    This is a convenience method to simplify comparing the expected output to the found output

    :param service: s3service
    :param expected_output: string
    :param found_output: string
    :return: boolean
    """
    creds_warning = "Warning: AWS credentials not configured - functionality will be limited. Run: `aws configure` to fix this."

    if service.aws_creds_configured:
        assert expected_output == found_output
    else:
        assert f"{creds_warning}{os.linesep}{os.linesep}{expected_output}" == found_output


def test_check_aws_creds():
    """
    Scenario checkAwsCreds.1 - Output of checkAwsCreds() matches a more intense check for creds
    """
    print("test_checkAwsCreds temporarily disabled.")

    # test_setup()
    #
    # # Check more thoroughly for creds being set.
    # vars = os.environ
    #
    # keyid = vars.get("AWS_ACCESS_KEY_ID")
    # key = vars.get("AWS_SECRET_ACCESS_KEY")
    # credsFile = os.path.expanduser("~") + "/.aws/credentials"
    #
    # if keyid is not None and len(keyid) == 20:
    #     if key is not None and len(key) == 40:
    #         credsActuallyConfigured = True
    #     else:
    #         credsActuallyConfigured = False
    # else:
    #     credsActuallyConfigured = False
    #
    # if os.path.exists(credsFile):
    #     print("credsFile path exists")
    #     if not credsActuallyConfigured:
    #         keyIdSet = None
    #         keySet = None
    #
    #         # Check the ~/.aws/credentials file
    #         with open(credsFile, "r") as f:
    #             for line in f:
    #                 line = line.strip()
    #                 if line[0:17].lower() == 'aws_access_key_id':
    #                     if len(line) >= 38:  # key + value = length of at least 38 if no spaces around equals
    #                         keyIdSet = True
    #                     else:
    #                         keyIdSet = False
    #
    #                 if line[0:21].lower() == 'aws_secret_access_key':
    #                     if len(line) >= 62:
    #                         keySet = True
    #                     else:
    #                         keySet = False
    #
    #         if keyIdSet and keySet:
    #             credsActuallyConfigured = True
    #
    # # checkAwsCreds.1
    # assert s3.checkAwsCreds() == credsActuallyConfigured

