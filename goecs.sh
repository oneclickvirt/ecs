#!/bin/sh
# From https://github.com/oneclickvirt/ecs
# 2025.10.08

# curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
# 或
# curl -L https://cnb.cool/oneclickvirt/ecs/-/git/raw/main/goecs.sh -o goecs.sh && chmod +x goecs.sh

cat <<"EOF"
  ,ad8888ba,     ,ad8888ba,    88888888888  ,ad8888ba,   ad88888ba
 d8"'    `"8b   d8"'    `"8b   88          d8"'    `"8b d8"     "8b
d8'            d8'        `8b  88         d8'           Y8a
88             88          88  88aaaaa    88             `"Y8aaaaa,
88      88888  88          88  88"""""    88               `"""""8b,
Y8,        88  Y8,        ,8P  88         Y8,                    `8b
 Y8a.    .a88   Y8a.    .a8P   88          Y8a.    .a8P  Y8a     a8P
  `"Y88888P"     `"Y8888Y"'    88888888888  `"Y8888Y"'    "Y88888P"
EOF
cd /root >/dev/null 2>&1
if [ ! -d "/usr/bin/" ]; then
    mkdir -p "/usr/bin/"
fi
_red() { printf "\033[31m\033[01m%s\033[0m\n" "$*"; }
_green() { printf "\033[32m\033[01m%s\033[0m\n" "$*"; }
_yellow() { printf "\033[33m\033[01m%s\033[0m\n" "$*"; }
_blue() { printf "\033[36m\033[01m%s\033[0m\n" "$*"; }
reading() { 
    printf "\033[32m\033[01m%s\033[0m" "$1"
    read "$2"
}

check_cdn() {
    local o_url="$1"
    local cdn_url
    for cdn_url in $cdn_urls; do
        if curl -4 -sL -k "$cdn_url$o_url" --max-time 6 | grep -q "success" >/dev/null 2>&1; then
            cdn_success_url="$cdn_url"
            return 0
        fi
        sleep 0.5
    done
    cdn_success_url=""
    return 1
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
    if ! wget -O "$output" "$url" 2>/dev/null; then
        _yellow "wget failed, trying curl..."
        if ! curl -L -o "$output" "$url" 2>/dev/null; then
            _red "Both wget and curl failed. Unable to download the file."
            return 1
        fi
    fi
    return 0
}

check_china() {
    _yellow "正在检测IP所在区域......"
    if [ -z "${CN}" ]; then
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
        local mem_kb
        mem_kb=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        echo $((mem_kb / 1024)) # Convert to MB
        return 0
    fi
    if command -v free >/dev/null 2>&1; then
        local mem_kb
        mem_kb=$(free -m | awk '/^Mem:/ {print $2}')
        echo "$mem_kb" # Already in MB
        return 0
    fi
    if command -v sysctl >/dev/null 2>&1; then
        local mem_bytes
        mem_bytes=$(sysctl -n hw.memsize 2>/dev/null || sysctl -n hw.physmem 2>/dev/null)
        if [ -n "$mem_bytes" ]; then
            echo $((mem_bytes / 1024 / 1024)) # Convert to MB
            return 0
        fi
    fi
    echo 0
    return 1
}

