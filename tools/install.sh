#!/usr/bin/env bash

set -e

DEFAULT_DRP_VERSION=${DEFAULT_DRP_VERSION:-"stable"}

usage() {
cat <<EOFUSAGE
Usage: $0 [--version=<Version to install>] [--nocontent]
          [--isolate] [--ipaddr=<ip>] install | remove

Options:
    --debug=[true|false]    # Enables debug output
    --force=[true|false]    # Forces an overwrite of local install binaries and content
    --upgrade=[true|false]  # Turns on 'force' option to overwrite local binaries/content
    --isolated              # Sets up current directory as install location for drpcli
                            # and dr-provision
    --nocontent             # Don't add content to the system
    --ipaddr=<ip>           # The IP to use for the system identified IP.  The system
                            # will attepmto to discover the value if not specified
    --version=<string>      # Version identifier if downloading.  stable, tip, or
                            # specific version label.  Defaults to: $DEFAULT_DRP_VERSION

    install                 # Sets up an insolated or system 'production' enabled install.
    remove                  # Removes the system enabled install.  Requires no other flags

Defaults are:
    version     = $DEFAULT_DRP_VERSION    (examples: 'tip', 'v3.6.0' or 'stable')
    isolated    = false
    nocontent   = false
    upgrade     = false
    force       = false
    debug       = false
EOFUSAGE

exit 0
}

ISOLATED=false
NO_CONTENT=false
DBG=false
UPGRADE=false

args=()
while (( $# > 0 )); do
    arg="$1"
    arg_key="${arg%%=*}"
    arg_data="${arg#*=}"
    case $arg_key in
        --help|-h)
            usage
            exit 0
            ;;
        --debug)
            DBG=true
            ;;
        --version|--drp-version)
            DRP_VERSION=${arg_data}
            ;;
        --isolated)
            ISOLATED=true
            ;;
        --force)
            force=true
            ;;
        --upgrade)
            UPGRADE=true
            force=true
            ;;
        --nocontent)
            NO_CONTENT=true
            ;;
        --*)
            arg_key="${arg_key#--}"
            arg_key="${arg_key//-/_}"
            # "^^" Paremeter Expansion is a bash v4.x feature; Mac by default is bash 3.x
            #arg_key="${arg_key^^}"
            arg_key=$(echo $arg_key | tr '[:lower:]' '[:upper:]')
            echo "Overriding $arg_key with $arg_data"
            export $arg_key="$arg_data"
            ;;
        *)
            args+=("$arg");;
    esac
    shift
done
set -- "${args[@]}"

DRP_VERSION=${DRP_VERSION:-"$DEFAULT_DRP_VERSION"}

[[ $DBG == true ]] && set -x

# Figure out what Linux distro we are running on.
export OS_TYPE= OS_VER= OS_NAME= OS_FAMILY=

if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    OS_TYPE=${ID,,}
    OS_VER=${VERSION_ID,,}
elif [[ -f /etc/lsb-release ]]; then
    . /etc/lsb-release
    OS_VER=${DISTRIB_RELEASE,,}
    OS_TYPE=${DISTRIB_ID,,}
elif [[ -f /etc/centos-release || -f /etc/fedora-release || -f /etc/redhat-release ]]; then
    for rel in centos-release fedora-release redhat-release; do
        [[ -f /etc/$rel ]] || continue
        OS_TYPE=${rel%%-*}
        OS_VER="$(egrep -o '[0-9.]+' "/etc/$rel")"
        break
    done

    if [[ ! $OS_TYPE ]]; then
        echo "Cannot determine Linux version we are running on!"
        exit 1
    fi
elif [[ -f /etc/debian_version ]]; then
    OS_TYPE=debian
    OS_VER=$(cat /etc/debian_version)
elif [[ $(uname -s) == Darwin ]] ; then
    OS_TYPE=darwin
    OS_VER=$(sw_vers | grep ProductVersion | awk '{ print $2 }')
fi
OS_NAME="$OS_TYPE-$OS_VER"

case $OS_TYPE in
    centos|redhat|fedora) OS_FAMILY="rhel";;
    debian|ubuntu) OS_FAMILY="debian";;
    *) OS_FAMILY=$OS_TYPE;;
esac

ensure_packages() {
    echo "Ensuring required tools are installed"
    if [[ $OS_FAMILY == darwin ]] ; then
        VER=$(tar -h | grep "bsdtar " | awk '{ print $2 }' | awk -F. '{ print $1 }')
        if [[ $VER != 3 ]] ; then
            echo "Please update tar to greater than 3.0.0"
            echo
            echo "E.g: "
            echo "  brew install libarchive --force"
            echo "  brew link libarchive --force"
            echo
            echo "Close current terminal and open a new terminal"
            echo
            exit 1
        fi
        if ! which 7z &>/dev/null; then
            echo "Must have 7z"
            echo "E.g: brew install p7zip"
            exit 1
        fi
    else
        if ! which bsdtar &>/dev/null; then
            echo "Installing bsdtar"
            if [[ $OS_FAMILY == rhel ]] ; then
                sudo yum install -y bsdtar
            elif [[ $OS_FAMILY == debian ]] ; then
                sudo apt-get install -y bsdtar
            fi
        fi
        if ! which 7z &>/dev/null; then
            echo "Installing 7z"
            if [[ $OS_FAMILY == rhel ]] ; then
                if [[ $OS_TYPE != fedora ]] ; then
                    sudo yum install -y epel-release
                fi
                sudo yum install -y p7zip
            elif [[ $OS_FAMILY == debian ]] ; then
                sudo apt-get install -y p7zip-full
            fi
        fi
    fi
}

