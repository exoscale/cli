#!/usr/bin/env sh

# to avoid permission errors we mount the /src folder read-only and create a snapshot
SNAPSHOT_DIR=/snapshot

rsync \
    --recursive \
    --links \
    --times \
    --delete \
    --exclude=/dist/ \
    /src/ $SNAPSHOT_DIR/

cd $SNAPSHOT_DIR

# TODO(sauterp) remove
git clean -f
git reset --hard HEAD

export CGO_ENABLED=1
make release
