# ecs

[![release](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml) [![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fecs&count_bg=%2357DEFF&title_bg=%23000000&icon=cliqz.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://www.spiritlhl.net/)

融合怪测评脚本 - GO重构版本 - 由于未正式发版，如有问题请issues反馈

Shell版本： https://github.com/spiritLHLS/ecs

## 语言

[中文文档](README.md) | [English Docs](README_EN.md)

## 适配系统和架构

编译支持的架构: amd64、arm、arm64、386、mips、mipsle、s390x、riscv64

测试支持的架构: amd64、arm64 

更多架构请自行测试

编译支持的系统: Linux、Windows、MacOS、FreeBSD、OpenBSD

测试支持的系统: Linux、Windows 

更多系统请自行测试

待支持的系统(存在硬件测试BUG未修复): MacOS、FreeBSD、OpenBSD

## 功能

- [x] 系统基础信息查询[自研[basics](https://github.com/oneclickvirt/basics)、[gostun](https://github.com/oneclickvirt/gostun)]
- [x] IP基础信息并发查询[自研[basics](https://github.com/oneclickvirt/basics)]
- [x] CPU测试[自研[cputest](https://github.com/oneclickvirt/cputest)支持sysbench、geekbench、winsat]
- [x] 内存测试[自研[memorytest](https://github.com/oneclickvirt/memorytest)支持sysbench、dd]
- [x] 硬盘测试[自研[disktest](https://github.com/oneclickvirt/disktest)支持dd、fio、winsat]
- [x] 御三家流媒体解锁信息并发查询[借鉴[netflix-verify](https://github.com/sjlleo/netflix-verify)、[VerifyDisneyPlus](https://github.com/sjlleo/VerifyDisneyPlus)、[TubeCheck](https://github.com/sjlleo/TubeCheck)逻辑开发至于[CommonMediaTests](https://github.com/oneclickvirt/CommonMediaTests)]
- [x] 常见流媒体测试并发查询[自研至于[UnlockTests](https://github.com/oneclickvirt/UnlockTests)，逻辑借鉴[RegionRestrictionCheck](https://github.com/lmc999/RegionRestrictionCheck)、[MediaUnlockTest](https://github.com/HsukqiLee/MediaUnlockTest)]
- [x] IP质量/安全信息并发查询[自研，由于测试含密钥信息，故而私有化开发，但二进制文件编译至于[securityCheck](https://github.com/oneclickvirt/securityCheck)]
- [x] 邮件端口测试[自研[portchecker](https://github.com/oneclickvirt/portchecker)]
- [x] 三网回程测试[借鉴[zhanghanyun/backtrace](https://github.com/zhanghanyun/backtrace)二次开发至于[oneclickvirt/backtrace](https://github.com/oneclickvirt/backtrace)]
- [x] 三网路由测试[借鉴[NTrace-core](https://github.com/nxtrace/NTrace-core)二次开发至于[nt3](https://github.com/oneclickvirt/nt3)]
- [x] 测试网速[基于[speedtest.net-爬虫](https://github.com/spiritLHLS/speedtest.net-CN-ID)、[speedtest.cn-爬虫](https://github.com/spiritLHLS/speedtest.cn-CN-ID)的数据，使用[speedtest-go](https://github.com/showwin/speedtest-go)开发至于[oneclickvirt/speedtest](https://github.com/oneclickvirt/speedtest)]
- [x] 测试三网Ping值[借鉴[ecsspeed](https://github.com/spiritLHLS/ecsspeed)的逻辑二次开发至于[pingtest](https://github.com/oneclickvirt/pingtest)]

## Linux/FreeBSD/MacOS上使用的说明

### 一键命令

```
export noninteractive=true && curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh && bash goecs.sh env && bash goecs.sh install && goecs
```

或

```
export noninteractive=true && curl -L https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh && bash goecs.sh env && bash goecs.sh install && goecs
```

### 详细说明

下载脚本

```
curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
```

或

```
curl -L https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
```

更新包管理器(可选择)和安装环境

```
./goecs.sh env
```

如果不想要互动，可使用```export noninteractive=true```再执行```env```命令

安装goecs

```
./goecs.sh install
```

升级goecs

```
./goecs.sh upgrade
```

卸载goecs

```
./goecs.sh uninstall
```

shell脚本的说明

```
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
```

goecs唤起菜单

```
goecs
```

或

```
./goecs
```

goecs命令参数化

```
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

## Windows上使用的说明

下载带exe文件的压缩文件

https://github.com/oneclickvirt/ecs/releases

找其中最新的版本，按照对应架构下载对应的 .zip 文件，解压后文件夹内有一个exe文件

选择该exe文件，右键点击选择管理员模式运行(非管理员模式运行无法进行硬件测试)，唤起菜单自选

windows测试无需进行环境安装

## 在Docker中使用的说明

地址：https://hub.docker.com/r/spiritlhl/goecs

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

## QA

#### Q: 为什么默认使用sysbench而不是geekbench

#### A: 比较二者特点

```
sysbench                          geekbench
轻量几乎所有服务器都能跑            重型小机器跑不动
测试无联网需求，无硬件需求          测试必须联网，且必须IPV4环境，且有内存大小1G的最低需求
LUA编写且开源，各架构系统可自行编译  仅官方二进制文件且不开源，无对应架构时无法自行编译
核心测试组件十多年不变              每次大版本更新对标的CPU，不同版本间得分互相之间难转化，你只能以对标的CPU为准
测试仅测试计算性能                  测试涵盖多种性能测试，得分以权重计算，但实际很多测试项目实际是用不到的
适合快速测试                       适合全面测试
```

且```goecs```测试使用何种CPU测试方式可使用参数指定，默认只是为了更多用户快速测试的需求

#### Q: 为什么使用Golang而不是Rust重构

#### A: 因为网络相关的项目目前以Golang语言为趋势，大多组件有开源生态维护，Rust很多得自己手搓，~~我懒得搞~~我没那个技术力

#### Q: 为什么不继续开发Shell版本而是选择重构

#### A: 因为太多千奇百怪的环境问题了，还是提前编译好测试的二进制文件比较容易解决环境问题(泛化性更好)

#### Q: 每个测试项目的说明有吗？

#### A: 每个测试项目有对应的维护仓库，自行点击查看仓库说明

