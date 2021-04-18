import sys
import subprocess
import os
import tempfile


def test_arguments():
    a = subprocess.run([sys.executable, 'scanner.py', '--version'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert a.stdout.decode('utf-8').strip() == '2.0.0'

    b = subprocess.run([sys.executable, 'scanner.py', 'scan', '--bucket', 'flaws.cloud'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert b.stdout.decode('utf-8').strip() == 'flaws.cloud | bucket_exists | AuthUsers: [], AllUsers: [Read]'

    c = subprocess.run([sys.executable, 'scanner.py', 'scan', '--bucket', 'asdfasdf---,'], stdout=subprocess.PIPE,
                       stderr=subprocess.PIPE)
    assert c.stdout.decode('utf-8').strip() == 'asdfasdf---, | bucket_invalid_name'

    d = subprocess.run([sys.executable, 'scanner.py', 'scan', '--bucket', 'isurehopethisbucketdoesntexistasdfasdf'],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert d.stdout.decode('utf-8').strip() == 'isurehopethisbucketdoesntexistasdfasdf | bucket_not_exist'

    e = subprocess.run([sys.executable, 'scanner.py', 'scan', '--bucket', 'flaws.cloud', '--dangerous'],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert e.stdout.decode('utf-8').strip() == f"INFO: Including dangerous checks. WARNING: This may change bucket ACL destructively{os.linesep}flaws.cloud | bucket_exists | AuthUsers: [], AllUsers: [Read]"

    f = subprocess.run([sys.executable, 'scanner.py', 'dump', '--bucket', 'flaws.cloud', '--dump-dir', './asfasdf'],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert f.stdout.decode('utf-8').strip() == f"Error: Given --dump-dir does not exist or is not a directory"

    f = subprocess.run([sys.executable, 'scanner.py', 'dump', '--bucket', 'flaws.cloud', '--dump-dir', tempfile.gettempdir()],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert f.stdout.decode('utf-8').strip() == f"flaws.cloud | Debug: Dumping without creds...{os.linesep}flaws.cloud | Enumerating bucket objects...{os.linesep}flaws.cloud | Total Objects: 7, Total Size: 25.0KB{os.linesep}flaws.cloud | Dumping contents...{os.linesep}flaws.cloud | Dumping completed"

    g = subprocess.run([sys.executable, 'scanner.py', 'dump', '--bucket', 'asdfasdf,asdfasd,', '--dump-dir', tempfile.gettempdir()],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert g.stdout.decode('utf-8').strip() == "asdfasdf,asdfasd, | bucket_name_invalid"

    h = subprocess.run([sys.executable, 'scanner.py', 'dump', '--bucket', 'isurehopethisbucketdoesntexistasdfasdf', '--dump-dir', tempfile.gettempdir()],
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    assert h.stdout.decode('utf-8').strip() == 'isurehopethisbucketdoesntexistasdfasdf | bucket_not_exist'




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

