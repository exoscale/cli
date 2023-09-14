#!/usr/bin/env sh

# a lot of inspiration for this installer was taken from the tailscale installer script at https://github.com/tailscale/tailscale/blob/827abbeeaaf21388ec0dc7046d61d8350afe98f7/scripts/installer.sh

# exit if one command exits with non-zero exit code and treat unset variables as an error
set -eu

# detect the latest version of the exoscale cli
LATEST_URL=$(curl -L -s -o /dev/null -w %{url_effective} https://github.com/exoscale/cli/releases/latest)
LATEST_TAG=$(basename "${LATEST_URL}")
LATEST_VERSION=$(echo "$LATEST_TAG" | cut -c 2-)

OSTYPE=""
CPUARCHITECTURE=""
FILEEXT=""
PACKAGETYPE=""

if [ -f /etc/os-release ]; then
    . /etc/os-release
    case "$ID" in
        debian | ubuntu | pop | neon | zorin | linuxmint | elementary | parrot | mendel | galliumos | pureos | raspian | kali | Deepin)
            PACKAGETYPE="dpkg"
            FILEEXT="deb"
            OSTYPE="linux"
            ;;
        fedora | rocky | almalinux | nobara | openmandriva | sangoma | risios)
            PACKAGETYPE="dnf"
            FILEEXT="rpm"
            OSTYPE="linux"
            ;;
        rhel)
            VERSION="$(echo "$VERSION_ID" | cut -f1 -d.)"
            PACKAGETYPE="dnf"
            FILEEXT="rpm"
            if [ "$VERSION" = "7" ]; then
                PACKAGETYPE="yum"
            fi
            ;;
        centos)
            PACKAGETYPE="dnf"
            FILEEXT="rpm"
            OSTYPE="linux"
            if [ "$VERSION_ID" = "7" ]; then
                PACKAGETYPE="yum"
            fi
            ;;
        amzn | xenenterprise)
            PACKAGETYPE="yum"
            FILEEXT="rpm"
            OSTYPE="linux"
            ;;
        *)
            echo "your OS is not supported by this script, please install manually https://community.exoscale.com/documentation/tools/exoscale-command-line-interface/#installation"
            exit 1
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

TEST_URL="https://www.exoscale.com/"
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

GITHUB_DOWNLOAD_URL="https://github.com/exoscale/cli/releases/download"

TEMPDIR=$(mktemp -d)
PKGPREFIX="exoscale-cli"
PKGFILE="${PKGPREFIX}_${LATEST_VERSION}_${OSTYPE}_${CPUARCHITECTURE}.${FILEEXT}"
PKGPATH=$TEMPDIR/$PKGFILE
$CURL "$GITHUB_DOWNLOAD_URL/${LATEST_TAG}/$PKGFILE" >$PKGPATH

# check the checksum
CHECKSUMSFILE="${PKGPREFIX}_${LATEST_VERSION}_checksums.txt"
CHECKSUMSPATH=$TEMPDIR/$CHECKSUMSFILE
$CURL "$GITHUB_DOWNLOAD_URL/${LATEST_TAG}/$CHECKSUMSFILE" >$CHECKSUMSPATH

COMPUTED_CHECKSUM=$(sha256sum "$PKGPATH" | cut -d " " -f 1)
EXPECTED_CHECKSUM=$(grep -m 1 $PKGFILE $CHECKSUMSPATH | cut -d " " -f 1)

if [ "$COMPUTED_CHECKSUM" != "$EXPECTED_CHECKSUM" ]; then
    echo "Error: Checksum of $PKGFILE does not match the expected checksum"
    echo $COMPUTED_CHECKSUM
    echo $EXPECTED_CHECKSUM
    exit 1
fi

if ! command -v gpg >/dev/null 2>&1; then
    echo "GPG is not installed. It is recommended to verify the authenticity of the exo cli package before installing it. Please install GPG."

    read -p "Would you like to install exo cli without verifying the package's authenticity? (N/y): " verify_signature
    if [ ! "$verify_signature" = "y" ]; then
        echo "Exiting."
        exit 1
    fi
else
    TOOLING_KEY_NAME="Exoscale Tooling <tooling@exoscale.ch>"
    TOOLING_KEY_FINGERPRINT="7100E8BFD6199CE0374CB7F003686F8CDE378D41"

    # Check if the tooling key is available
    if gpg --list-keys | grep -q $TOOLING_KEY_FINGERPRINT; then
        # verity sig
        echo "the key is available"
        exit 1
    else
        read -p "The GPG key $TOOLING_KEY_NAME ($TOOLING_KEY_FINGERPRINT) is missing, would you like to import it? (N/y): " import_key
        if [ "$import_key" = "y" ]; then
            echo "Importing key"
            gpg --recv-keys "$TOOLING_KEY_FINGERPRINT"
            if [ $? -eq 0 ]; then
                echo "Import successful."
                echo "the key is available"
                # verity sig
            else
                echo "Import failed. Exiting."
                exit 1
            fi
        else
            echo "Exiting."
        fi
    fi
fi

echo "Installing exo CLI, using $PACKAGETYPE"
case "$PACKAGETYPE" in
    dpkg)
        $SUDO dpkg -i $PKGPATH
        ;;
    yum)
        $SUDO yum install -y $PKGPATH
        ;;
    dnf)
        $SUDO dnf install -y $PKGPATH
        ;;
esac
