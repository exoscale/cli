#!/usr/bin/env sh

# a lot of inspiration for this installer was taken from the tailscale installer script at https://github.com/tailscale/tailscale/blob/827abbeeaaf21388ec0dc7046d61d8350afe98f7/scripts/installer.sh

# exit if one command exits with non-zero exit code and treat unset variables as an error
set -eu

# detect the latest version of the exoscale cli
LATEST_URL=$(curl -L -s -o /dev/null -w %{url_effective} https://github.com/exoscale/cli/releases/latest)
LATEST_TAG=$(basename "${LATEST_URL}")
LATEST_VERSION=$(echo "$LATEST_TAG" | cut -c 2-)

OS=""
OSTYPE=""
CPUARCHITECTURE=""
FILEEXT=""
VERSION=""
PACKAGETYPE=""
APT_KEY_TYPE="" # Only for apt-based distros

if [ -f /etc/os-release ]; then
    . /etc/os-release
    case "$ID" in
        debian | ubuntu | pop | neon | zorin | linuxmint | elementary | parrot | mendel | galliumos | pureos | raspian | kali | Deepin)
            OS="$ID"
            PACKAGETYPE="deb"
            FILEEXT="deb"
            ;;
        fedora)
            OS="$ID"
            VERSION=""
            PACKAGETYPE="dnf"
            FILEEXT="rpm"
            OSTYPE="linux"
            ;;
        centos)
            OS="$ID"
            VERSION="$VERSION_ID"
            PACKAGETYPE="dnf"
            FILEEXT="rpm"
            OSTYPE="linux"
            if [ "$VERSION" = "7" ]; then
                PACKAGETYPE="yum"
            fi
            ;;
    esac
fi

# detect the cpu architecture
CPUARCHITECTURE="$(uname -m)"
case "$(uname -m)" in
    x86_64)
        CPUARCHITECTURE="amd64"
        ;;
    armv7l)
        CPUARCHITECTURE="armv7"
        ;;
esac

# Ideally we want to use curl, but on some installs we
# only have wget. Detect and use what's available.
CURL=
if type curl >/dev/null; then
    CURL="curl -fsSL"
elif type wget >/dev/null; then
    CURL="wget -q -O-"
fi
if [ -z "$CURL" ]; then
    echo "The installer needs either curl or wget to download files."
    echo "Please install either curl or wget to proceed."
    exit 1
fi

TEST_URL="https://exoscale.com/"
RC=0
TEST_OUT=$($CURL "$TEST_URL" 2>&1) || RC=$?
if [ $RC != 0 ]; then
    echo "The installer cannot reach $TEST_URL"
    echo "Please make sure that your machine has internet access."
    echo "Test output:"
    echo $TEST_OUT
    exit 1
fi

# work out if we can run privileged commands, and if so, how
CAN_ROOT=
SUDO=
if [ "$(id -u)" = 0 ]; then
    CAN_ROOT=1
    SUDO=""
elif type sudo >/dev/null; then
    CAN_ROOT=1
    SUDO="sudo"
elif type doas >/dev/null; then
    CAN_ROOT=1
    SUDO="doas"
fi
if [ "$CAN_ROOT" != "1" ]; then
    echo "This installer needs to run commands as root."
    echo "We tried looking for 'sudo' and 'doas', but couldn't find them."
    echo "Either re-run this script as root, or set up sudo/doas."
    exit 1
fi

TEMPDIR=$(mktemp -d)
PKGFILE="exoscale-cli_${LATEST_VERSION}_${OSTYPE}_${CPUARCHITECTURE}.${FILEEXT}"
PKGPATH=$TEMPDIR/$PKGFILE
$CURL "https://github.com/exoscale/cli/releases/download/${LATEST_TAG}/$PKGFILE" >$PKGPATH

OSVERSION="$OS"
echo "Installing exo CLI for $OSVERSION, using method $PACKAGETYPE"
case "$PACKAGETYPE" in
    deb)
        set -x
        $SUDO dpkg -i $PKGPATH
        set -x
        ;;
    yum)
        set -x
        $SUDO yum install -y $PKGPATH
        set +x
        ;;
    dnf)
        set -x
        $SUDO dnf install -y $PKGPATH
        set +x
        ;;
esac
