#!/bin/bash
#From https://github.com/oneclickvirt/ecs
#2024.06.29

# curl -L https://github.com/oneclickvirt/ecs/raw/main/goecs_install.sh -o goecs_install.sh && chmod +x goecs_install.sh && bash goecs_install.sh

cd /root >/dev/null 2>&1
rm -rf /usr/bin/goecs
rm -rf goecs
os=$(uname -s)
arch=$(uname -m)

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

cdn_urls=("https://cdn0.spiritlhl.top/" "http://cdn3.spiritlhl.net/" "http://cdn1.spiritlhl.net/" "http://cdn2.spiritlhl.net/")
check_cdn_file
ECS_VERSION=$(curl -m 6 -sSL "https://api.github.com/repos/oneclickvirt/ecs/releases/latest" | awk -F \" '/tag_name/{gsub(/^v/,"",$4); print $4}')
# 如果 https://api.github.com/ 请求失败，则使用 https://githubapi.spiritlhl.workers.dev/ ，此时可能宿主机无IPV4网络
if [ -z "$ECS_VERSION" ]; then
    ECS_VERSION=$(curl -m 6 -sSL "https://githubapi.spiritlhl.workers.dev/repos/oneclickvirt/ecs/releases/latest" | awk -F \" '/tag_name/{gsub(/^v/,"",$4); print $4}')
fi
# 如果 https://githubapi.spiritlhl.workers.dev/ 请求失败，则使用 https://githubapi.spiritlhl.top/ ，此时可能宿主机在国内
if [ -z "$ECS_VERSION" ]; then
    ECS_VERSION=$(curl -m 6 -sSL "https://githubapi.spiritlhl.top/repos/oneclickvirt/ecs/releases/latest" | awk -F \" '/tag_name/{gsub(/^v/,"",$4); print $4}')
fi

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
goecs -v