arch=$(uname -m)
case $arch in
	x86_64|amd64) arch=amd64  ;;
	aarch64)      arch=arm64  ;;
  armv7l)       arch=arm_v7 ;;
	*) 	echo "FATAL: architecture ('$arch') not supported"
		exit 1 
	;;
esac

case $(uname -s) in
    Darwin)
        binpath="bin/darwin/$arch"
        bindest="/usr/local/bin"
        tar="command bsdtar"
        # Someday, handle adding all the launchd stuff we will need.
        shasum="command shasum -a 256";;
    Linux)
        binpath="bin/linux/$arch"
        bindest="/usr/local/bin"
        tar="command bsdtar"
        if [[ -d /etc/systemd/system ]]; then
            # SystemD
            initfile="assets/startup/dr-provision.service"
            initdest="/etc/systemd/system/dr-provision.service"
            starter="sudo systemctl daemon-reload && sudo systemctl start dr-provision"
            enabler="sudo systemctl daemon-reload && sudo systemctl enable dr-provision"
        elif [[ -d /etc/init ]]; then
            # Upstart
            initfile="assets/startup/dr-provision.unit"
            initdest="/etc/init/dr-provision.conf"
            starter="sudo service dr-provision start"
            enabler="sudo service dr-provision enable"
        elif [[ -d /etc/init.d ]]; then
            # SysV
            initfile="assets/startup/dr-provision.sysv"
            initdest="/etc/init.d/dr-provision"
            starter="/etc/init.d/dr-provision start"
            enabler="/etc/init.d/dr-provision enable"
        else
            echo "No idea how to install startup stuff -- not using systemd, upstart, or sysv init"
            exit 1
        fi
        shasum="command sha256sum";;
    *)
        # Someday, support installing on Windows.  Service creation could be tricky.
        echo "No idea how to check sha256sums"
        exit 1;;
esac

