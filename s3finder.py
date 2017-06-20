#!/usr/bin/env python
#
# AWS S3scanner - Scans domain names for A records that resolve to *.amazonaws.com
# 
# Author:  Dan Salmon (twitter.com/bltjetpack, github.com/sa7mon)
# Created: 6/19/17
#

import dns.resolver
import socket
import dns.reversename
import subprocess
import re
import argparse

resolver = dns.resolver.Resolver()
resolver.nameservers=['8.8.8.8', '8.8.4.4', '208.67.222.222',
					'208.67.220.220', '216.146.35.35',
					'216.146.36.36', socket.gethostbyname('ns1.cisco.com')]

serverCountLimit = 2
logFileName = "results.txt"


def pprint(good, message):
    if good:
		# print in green
		print("\033[0;32m" + message + "\033[0;m")
		logFile.write(message + "\n")
    else:
    	# print in red
    	print("\033[0;91m" + message + "\033[0;m")


def checkSite(site):
	i = 1
	try:
		for rdata in resolver.query(site, 'A'):
			if (i > serverCountLimit):
				pprint(False, "Skipping " + site + " additional A server(s)...")
				continue
			resultText = "	A record: " + str(rdata) + " - "

			nslookup = subprocess.Popen("nslookup " + str(rdata),stdout=subprocess.PIPE,shell=True)
			match = re.search("(?:name = s3-website-)(.*)(?:\.amazonaws\.com)",nslookup.stdout.read())
			if match:
				aws = str(match.group(1))
				printLine = site + ":" + aws
				print("\033[0;32m" + printLine + "\033[0;m")
				logFile.write(printLine + "\n")
			else:
				pprint(False, site + " : " + str(rdata) + " : Not")
			i += 1
	except dns.resolver.NoAnswer as err:
		pprint(False, "Caught 'No A Record' error. Skipping...")
	except dns.resolver.NXDOMAIN as err:
		pprint(False, "Caught 'No NX Domain' error. Skipping...")
	except dns.resolver.NoNameservers as err:
		pprint(False, "Caught 'All nameservers failed' error. Skipping...")
	except dns.exception.Timeout as err:
		pprint(False, "Caught 'DNS timout' error. Skipping...")

# Instantiate the parser
parser = argparse.ArgumentParser(description='Find S3 sites!')

# Declare arguments
parser.add_argument('-g', '--goodDomains', required=True, help='Name of file to save the success domains in')
parser.add_argument('domains', help='Name of file containing domains to check')

# Parse the args
args = parser.parse_args()


# Open log file for writing
logFile = open(args.goodDomains, 'a+')

with open(args.domains, 'r') as f:
	for line in f:
		site = line.rstrip()
		checkSite(site)


