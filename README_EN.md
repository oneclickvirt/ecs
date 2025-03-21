# ecs

[![release](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/ecs/actions/workflows/main.yaml) 

[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fecs&count_bg=%2357DEFF&title_bg=%23000000&icon=cliqz.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://www.spiritlhl.net/)

Fusion Monster Evaluation Project - GO Refactored Version

Please report any issues via [issues](https://github.com/oneclickvirt/ecs/issues).

Go version: [https://github.com/oneclickvirt/ecs](https://github.com/oneclickvirt/ecs)

Shell version: [https://github.com/spiritLHLS/ecs/blob/main/README_EN.md](https://github.com/spiritLHLS/ecs/blob/main/README_EN.md)

---

## **Language**

[中文文档](README.md) | [English Docs](README_EN.md)

---

## **Supported Systems and Architectures**

### **Compilation and Testing Support**
| Supported for Compilation | Tested on | Supported OS for Compilation | Tested OS |
|---------------------------|-----------|------------------------------|-----------|
| amd64                     | amd64     | Linux                        | Linux     |
| arm                       | arm       | Windows                      | Windows   |
| arm64                     | arm64     | FreeBSD                      | FreeBSD   |
| 386                       | 386       | OpenBSD                      | OpenBSD   |
| mips                      |           | MacOS                        |           |
| mipsle                    |           |                              |           |
| s390x                     | s390x     |                              |           |
| riscv64                   |           |                              |           |

> Please test additional architectures and systems yourself. If you encounter any issues, please open an issue.

### **Systems Pending Support**
| OS     | Notes                                                  |
|--------|--------------------------------------------------------|
| MacOS  | Hardware testing bugs and environment dependencies unresolved |

---

## **Features**

- System basic information query and concurrent IP basic information query:  
  Self-developed [basics](https://github.com/oneclickvirt/basics), [gostun](https://github.com/oneclickvirt/gostun)
- CPU test:  
  Self-developed [cputest](https://github.com/oneclickvirt/cputest) supporting sysbench(lua/golang version), geekbench, winsat
- Memory test:  
  Self-developed [memorytest](https://github.com/oneclickvirt/memorytest) supporting sysbench, dd
- Disk test:  
  Self-developed [disktest](https://github.com/oneclickvirt/disktest) supporting dd, fio, winsat
- Streaming media unlock information concurrent query:  
  Modified from [netflix-verify](https://github.com/sjlleo/netflix-verify) and more to [CommonMediaTests](https://github.com/oneclickvirt/CommonMediaTests)
- Common streaming media tests concurrent query:  
  Self-developed to [UnlockTests](https://github.com/oneclickvirt/UnlockTests), logic modified from [RegionRestrictionCheck](https://github.com/lmc999/RegionRestrictionCheck) and others
- IP quality/security information concurrent query:  
  Self-developed, binary files compiled in [securityCheck](https://github.com/oneclickvirt/securityCheck)
- Email port test:  
  Self-developed [portchecker](https://github.com/oneclickvirt/portchecker)
- Three-network return path test:  
  Modified from [zhanghanyun/backtrace](https://github.com/zhanghanyun/backtrace) to [oneclickvirt/backtrace](https://github.com/oneclickvirt/backtrace)
- Three-network route test:  
  Modified from [NTrace-core](https://github.com/nxtrace/NTrace-core) to [nt3](https://github.com/oneclickvirt/nt3)
- Speed test:  
  Based on data from [speedtest.net](https://github.com/spiritLHLS/speedtest.net-CN-ID) and [speedtest.cn](https://github.com/spiritLHLS/speedtest.cn-CN-ID), developed to [oneclickvirt/speedtest](https://github.com/oneclickvirt/speedtest)
- Three-network Ping test:  
  Modified from [ecsspeed](https://github.com/spiritLHLS/ecsspeed) to [pingtest](https://github.com/oneclickvirt/pingtest)

**For first-time users of this project, it is recommended to check the instructions: [Jump to](https://github.com/oneclickvirt/ecs/blob/master/README_NEW_USER.md)**

---

## **Instructions for Use**

### **Linux/FreeBSD/MacOS**

#### **One-click command**

Will install dependencies by default, update package manager by default, non-interactive mode by default. The non-one-click version below allows control of dependency installation/package manager updates/interactive mode by default.

- **International users without acceleration:**

  ```bash
  export noninteractive=true && curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh && bash goecs.sh env && bash goecs.sh install && goecs -l en
  ```

- **International/domestic users with CDN acceleration:**

  ```bash
  export noninteractive=true && curl -L https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh && bash goecs.sh env && bash goecs.sh install && goecs -l en
  ```

- **Domestic users with CNB acceleration:**

  ```bash
  export noninteractive=true && curl -L https://cnb.cool/oneclickvirt/ecs/-/git/raw/main/goecs.sh -o goecs.sh && chmod +x goecs.sh && bash goecs.sh env && bash goecs.sh install && goecs -l en
  ```

#### **Detailed instructions**

<details>
<summary>Expand to view detailed instructions</summary>

1. **Download the script**

   **International users without acceleration:**

   ```bash
   curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
   ```

   **International/domestic users with CDN acceleration:**

   ```bash
   curl -L https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && chmod +x goecs.sh
   ```

   **Domestic users with CNB acceleration:**

   ```bash
   curl -L https://cnb.cool/oneclickvirt/ecs/-/git/raw/main/goecs.sh -o goecs.sh && chmod +x goecs.sh
   ```

2. **Update package manager (optional) and install environment**

   ```bash
   ./goecs.sh env
   ```

   **Non-interactive mode:**

   ```bash
   export noninteractive=true && ./goecs.sh env
   ```

3. **Install `goecs`**

   ```bash
   ./goecs.sh install
   ```

4. **Upgrade `goecs`**

   ```bash
   ./goecs.sh upgrade
   ```

5. **Uninstall `goecs`**

   ```bash
   ./goecs.sh uninstall

6. **help command**

   ```bash
   ./goecs.sh -h
   ```

7. **Invoke the menu**

   ```bash
   goecs -l en
   ```

</details>

---

#### **Command parameterization**

<details>
<summary>Expand to view parameter descriptions</summary>

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

1. Download the compressed file with the .exe file: [Releases](https://github.com/oneclickvirt/ecs/releases)
2. After unzipping, right-click and run as administrator.

---

### **Docker**

<details>
<summary>Expand to view how to use it</summary>

International image: https://hub.docker.com/r/spiritlhl/goecs

Please ensure Docker is installed on your machine before executing the following commands

Privileged mode + host network

```shell
docker run --rm --privileged --network host spiritlhl/goecs:latest -menu=false -l en
```

Unprivileged mode + non-host network

```shell
docker run --rm spiritlhl/goecs:latest -menu=false -l en
```

Using Docker to execute tests will result in some hardware testing bias and virtualization architecture detection failure. Direct testing is recommended over Docker testing.

Mirror image: https://cnb.cool/oneclickvirt/ecs/-/packages/docker/ecs

Please ensure Docker is installed on your machine before executing the following commands

Privileged mode + host network

```shell
docker run --rm --privileged --network host docker.cnb.cool/oneclickvirt/ecs:latest -menu=false -l en
```

Unprivileged mode + non-host network

```shell
docker run --rm docker.cnb.cool/oneclickvirt/ecs:latest -menu=false -l en
```

</details>

---

### Compiling from source code

<details>
<summary>Expand to view compilation instructions</summary>

1. Clone the public branch of the repository (without private dependencies)
```bash
git clone -b public https://github.com/oneclickvirt/ecs.git
cd ecs
```

2. Install Go environment (skip if already installed)
```bash
# Download and install Go
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

3. Compile
```bash
go build -o goecs
```

4. Run test
```bash
./goecs -menu=false -l en
```

Supported compilation parameters:
- GOOS: supports linux, windows, darwin, freebsd, openbsd
- GOARCH: supports amd64, arm, arm64, 386, mips, mipsle, s390x, riscv64

Cross-platform compilation examples:
```bash
# Compile Windows version
GOOS=windows GOARCH=amd64 go build -o goecs.exe
# Compile MacOS version
GOOS=darwin GOARCH=amd64 go build -o goecs_darwin
```
</details>

---

## QA

#### Q: Why is sysbench used by default instead of geekbench?

#### A: Comparing the characteristics of both:

| Comparison | sysbench | geekbench |
|------------|----------|-----------|
| Application scope | Lightweight, runs on almost any server | Heavyweight, won't run on small machines |
| Test requirements | No network needed, no special hardware requirements | Requires internet, IPv4 environment, minimum 1GB memory |
| Open source status | Based on LUA, open source, can compile for various architectures | Official binaries are closed source, cannot compile your own version |
| Test stability | Core test components unchanged for 10+ years | Each major version updates test items, making scores hard to compare between versions (each version benchmarks against current best CPUs) |
| Test content | Only tests computing performance | Covers multiple performance aspects with weighted scores, though some tests aren't commonly used |
| Suitable scenarios | Good for quick tests, focuses on computing performance | Good for comprehensive testing |

Note that `goecs` allows you to specify CPU test method via parameters. The default is chosen for faster testing across more systems.

#### Q: Why use Golang instead of Rust for refactoring?

#### A: Because network-related projects currently trend toward Golang, with many components maintained by open source communities. Many Rust components would require building from scratch, ~~I'm too lazy~~ I don't have that technical capability.

#### Q: Why not continue developing the Shell version instead of refactoring?

#### A: Because there were too many varied environment issues. Pre-compiled binary files are easier for solving environment problems (better generalization).

#### Q: Are there explanations for each test item?

#### A: Each test project has its own maintenance repository. Click through to view the repository description.

#### Q: How do I manually terminate a test halfway through?

#### A: Press Ctrl+C to terminate the program. After termination, a goecs.txt file and share link will still be generated in the current directory containing information tested so far.

#### Q: How do I test in a non-Root environment?

#### A: Execute the installation command manually. If you can't install it, simply download the appropriate architecture package from releases, extract it, and run the file if you have execution permissions. Alternatively, use Docker if you can.

## Thanks

Thank [ipinfo.io](https://ipinfo.io) [ip.sb](https://ip.sb) [cheervision.co](https://cheervision.co) [scamalytics.com](https://scamalytics.com) [abuseipdb.com](https://www.abuseipdb.com/) [virustotal.com](https://www.virustotal.com/) [ip2location.com](ip2location.com/) [ip-api.com](https://ip-api.com) [ipregistry.co](https://ipregistry.co/) [ipdata.co](https://ipdata.co/) [ipgeolocation.io](https://ipgeolocation.io) [ipwhois.io](https://ipwhois.io) [ipapi.com](https://ipapi.com/) [ipapi.is](https://ipapi.is/) [ipqualityscore.com](https://www.ipqualityscore.com/) [bigdatacloud.com](https://www.bigdatacloud.com/) and others for providing APIs for testing, and thanks to various websites on the Internet for providing query resources.

Thank

<a href="https://h501.io/?from=69" target="_blank">
  <img src="https://github.com/spiritLHLS/ecs/assets/103393591/dfd47230-2747-4112-be69-b5636b34f07f" alt="h501">
</a>

provided free hosting support for this open source project's shared test results storage

Thanks also to the following platforms for editorial and testing support

<a href="https://www.jetbrains.com/go/" target="_blank">
  <img src="https://resources.jetbrains.com/storage/products/company/brand/logos/GoLand.png" alt="goland">
</a>

<a href="https://community.ibm.com/zsystems/form/l1cc-oss-vm-request/" target="_blank">
  <img src="https://linuxone.cloud.marist.edu/oss/resources/images/linuxonelogo03.png" alt="ibm">
</a>

## Stargazers over time

[![Stargazers over time](https://starchart.cc/oneclickvirt/ecs.svg?variant=adaptive)](https://starchart.cc/oneclickvirt/ecs)
