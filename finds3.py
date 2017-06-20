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
# resolver.nameservers=[socket.gethostbyname('ns1.cisco.com')]

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
	# try:
	for rdata in resolver.query(site, 'A'):
		if (i > serverCountLimit):
			pprint(False, "Skipping " + site + " additional A server(s)...")
			continue
		resultText = "	A record: " + str(rdata) + " - "

		nslookup = subprocess.Popen("nslookup " + str(rdata),stdout=subprocess.PIPE,shell=True)
		match = re.search("(name = )\w(.*amazonaws\.com)",nslookup.stdout.read())
		if match:
			pprint(True, site + " : " + str(rdata) + " : S3!")
		else:
			pprint(False, site + " : " + str(rdata) + " : Not")
		i += 1
	# except:
		# pprint(False, "Caught error. Skipping...")

# Instantiate the parser
parser = argparse.ArgumentParser(description='Find S3 sites!')

# Declare arguments
parser.add_argument('-g', '--goodDomains', required=True, help='Name of file to save the success domains in')
# parser.add_argument('-g', '--email', required=False, help='Optional email to resume at')
parser.add_argument('domains', help='Name of file containing domains to check')
# parser.add_argument('-t', '--threads', type=int, required=False, help='Number of threads to use')

# Parse the args
args = parser.parse_args()


# Open log file for writing
logFile = open(args.goodDomains, 'a+')

with open(args.domains, 'r') as f:
	for line in f:
		site = line.rstrip()
		checkSite(site)


