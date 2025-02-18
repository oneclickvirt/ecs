#!/bin/bash
# From https://github.com/oneclickvirt/ecs
# 2024.12.08

# curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
# 或
# curl -L https://cnb.cool/oneclickvirt/ecs/-/git/raw/main/goecs.sh -o goecs.sh && chmod +x goecs.sh

cat <<"EOF"
       GGGGGGGG        OOOOOOO         EEEEEEEE      CCCCCCCCC    SSSSSSSSSS
     GG        GG    OO       OO      EE           CC           SS
    GG              OO         OO     EE          CC           SS
    GG              OO         OO     EE          CC            SS
    GG              OO         OO     EEEEEEEE    CC             SSSSSSSSSS
    GG     GGGGGG   OO         OO     EE          CC                      SS
    GG        GG    OO         OO     EE          CC                       SS
     GG      GG      OO       OO      EE           CC                     SS
      GGGGGGGG         OOOOOOO         EEEEEEEE     CCCCCCCCC    SSSSSSSSSS
EOF

cd /root >/dev/null 2>&1
if [ ! -d "/usr/bin/" ]; then
    mkdir -p "/usr/bin/"
fi
_red() { echo -e "\033[31m\033[01m$@\033[0m"; }
_green() { echo -e "\033[32m\033[01m$@\033[0m"; }
_yellow() { echo -e "\033[33m\033[01m$@\033[0m"; }
_blue() { echo -e "\033[36m\033[01m$@\033[0m"; }
reading() { read -rp "$(_green "$1")" "$2"; }

check_cdn() {
    local o_url=$1
    for cdn_url in "${cdn_urls[@]}"; do
        if curl -sL -k "$cdn_url$o_url" --max-time 6 | grep -q "success" >/dev/null 2>&1; then
            export cdn_success_url="$cdn_url"
            return
        fi
        sleep 0.5
    done
    export cdn_success_url=""
}

check_cdn_file() {
    check_cdn "https://raw.githubusercontent.com/spiritLHLS/ecs/main/back/test"
    if [ -n "$cdn_success_url" ]; then
        _green "CDN available, using CDN"
    else
        _yellow "No CDN available, no use CDN"
    fi
}

download_file() {
    local url="$1"
    local output="$2"
    if ! wget -O "$output" "$url"; then
        _yellow "wget failed, trying curl..."
        if ! curl -L -o "$output" "$url"; then
            _red "Both wget and curl failed. Unable to download the file."
            return 1
        fi
    fi
    return 0
}

check_china() {
    _yellow "正在检测IP所在区域......"
    if [[ -z "${CN}" ]]; then
        # 首先尝试通过 ipapi.co 检测
        if curl -m 6 -s https://ipapi.co/json | grep -q 'China'; then
            _yellow "根据ipapi.co提供的信息，当前IP可能在中国"
            if [ "$noninteractive" != "true" ]; then
                reading "是否使用中国镜像完成安装? ([y]/n) " input
                case $input in
                    [yY][eE][sS] | [yY] | "")
                        _green "已选择使用中国镜像"
                        CN=true
                        ;;
                    [nN][oO] | [nN])
                        _yellow "已选择不使用中国镜像"
                        CN=false
                        ;;
                    *)
                        _green "已选择使用中国镜像"
                        CN=true
                        ;;
                esac
            else
                # 在非交互模式下默认不使用中国镜像
                CN=false
            fi
        else
            CN=false
        fi
    fi
}

get_memory_size() {
    if [ -f /proc/meminfo ]; then
        local mem_kb=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        echo $((mem_kb / 1024)) # Convert to MB
        return
    fi
    if command -v free >/dev/null 2>&1; then
        local mem_kb=$(free -m | awk '/^Mem:/ {print $2}')
        echo "$mem_kb" # Already in MB
        return
    fi
    if command -v sysctl >/dev/null 2>&1; then
        local mem_bytes=$(sysctl -n hw.memsize 2>/dev/null || sysctl -n hw.physmem 2>/dev/null)
        if [ -n "$mem_bytes" ]; then
            echo $((mem_bytes / 1024 / 1024)) # Convert to MB
            return
        fi
    fi
}

