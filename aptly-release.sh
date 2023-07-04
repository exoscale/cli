#!/usr/bin/env sh

set -e

reponame=myrepo

aptly -config=.aptly.conf repo create $reponame
aptly -config=.aptly.conf repo add $reponame $1
aptly -config=.aptly.conf publish update ubuntu s3:sauterp-aptly-test-repo:
