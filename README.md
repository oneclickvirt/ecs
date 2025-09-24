# ECS

[![Build and Release](https://github.com/oneclickvirt/ecs/actions/workflows/build_binary.yaml/badge.svg)](https://github.com/oneclickvirt/ecs/actions/workflows/build_binary.yaml)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Foneclickvirt%2Fecs.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Foneclickvirt%2Fecs?ref=badge_shield)

[![Hits](https://hits.spiritlhl.net/goecs.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false)](https://hits.spiritlhl.net) [![Downloads](https://ghdownload.spiritlhl.net/oneclickvirt/ecs?color=36c600)](https://github.com/oneclickvirt/ecs/releases)

融合怪测评项目 - GO版本

(仅环境安装[非必须]使用shell外无额外shell文件依赖，环境安装只是为了测的更准，极端情况下无环境依赖安装也可全测项目)

如有问题请 [issues](https://github.com/oneclickvirt/ecs/issues) 反馈。

Go 版本：[https://github.com/oneclickvirt/ecs](https://github.com/oneclickvirt/ecs)

Shell 版本：[https://github.com/spiritLHLS/ecs](https://github.com/spiritLHLS/ecs)

---

## **语言**

[中文文档](README.md) | [English Docs](README_EN.md)

---

## **适配系统和架构**

### **编译与测试支持情况**
| 编译支持的架构             | 测试支持的架构 | 编译支持的系统             | 测试支持的系统 |
|---------------------------|--------------|---------------------------|---------------|
| amd64                     | amd64        | Linux                     | Linux         |
| arm64                     | arm64        | Windows                   | Windows       |
| arm                       |              | MacOS(Darwin)             | MacOS         |
| 386                       |              | FreeBSD                   |               |
| mips,mipsle               |              | Android                   |               |
| mips64,mips64le           |              |                           |               | 
| ppc64,ppc64le             |              |                           |               |
| s390x                     | s390x        |                           |               |
| riscv64                   |              |                           |               |

> 更多架构与系统请自行测试或编译，如有问题请开 issues。

### **待支持的系统**

| 系统           | 说明                       |
|----------------|---------------------------|
| Android(arm64) | 存在权限问题未修复，非安卓系统的ARM架构无问题      |
| OpenBSD/NetBSD | 部分Goalng的官方库未支持本系统(尤其是net相关项目)  |

---

## **功能**

- 系统基础信息查询，IP基础信息并发查询：[basics](https://github.com/oneclickvirt/basics)、[gostun](https://github.com/oneclickvirt/gostun)
- CPU 测试：[cputest](https://github.com/oneclickvirt/cputest)，支持 sysbench(lua/golang版本)、geekbench、winsat
- 内存测试：[memorytest](https://github.com/oneclickvirt/memorytest)，支持 sysbench、dd、winsat、mbw、stream
- 硬盘测试：[disktest](https://github.com/oneclickvirt/disktest)，支持 dd、fio、winsat
- 流媒体解锁信息并发查询：[netflix-verify](https://github.com/sjlleo/netflix-verify) 等逻辑，开发至 [CommonMediaTests](https://github.com/oneclickvirt/CommonMediaTests)
- 常见流媒体测试并发查询：[UnlockTests](https://github.com/oneclickvirt/UnlockTests)，逻辑借鉴 [RegionRestrictionCheck](https://github.com/lmc999/RegionRestrictionCheck) 等
- IP 质量/安全信息并发查询：二进制文件编译至 [securityCheck](https://github.com/oneclickvirt/securityCheck)
- 邮件端口测试：[portchecker](https://github.com/oneclickvirt/portchecker)
- 上游及回程路由线路检测：借鉴 [zhanghanyun/backtrace](https://github.com/zhanghanyun/backtrace)，二次开发至 [oneclickvirt/backtrace](https://github.com/oneclickvirt/backtrace)
- 三网路由测试：基于 [NTrace-core](https://github.com/nxtrace/NTrace-core)，二次开发至 [nt3](https://github.com/oneclickvirt/nt3)
- 网速测试：基于 [speedtest.net](https://github.com/spiritLHLS/speedtest.net-CN-ID) 和 [speedtest.cn](https://github.com/spiritLHLS/speedtest.cn-CN-ID) 数据，开发至 [oneclickvirt/speedtest](https://github.com/oneclickvirt/speedtest)
- 三网 Ping 值测试：借鉴 [ecsspeed](https://github.com/spiritLHLS/ecsspeed)，二次开发至 [pingtest](https://github.com/oneclickvirt/pingtest)
- 支持root或admin环境下测试，支持非root或非admin环境下测试，支持离线环境下进行测试，**暂未**支持无DNS环境下进行测试

**本项目初次使用建议查看说明：[跳转](https://github.com/oneclickvirt/ecs/blob/master/README_NEW_USER.md)**

---

## **使用说明**

### **Linux/FreeBSD/MacOS**

#### **一键命令**

**一键命令**将**默认安装依赖**，**默认更新包管理器**，**默认非互动模式**

- **国际用户无加速：**

  ```bash
  export noninteractive=true && curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh && ./goecs.sh env && ./goecs.sh install && goecs
  ```

- **国际/国内使用 CDN 加速：**

  ```bash
  export noninteractive=true && curl -L https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh && ./goecs.sh env && ./goecs.sh install && goecs
  ```

- **国内用户使用 CNB 加速：**

  ```bash
  export noninteractive=true && curl -L https://cnb.cool/oneclickvirt/ecs/-/git/raw/main/goecs.sh -o goecs.sh && chmod +x goecs.sh && ./goecs.sh env && ./goecs.sh install && goecs
  ```

- **短链接：**

  ```bash
  export noninteractive=true && curl -L https://bash.spiritlhl.net/goecs -o goecs.sh && chmod +x goecs.sh && ./goecs.sh env && ./goecs.sh install && goecs
  ```

#### **详细说明**

**详细说明**中的命令**可控制是否安装依赖**，**是否更新包管理器**，**默认互动模式可进行选择**

<details>
<summary>展开查看详细说明</summary>

1. **下载脚本**

   **国际用户无加速：**

   ```bash
   curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
   ```

   **国际/国内使用 CDN 加速：**

   ```bash
   curl -L https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
   ```

   **国内用户使用 CNB 加速：**

   ```bash
   curl -L https://cnb.cool/oneclickvirt/ecs/-/git/raw/main/goecs.sh -o goecs.sh && chmod +x goecs.sh
   ```

2. **更新包管理器（可选择）并安装环境**

   ```bash
   ./goecs.sh env
   ```

   **非互动模式：**

   ```bash
   export noninteractive=true && ./goecs.sh env
   ```

3. **安装 `goecs` 本体(仅下载二进制文件无依赖安装)**

   ```bash
   ./goecs.sh install
   ```

4. **升级 `goecs` 本体**

   ```bash
   ./goecs.sh upgrade
   ```

5. **卸载 `goecs` 本体**

   ```bash
   ./goecs.sh uninstall
   ```

6. **帮助命令**

   ```bash
   ./goecs.sh -h
   ```

7. **唤起菜单**

   ```bash
   goecs
   ```

</details>

---

#### **命令参数化**

<details>
<summary>展开查看各参数说明</summary>

```bash
Usage: goecs [options]
  -backtrace
        Enable/Disable backtrace test (in 'en' language or on windows it always false) (default true)
  -basic
        Enable/Disable basic test (default true)
  -comm
        Enable/Disable common media test (default true)
  -cpu
        Enable/Disable CPU test (default true)
  -cpum string
        Set CPU test method (supported: sysbench, geekbench, winsat) (default "sysbench")
  -cput string
        Set CPU test thread mode (supported: single, multi) (default "multi")
  -disk
        Enable/Disable disk test (default true)
  -diskm string
        Set disk test method (supported: fio, dd, winsat) (default "fio")
  -diskmc
        Enable/Disable multiple disk checks, e.g., -diskmc=false
  -diskp string
        Set disk test path, e.g., -diskp /root
  -email
        Enable/Disable email port test (default true)
  -h    Show help information
  -l string
        Set language (supported: en, zh) (default "zh")
  -log
        Enable/Disable logging in the current path
  -memory
        Enable/Disable memory test (default true)
  -memorym string
        Set memory test method (supported: sysbench, dd, winsat) (default "sysbench")
  -menu
        Enable/Disable menu mode, disable example: -menu=false (default true)
  -nt3
        Enable/Disable NT3 test (in 'en' language or on windows it always false) (default true)
  -nt3loc string
        Specify NT3 test location (supported: GZ, SH, BJ, CD for Guangzhou, Shanghai, Beijing, Chengdu) (default "GZ")
  -nt3t string
        Set NT3 test type (supported: both, ipv4, ipv6) (default "ipv4")
  -security
        Enable/Disable security test (default true)
  -speed
        Enable/Disable speed test (default true)
  -spnum int
        Set the number of servers per operator for speed test (default 2)
  -upload
        Enable/Disable upload the result (default true)
  -ut
        Enable/Disable unlock media test (default true)
  -v    Display version information
```
</details>

---

### **Windows**

1. 下载带 exe 文件的压缩包：[Releases](https://github.com/oneclickvirt/ecs/releases)
2. 解压后，右键以管理员模式运行。

---

### **Docker**

<details>
<summary>展开查看使用说明</summary>

国际镜像地址：https://hub.docker.com/r/spiritlhl/goecs

请确保执行下述命令前本机已安装Docker

特权模式+host网络

```shell
docker run --rm --privileged --network host spiritlhl/goecs:latest -menu=false -l zh
```

非特权模式+非host网络

```shell
docker run --rm spiritlhl/goecs:latest -menu=false -l zh
```

使用Docker执行测试，硬件测试会有一些偏差和虚拟化架构判断失效，还是推荐直接测试而不使用Docker测试。

国内阿里云镜像加速

请确保执行下述命令前本机已安装Docker

特权模式+host网络

```shell
docker run --rm --privileged --network host crpi-8tmognxgyb86bm61.cn-guangzhou.personal.cr.aliyuncs.com/oneclickvirt/ecs:latest -menu=false -l zh
```

非特权模式+非host网络

```shell
docker run --rm crpi-8tmognxgyb86bm61.cn-guangzhou.personal.cr.aliyuncs.com/oneclickvirt/ecs:latest -menu=false -l zh
```

实际上还有CNB镜像地址 https://cnb.cool/oneclickvirt/ecs/-/packages/docker/ecs 但很可惜组织空间不足无法推送了，更推荐使用阿里云镜像加速

</details>

---

### 从源码进行编译

<details>
<summary>展开查看编译说明</summary>

1. 克隆仓库的 public 分支（不含私有依赖）
```bash
git clone -b public https://github.com/oneclickvirt/ecs.git
cd ecs
```

2. 安装 Go 环境（如已安装可跳过）

选择 go 1.24.5 的版本进行安装

```bash
curl -L https://cdn.spiritlhl.net/https://raw.githubusercontent.com/spiritLHLS/one-click-installation-script/main/install_scripts/go.sh -o go.sh && chmod +x go.sh && bash go.sh 
```

3. 编译
```bash
go build -o goecs
```

4. 运行测试
```bash
./goecs -menu=false -l zh
```

支持的编译参数：
- GOOS：支持 linux、windows、darwin、freebsd、openbsd
- GOARCH：支持 amd64、arm、arm64、386、mips、mipsle、s390x、riscv64

跨平台编译示例：
```bash
# 编译 Windows 版本
GOOS=windows GOARCH=amd64 go build -o goecs.exe
# 编译 MacOS 版本
GOOS=darwin GOARCH=amd64 go build -o goecs_darwin
```
</details>

---

## QA

#### Q: 为什么默认使用sysbench而不是geekbench

#### A: 比较二者特点

| 比较项             | sysbench | geekbench |
|------------------|----------|-----------|
| 适用范围         | 轻量级，几乎可在任何服务器上运行 | 重量级，小型机器无法运行 |
| 测试要求         | 无需网络，无特殊硬件需求 | 需联网，IPV4环境，至少1G内存 |
| 开源情况         | 基于LUA，开源，可自行编译各架构版本 | 官方二进制闭源代码，不支持自行编译 |
| 测试稳定性       | 核心测试组件10年以上未变 | 每个大版本更新测试项，分数不同版本间难以对比(每个版本对标当前最好的CPU) |
| 测试内容         | 仅测试计算性能 | 覆盖多种性能测试，分数加权计算，但部分测试实际不常用 |
| 适用场景         | 适合快速测试，仅测试计算性能 | 适合综合全面的测试 |
| 排行榜         | [sysbench.spiritlhl.net](https://sysbench.spiritlhl.net/) | [browser.geekbench.com](https://browser.geekbench.com/) |

且```goecs```测试使用何种CPU测试方式可使用参数指定，默认只是为了更多用户快速测试的需求

#### Q: 为什么使用Golang而不是Rust重构

#### A: 因为网络相关的项目目前以Golang语言为趋势，大多组件有开源生态维护，Rust很多得自己手搓，~~我懒得搞~~我没那个技术力

#### Q: 为什么不继续开发Shell版本而是选择重构

#### A: 因为太多千奇百怪的环境问题了，还是提前编译好测试的二进制文件比较容易解决环境问题(泛化性更好)

#### Q: 每个测试项目的说明有吗？

#### A: 每个测试项目有对应的维护仓库，自行点击查看仓库说明

#### Q: 测试进行到一半如何手动终止？

#### A: 按ctrl键和c键终止程序，终止后依然会在当前目录下生成goecs.txt文件和分享链接，里面是已经测试到的信息。

#### Q: 非Root环境如何进行测试？

#### A: 手动执行安装命令，实在装不上也没问题，直接在release中下载对应架构的压缩包解压后执行即可，只要你能执行的了文件。或者你能使用docker的话用docker执行。

## 致谢

感谢 [he.net](https://he.net) [bgp.tools](https://bgp.tools) [ipinfo.io](https://ipinfo.io) [maxmind.com](https://www.maxmind.com/en/home) [cloudflare.com](https://www.cloudflare.com/) [ip.sb](https://ip.sb) [scamalytics.com](https://scamalytics.com) [abuseipdb.com](https://www.abuseipdb.com/) [ip2location.com](https://ip2location.com/) [ip-api.com](https://ip-api.com) [ipregistry.co](https://ipregistry.co/) [ipdata.co](https://ipdata.co/) [ipgeolocation.io](https://ipgeolocation.io) [ipwhois.io](https://ipwhois.io) [ipapi.com](https://ipapi.com/) [ipapi.is](https://ipapi.is/) [ipqualityscore.com](https://www.ipqualityscore.com/) [bigdatacloud.com](https://www.bigdatacloud.com/) [dkly.net](https://data.dkly.net) [virustotal.com](https://www.virustotal.com/) 等网站提供的API进行检测，感谢互联网各网站提供的查询资源

感谢

<a href="https://h501.io/?from=69" target="_blank">
  <img src="https://github.com/spiritLHLS/ecs/assets/103393591/dfd47230-2747-4112-be69-b5636b34f07f" alt="h501" style="height: 50px;">
</a>

提供的免费托管支持本开源项目的共享测试结果存储

同时感谢以下平台提供编辑和测试支持

<a href="https://www.jetbrains.com/go/" target="_blank">
  <img src="https://resources.jetbrains.com/storage/products/company/brand/logos/GoLand.png" alt="goland" style="height: 50px;">
</a>

<a href="https://community.ibm.com/zsystems/form/l1cc-oss-vm-request/" target="_blank">
  <img src="https://linuxone.cloud.marist.edu/oss/resources/images/linuxonelogo03.png" alt="ibm" style="height: 50px;">
</a>

<a href="https://console.zmto.com/?affid=1524" target="_blank">
  <img src="https://console.zmto.com/templates/2019/dist/images/logo_dark.svg" alt="zmto" style="height: 50px;">
</a>

## History Usage

![goecs](https://hits.spiritlhl.net/chart/goecs.svg)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/oneclickvirt/ecs.svg?variant=adaptive)](https://www.spiritlhl.net)

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Foneclickvirt%2Fecs.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Foneclickvirt%2Fecs?ref=badge_large)