cleanup_epel() {
    _yellow "Cleaning up EPEL repositories..."
    rm -f /etc/yum.repos.d/*epel*
    yum clean all
}

goecs_check() {
    # Get system and architecture info with error handling
    os=$(uname -s 2>/dev/null || echo "Unknown")
    arch=$(uname -m 2>/dev/null || echo "Unknown")
    # First check for China IP
    check_china
    # Get latest version number with multiple backup sources
    ECS_VERSION=""
    for api in \
        "https://api.github.com/repos/oneclickvirt/ecs/releases/latest" \
        "https://githubapi.spiritlhl.workers.dev/repos/oneclickvirt/ecs/releases/latest" \
        "https://githubapi.spiritlhl.top/repos/oneclickvirt/ecs/releases/latest"; do
        ECS_VERSION=$(curl -m 6 -sSL "$api" | awk -F \" '/tag_name/{gsub(/^v/,"",$4); print $4}')
        if [ -n "$ECS_VERSION" ]; then
            break
        fi
        sleep 1
    done
    if [ -z "$ECS_VERSION" ]; then
        _yellow "Unable to get version info, using default version 0.1.8"
        ECS_VERSION="0.1.8"
    fi
    # Check if original goecs command exists
    version_output=""
    for cmd_path in "goecs" "./goecs" "/usr/bin/goecs" "/usr/local/bin/goecs"; do
        if [ -x "$(command -v $cmd_path 2>/dev/null)" ]; then
            version_output=$($cmd_path -v command 2>/dev/null)
            break
        fi
    done
    if [ -n "$version_output" ]; then
        extracted_version=${version_output//v/}
        if [ -n "$extracted_version" ]; then
            ecs_version=$ECS_VERSION
            if [[ "$(echo -e "$extracted_version\n$ecs_version" | sort -V | tail -n 1)" == "$extracted_version" ]]; then
                _green "goecs version ($extracted_version) is up to date, no upgrade needed"
                return
            else
                _yellow "goecs version ($extracted_version) < $ecs_version, upgrade needed, starting in 5 seconds"
                rm -rf /usr/bin/goecs /usr/local/bin/goecs ./goecs
            fi
        fi
    else
        _green "goecs not found, installation needed, starting in 5 seconds"
    fi
    sleep 5
    # Download corresponding version with error handling
    if [[ "$CN" == true ]]; then
        _yellow "Using China mirror for download..."
        base_url="https://cnb.cool/oneclickvirt/ecs/-/git/raw/main"
    else
        cdn_urls=("https://cdn0.spiritlhl.top/" "http://cdn3.spiritlhl.net/" "http://cdn1.spiritlhl.net/" "http://cdn2.spiritlhl.net/")
        check_cdn_file
        if [ -n "$cdn_success_url" ]; then
            base_url="${cdn_success_url}https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}"
        else
            base_url="https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}"
        fi
    fi
    # Build download URL with architecture support
    local zip_file=""
    case $os in
        Linux|linux|LINUX)
            case $arch in
                x86_64|amd64|x64) zip_file="goecs_linux_amd64.zip" ;;
                i386|i686) zip_file="goecs_linux_386.zip" ;;
                aarch64|arm64|armv8|armv8l) zip_file="goecs_linux_arm64.zip" ;;
                arm|armv7l) zip_file="goecs_linux_arm.zip" ;;
                mips) zip_file="goecs_linux_mips.zip" ;;
                mipsle) zip_file="goecs_linux_mipsle.zip" ;;
                s390x) zip_file="goecs_linux_s390x.zip" ;;
                riscv64) zip_file="goecs_linux_riscv64.zip" ;;
                *) zip_file="goecs_linux_amd64.zip" ;;
            esac
            ;;
        FreeBSD|freebsd)
            case $arch in
                x86_64|amd64) zip_file="goecs_freebsd_amd64.zip" ;;
                i386|i686) zip_file="goecs_freebsd_386.zip" ;;
                arm64|aarch64) zip_file="goecs_freebsd_arm64.zip" ;;
                *) zip_file="goecs_freebsd_amd64.zip" ;;
            esac
            ;;
        Darwin|darwin)
            case $arch in
                x86_64|amd64) zip_file="goecs_darwin_amd64.zip" ;;
                arm64|aarch64) zip_file="goecs_darwin_arm64.zip" ;;
                *) zip_file="goecs_darwin_amd64.zip" ;;
            esac
            ;;
        *)
            _yellow "Unknown system $os, trying amd64 version"
            zip_file="goecs_linux_amd64.zip"
            ;;
    esac
    download_url="${base_url}/${zip_file}"
    _green "Downloading $download_url"
    # Download file with retry mechanism
    local max_retries=3
    local retry_count=0
    while [ $retry_count -lt $max_retries ]; do
        if download_file "$download_url" "goecs.zip"; then
            break
        fi
        _yellow "Download failed, retrying (${retry_count}/${max_retries})..."
        retry_count=$((retry_count + 1))
        sleep 2
    done
    if [ $retry_count -eq $max_retries ]; then
        _red "Download failed, please check your network connection or download manually"
        return 1
    fi
    # Extract and install with error handling
    if ! unzip -o goecs.zip >/dev/null 2>&1; then
        _red "Extraction failed"
        return 1
    fi
    rm -f goecs.zip README.md LICENSE README_EN.md
    # Set execution permissions and install
    chmod 777 goecs
    for install_path in "/usr/bin" "/usr/local/bin"; do
        if [ -d "$install_path" ]; then
            cp -f goecs "$install_path/"
            break
        fi
    done
    # Set system parameters
    if [ "$os" != "Darwin" ]; then
        PARAM="net.ipv4.ping_group_range"
        NEW_VALUE="0 2147483647"
        if [ -f /etc/sysctl.conf ]; then
            if grep -q "^$PARAM" /etc/sysctl.conf; then
                sed -i "s/^$PARAM.*/$PARAM = $NEW_VALUE/" /etc/sysctl.conf
            else
                echo "$PARAM = $NEW_VALUE" >> /etc/sysctl.conf
            fi
            sysctl -p >/dev/null 2>&1
        fi
    fi
    # Set special permissions
    setcap cap_net_raw=+ep goecs 2>/dev/null
    setcap cap_net_raw=+ep /usr/bin/goecs 2>/dev/null
    setcap cap_net_raw=+ep /usr/local/bin/goecs 2>/dev/null
    _green "goecs installation complete, current version:"
    goecs -v || ./goecs -v
}

