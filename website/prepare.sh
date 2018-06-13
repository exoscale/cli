#!/bin/sh
set -xe

cd ../
dep ensure -vendor-only

cd cmd/cs
dep ensure -vendor-only
go build
./cs gen-doc
cp README.md ../../website/content/cs/_index.md

cd ../exo
dep ensure -vendor-only
go run doc/main.go
cp README.md ../../website/content/exo/_index.md

set +xe
echo "we are now ready to run hugo"
