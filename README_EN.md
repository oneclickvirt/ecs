# ecs

[![release](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml) [![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fecs&count_bg=%2357DEFF&title_bg=%23000000&icon=cliqz.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://www.spiritlhl.net/)

Fusion Monster Evaluation Script - GO Refactored Version - Since it has not been officially released, please report any issues via issues.

Shell version: https://github.com/spiritLHLS/ecs/blob/main/README_EN.md

## Language

[中文文档](README.md) | [English Docs](README_EN.md)

## Supported Systems and Architectures

Architectures supported for compilation: amd64、arm、arm64、386、s390x、mips、mipsle、s390x、riscv64

Tested architectures: amd64, arm64

More architectures please test by yourself

Compilation support: Linux, Windows、FreeBSD、OpenBSD

Tested on: Linux, Windows

More systems to be tested

Systems to be supported (hardware testing bugs not yet fixed): MacOS、FreeBSD、OpenBSD

## Features

- [x] System basic information query [Self-developed [basics](https://github.com/oneclickvirt/basics), [gostun](https://github.com/oneclickvirt/gostun)]
- [x] Concurrent IP basic information query [Self-developed [basics](https://github.com/oneclickvirt/basics)]
- [x] CPU test [Self-developed [cputest](https://github.com/oneclickvirt/cputest) supporting sysbench, geekbench, winsat]
- [x] Memory test [Self-developed [memorytest](https://github.com/oneclickvirt/memorytest) supporting sysbench, dd]
- [x] Disk test [Self-developed [disktest](https://github.com/oneclickvirt/disktest) supporting dd, fio, winsat]
- [x] Concurrent streaming media unlock information query for three major platforms [Modified from [netflix-verify](https://github.com/sjlleo/netflix-verify), [VerifyDisneyPlus](https://github.com/sjlleo/VerifyDisneyPlus), [TubeCheck](https://github.com/sjlleo/TubeCheck) to [CommonMediaTests](https://github.com/oneclickvirt/CommonMediaTests)]
- [x] Concurrent common streaming media tests [Self-developed code, logic modified from [RegionRestrictionCheck](https://github.com/lmc999/RegionRestrictionCheck), [MediaUnlockTest](https://github.com/HsukqiLee/MediaUnlockTest) to [UnlockTests](https://github.com/oneclickvirt/UnlockTests)]
- [x] Concurrent IP quality/security information query [Self-developed, due to testing with key information, privately developed, but binary files compiled in [securityCheck](https://github.com/oneclickvirt/securityCheck)]
- [x] Email port test [Self-developed [portchecker](https://github.com/oneclickvirt/portchecker)]
- [x] Three-network return path test [Modified from [zhanghanyun/backtrace](https://github.com/zhanghanyun/backtrace) to [oneclickvirt/backtrace](https://github.com/oneclickvirt/backtrace)]
- [x] Three-network route test [Modified from [NTrace-core](https://github.com/nxtrace/NTrace-core) to [nt3](https://github.com/oneclickvirt/nt3)]
- [x] Speed test [Based on data from [speedtest.net-crawler](https://github.com/spiritLHLS/speedtest.net-CN-ID), [speedtest.cn-crawler](https://github.com/spiritLHLS/speedtest.cn-CN-ID), modified from [speedtest-go](https://github.com/showwin/speedtest-go) to [oneclickvirt/speedtest](https://github.com/oneclickvirt/speedtest)]
- [x] Three-network Ping test [Modified from [ecsspeed](https://github.com/spiritLHLS/ecsspeed) logic to [pingtest](https://github.com/oneclickvirt/pingtest)]

## Instructions for Use on Linux/FreeBSD/MacOS

### one-click command

```
curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh && bash goecs.sh env && bash goecs.sh install && goecs
```

### explain in detail

Download the script

```
curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
```

Install environment

```
./goecs.sh env
```

Install goecs

```
./goecs.sh install
```

Upgrade goecs

```
./goecs.sh upgrade
```

Uninstall goecs

```
./goecs.sh delete
```

Explanation of the shell script

```
Available commands:

./goecs.sh env             Check and Install package:
                                sudo (Almost all unix-like systems have it.)
                                tar (Almost all unix-like systems have it.)
                                unzip (Almost all unix-like systems have it.)
                                dd (Almost all unix-like systems have it.)
                                fio (Almost all unix-like systems can be installed through the system's package manager.)
                                sysbench (Almost all unix-like systems can be installed through the system's package manager.)
                                geekbench (geekbench5)(Only support IPV4 environment, and memory greater than 1GB network detection, only support amd64 and arm64 architecture.)
                                speedtest (Use the officially provided binaries for more accurate test results.)
                                ping (Use the officially provided binaries for more accurate test results.)
                           In fact, sysbench/geekbench is the only one of the above dependencies that must be installed, without which the CPU score cannot be tested.
./goecs.sh install         Install goecs command
./goecs.sh upgrade         Upgrade goecs command
./goecs.sh delete          Uninstall goecs command
./goecs.sh help            Show this message
```

Invoke the goecs menu

```
goecs
```

or

```
./goecs
```

Parameterized goecs command

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
  -upload
        Enable/Disable upload the result (default true)
  -ut
        Enable/Disable unlock media test (default true)
  -v    Display version information
```

## Instructions for Use on Windows

Download the compressed file with the exe file

https://github.com/oneclickvirt/ecs/releases

Find the latest version, download the .zip file corresponding to your architecture, and unzip it to get an exe file.

Right-click the exe file and select Run as administrator (running without administrator mode will not allow hardware testing), and invoke the menu to choose.

No environment installation is required for Windows testing.