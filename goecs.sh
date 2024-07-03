#!/bin/bash
#From https://github.com/oneclickvirt/ecs
#2024.06.29

# curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh

cat <<"EOF"
       GGGGGGGG        OOOOOOO         EEEEEEEE      CCCCCCCCC    SSSSSSSSSS
     GG        GG    OO       OO      EE           CC           SS
    GG              OO         OO     EE          CC           SS
    GG              OO         OO     EE          CC            SS
    GG              OO         OO     EEEEEEEE    CC             SSSSSSSSSS
    GG    GGGGGG    OO         OO     EE          CC                      SS
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
        echo "CDN available, using CDN"
    else
        echo "No CDN available, no use CDN"
    fi
}

goecs_check() {
    os=$(uname -s)
    arch=$(uname -m)
    ECS_VERSION=$(curl -m 6 -sSL "https://api.github.com/repos/oneclickvirt/ecs/releases/latest" | awk -F \" '/tag_name/{gsub(/^v/,"",$4); print $4}')
    # 如果 https://api.github.com/ 请求失败，则使用 https://githubapi.spiritlhl.workers.dev/ ，此时可能宿主机无IPV4网络
    if [ -z "$ECS_VERSION" ]; then
        ECS_VERSION=$(curl -m 6 -sSL "https://githubapi.spiritlhl.workers.dev/repos/oneclickvirt/ecs/releases/latest" | awk -F \" '/tag_name/{gsub(/^v/,"",$4); print $4}')
    fi
    # 如果 https://githubapi.spiritlhl.workers.dev/ 请求失败，则使用 https://githubapi.spiritlhl.top/ ，此时可能宿主机在国内
    if [ -z "$ECS_VERSION" ]; then
        ECS_VERSION=$(curl -m 6 -sSL "https://githubapi.spiritlhl.top/repos/oneclickvirt/ecs/releases/latest" | awk -F \" '/tag_name/{gsub(/^v/,"",$4); print $4}')
    fi
    # 检测原始goecs命令是否存在，若存在则升级，不存在则安装
    version_output=$(goecs -v command 2>/dev/null || ./goecs -v command 2>/dev/null)
    if [ $? -eq 0 ]; then
        extracted_version=$(echo "${version_output//v/}")
        if [ -n "$extracted_version" ]; then
            ecs_version=$ECS_VERSION
            if [[ "$(echo -e "$extracted_version\n$ecs_version" | sort -V | tail -n 1)" == "$extracted_version" ]]; then
                echo "goecs version ($extracted_version) is latest, no need to upgrade."
                return
            else
                echo "goecs version ($extracted_version) < $ecs_version, need to upgrade, 5 seconds later will start to upgrade"
                rm -rf /usr/bin/goecs
                rm -rf goecs
            fi
        fi
    else
        echo "Can not find goecs, need to download and install, 5 seconds later will start to install"
    fi
    sleep 5
    cdn_urls=("https://cdn0.spiritlhl.top/" "http://cdn3.spiritlhl.net/" "http://cdn1.spiritlhl.net/" "http://cdn2.spiritlhl.net/")
    check_cdn_file
    case $os in
    Linux)
        case $arch in
        "x86_64" | "x86" | "amd64" | "x64")
            wget -O goecs.tar.gz "${cdn_success_url}https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}/ecs_${ECS_VERSION}_linux_amd64.tar.gz"
            ;;
        "i386" | "i686")
            wget -O goecs.tar.gz "${cdn_success_url}https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}/ecs_${ECS_VERSION}_linux_386.tar.gz"
            ;;
        "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
            wget -O goecs.tar.gz "${cdn_success_url}https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}/ecs_${ECS_VERSION}_linux_arm64.tar.gz"
            ;;
        *)
            echo "Unsupported architecture: $arch"
            exit 1
            ;;
        esac
        ;;
    Darwin)
        case $arch in
        "x86_64" | "x86" | "amd64" | "x64")
            wget -O goecs.tar.gz "${cdn_success_url}https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}/ecs_${ECS_VERSION}_linux_amd64.tar.gz"
            ;;
        "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
            wget -O goecs.tar.gz "${cdn_success_url}https://github.com/oneclickvirt/ecs/releases/download/v${ECS_VERSION}/ecs_${ECS_VERSION}_linux_arm64.tar.gz"
            ;;
        *)
            echo "Unsupported architecture: $arch"
            exit 1
            ;;
        esac
        ;;
    *)
        echo "Unsupported operating system: $os"
        exit 1
        ;;
    esac
    tar -xvf goecs.tar.gz
    rm -rf goecs.tar.gz
    rm -rf README.md
    rm -rf LICENSE
    mv ecs goecs
    sleep 1
    chmod 777 goecs
    rm -rf /usr/bin/goecs
    sleep 1
    cp goecs /usr/bin/goecs
    echo "goecs version:"
    goecs -v || ./goecs -v
}

