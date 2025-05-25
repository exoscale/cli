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

is_apt_newer_than_v2_2() {
    apt_version=$(apt-get --version 2>&1 | grep -oP '(?<=apt )[0-9]+\.[0-9]+\.[0-9]+')

    if [ -z "$apt_version" ]; then
        # can't determine apt version
        return 1
    elif [ "$(
        dpkg --compare-versions "$apt_version" ge "2.2.0"
        echo $?
    )" -eq 0 ]; then
        # apt is >= 2.2.0
        return 0
    else
        # apt is < 2.2.0
        return 1
    fi
}

if [ -f /etc/os-release ]; then
    . /etc/os-release
    case "$ID" in
        debian | ubuntu | pop | neon | zorin | linuxmint | elementary | parrot | mendel | galliumos | pureos | raspian | kali | Deepin)
            if is_apt_newer_than_v2_2; then
                PACKAGETYPE="apt"
            else
                PACKAGETYPE="dpkg"
            fi
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
    aarch64)
        CPUARCHITECTURE="arm64"
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
PKGSIGFILE=$PKGFILE.sig
PKGPATH=$TEMPDIR/$PKGFILE
PKGSIGPATH=$TEMPDIR/$PKGSIGFILE

download_pkg() {
    $CURL "$GITHUB_DOWNLOAD_URL/${LATEST_TAG}/$PKGFILE" >$PKGPATH

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
}

TOOLING_KEY_NAME="Exoscale Tooling <tooling@exoscale.ch>"
TOOLING_KEY_FINGERPRINT="7100E8BFD6199CE0374CB7F003686F8CDE378D41"

GPG_AVAILABLE=no
# downloads and verifies the signature file for the package if gpg is available
verify_pkg() {
    if [ "$GPG_AVAILABLE" = "yes" ]; then
        $CURL "$GITHUB_DOWNLOAD_URL/${LATEST_TAG}/$PKGSIGFILE" >$PKGSIGPATH
        gpg --verify $PKGSIGPATH $PKGPATH
    fi
}

if command -v gpg >/dev/null 2>&1 && [ "$PACKAGETYPE" != "yum" ]; then
    if ! gpg --list-keys | grep -q $TOOLING_KEY_FINGERPRINT; then
        gpg --keyserver hkps://keys.openpgp.org:443 --recv-keys "$TOOLING_KEY_FINGERPRINT"
    fi

    GPG_AVAILABLE=yes
else
    if [ "$PACKAGETYPE" = "apt" ]; then
        # since gpg is not available, we have to fall back to installing via dpkg
        PACKAGETYPE="dpkg"
    fi
fi

install_rpm_pkg() {
    repofile=/etc/yum.repos.d/exoscale-cli.repo
    cat <<EOF | $SUDO tee $repofile
[exoscale-cli-repo]
name=exoscale-cli-repo
baseurl=https://sos-ch-gva-2.exo.io/exoscale-packages/rpm/cli
enabled=1
repo_gpgcheck=1
gpgcheck=0
gpgkey=https://keys.openpgp.org/vks/v1/by-fingerprint/7100E8BFD6199CE0374CB7F003686F8CDE378D41
EOF

    if [ "$PACKAGETYPE" = "yum" ]; then
        REPOFLAG=""
    else
        REPOFLAG="--repo=exoscale-cli-repo"
    fi
    if $PACKAGETYPE list installed exoscale-cli >/dev/null 2>&1; then
        $SUDO $PACKAGETYPE makecache -y $REPOFLAG
        $SUDO $PACKAGETYPE $REPOFLAG upgrade -y exoscale-cli
    else
        $SUDO $PACKAGETYPE makecache -y $REPOFLAG
        $SUDO $PACKAGETYPE $REPOFLAG install -y exoscale-cli
    fi
}

echo "Installing exo CLI, using $PACKAGETYPE"
case "$PACKAGETYPE" in
    apt)
        $SUDO mkdir -p /etc/apt/keyrings
        gpg --export $TOOLING_KEY_FINGERPRINT | $SUDO tee /etc/apt/keyrings/exoscale.gpg >/dev/null
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/exoscale.gpg] https://sos-ch-gva-2.exo.io/exoscale-packages/deb/cli stable main" | $SUDO tee /etc/apt/sources.list.d/exoscale.list >/dev/null
        $SUDO apt-get update
        $SUDO apt-get install -y exoscale-cli
        ;;
    dpkg)
        download_pkg
        verify_pkg
        $SUDO dpkg -i $PKGPATH
        ;;
    yum)
        install_rpm_pkg
        ;;
    dnf)
        install_rpm_pkg
        ;;
esac