InstallSysbench() {
    if [ -f "/etc/opencloudos-release" ]; then # OpenCloudOS
        Var_OSRelease="opencloudos"
    elif [ -f "/etc/centos-release" ]; then # CentOS
        Var_OSRelease="centos"
    elif [ -f "/etc/fedora-release" ]; then # Fedora
        Var_OSRelease="fedora"
    elif [ -f "/etc/redhat-release" ]; then # RedHat
        Var_OSRelease="rhel"
    elif [ -f "/etc/astra_version" ]; then # Astra
        Var_OSRelease="astra"
    elif [ -f "/etc/lsb-release" ]; then # Ubuntu
        Var_OSRelease="ubuntu"
    elif [ -f "/etc/debian_version" ]; then # Debian
        Var_OSRelease="debian"
    elif [ -f "/etc/alpine-release" ]; then # Alpine Linux
        Var_OSRelease="alpinelinux"
    elif [ -f "/etc/almalinux-release" ]; then # almalinux
        Var_OSRelease="almalinux"
    elif [ -f "/etc/arch-release" ]; then # archlinux
        Var_OSRelease="arch"
    elif [ -f "/etc/freebsd-update.conf" ]; then # freebsd
        Var_OSRelease="freebsd"
    else
        Var_OSRelease="unknown" # 未知系统分支
    fi
    local mem_size=$(get_memory_size)
    if [ -z "$mem_size" ] || [ "$mem_size" -eq 0 ]; then
        echo "Error: Unable to determine memory size or memory size is zero."
    elif [ $mem_size -lt 1024 ]; then
        _red "Warning: Your system has less than 1GB RAM (${mem_size}MB)"
        if [ "$noninteractive" != "true" ]; then
            reading "Do you want to continue with EPEL installation? (y/N): " confirm
            if [[ ! $confirm =~ ^[Yy]$ ]]; then
                _yellow "Skipping EPEL installation"
                return 1
            fi
        fi
        case "$Var_OSRelease" in
        ubuntu | debian | astra) 
            ! apt-get install -y sysbench && apt-get --fix-broken install -y && apt-get install --no-install-recommends -y sysbench ;;
        centos | rhel | almalinux | redhat | opencloudos) 
            (yum -y install epel-release && yum -y install sysbench) || (dnf install epel-release -y && dnf install sysbench -y) ;;
        fedora) 
            dnf -y install sysbench ;;
        arch) 
            pacman -S --needed --noconfirm sysbench && pacman -S --needed --noconfirm libaio && ldconfig ;;
        freebsd) 
            pkg install -y sysbench ;;
        alpinelinux)
            if [ "$noninteractive" != "true" ]; then
                reading "Do you want to continue with sysbench installation? (y/N): " confirm
                if [[ ! $confirm =~ ^[Yy]$ ]]; then
                    _yellow "Skipping sysbench installation"
                    return 1
                fi
            fi
            ALPINE_VERSION=$(grep -o '^[0-9]\+\.[0-9]\+' /etc/alpine-release)
            COMMUNITY_REPO="http://dl-cdn.alpinelinux.org/alpine/v${ALPINE_VERSION}/community"
            if grep -q "^${COMMUNITY_REPO}" /etc/apk/repositories; then
                echo "Community repository is already enabled."
            else
                echo "Enabling community repository..."
                echo "${COMMUNITY_REPO}" >> /etc/apk/repositories
                echo "Community repository has been added."
                echo "Updating apk package index..."
                apk update && echo "Package index updated successfully."
            fi
            if apk info sysbench >/dev/null 2>&1; then
                echo -e "${Msg_Info}Sysbench already installed."
            else
                apk add --no-cache sysbench
                if [ $? -ne 0 ]; then
                    echo -e "${Msg_Warning}Sysbench Module not found, installing ..." && echo -e "${Msg_Warning}SysBench Current not support Alpine Linux, Skipping..." && Var_Skip_SysBench="1"
                else
                    echo -e "${Msg_Success}Sysbench installed successfully."
                fi
            fi ;;
        *) 
            _red "Sysbench Install Error: Unknown OS release: $Var_OSRelease" ;;
        esac
        if [[ $SYSTEM =~ ^(CentOS|RHEL|AlmaLinux)$ ]]; then
        _yellow "Installing EPEL repository..."
            if ! yum -y install epel-release; then
                _red "EPEL installation failed!"
                cleanup_epel
                _yellow "Attempting to continue without EPEL..."
            fi
        fi
    fi
}