InstallSysbench() {
    if [ -f "/etc/centos-release" ]; then # CentOS
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
    # rockylinux
    elif [ -f "/etc/arch-release" ]; then # archlinux
        Var_OSRelease="arch"
    elif [ -f "/etc/freebsd-update.conf" ]; then # freebsd
        Var_OSRelease="freebsd"
    else
        Var_OSRelease="unknown" # 未知系统分支
    fi
    case "$Var_OSRelease" in
    ubuntu | debian | astra) ! apt-get install -y sysbench && apt-get --fix-broken install -y && apt-get install --no-install-recommends -y sysbench ;;
    centos | rhel | almalinux | redhat) (yum -y install epel-release && yum -y install sysbench) || (dnf install epel-release -y && dnf install sysbench -y) ;;
    fedora) dnf -y install sysbench ;;
    arch) pacman -S --needed --noconfirm sysbench && pacman -S --needed --noconfirm libaio && ldconfig ;;
    freebsd) pkg install -y sysbench ;;
    alpinelinux) echo -e "${Msg_Warning}Sysbench Module not found, installing ..." && echo -e "${Msg_Warning}SysBench Current not support Alpine Linux, Skipping..." && Var_Skip_SysBench="1" ;;
    *) echo "Error: Unknown OS release: $os_release" ;;
    esac
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
        _blue "Will try to test with geekbench5 instead later on"
        test_cpu_type="gb5"
    fi
    sleep 3
}

