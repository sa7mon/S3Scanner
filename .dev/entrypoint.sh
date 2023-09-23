#!/bin/sh

apk add curl ca-certificates
http_proxy=mitmproxy:8080 curl http://mitm.it/cert/pem -o /usr/local/share/ca-certificates/mitmproxy-ca-cert.pem
update-ca-certificates
tail -f /dev/null