cleanup_epel() {
    _yellow "Cleaning up EPEL repositories..."
    rm -f /etc/yum.repos.d/*epel*
    yum clean all >/dev/null 2>&1
}

goecs_check() {
    if command -v apt-get >/dev/null 2>&1; then
        INSTALL_CMD="apt-get -y install"
    elif command -v yum >/dev/null 2>&1; then
        INSTALL_CMD="yum -y install"
    elif command -v dnf >/dev/null 2>&1; then
        INSTALL_CMD="dnf -y install"
    elif command -v pacman >/dev/null 2>&1; then
        INSTALL_CMD="pacman -S --noconfirm"
    elif command -v apk >/dev/null 2>&1; then
        INSTALL_CMD="apk add"
    elif command -v zypper >/dev/null 2>&1; then
        INSTALL_CMD="zypper install -y"
    fi
    if ! command -v unzip >/dev/null 2>&1; then
        _green "Installing unzip"
        ${INSTALL_CMD} unzip
    fi
    if ! command -v curl >/dev/null 2>&1; then
        _green "Installing curl"
        ${INSTALL_CMD} curl
    fi
    os=$(uname -s 2>/dev/null || echo "Unknown")
    arch=$(uname -m 2>/dev/null || echo "Unknown")
    check_china
    ECS_VERSION="0.1.104"
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
        _yellow "Unable to get version info, using default version 0.1.104"
        ECS_VERSION="0.1.104"
    fi
    version_output=""
    for cmd_path in "goecs" "./goecs" "/usr/bin/goecs" "/usr/local/bin/goecs"; do
        if command -v "$cmd_path" >/dev/null 2>&1; then
            version_output=$($cmd_path -v command 2>/dev/null)
            break
        fi
    done
    if [ -n "$version_output" ]; then
        extracted_version=${version_output#*v}
        extracted_version=${extracted_version#v}
        if [ -n "$extracted_version" ]; then
            ecs_version=$ECS_VERSION
            if [ "$(printf '%s\n%s\n' "$extracted_version" "$ecs_version" | sort -V | tail -n 1)" = "$extracted_version" ]; then
                _green "goecs version ($extracted_version) is up to date, no upgrade needed"
                return 0
            else
                _yellow "goecs version ($extracted_version) < $ecs_version, upgrade needed, starting in 5 seconds"
                rm -rf /usr/bin/goecs /usr/local/bin/goecs ./goecs
            fi
        fi
    else
        _green "goecs not found, installation needed, starting in 5 seconds"
    fi
    sleep 5
    if [ "$CN" = "true" ]; then
        _yellow "Using China mirror for download..."
        base_url="https://cnb.cool/oneclickvirt/ecs/-/git/raw/main"
    else
        cdn_urls="https://cdn0.spiritlhl.top/ http://cdn3.spiritlhl.net/ http://cdn1.spiritlhl.net/ http://cdn2.spiritlhl.net/"
        check_cdn_file
        if [ -n "$cdn_success_url" ]; then
            base_url="${cdn_success_url}https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}"
        else
            base_url="https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}"
        fi
    fi
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
    if ! unzip -o goecs.zip >/dev/null 2>&1; then
        _red "Extraction failed"
        return 1
    fi
    rm -f goecs.zip README.md LICENSE README_EN.md
    chmod 777 goecs
    for install_path in "/usr/bin" "/usr/local/bin"; do
        if [ -d "$install_path" ]; then
            cp -f goecs "$install_path/"
            break
        fi
    done
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
    local mem_size
    mem_size=$(get_memory_size)
    if [ -z "$mem_size" ] || [ "$mem_size" -eq 0 ]; then
        echo "Error: Unable to determine memory size or memory size is zero."
    elif [ "$mem_size" -lt 1024 ]; then
        _red "Warning: Your system has less than 1GB RAM (${mem_size}MB)"
        if [ "$noninteractive" != "true" ]; then
            reading "Do you want to continue with EPEL installation? (y/N): " confirm
            case "$confirm" in
                [Yy]*)
                    ;;
                *)
                    _yellow "Skipping EPEL installation"
                    return 1
                    ;;
            esac
        fi
        case "$Var_OSRelease" in
        ubuntu | debian | astra)
            if ! apt-get install -y sysbench; then
                apt-get --fix-broken install -y
                apt-get install --no-install-recommends -y sysbench
            fi
            ;;
        centos | rhel | almalinux | redhat | opencloudos)
            if ! yum -y install epel-release || ! yum -y install sysbench; then
                if command -v dnf >/dev/null 2>&1; then
                    dnf install epel-release -y
                    dnf install sysbench -y
                fi
            fi
            ;;
        fedora)
            dnf -y install sysbench ;;
        arch)
            pacman -S --needed --noconfirm sysbench
            pacman -S --needed --noconfirm libaio
            ldconfig
            ;;
        freebsd)
            pkg install -y sysbench ;;
        alpinelinux)
            if [ "$noninteractive" != "true" ]; then
                reading "Do you want to continue with sysbench installation? (y/N): " confirm
                case "$confirm" in
                    [Yy]*)
                        ;;
                    *)
                        _yellow "Skipping sysbench installation"
                        return 1
                        ;;
                esac
            fi
            ALPINE_VERSION=$(grep -o '^[0-9]\+\.[0-9]\+' /etc/alpine-release)
            COMMUNITY_REPO="http://dl-cdn.alpinelinux.org/alpine/v${ALPINE_VERSION}/community"
            if ! grep -q "^${COMMUNITY_REPO}" /etc/apk/repositories; then
                echo "Enabling community repository..."
                echo "${COMMUNITY_REPO}" >> /etc/apk/repositories
                echo "Community repository has been added."
                echo "Updating apk package index..."
                apk update && echo "Package index updated successfully."
            else
                echo "Community repository is already enabled."
            fi
            if apk info sysbench >/dev/null 2>&1; then
                echo "Sysbench already installed."
            else
                if ! apk add --no-cache sysbench; then
                    echo "Sysbench Module not found, installing ..."
                    echo "SysBench Current not support Alpine Linux, Skipping..."
                    Var_Skip_SysBench="1"
                else
                    echo "Sysbench installed successfully."
                fi
            fi
            ;;
        *)
            _red "Sysbench Install Error: Unknown OS release: $Var_OSRelease" ;;
        esac
        case "$SYSTEM" in
            CentOS|RHEL|AlmaLinux)
                _yellow "Installing EPEL repository..."
                if ! yum -y install epel-release; then
                    _red "EPEL installation failed!"
                    cleanup_epel
                    _yellow "Attempting to continue without EPEL..."
                fi
                ;;
        esac
    fi
}

env_check() {
    # 检测是否为 macOS 系统
    if [ "$(uname -s)" = "Darwin" ]; then
        _green "Detected macOS system"
        _green "macOS has built-in tools, skipping dependency installation"
        _green "Environment preparation complete."
        _green "Next command is: ./goecs.sh install"
        return 0
    fi
    
    if [ -f /etc/opencloudos-release ]; then
        SYS="opencloudos"
    elif [ -s /etc/os-release ]; then
        SYS="$(grep -i pretty_name /etc/os-release | cut -d \" -f2)"
    elif command -v hostnamectl >/dev/null 2>&1; then
        SYS="$(hostnamectl | grep -i system | cut -d : -f2 | sed 's/^ *//')"
    elif command -v lsb_release >/dev/null 2>&1; then
        SYS="$(lsb_release -sd)"
    elif [ -s /etc/lsb-release ]; then
        SYS="$(grep -i description /etc/lsb-release | cut -d \" -f2)"
    elif [ -s /etc/redhat-release ]; then
        SYS="$(cat /etc/redhat-release)"
    elif [ -s /etc/issue ]; then
        SYS="$(head -n1 /etc/issue | cut -d '\' -f1 | sed '/^[ ]*$/d')"
    else
        SYS="$(uname -s)"
    fi
    SYSTEM=""
    sys_lower=$(echo "$SYS" | tr '[:upper:]' '[:lower:]')
    if echo "$sys_lower" | grep -E "debian|astra" >/dev/null 2>&1; then
        SYSTEM="Debian"
        UPDATE_CMD="apt-get update"
        INSTALL_CMD="apt-get -y install"
        REMOVE_CMD="apt-get -y remove"
        UNINSTALL_CMD="apt-get -y autoremove"
    elif echo "$sys_lower" | grep -E "ubuntu" >/dev/null 2>&1; then
        SYSTEM="Ubuntu"
        UPDATE_CMD="apt-get update"
        INSTALL_CMD="apt-get -y install"
        REMOVE_CMD="apt-get -y remove"
        UNINSTALL_CMD="apt-get -y autoremove"
    elif echo "$sys_lower" | grep -E "centos|red hat|kernel|oracle linux|alma|rocky" >/dev/null 2>&1; then
        SYSTEM="CentOS"
        UPDATE_CMD="yum -y update"
        INSTALL_CMD="yum -y install"
        REMOVE_CMD="yum -y remove"
        UNINSTALL_CMD="yum -y autoremove"
    elif echo "$sys_lower" | grep -E "amazon linux" >/dev/null 2>&1; then
        SYSTEM="CentOS"
        UPDATE_CMD="yum -y update"
        INSTALL_CMD="yum -y install"
        REMOVE_CMD="yum -y remove"
        UNINSTALL_CMD="yum -y autoremove"
    elif echo "$sys_lower" | grep -E "fedora" >/dev/null 2>&1; then
        SYSTEM="Fedora"
        UPDATE_CMD="yum -y update"
        INSTALL_CMD="yum -y install"
        REMOVE_CMD="yum -y remove"
        UNINSTALL_CMD="yum -y autoremove"
    elif echo "$sys_lower" | grep -E "arch" >/dev/null 2>&1; then
        SYSTEM="Arch"
        UPDATE_CMD="pacman -Sy"
        INSTALL_CMD="pacman -Sy --noconfirm --needed"
        REMOVE_CMD="pacman -Rsc --noconfirm"
        UNINSTALL_CMD="pacman -Rns --noconfirm"
    elif echo "$sys_lower" | grep -E "freebsd" >/dev/null 2>&1; then
        SYSTEM="FreeBSD"
        UPDATE_CMD="pkg update"
        INSTALL_CMD="pkg install -y"
        REMOVE_CMD="pkg delete"
        UNINSTALL_CMD="pkg autoremove"
    elif echo "$sys_lower" | grep -E "alpine" >/dev/null 2>&1; then
        SYSTEM="Alpine"
        UPDATE_CMD="apk update"
        INSTALL_CMD="apk add --no-cache"
        REMOVE_CMD="apk del"
        UNINSTALL_CMD="apk autoremove"
    elif echo "$sys_lower" | grep -E "openbsd" >/dev/null 2>&1; then
        SYSTEM="OpenBSD"
        UPDATE_CMD="pkg_add -qu"
        INSTALL_CMD="pkg_add -I"
        REMOVE_CMD="pkg_delete -I"
        UNINSTALL_CMD="pkg_delete -a"
    elif echo "$sys_lower" | grep -E "opencloudos" >/dev/null 2>&1; then
        SYSTEM="OpenCloudOS"
        UPDATE_CMD="yum -y update"
        INSTALL_CMD="yum -y install"
        REMOVE_CMD="yum -y remove"
        UNINSTALL_CMD="yum -y autoremove"
    fi
    if [ -z "$SYSTEM" ]; then
        _yellow "Unable to recognize system, trying common package managers..."
        if command -v apt-get >/dev/null 2>&1; then
            SYSTEM="Unknown-Debian"
            UPDATE_CMD="apt-get update"
            INSTALL_CMD="apt-get -y install"
            REMOVE_CMD="apt-get -y remove"
            UNINSTALL_CMD="apt-get -y autoremove"
        elif command -v yum >/dev/null 2>&1; then
            SYSTEM="Unknown-RHEL"
            UPDATE_CMD="yum -y update"
            INSTALL_CMD="yum -y install"
            REMOVE_CMD="yum -y remove"
            UNINSTALL_CMD="yum -y autoremove"
        elif command -v dnf >/dev/null 2>&1; then
            SYSTEM="Unknown-Fedora"
            UPDATE_CMD="dnf -y update"
            INSTALL_CMD="dnf -y install"
            REMOVE_CMD="dnf -y remove"
            UNINSTALL_CMD="dnf -y autoremove"
        elif command -v pacman >/dev/null 2>&1; then
            SYSTEM="Unknown-Arch"
            UPDATE_CMD="pacman -Sy"
            INSTALL_CMD="pacman -S --noconfirm"
            REMOVE_CMD="pacman -R --noconfirm"
            UNINSTALL_CMD="pacman -Rns --noconfirm"
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
    cdn_urls="https://cdn0.spiritlhl.top/ http://cdn3.spiritlhl.net/ http://cdn1.spiritlhl.net/ http://cdn2.spiritlhl.net/"
    check_cdn_file
    _yellow "Warning: System update will be performed"
    _yellow "This operation may:"
    _yellow "1. Take considerable time"
    _yellow "2. Cause temporary network interruptions"
    _yellow "3. Impact system stability"
    _yellow "4. Affect subsequent system startups"
    if [ "$noninteractive" != "true" ]; then
        reading "Continue with system update? (y/N): " update_confirm
        case "$update_confirm" in
            [Yy]*)
                _green "Updating system package manager..."
                if ! ${UPDATE_CMD} 2>/dev/null; then
                    _red "System update failed!"
                fi
                ;;
            *)
                _yellow "Skipping system update"
                _yellow "Note: Some packages may fail to install"
                ;;
        esac
    fi
    for cmd in sudo wget tar unzip iproute2 systemd-detect-virt dd fio; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            _green "Installing $cmd"
            ${INSTALL_CMD} "$cmd"
        fi
    done
    if ! command -v sysbench >/dev/null 2>&1; then
        _green "Installing sysbench"
        if ! ${INSTALL_CMD} sysbench; then
            _red "Unable to install sysbench through package manager"
            _yellow "Sysbench installation skipped"
        fi
    fi
    if ! command -v geekbench >/dev/null 2>&1; then
        _green "Installing geekbench"
        curl -L "${cdn_success_url}https://raw.githubusercontent.com/oneclickvirt/cputest/main/dgb.sh" -o dgb.sh && chmod +x dgb.sh
        sh dgb.sh -v gb5
        rm -rf dgb.sh
    fi
    if ! command -v speedtest >/dev/null 2>&1; then
        _green "Installing speedtest"
        curl -L "${cdn_success_url}https://raw.githubusercontent.com/oneclickvirt/speedtest/main/dspt.sh" -o dspt.sh && chmod +x dspt.sh
        sh dspt.sh
        rm -rf dspt.sh
        rm -rf speedtest.tar.gz
    fi
    if ! command -v ping >/dev/null 2>&1; then
        _green "Installing ping"
        ${INSTALL_CMD} iputils-ping >/dev/null 2>&1 || ${INSTALL_CMD} ping >/dev/null 2>&1
    fi
    if ! grep -q "^net.ipv4.ping_group_range = 0 2147483647$" /etc/sysctl.conf 2>/dev/null; then
        echo "net.ipv4.ping_group_range = 0 2147483647" >> /etc/sysctl.conf 2>/dev/null
        sysctl -p >/dev/null 2>&1
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
                          注意: macOS系统会自动跳过依赖安装
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

                          可选组件:
                          sysbench/geekbench (CPU性能测试)
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
                           Note: macOS systems will skip dependency installation
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

                           Optional components:
                           sysbench/geekbench (CPU testing)
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
