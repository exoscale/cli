#!/usr/bin/env sh

set -e

aptly -config=.aptly.conf repo create my-first-repo
aptly -config=.aptly.conf repo import s3:sauterp-aptly-test-repo: my-first-repo
aptly -config=.aptly.conf repo add my-first-repo $1
aptly -config=.aptly.conf publish update ubuntu s3:sauterp-aptly-test-repo:
