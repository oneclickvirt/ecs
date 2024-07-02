# ecs

[![release](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml) [![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fecs&count_bg=%2357DEFF&title_bg=%23000000&icon=cliqz.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://www.spiritlhl.net/)

融合怪测评脚本 - GO重构版本 - 由于未正式发版，如有问题请issues反馈

Shell版本： https://github.com/spiritLHLS/ecs

## 适配系统

Linux、MacOS、Windows

## 功能

- [x] 系统基础信息查询[自研[basics](https://github.com/oneclickvirt/basics)、[gostun](https://github.com/oneclickvirt/gostun)]
- [x] IP基础信息并发查询[自研[basics](https://github.com/oneclickvirt/basics)]
- [x] CPU测试[自研[cputest](https://github.com/oneclickvirt/cputest)支持sysbench、geekbench、winsat]
- [x] 内存测试[自研[memorytest](https://github.com/oneclickvirt/memorytest)支持sysbench、dd]
- [x] 硬盘测试[自研[disktest](https://github.com/oneclickvirt/disktest)支持dd、fio、winsat]
- [x] 御三家流媒体解锁信息并发查询[借鉴[netflix-verify](https://github.com/sjlleo/netflix-verify)、[VerifyDisneyPlus](https://github.com/sjlleo/VerifyDisneyPlus)、[TubeCheck](https://github.com/sjlleo/TubeCheck)二次开发至于[CommonMediaTests](https://github.com/oneclickvirt/CommonMediaTests)]
- [x] 常见流媒体测试并发查询[自研代码，逻辑借鉴[RegionRestrictionCheck](https://github.com/lmc999/RegionRestrictionCheck)、[MediaUnlockTest](https://github.com/HsukqiLee/MediaUnlockTest)并自行修复错漏至于[UnlockTests](https://github.com/oneclickvirt/UnlockTests)]
- [x] IP质量/安全信息并发查询[自研，由于测试含密钥信息，故而私有化开发，但二进制文件编译至于[securityCheck](https://github.com/oneclickvirt/securityCheck)]
- [x] 邮件端口测试[自研[portchecker](https://github.com/oneclickvirt/portchecker)]
- [x] 三网回程测试[借鉴[zhanghanyun/backtrace](https://github.com/zhanghanyun/backtrace)二次开发至于[oneclickvirt/backtrace](https://github.com/oneclickvirt/backtrace)]
- [x] 三网路由测试[借鉴[NTrace-core](https://github.com/nxtrace/NTrace-core)二次开发至于[nt3](https://github.com/oneclickvirt/nt3)]
- [x] 测试网速[基于[speedtest.net-爬虫](https://github.com/spiritLHLS/speedtest.net-CN-ID)、[speedtest.cn-爬虫](https://github.com/spiritLHLS/speedtest.cn-CN-ID)的数据，借鉴[speedtest-go](https://github.com/showwin/speedtest-go)二次开发至于[oneclickvirt/speedtest](https://github.com/oneclickvirt/speedtest)]

## TODO

- [ ] 测试三网Ping值[借鉴[ecsspeed](https://github.com/spiritLHLS/ecsspeed)的逻辑二次开发]

## Linux/macOS上使用的说明

下载脚本

```
curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
```

或

```
curl -L https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
```

安装环境

```
./goecs.sh env
```

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
./goecs.sh delete
```

shell脚本的说明

```
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
        Set memory test method (supported: sysbench, dd, winsat) (default "dd")
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
  -ut
        Enable/Disable unlock media test (default true)
  -v    Display version information
```

## Windows上使用的说明

下载带exe文件的压缩文件

https://github.com/oneclickvirt/ecs/releases

找其中最新的版本，按照对应架构下载对应的 .tar.gz 文件，解压后文件夹内有一个exe文件

选择该exe文件，右键点击选择管理员模式运行(非管理员模式运行无法进行硬件测试)，唤起菜单自选

windows测试无需进行环境安装