Check_Sysbench_InstantBuild() {
    if [ "${Var_OSRelease}" = "centos" ] || [ "${Var_OSRelease}" = "rhel" ] || [ "${Var_OSRelease}" = "almalinux" ] || [ "${Var_OSRelease}" = "ubuntu" ] || [ "${Var_OSRelease}" = "debian" ] || [ "${Var_OSRelease}" = "fedora" ] || [ "${Var_OSRelease}" = "arch" ] || [ "${Var_OSRelease}" = "astra" ]; then
        local os_sysbench=${Var_OSRelease}
        if [ "$os_sysbench" = "astra" ]; then
            os_sysbench="debian"
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
    REGEX=("debian|astra" "ubuntu" "centos|red hat|kernel|oracle linux|alma|rocky" "'amazon linux'" "fedora" "arch" "freebsd" "alpine" "openbsd")
    RELEASE=("Debian" "Ubuntu" "CentOS" "CentOS" "Fedora" "Arch" "FreeBSD" "Alpine" "OpenBSD")
    PACKAGE_UPDATE=("apt-get update" "apt-get update" "yum -y update" "yum -y update" "yum -y update" "pacman -Sy" "pkg update" "apk update" "pkg_add -u")
    PACKAGE_INSTALL=("apt-get -y install" "apt-get -y install" "yum -y install" "yum -y install" "yum -y install" "pacman -Sy --noconfirm --needed" "pkg install -y" "apk add")
    PACKAGE_REMOVE=("apt-get -y remove" "apt-get -y remove" "yum -y remove" "yum -y remove" "yum -y remove" "pacman -Rsc --noconfirm" "pkg delete" "apk del")
    PACKAGE_UNINSTALL=("apt-get -y autoremove" "apt-get -y autoremove" "yum -y autoremove" "yum -y autoremove" "yum -y autoremove" "" "pkg autoremove" "apk autoremove")
    CMD=("$(grep -i pretty_name /etc/os-release 2>/dev/null | cut -d \" -f2)" "$(hostnamectl 2>/dev/null | grep -i system | cut -d : -f2)" "$(lsb_release -sd 2>/dev/null)" "$(grep -i description /etc/lsb-release 2>/dev/null | cut -d \" -f2)" "$(grep . /etc/redhat-release 2>/dev/null)" "$(grep . /etc/issue 2>/dev/null | cut -d \\ -f1 | sed '/^[ ]*$/d')" "$(grep -i pretty_name /etc/os-release 2>/dev/null | cut -d \" -f2)" "$(uname -s)" "$(uname -s)")
    SYS="${CMD[0]}"
    [[ -n $SYS ]] || exit 1
    for ((int = 0; int < ${#REGEX[@]}; int++)); do
        if [[ $(echo "$SYS" | tr '[:upper:]' '[:lower:]') =~ ${REGEX[int]} ]]; then
            SYSTEM="${RELEASE[int]}"
            [[ -n $SYSTEM ]] && break
        fi
    done
    cdn_urls=("https://cdn0.spiritlhl.top/" "http://cdn3.spiritlhl.net/" "http://cdn1.spiritlhl.net/" "http://cdn2.spiritlhl.net/")
    check_cdn_file
    _green "Update system manager."
    $PACKAGE_UPDATE command 2>/dev/null
    if ! command -v tar >/dev/null 2>&1; then
        _green "Installing tar"
        $PACKAGE_INSTALL tar
    fi
    if ! command -v unzip >/dev/null 2>&1; then
        _green "Installing unzip"
        $PACKAGE_INSTALL unzip
    fi
    if ! command -v dd >/dev/null 2>&1; then
        _green "Installing dd"
        $PACKAGE_INSTALL dd
    fi
    if ! command -v fio >/dev/null 2>&1; then
        _green "Installing fio"
        $PACKAGE_INSTALL fio
    fi
    if ! command -v sysbench >/dev/null 2>&1; then
        _green "Installing sysbench"
        $PACKAGE_INSTALL sysbench
        if [ $? -ne 0 ]; then
            echo "Unable to download sysbench through the system's package manager, speak to try compiling and installing it..."
            if ! wget -O /tmp/sysbench.zip "${cdn_success_url}https://github.com/akopytov/sysbench/archive/1.0.20.zip"; then
                echo "wget failed, trying with curl"
                curl -Lk -o /tmp/sysbench.zip "${cdn_success_url}https://github.com/akopytov/sysbench/archive/1.0.20.zip"
            fi
            if [ ! -f /tmp/sysbench.zip ]; then
                wget -q -O /tmp/sysbench.zip "https://hub.fgit.cf/akopytov/sysbench/archive/1.0.20.zip"
            fi
            chmod +x /tmp/sysbench.zip
            unzip /tmp/sysbench.zip -d /tmp
        fi
    fi
    if ! command -v geekbench >/dev/null 2>&1; then
        _green "Installing geekbench"
        curl -L "${cdn_success_url}https://raw.githubusercontent.com/oneclickvirt/cputest/main/dgb.sh" -o dgb.sh && chmod +x dgb.sh
        bash dgb.sh -v gb5
        _blue "If you not want to use geekbench5, you can use"
        echo "bash dgb.sh -v gb6"
        echo "bash dgb.sh -v gb4"
        _blue "to change version, or use"
        echo "rm -rf /usr/bin/geekbench*"
        _blue "to uninstall geekbench"
        rm -rf dgb.sh
    fi
    if ! command -v speedtest >/dev/null 2>&1; then
        _green "Installing geekbench"
        curl -L "${cdn_success_url}https://raw.githubusercontent.com/oneclickvirt/speedtest/main/dspt.sh" -o dspt.sh && chmod +x dspt.sh
        bash dspt.sh
        rm -rf dspt.sh
        _blue "if you want to use golang origin speedtest, you can use"
        echo "rm -rf /usr/bin/speedtest"
        echo "rm -rf /usr/bin/speedtest-go"
        _blue "to uninstall speedtest and speedtest-go"
    fi
    if [ "$(uname -s)" = "Darwin" ]; then
        echo "Detected MacOS. Installing sysbench and fio..."
        brew install --force sysbench fio dd
        # 有问题，需要修复，root环境不能brew，brew安装完毕后可能路径不在环境变量中
    else
        sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
    fi
    _green "The environment is ready."
}

delete_goecs() {
  rm -rf /root/goecs
  rm -rf /usr/bin/goecs
}

show_help() {
    cat <<"EOF"
Available commands:

./goecs.sh env             Check and Install package:
                                tar (Almost all unix-like systems have it.)
                                unzip (Almost all unix-like systems have it.)
                                dd (Almost all unix-like systems have it.)
                                fio (Almost all unix-like systems can be installed through the system's package manager.)
                                sysbench (Almost all unix-like systems can be installed through the system's package manager.)
                                geekbench (geekbench5)(Only support IPV4 environment, and memory greater than 1GB network detection, only support amd64 and arm64 architecture.)
                                speedtest (Use the officially provided binaries for more accurate test results.)
                           In fact, sysbench/geekbench is the only one of the above dependencies that must be installed, without which the CPU score cannot be tested.
./goecs.sh install         Install goecs command
./goecs.sh upgrade         Upgrade goecs command
./goecs.sh delete          Uninstall goecs command
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
"delete")
    delete_goecs
    ;;
*)
    echo "No command found."
    echo
    show_help
    ;;
esac
