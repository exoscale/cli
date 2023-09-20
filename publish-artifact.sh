#!/usr/bin/env sh

set -e

artifact=$1
aptlyrepo=cli-release
aptlyremote=s3:sauterp-packages:
archiveurl=https://sos-ch-dk-2.exo.io/sauterp-packages
aptlymirror=cli-release-mirror
aptlydistro=stable
aptlycmd="aptly -config=.aptly.conf"
gpgkeyflag='-gpg-key=7100E8BFD6199CE0374CB7F003686F8CDE378D41'

if $aptlycmd mirror show $aptlymirror >/dev/null 2>&1; then
    echo "repo $aptlymirror already exists"
else
    $aptlycmd mirror create \
        $aptlymirror \
        $archiveurl \
        $aptlydistro
fi

$aptlycmd mirror update \
    $aptlymirror

if $aptlycmd repo show $aptlyrepo >/dev/null 2>&1; then
    echo "repo $aptlyrepo already exists"
else
    $aptlycmd repo import $aptlymirror $aptlyrepo $aptlydistro
fi

if echo $artifact | grep -q '.*.deb'; then
    $aptlycmd repo add $aptlyrepo $artifact
    $aptlycmd publish update \
        $gpgkeyflag \
        $aptlydistro \
        $aptlyremote
fi
