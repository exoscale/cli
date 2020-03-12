#!/bin/sh
set -xe

cp ../README.md content/_index.md
mkdir -p static

cd ..
go run -mod vendor doc/main.go

set +xe
echo "we are now ready to: hugo serve"
