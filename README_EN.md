# ecs

[![release](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml) [![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fecs&count_bg=%2357DEFF&title_bg=%23000000&icon=cliqz.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://www.spiritlhl.net/)

Fusion Monster Evaluation Script - GO Refactored Version

Please report any issues via issues.

Go version：[https://github.com/oneclickvirt/ecs](https://github.com/oneclickvirt/ecs)

Shell version: [https://github.com/spiritLHLS/ecs/blob/main/README_EN.md](https://github.com/spiritLHLS/ecs/blob/main/README_EN.md)

## Language

[中文文档](README.md) | [English Docs](README_EN.md)

## Supported Systems and Architectures

Architectures supported for compilation: amd64、arm、arm64、386、mips、mipsle、s390x、riscv64

Tested architectures: amd64, arm64

More architectures please test by yourself

Compilation support: Linux, Windows、MacOS、FreeBSD、OpenBSD

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
export noninteractive=true && curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh && bash goecs.sh env && bash goecs.sh install && goecs -l en
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

If you don't want interaction, use ```export noninteractive=true``` and then execute the ```env``` command

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
./goecs.sh uninstall
```

Explanation of the shell script

```
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
```

Invoke the goecs menu

```
goecs -l en
```

or

```
./goecs -l en
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

## Instructions for Use on Windows

Download the compressed file with the exe file

https://github.com/oneclickvirt/ecs/releases

Find the latest version, download the .zip file corresponding to your architecture, and unzip it to get an exe file.

Right-click the exe file and select Run as administrator (running without administrator mode will not allow hardware testing), and invoke the menu to choose.

No environment installation is required for Windows testing.

## Instructions for Use in Docker

Link: https://hub.docker.com/r/spiritlhl/goecs

Please make sure that Docker is installed on your machine before executing the following commands

Privileged Mode + host network

```shell
docker run --rm --privileged --network host spiritlhl/goecs:latest -menu=false -l en
```

Unprivileged mode + non-host network

```shell
docker run --rm spiritlhl/goecs:latest
```

Using Docker to execute tests, hardware testing will have some bias and virtualization architecture to determine the failure.

Recommended direct testing without using Docker testing.
