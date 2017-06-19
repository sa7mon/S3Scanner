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

resolver = dns.resolver.Resolver()
# resolver.nameservers=[socket.gethostbyname('ns1.cisco.com')]

serverCountLimit = 2
logFileName = "results.txt"


def pprint(message):
    logFile.write(message + "\n")
    print(message)


def checkSite(site):
	i = 1
	try:
		for rdata in resolver.query(site, 'A'):
			if (i > serverCountLimit):
				pprint("Skipping " + site + " additional A server(s)...")
				continue
			resultText = "	A record: " + str(rdata) + " - "

			nslookup = subprocess.Popen("nslookup " + str(rdata),stdout=subprocess.PIPE,shell=True)
			match = re.search("(name = )\w(.*amazonaws\.com)",nslookup.stdout.read())
			if match:
				pprint("\033[0;32m" + site + " : " + str(rdata) + " : S3!\033[0;m")
			else:
				pprint("\033[0;91m" + site + " : " + str(rdata) + " : Not\033[0;m")
			i += 1
	except:
		pprint("Caught error. Skipping...")

# Open log file for writing
logFile = open(logFileName, 'a+')

with open("sites.txt", 'r') as f:
	for line in f:
		site = line.rstrip()
		checkSite(site)