Check_SysBench() {
    if [ ! -f "/usr/bin/sysbench" ] && [ ! -f "/usr/local/bin/sysbench" ]; then
        InstallSysbench
    fi
    # 尝试编译安装
    if [ ! -f "/usr/bin/sysbench" ] && [ ! -f "/usr/local/bin/sysbench" ]; then
        echo -e "${Msg_Warning}Sysbench Module install Failure, trying compile modules ..."
        Check_Sysbench_InstantBuild
    fi
    source ~/.bashrc
    # 最终检测
    if [ "$(command -v sysbench)" ] || [ -f "/usr/bin/sysbench" ] || [ -f "/usr/local/bin/sysbench" ]; then
        _yellow "Install sysbench successfully!"
    else
        _red "SysBench Moudle install Failure! Try Restart Bench or Manually install it! (/usr/bin/sysbench)"
        _blue "Will try to test with geekbench5 instead later."
    fi
    sleep 3
}

Check_Sysbench_InstantBuild() {
    if [ "${Var_OSRelease}" = "centos" ] || [ "${Var_OSRelease}" = "rhel" ] || [ "${Var_OSRelease}" = "almalinux" ] || [ "${Var_OSRelease}" = "ubuntu" ] || [ "${Var_OSRelease}" = "debian" ] || [ "${Var_OSRelease}" = "fedora" ] || [ "${Var_OSRelease}" = "arch" ] || [ "${Var_OSRelease}" = "astra" ]; then
        local os_sysbench=${Var_OSRelease}
        if [ "$os_sysbench" = "astra" ]; then
            os_sysbench="debian"
        fi
        if [ "$os_sysbench" = "opencloudos" ]; then
            os_sysbench="centos"
        fi
        echo -e "${Msg_Info}Release Detected: ${os_sysbench}"
        echo -e "${Msg_Info}Preparing compile enviorment ..."
        prepare_compile_env "${os_sysbench}"
        echo -e "${Msg_Info}Downloading Source code (Version 1.0.20)..."
        mkdir -p /tmp/sysbench_install/src/
        mv /tmp/sysbench-1.0.20 /tmp/sysbench_install/src/
        echo -e "${Msg_Info}Compiling Sysbench Module ..."
        cd /tmp/sysbench_install/src/sysbench-1.0.20
        ./autogen.sh && ./configure --without-mysql && make -j8 && make install
        echo -e "${Msg_Info}Cleaning up ..."
        cd /tmp && rm -rf /tmp/sysbench_install/src/sysbench*
    else
        echo -e "${Msg_Warning}Unsupported operating system: ${Var_OSRelease}"
    fi
}

prepare_compile_env() {
    local system="$1"
    if [ "${system}" = "centos" ] || [ "${system}" = "rhel" ] || [ "${system}" = "almalinux" ]; then
        yum install -y epel-release
        yum install -y wget curl make gcc gcc-c++ make automake libtool pkgconfig libaio-devel
    elif [ "${system}" = "ubuntu" ] || [ "${system}" = "debian" ]; then
        ! apt-get update && apt-get --fix-broken install -y && apt-get update
        ! apt-get -y install --no-install-recommends curl wget make automake libtool pkg-config libaio-dev unzip && apt-get --fix-broken install -y && apt-get -y install --no-install-recommends curl wget make automake libtool pkg-config libaio-dev unzip
    elif [ "${system}" = "fedora" ]; then
        dnf install -y wget curl gcc gcc-c++ make automake libtool pkgconfig libaio-devel
    elif [ "${system}" = "arch" ]; then
        pacman -S --needed --noconfirm wget curl gcc gcc make automake libtool pkgconfig libaio lib32-libaio
    else
        echo -e "${Msg_Warning}Unsupported operating system: ${system}"
    fi
}

env_check() {
    REGEX=("debian|astra" "ubuntu" "centos|red hat|kernel|oracle linux|alma|rocky" "'amazon linux'" "fedora" "arch" "freebsd" "alpine" "openbsd" "opencloudos")
    RELEASE=("Debian" "Ubuntu" "CentOS" "CentOS" "Fedora" "Arch" "FreeBSD" "Alpine" "OpenBSD" "OpenCloudOS")
    PACKAGE_UPDATE=("apt-get update" "apt-get update" "yum -y update" "yum -y update" "yum -y update" "pacman -Sy" "pkg update" "apk update" "pkg_add -qu" "yum -y update")
    PACKAGE_INSTALL=("apt-get -y install" "apt-get -y install" "yum -y install" "yum -y install" "yum -y install" "pacman -Sy --noconfirm --needed" "pkg install -y" "apk add --no-cache" "pkg_add -I" "yum -y install")
    PACKAGE_REMOVE=("apt-get -y remove" "apt-get -y remove" "yum -y remove" "yum -y remove" "yum -y remove" "pacman -Rsc --noconfirm" "pkg delete" "apk del" "pkg_delete -I" "yum -y remove")
    PACKAGE_UNINSTALL=("apt-get -y autoremove" "apt-get -y autoremove" "yum -y autoremove" "yum -y autoremove" "yum -y autoremove" "pacman -Rns --noconfirm" "pkg autoremove" "apk autoremove" "pkg_delete -a" "yum -y autoremove")
    # Check system information
    if [ -f /etc/opencloudos-release ]; then
        SYS="opencloudos"
    elif [ -s /etc/os-release ]; then
        SYS="$(grep -i pretty_name /etc/os-release | cut -d \" -f2)"
    elif [ -x "$(type -p hostnamectl)" ]; then
        SYS="$(hostnamectl | grep -i system | cut -d : -f2 | xargs)"
    elif [ -x "$(type -p lsb_release)" ]; then
        SYS="$(lsb_release -sd)"
    elif [ -s /etc/lsb-release ]; then
        SYS="$(grep -i description /etc/lsb-release | cut -d \" -f2)"
    elif [ -s /etc/redhat-release ]; then
        SYS="$(grep . /etc/redhat-release)"
    elif [ -s /etc/issue ]; then
        SYS="$(grep . /etc/issue | cut -d '\' -f1 | sed '/^[ ]*$/d')"
    else
        SYS="$(uname -s)"
    fi
    # Match operating system
    SYSTEM=""
    for ((int = 0; int < ${#REGEX[@]}; int++)); do
        if [[ $(echo "$SYS" | tr '[:upper:]' '[:lower:]') =~ ${REGEX[int]} ]]; then
            SYSTEM="${RELEASE[int]}"
            UPDATE_CMD=${PACKAGE_UPDATE[int]}
            INSTALL_CMD=${PACKAGE_INSTALL[int]}
            REMOVE_CMD=${PACKAGE_REMOVE[int]}
            UNINSTALL_CMD=${PACKAGE_UNINSTALL[int]}
            break
        fi
    done
    # If system is unrecognized, try common package managers
    if [ -z "$SYSTEM" ]; then
        _yellow "Unable to recognize system, trying common package managers..."
        # Try apt
        if command -v apt-get >/dev/null 2>&1; then
            SYSTEM="Unknown-Debian"
            UPDATE_CMD="apt-get update"
            INSTALL_CMD="apt-get -y install"
            REMOVE_CMD="apt-get -y remove"
            UNINSTALL_CMD="apt-get -y autoremove"
        # Try yum
        elif command -v yum >/dev/null 2>&1; then
            SYSTEM="Unknown-RHEL"
            UPDATE_CMD="yum -y update"
            INSTALL_CMD="yum -y install"
            REMOVE_CMD="yum -y remove"
            UNINSTALL_CMD="yum -y autoremove"
        # Try dnf
        elif command -v dnf >/dev/null 2>&1; then
            SYSTEM="Unknown-Fedora"
            UPDATE_CMD="dnf -y update"
            INSTALL_CMD="dnf -y install"
            REMOVE_CMD="dnf -y remove"
            UNINSTALL_CMD="dnf -y autoremove"
        # Try pacman
        elif command -v pacman >/dev/null 2>&1; then
            SYSTEM="Unknown-Arch"
            UPDATE_CMD="pacman -Sy"
            INSTALL_CMD="pacman -S --noconfirm"
            REMOVE_CMD="pacman -R --noconfirm"
            UNINSTALL_CMD="pacman -Rns --noconfirm"
        # Try apk
        elif command -v apk >/dev/null 2>&1; then
            SYSTEM="Unknown-Alpine"
            UPDATE_CMD="apk update"
            INSTALL_CMD="apk add"
            REMOVE_CMD="apk del"
            UNINSTALL_CMD="apk del"
        elif command -v zypper >/dev/null 2>&1; then
            SYSTEM="Unknown-SLES"
            UPDATE_CMD="zypper refresh"
            INSTALL_CMD="zypper install -y"
            REMOVE_CMD="zypper remove -y"
            UNINSTALL_CMD="zypper remove -y"
        else
            _red "Unable to recognize package manager, exiting installation"
            exit 1
        fi
    fi
    _green "System information: $SYSTEM"
    _green "Update command: $UPDATE_CMD"
    _green "Install command: $INSTALL_CMD"
    cdn_urls=("https://cdn0.spiritlhl.top/" "http://cdn3.spiritlhl.net/" "http://cdn1.spiritlhl.net/" "http://cdn2.spiritlhl.net/")
    check_cdn_file
    _yellow "Warning: System update will be performed"
    _yellow "This operation may:"
    _yellow "1. Take considerable time"
    _yellow "2. Cause temporary network interruptions"
    _yellow "3. Impact system stability"
    _yellow "4. Affect subsequent system startups"
    if [ "$noninteractive" != "true" ]; then
        reading "Continue with system update? (y/N): " update_confirm
        if [[ ! $update_confirm =~ ^[Yy]$ ]]; then
            _yellow "Skipping system update"
            _yellow "Note: Some packages may fail to install"
        else
            _green "Updating system package manager..."
            if ! ${UPDATE_CMD} 2>/dev/null; then
                _red "System update failed!"
            fi
        fi
    fi
    # Install necessary commands
    for cmd in sudo wget tar unzip iproute2 systemd-detect-virt dd fio; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            _green "Installing $cmd"
            ${INSTALL_CMD} "$cmd"
        fi
    done
    # sysbench installation
    if ! command -v sysbench >/dev/null 2>&1; then
        _green "Installing sysbench"
        ${INSTALL_CMD} sysbench
        if [ $? -ne 0 ]; then
            echo "Unable to download sysbench through package manager, attempting compilation..."
            wget -O /tmp/sysbench.zip "${cdn_success_url}https://github.com/akopytov/sysbench/archive/1.0.20.zip" || curl -Lk -o /tmp/sysbench.zip "${cdn_success_url}https://github.com/akopytov/sysbench/archive/1.0.20.zip"
            if [ ! -f /tmp/sysbench.zip ]; then
                wget -q -O /tmp/sysbench.zip "https://hub.fgit.cf/akopytov/sysbench/archive/1.0.20.zip"
            fi
            chmod +x /tmp/sysbench.zip
            unzip /tmp/sysbench.zip -d /tmp
            Check_SysBench
        fi
    fi
    # geekbench and speedtest installation
    if ! command -v geekbench >/dev/null 2>&1; then
        _green "Installing geekbench"
        curl -L "${cdn_success_url}https://raw.githubusercontent.com/oneclickvirt/cputest/main/dgb.sh" -o dgb.sh && chmod +x dgb.sh
        bash dgb.sh -v gb5
        rm -rf dgb.sh
    fi
    if ! command -v speedtest >/dev/null 2>&1; then
        _green "Installing speedtest"
        curl -L "${cdn_success_url}https://raw.githubusercontent.com/oneclickvirt/speedtest/main/dspt.sh" -o dspt.sh && chmod +x dspt.sh
        bash dspt.sh
        rm -rf dspt.sh
        rm -rf speedtest.tar.gz
    fi
    if ! command -v ping >/dev/null 2>&1; then
        _green "Installing ping"
        ${INSTALL_CMD} iputils-ping >/dev/null 2>&1
        ${INSTALL_CMD} ping >/dev/null 2>&1
    fi
    # MacOS support
    if [ "$(uname -s)" = "Darwin" ]; then
        echo "Detected MacOS, installing sysbench iproute2mac fio..."
        brew install --force sysbench iproute2mac fio
    else
        if ! grep -q "^net.ipv4.ping_group_range = 0 2147483647$" /etc/sysctl.conf; then
            echo "net.ipv4.ping_group_range = 0 2147483647" >> /etc/sysctl.conf
            sysctl -p
        fi
    fi
    _green "Environment preparation complete."
    _green "Next command is: ./goecs.sh install"
}

uninstall_goecs() {
    rm -rf /root/goecs
    rm -rf /usr/bin/goecs
    _green "The command (goecs) has been uninstalled."
}

show_help() {
    cat <<"EOF"
可用命令：

./goecs.sh env            检查并安装依赖包
                          警告: 此命令会执行系统更新(可选择)，可能:
                          1. 耗时较长
                          2. 导致网络短暂中断
                          3. 影响系统稳定性
                          4. 影响后续系统启动
                          对于内存小于1GB的系统，还可能导致:
                          1. 系统卡死
                          2. SSH连接中断
                          3. 关键服务失败
                          推荐：
                          环境依赖安装过程中挂起执行

                          必需组件:
                          sysbench/geekbench (CPU性能测试必需)
                          
                          可选组件:
                          sudo, tar, unzip, dd, fio
                          speedtest (网络测试)
                          ping (网络连通性测试)
                          systemd-detect-virt/dmidecode (系统信息检测)

./goecs.sh install        安装 goecs 命令
./goecs.sh upgrade        升级 goecs 命令
./goecs.sh uninstall      卸载 goecs 命令
./goecs.sh help           显示此消息

Available commands:

./goecs.sh env             Check and Install dependencies
                           Warning: This command performs system update(optional), which may:
                           1. Take considerable time
                           2. Cause temporary network interruptions
                           3. Impact system stability
                           4. Affect subsequent system startups
                           For systems with less than 1GB RAM, additional risks:
                           1. System freeze
                           2. SSH connection loss
                           3. Critical service failures
                           Recommended:
                           Hanging execution during environment dependency installation
                           
                           Required components:
                           sysbench/geekbench (Required for CPU testing)
                           
                           Optional components:
                           sudo, tar, unzip, dd, fio
                           speedtest (Network testing)
                           ping (Network connectivity)
                           systemd-detect-virt/dmidecode (System info detection)

./goecs.sh install         Install goecs command
./goecs.sh upgrade         Upgrade goecs command
./goecs.sh uninstall       Uninstall goecs command
./goecs.sh help            Show this message
EOF
}

case "$1" in
"help")
    show_help
    ;;
"env")
    env_check
    ;;
"install" | "upgrade")
    goecs_check
    ;;
"uninstall")
    uninstall_goecs
    ;;
*)
    echo "No command found."
    echo
    show_help
    ;;
esac

