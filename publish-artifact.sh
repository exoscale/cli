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

# Get the 10 latest Git tags
latest_tags=$(git tag --sort=-v:refname | head -n 10)

# Create a package query filter for aptly
package_filter=""
for tag in $latest_tags; do
    stripped_tag=$(expr "$tag" : '.\(.*\)')
    if [ $first_tag_set ]; then
        package_filter+=" | "
    fi
    package_filter+="exoscale-cli (= $stripped_tag)"
    first_tag_set=1
done

mirrorrepo() {
    $aptlycmd mirror create \
        $aptlymirror \
        $archiveurl \
        $aptlydistro

    $aptlycmd mirror update \
        $aptlymirror
}

$aptlycmd repo create $aptlyrepo

if mirrorrepo; then
    $aptlycmd repo import $aptlymirror $aptlyrepo "$package_filter"
fi

$aptlycmd repo add $aptlyrepo $artifact

$aptlycmd publish repo \
    $gpgkeyflag \
    -distribution=$aptlydistro \
    $aptlyrepo \
    $aptlyremote
