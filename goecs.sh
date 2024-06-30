#!/bin/bash
#From https://github.com/oneclickvirt/ecs
#2024.06.29

# curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh

cat <<"EOF"
  GGG    OOO   EEEE  CCCC   SSS
 G   G  O   O  E     C     S
 G      O   O  EEE   C      SSS
 G  GG  O   O  E     C         S
  GGG    OOO   EEEE  CCCC   SSS
EOF
cd /root >/dev/null 2>&1

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
  version_output=$(goecs -v || ./goecs -v)
  if [ $? -eq 0 ]; then
    extracted_version=$(echo "$version_output" | grep -oP '^v\d+(\.\d+)+')
    if [ -n "$extracted_version" ]; then
      current_version=$(echo "$extracted_version" | cut -c 2-)
      ecs_version=$ECS_VERSION
      if [[ "$(echo -e "$current_version\n$ecs_version" | sort -V | tail -n 1)" == "$current_version" ]]; then
        echo "goecs version ($current_version) is latest, no need to upgrade."
        return
      else
        echo "goecs version ($current_version) < $ecs_version, need to upgrade, 5 seconds later will start to upgrade"
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
  chmod 777 goecs
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
  *) echo "Error: Unknown OS release: $os_release" && exit 1 ;;
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
  PACKAGE_UPDATE=("! apt-get update && apt-get --fix-broken install -y && apt-get update" "apt-get update" "yum -y update" "yum -y update" "yum -y update" "pacman -Sy" "pkg update" "apk update" "pkg_add -u")
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
  if ! commadn -v geekbench >/dev/null 2>&1; then
    _green "Installing geekbench"
    curl -L "${cdn_success_url}https://raw.githubusercontent.com/oneclickvirt/cputest/main/dgb.sh" -o dgb.sh && chmod +x dgb.sh
    bash dgb.sh -v gb5
    _blue "If you not want to use geekbench5, you can use"
    echo "bash dgb.sh -v gb6"
    echo "bash dgb.sh -v gb4"
    _blue "to change version, or use"
    echo "rm -rf /usr/bin/geekbench*"
    _blue "to uninstall geekbench"
  fi
}

show_help() {
  cat <<"EOF"
Available commands:

env             Check and install dd fio sysbench geekbench5
install         Install goecs command
upgrade         Upgrade goecs command
help            Show this message

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
*)
  echo "No command found."
  echo
  show_help
  ;;
esac