case $1 in
     install)
             if pgrep dr-provision; then
                 echo "'dr-provision' service is running, CAN NOT upgrade ... please stop service first"
                 exit 9
             else
                 echo "'dr-provision' service is not running, beginning install process ... "
             fi

             ensure_packages
             # Are we in a build tree
             if [ -e server ] ; then
                 if [ ! -e bin/linux/amd64/drpcli ] ; then
                     echo "It appears that nothing has been built."
                     echo "Please run tools/build.sh and then rerun this command".
                     exit 1
                 fi
             else
                 # We aren't a build tree, but are we extracted install yet?
                 # If not, get the requested version.
                 if [[ ! -e sha256sums || $force ]] ; then
                     echo "Installing Version $DRP_VERSION of Digital Rebar Provision"
                     curl -sfL -o dr-provision.zip https://github.com/digitalrebar/provision/releases/download/$DRP_VERSION/dr-provision.zip
                     curl -sfL -o dr-provision.sha256 https://github.com/digitalrebar/provision/releases/download/$DRP_VERSION/dr-provision.sha256

                     $shasum -c dr-provision.sha256
                     $tar -xf dr-provision.zip
                 fi
                 $shasum -c sha256sums || exit 1
             fi

             if [[ $NO_CONTENT == false ]] ; then
                 echo "Installing Version $DRP_VERSION of Digital Rebar Provision Community Content"
                 curl -sfL -o drp-community-content.yaml https://github.com/digitalrebar/provision-content/releases/download/$DRP_VERSION/drp-community-content.yaml || echo "Failed to dowload content."
                 curl -sfL -o drp-community-content.sha256 https://github.com/digitalrebar/provision-content/releases/download/$DRP_VERSION/drp-community-content.sha256 || echo "Failed to download sha of content."
                 $shasum -c drp-community-content.sha256
             fi

             if [[ $ISOLATED == false ]] ; then
                 sudo cp "$binpath"/* "$bindest"
                 if [[ $initfile ]]; then
                     if [[ -r $initdest ]]
                     then
                         echo "WARNING ... WARNING ... WARNING"
                         echo "initfile ('$initfile') exists already, not overwriting it"
                         echo "please verify 'dr-provision' startup options are correct"
                         echo "for your environment and the new version .. "
                         echo ""
                         echo "specifically verify: '--file-root=<tftpboot directory>'"
                     else
                         sudo cp "$initfile" "$initdest"
                     fi
                     echo "# You can start the DigitalRebar Provision service with:"
                     echo "$starter"
                     echo "# You can enable the DigitalRebar Provision service with:"
                     echo "$enabler"
                 fi

                 # handle the v3.0.X to v3.1.0 directory structure.
                 if [[ ! -e /var/lib/dr-provision/digitalrebar && -e /var/lib/dr-provision ]] ; then
                     DIR=$(mktemp -d)
                     sudo mv /var/lib/dr-provision $DIR
                     sudo mkdir -p /var/lib/dr-provision
                     sudo mv $DIR/* /var/lib/dr-provision/digitalrebar
                 fi

                 if [[ ! -e /var/lib/dr-provision/digitalrebar/tftpboot && -e /var/lib/tftpboot ]] ; then
                     echo "MOVING /var/lib/tftpboot to /var/lib/dr-provision/tftpboot location ... "
                     sudo mv /var/lib/tftpboot /var/lib/dr-provision/
                 fi

                 sudo mkdir -p /usr/share/dr-provision
                 if [[ $NO_CONTENT == false ]] ; then
                     DEFAULT_CONTENT_FILE="/usr/share/dr-provision/default.yaml"
                     sudo mv drp-community-content.yaml $DEFAULT_CONTENT_FILE
                 fi
             else
                 mkdir -p drp-data

                 # Make local links for execs
                 rm -f drpcli dr-provision drbundler
                 ln -s $binpath/drpcli drpcli
                 ln -s $binpath/dr-provision dr-provision
                 if [[ -e $binpath/drbundler ]] ; then
                     ln -s $binpath/drbundler drbundler
                 fi

                 echo "# Run the following commands to start up dr-provision in a local isolated way."
                 echo "# The server will store information and serve files from the drp-data directory."
                 echo

                 if [[ $IPADDR == "" ]] ; then
                     if [[ $OS_FAMILY == darwin ]]; then
                         ifdefgw=$(netstat -rn -f inet | grep default | awk '{ print $6 }')
                         if [[ $ifdefgw ]] ; then
                                 IPADDR=$(ifconfig en0 | grep 'inet ' | awk '{ print $2 }')
                         else
                                 IPADDR=$(ifconfig -a | grep "inet " | grep broadcast | head -1 | awk '{ print $2 }')
                         fi
                     else
                         gwdev=$(/sbin/ip -o -4 route show default |head -1 |awk '{print $5}')
                         if [[ $gwdev ]]; then
                             # First, advertise the address of the device with the default gateway
                             IPADDR=$(/sbin/ip -o -4 addr show scope global dev "$gwdev" |head -1 |awk '{print $4}')
                         else
                             # Hmmm... we have no access to the Internet.  Pick an address with
                             # global scope and hope for the best.
                             IPADDR=$(/sbin/ip -o -4 addr show scope global |head -1 |awk '{print $4}')
                         fi
                     fi
                 fi

                 if [[ $IPADDR ]] ; then
                     IPADDR="${IPADDR///*}"
                 fi

                 if [[ $OS_FAMILY == darwin ]]; then
                     bcast=$(netstat -rn | grep "255.255.255.255 " | awk '{ print $6 }')
                     if [[ $bcast == "" && $IPADDR ]] ; then
                             echo "# No broadcast route set - this is required for Darwin < 10.9."
                             echo "sudo route add 255.255.255.255 $IPADDR"
                             echo "# No broadcast route set - this is required for Darwin > 10.9."
                             echo "sudo route -n add -net 255.255.255.255 $IPADDR"
                     fi
                 fi

                 if [[ $IPADDR ]] ; then
                     IPADDR="--static-ip=${IPADDR}"
                 fi

                 set +e
                 ./dr-provision --help | grep -q base-root
                 if [[ $? -eq 0 ]] ; then
                     echo "sudo ./dr-provision $IPADDR --base-root=`pwd`/drp-data --local-content=\"\" --default-content=\"\" &"
                 else
                     echo "sudo ./dr-provision $IPADDR --file-root=`pwd`/drp-data/tftpboot --data-root=drp-data/digitalrebar --local-store=\"\" --default-store=\"\" &"
                 fi
                 set -e
                 mkdir -p "`pwd`/drp-data/saas-content"
                 if [[ $NO_CONTENT == false ]] ; then
                     DEFAULT_CONTENT_FILE="`pwd`/drp-data/saas-content/default.yaml"
                     mv drp-community-content.yaml $DEFAULT_CONTENT_FILE
                 fi

                 EP="./"
             fi

             echo
             echo "# Once dr-provision is started, these commands will install the isos for the community defaults"
             echo "  ${EP}drpcli bootenvs uploadiso ubuntu-16.04-install"
             echo "  ${EP}drpcli bootenvs uploadiso centos-7-install"
             echo "  ${EP}drpcli bootenvs uploadiso sledgehammer"
             echo

             ;;
     remove)
         sudo rm -f "$bindest/dr-provision" "$bindest/drpcli" "$initdest";;
     *)
         echo "Unknown action \"$1\". Please use 'install' or 'remove'";;
esac
