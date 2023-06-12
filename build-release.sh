#!/usr/bin/env sh

SNAPSHOT_DIR=/snapshot
rsync -rltD --delete /src/ $SNAPSHOT_DIR/

cd $SNAPSHOT_DIR

export CGO_ENABLED=1
make test
