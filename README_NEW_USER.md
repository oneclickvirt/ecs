## 目录 / Table of Contents / 目次

[![Hits](https://hits.spiritlhl.net/goecs.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false)](https://hits.spiritlhl.net)

## 语言 / Languages / 言語
- [中文](#中文)
- [English](#English)
- [日本語](#日本語)

## 中文
- [系统基础信息](#系统基础信息)
- [CPU测试](#CPU测试)
- [内存测试](#内存测试)
- [硬盘测试](#硬盘测试)
- [流媒体解锁](#流媒体解锁)
- [IP质量检测](#IP质量检测)
- [邮件端口检测](#邮件端口检测)
- [三网回城线路检测](#三网回城线路检测)
- [三网回程路由检测](#三网回程路由检测)
- [就近测速](#就近测速)

## English
- [Basic System Information](#Basic-System-Information)
- [CPU Testing](#CPU-Testing)
- [Memory Testing](#Memory-Testing)
- [Disk Testing](#Disk-Testing)
- [Streaming Media Unlocking](#Streaming-Media-Unlocking)
- [IP Quality Detection](#IP-Quality-Detection)
- [Email Port Detection](#Email-Port-Detection)

## 日本語
- [システム基本情報](#システム基本情報)
- [CPUテスト](#CPUテスト)
- [メモリテスト](#メモリテスト)
- [ディスクテスト](#ディスクテスト)
- [ストリーミングメディアロック解除](#ストリーミングメディアロック解除)
- [IP品質検出](#IP品質検出)
- [メールポート検出](#メールポート検出)

---

## 中文

### **系统基础信息**

CPU型号: 不必多说，大概的说，按CPU的发布时间，都是新款则AMD好于Intel，都是旧款则Intel好于AMD。

CPU数量: 会检测是物理核心还是逻辑核心，优先展示物理核心，查不到物理核心才去展示逻辑核心。在服务器实际使用过程中，程序一般是按逻辑核心分配执行的，非视频转码和科学计算，物理核心一般都是开超线程成逻辑核心用，横向比较的时候，对应类型的核心数量才有比较的意义。

CPU缓存：显示的宿主机的CPU三级缓存信息。

AES-NI: 指令集是加密解密加速用的，有的话常规网络请求会更快一些，性能更高一些，没有的话会影响网络请求(含代理用途)。

VM-x/AMD-V/Hyper-V: 是当前测试宿主机是否支持嵌套虚拟化的指标，如果测试环境是套在docker里测或者没有root权限，那么这个默认就是检测不到显示不支持嵌套虚拟化。这个指标在你需要在宿主机上开设虚拟机(如 KVM、VirtualBox、VMware)的时候有用，其他用途该指标用处不大。

内存: 显示内存 正在使用的大小/总大小 ，不含虚拟内存。

气球驱动: 显示宿主机是否使用了气球驱动，使用了证明母机有共享内存使用，需要结合下面的内存读写测试查看是否有超售/严格的限制。

内核页合并：显示宿主机是否使用了KSM内存融合，使用了证明母机有共享内存使用，需要结合下面的内存读写测试查看是否有超售/严格的限制。

虚拟内存: 显示 SWAP虚拟内存

硬盘空间: 显示硬盘 正在使用的大小/总大小

启动盘路径：显示启动盘的路径

系统: 显示系统名字和架构

内核: 显示系统内核版本

系统在线时间: 显示宿主机自从开机到测试时已在线时长

时区: 显示宿主机系统时区

负载: 显示系统负载

虚拟化架构: 显示宿主机来自什么虚拟化架构，一般来说推荐```Dedicated > KVM > Xen```虚拟化，其他虚拟化都会存在性能损耗，导致使用的时候存在性能共享/损耗，但这个也说不准，独立服务器才拥有完全独立的资源占用，其他虚拟化基本都会有资源共享，取决于宿主机的售卖者是否有良心，具体性能优劣还是得看后面的专项测试。

NAT类型: 显示NAT类型，具体推荐```Full Cone > Restricted Cone > Port Restricted Cone > Symmetric```，测不出来时会显示```Inconclusive```，一般来说不拿来做特殊用途(有关于特殊的代理和实时通讯需求的)，都不用关注本指标。

TCP加速方式：一般是```cubic/bbr```拥塞控制协议，一般来说做代理服务器用bbr可以改善网速，普通用途不必关注此指标。

IPV4/IPV6 ASN: 显示宿主机IP所属的ASN组织ID和名字，同一个IDC可能会有多个ASN，ASN下可能有多个商家售卖不同段的IP的服务器，具体的上下游关系错综复杂，可使用 bgp.tool 进一步查看。

IPV4/IPV6 Location: 显示对应协议的IP在数据库中的地理位置。

IPV4 Active IPs: 根据 bgp.tools 信息查询当前CIDR分块中 活跃邻居数量/总邻居数量

IPV6 子网掩码：根据宿主机信息查询的本机IPV6子网大小

### **CPU测试**

支持通过命令行参数选择```GeekBench```和```Sysbench```进行测试：

| 比较项             | sysbench | geekbench |
|------------------|----------|-----------|
| 适用范围         | 轻量级，几乎可在任何服务器上运行 | 重量级，小型机器无法运行 |
| 测试要求         | 无需网络，无特殊硬件需求 | 需联网，IPV4环境，至少1G内存 |
| 开源情况         | 基于LUA，开源，可自行编译各架构版本(本项目有重构为Go版本内置) | 官方二进制闭源代码，不支持自行编译 |
| 测试稳定性       | 核心测试组件10年以上未变 | 每个大版本更新测试项，分数不同版本间难以对比(每个版本对标当前最好的CPU) |
| 测试内容         | 仅测试计算性能，基于素数计算 | 覆盖多种性能测试，分数加权计算，但部分测试实际不常用 |
| 适用场景         | 适合快速测试，仅测试计算性能 | 适合综合全面的测试 |

默认使用```Sysbench```进行测试，基准大致如下：

CPU测试单核```Sysbench```得分在5000以上的可以算第一梯队，4000到5000分算第二梯队，每1000分大致算一档。

AMD的7950x单核满血性能得分在6500左右，AMD的5950x单核满血性能得分5700左右，Intel普通的CPU(E5之类的)在1000~800左右，低于500的单核CPU可以说是性能比较差的了。

有时候多核得分和单核得分一样，证明商家在限制程序并发使用CPU，典型例子腾讯云。

```Sysbench```的基准可见 [CPU Performance Ladder For Sysbench](https://sysbench.spiritlhl.net/) 天梯图，具体得分不分测试的sysbench的版本。

```GeekBench```的基准可见 [官方网站](https://browser.geekbench.com/processor-benchmarks/) 天梯图，具体得分每个```GeekBench```版本都不一样，注意使用时测试的```GeekBench```版本是什么。

多说一句，```GeekBench```测的很多内容，实际在服务器使用过程中根本用不到，测试仅供参考。当然```Sysbench```非常不全面，但它基于最基础的计算性能可以大致比较CPU的性能。

实际上CPU性能测试够用就行，除非是科学计算以及视频转码，一般不需要特别追求高性能CPU。

### **内存测试**

一般来说，只需要判断IO速度是否低于```10240MB/s```，如果低于这个值那么证明内存性能不佳，极大概率存在超售超卖问题。

至于超开的原因可能是开了虚拟内存(硬盘当内存用)、可能是开了ZRAM(牺牲CPU性能)、可能是开了气球驱动、可能是开了KSM内存融合，原因多种多样。

### **硬盘测试**

```dd```测试可能误差偏大但测试速度快无硬盘大小限制，```fio```测试真实一些但测试速度慢有硬盘以及内存大小的最低需求。

同时，服务器可能有不同的文件系统，某些文件系统的IO引擎在同样的硬件条件下测试的读写速度更快，这是正常的。项目默认使用```fio```进行测试，测试使用的IO引擎优先级为```libaio > posixaio > psync```，备选项```dd```测试在```fio```测试不可用时自动替换。

以```fio```测试结果为例基准如下：

| 操作系统类型 | 主要指标 | 次要指标 |
|---------|-------------------|---------------------|
| Windows/Mac | 4K读 → 64K读 → 写入测试 | 图形界面系统优先考虑读取性能 |
| Linux (无图形界面) | 4K读 + 4K写 + 1M读写| 读/写值通常相似 |

以下硬盘类型对于指标值指 常规~满血 性能状态，指```libaio```作为IO测试引擎，指在```Linux```下进行测试

| 驱动类型 | 4K(IOPS)性能 | 1M(IOPS)性能 |
|------------|--------------------------|----------------------|
| NVMe SSD | ≥ 200 MB/s | 5-10 GB/s |
| 标准SSD | 50-100 MB/s | 2-3 GB/s |
| HDD (机械硬盘) | 10-40 MB/s | 500-600 MB/s |
| 性能不佳 | < 10 MB/s | < 200 MB/s |

快速评估：

1. **主要检查**: 4K读(IOPS) 4K写(IOPS)
    - 几乎相同差别不大
    - ≥ 200 MB/s = NVMe SSD
    - 50-100 MB/s = 标准SSD
    - 10-40 MB/s = HDD (机械硬盘)
    - < 10 MB/s = 垃圾性能，超售/限制严重

2. **次要检查**: 1M总和(IOPS)
    - 提供商设置的IO限制
    - 资源超开超售情况
    - 数值越高越好
    - NVMe SSD通常达到4-6 GB/s
    - 标准SSD通常达到1-2 GB/s

如果 NVMe SSD的1M(IOPS)值 < 1GB/s 表明存在严重的资源超开超售。

注意，这里测试的是真实的IO，仅限本项目，非本项目测试的IO不保证基准通用，因为他们测试的时候可能用的不是同样的参数，可能未设置IO直接读写，可能设置IO引擎不一致，可能设置测试时间不一致，都会导致基准有偏差。

### **流媒体解锁**

检索常见的流媒体平台解锁，当然也不全是流媒体，还有一些常见的别的平台的解锁也纳入了。一般来说，IP解锁地区都是一致的不会到处乱飘，如果发现多家平台解锁地区不一致，那么IP大概率是租赁的IPXO等平台的，各平台数据库识别缓慢，IP质量一般来说也好不到哪里去。

### **IP质量检测**

检测14个数据库的IP相关信息，一般来说看使用类型和公司类型还有安全信息的其他判别足矣，安全得分真的图一乐。多个平台比较对应检测项目都为对应值，证明当前IP确实如此，不要仅相信一个数据库源的信息。

### **邮件端口检测**

- **SMTP（25）**：用于邮件服务器之间传输邮件（发送邮件）。
- **SMTPS（465）**：用于加密的 SMTP 发送邮件（SSL/TLS 方式）。
- **SMTP（587）**：用于客户端向邮件服务器发送邮件，支持 STARTTLS 加密。
- **POP3（110）**：用于邮件客户端从服务器下载邮件，不加密。
- **POP3S（995）**：用于加密的 POP3，安全地下载邮件（SSL/TLS 方式）。
- **IMAP（143）**：用于邮件客户端在线管理邮件（查看、同步邮件），不加密。
- **IMAPS（993）**：用于加密的 IMAP，安全地管理邮件（SSL/TLS 方式）。

具体当前宿主机不做邮局或者不收电子邮件，那么该项目指标不用理会。

### **三网回程线路检测**

检测当前的宿主机的IP地址 到 四个主要POP点城市的三个主要运营商的接入点的IP地址 的线路，具体来说

电信163、联通4837、移动CMI 是常见的线路

电信CN2GIA > 电信CN2GT 移动CMIN2 联通9929 算优质的线路

用什么运营商连宿主机的IP就看哪个运营商的线路就行了，具体线路的路由情况，看在下一个检测项看到对应的ICMP检测路由信息。

### **三网回程路由检测**

默认检测广州为目的地，实际可使用命令行参数指定目的地，见对应的说明。

主要就是看是不是直连，是不是延迟低，是不是没有隐藏路由信息。如果路由全球跑，延迟起飞，那么线路自然不会好到哪里去。

### **就近测速**

先测的官方推荐的测速点，然后测有代表性的国际测速点，最后测国内三大运营商ping值最低的测速点。

境内使用为主就看境内测速即可，境外使用看境外测速，官方测速点可以代表受测的宿主机本地带宽基准。

一般来说境外的服务器的带宽100Mbps起步，境内的服务器1Mbps带宽起步，具体看线路优劣，带宽特别大有时候未必用得上，够用就行了。

---

## English

### **Basic System Information**

CPU Model: Simply put, generally speaking, based on CPU release dates, newer AMD models are better than Intel, while for older models, Intel is better than AMD.

CPU Count: It will detect whether these are physical cores or logical cores, prioritizing display of physical cores, only showing logical cores if physical core information is unavailable. In actual server usage, programs are generally allocated by logical cores. Except for video transcoding and scientific computing, physical cores are usually enabled with hyperthreading to function as logical cores. When making comparisons, only cores of the corresponding type have meaningful comparison value.

CPU Cache: Displays the host machine's three-level CPU cache information.

AES-NI: This instruction set is used for encryption/decryption acceleration. With it, normal network requests will be faster and performance will be higher. Without it, network requests (including proxy usage) will be affected.

VM-x/AMD-V/Hyper-V: This indicates whether the current host machine supports nested virtualization. If the test environment is running inside Docker or doesn't have root privileges, then by default this will be undetectable and will show as not supporting nested virtualization. This metric is useful when you need to set up virtual machines (such as KVM, VirtualBox, VMware) on the host machine; for other purposes, this metric is not very useful.

Memory: Displays memory size in format "currently used size/total size", not including virtual memory.

Balloon Driver: Shows whether the host machine is using a balloon driver. If used, it proves the parent machine has shared memory usage, which should be examined alongside the memory read/write test below to check for overselling/strict limitations.

Kernel Same-page Merging: Shows whether the host machine is using KSM memory fusion. If used, it proves the parent machine has shared memory usage, which should be examined alongside the memory read/write test below to check for overselling/strict limitations.

Virtual Memory: Displays SWAP virtual memory.

Disk Space: Displays disk usage in format "currently used size/total size".

Boot Disk Path: Shows the path of the boot disk.

System: Displays system name and architecture.

Kernel: Displays system kernel version.

System Uptime: Shows how long the host machine has been online since boot until testing time.

Timezone: Displays the host machine's system timezone.

Load: Displays system load.

Virtualization Architecture: Shows what virtualization architecture the host machine uses. Generally speaking, the recommended order is `Dedicated > KVM > Xen` virtualization. Other virtualization will have performance losses, leading to shared/degraded performance during use. However, this is not definitive. Only dedicated servers have completely independent resource usage; other virtualization methods basically all have resource sharing, depending on whether the host machine seller has a conscience. The specific performance merits still depend on the specialized tests that follow.

NAT Type: Displays NAT type. Specifically recommended in order: `Full Cone > Restricted Cone > Port Restricted Cone > Symmetric`. When not detectable, it will show `Inconclusive`. Generally speaking, if you're not using it for special purposes (related to special proxy and real-time communication needs), you don't need to pay attention to this metric.

TCP Acceleration Method: Generally this is the `cubic/bbr` congestion control protocol. Generally speaking, using bbr for proxy servers can improve network speed; for ordinary purposes, you don't need to pay attention to this indicator.

IPv4/IPv6 ASN: Displays the ASN organization ID and name that the host machine's IP belongs to. The same IDC may have multiple ASNs, and an ASN may have multiple vendors selling servers with different IP segments. The specific upstream and downstream relationships are complex and can be further viewed using bgp.tool.

IPv4/IPv6 Location: Shows the geographic location of the corresponding protocol's IP in the database.

IPV4 Active IPs: Query the number of active neighbours/total number of neighbours in the current CIDR chunk based on the bgp.tools information.

### **CPU Testing**

Supports selecting `GeekBench` and `Sysbench` for testing through command line parameters:

| Comparison Item | sysbench | geekbench |
|------------------|----------|-----------|
| Application Range | Lightweight, can run on almost any server | Heavyweight, cannot run on small machines |
| Test Requirements | No network needed, no special hardware requirements | Requires network, IPv4 environment, at least 1GB memory |
| Open Source Status | Based on LUA, open source, can compile versions for various architectures (this project has been rebuilt in Go version built-in) | Official binary closed source code, does not support self-compilation |
| Test Stability | Core test components unchanged for over 10 years | Test items updated with each major version, scores difficult to compare between different versions (each version benchmarks against current best CPUs) |
| Test Content | Only tests computational performance, based on prime number calculation | Covers multiple performance tests, weighted score calculation, but some tests are not commonly used in practice |
| Applicable Scenarios | Suitable for quick testing, only tests computational performance | Suitable for comprehensive testing |

By default, `Sysbench` is used for testing, with the baseline roughly as follows:

CPU test single-core `Sysbench` scores above 5000 can be considered first tier, 4000 to 5000 points second tier, with roughly one tier per 1000 points.

AMD's 7950x single-core full performance score is around 6500, AMD's 5950x single-core full performance score is around 5700, Intel's ordinary CPUs (E5 series, etc.) are around 1000~800, and single-core CPUs scoring below 500 can be said to have relatively poor performance.

Sometimes multi-core scores are the same as single-core scores, proving that the vendor is limiting program concurrent use of CPU, a typical example being Tencent Cloud.

Benchmarks for ```Sysbench`` can be found in the [CPU Performance Ladder For Sysbench](https://sysbench.spiritlhl.net/) ladder chart, with specific scores regardless of the version of sysbench tested.

For `GeekBench` baselines, see the [official website](https://browser.geekbench.com/processor-benchmarks/) ladder chart. Specific scores differ for each `GeekBench` version, so note which `GeekBench` version is being used when testing.

As an additional note, many things tested by `GeekBench` are not actually used in server usage processes, so the test is for reference only. Of course, `Sysbench` is very incomplete, but it can roughly compare CPU performance based on the most basic computational performance.

In practice, CPU performance just needs to be sufficient. Unless you're doing scientific computing or video transcoding, you generally don't need to pursue high-performance CPUs.

### **Memory Testing**

Generally speaking, you only need to determine whether the IO speed is below `10240MB/s`. If it's below this value, it proves that memory performance is poor, with an extremely high probability of overselling issues.

As for the reasons for oversubscription, it could be that virtual memory is enabled (using disk as memory), ZRAM might be enabled (sacrificing CPU performance), balloon drivers might be enabled, or KSM memory fusion might be enabled - there are various possible reasons.

### **Disk Testing**

The `dd` test may have larger errors but is faster to test with no disk size limitations. The `fio` test is more realistic but slower to test and has minimum requirements for disk and memory size.

At the same time, servers may have different file systems, and certain file systems' IO engines may test faster read/write speeds under the same hardware conditions, which is normal. The project uses `fio` for testing by default, with IO engine priority being `libaio > posixaio > psync`. The alternative `dd` test automatically replaces when `fio` testing is not available.

Using `fio` test results as an example, the baseline is as follows:

| OS Type | Primary Metrics | Secondary Metrics |
|---------|-------------------|---------------------|
| Windows/Mac | 4K read → 64K read → Write test | Graphical systems prioritize read performance |
| Linux (without GUI) | 4K read + 4K write + 1M read/write | Read/write values usually similar |

The following disk types refer to metric values indicating normal~full-power performance states, using `libaio` as the IO test engine, testing under `Linux`

| Drive Type | 4K(IOPS) Performance | 1M(IOPS) Performance |
|------------|--------------------------|----------------------|
| NVMe SSD | ≥ 200 MB/s | 5-10 GB/s |
| Standard SSD | 50-100 MB/s | 2-3 GB/s |
| HDD (Mechanical) | 10-40 MB/s | 500-600 MB/s |
| Poor Performance | < 10 MB/s | < 200 MB/s |

Quick assessment:

1. **Primary Check**: 4K read(IOPS) 4K write(IOPS)
    - Almost identical with little difference
    - ≥ 200 MB/s = NVMe SSD
    - 50-100 MB/s = Standard SSD
    - 10-40 MB/s = HDD (Mechanical)
    - < 10 MB/s = Poor performance, severe overselling/restriction

2. **Secondary Check**: 1M total(IOPS)
    - IO limit set by provider
    - Resource overselling situation
    - Higher value is better
    - NVMe SSD typically reaches 4-6 GB/s
    - Standard SSD typically reaches 1-2 GB/s

If NVMe SSD's 1M(IOPS) value < 1GB/s, it indicates severe resource overselling.

Note that this is testing real IO, limited to this project only. The baseline may not be universal for tests not from this project, because they might not use the same parameters when testing, might not set direct IO reading/writing, might use inconsistent IO engines, or might set inconsistent test times, all of which will cause baseline deviations.

### **Streaming Media Unlocking**

Checks common streaming media platform unlocking, though not all are streaming media - some other common platform unlocks are also included. Generally speaking, IP unlocking regions are consistent and don't randomly fluctuate. If you find that multiple platforms have inconsistent unlocking regions, then the IP is likely rented from platforms like IPXO, with slow recognition in various platform databases. Generally speaking, the IP quality won't be good either.

### **IP Quality Detection**

Checks IP-related information from 14 databases. Generally speaking, it's sufficient to look at usage type, company type, and other security information judgments. The security score is really just for amusement. When multiple platforms compare corresponding detection items to corresponding values, it proves that the current IP is indeed as such. Don't just trust information from a single database source.

### **Email Port Detection**

- **SMTP (25)**: Used for email transmission between mail servers (sending mail).
- **SMTPS (465)**: Used for encrypted SMTP mail sending (SSL/TLS method).
- **SMTP (587)**: Used for clients to send email to mail servers, supports STARTTLS encryption.
- **POP3 (110)**: Used for email clients to download mail from servers, unencrypted.
- **POP3S (995)**: Used for encrypted POP3, securely downloading mail (SSL/TLS method).
- **IMAP (143)**: Used for email clients to manage mail online (view, sync mail), unencrypted.
- **IMAPS (993)**: Used for encrypted IMAP, securely managing mail (SSL/TLS method).

Specifically, if the current host machine is not being used as a mail server or not receiving electronic mail, then this project metric can be disregarded.

---

## 日本語

### **システム基本情報**

CPU型番: 簡単に言えば、CPUの発売時期によって、新しいモデルならAMDがIntelより優れ、古いモデルならIntelがAMDより優れています。

CPUコア数: 物理コアか論理コアかを検出し、優先的に物理コアを表示します。物理コアが検出できない場合のみ論理コアを表示します。サーバーの実際の使用では、プログラムは通常、論理コアに基づいて実行されます。ビデオエンコードや科学計算以外では、物理コアは通常ハイパースレッディングを有効にして論理コアとして使用されます。比較する際は、同じタイプのコア数を比較することが意味を持ちます。

CPUキャッシュ：ホストマシンのCPU L1/L2/L3キャッシュ情報を表示します。

AES-NI: 暗号化/復号化を高速化する命令セットです。これがあれば通常のネットワークリクエストがより速く、パフォーマンスが高くなります。ない場合はネットワークリクエスト（プロキシ用途を含む）に影響します。

VM-x/AMD-V/Hyper-V: 現在のテスト環境がネステッド仮想化をサポートしているかどうかを示す指標です。テスト環境がDockerコンテナ内にあるか、root権限がない場合、デフォルトでは検出できず、ネステッド仮想化をサポートしていないと表示されます。この指標は、ホストマシン上で仮想マシン（KVM、VirtualBox、VMwareなど）を設定する必要がある場合に役立ちますが、他の用途ではあまり重要ではありません。

メモリ: 使用中サイズ/総サイズ のメモリを表示します。仮想メモリは含まれません。

バルーンドライバ: ホストマシンがバルーンドライバを使用しているかどうかを表示します。使用している場合は、親マシンがメモリを共有していることを示し、以下のメモリ読み書きテストと合わせて、オーバーセリング/厳しい制限があるかどうかを確認する必要があります。

Kernel Same-page Merging: ホストマシンがKSMメモリマージを使用しているかどうかを表示します。使用している場合は、親マシンがメモリを共有していることを示し、以下のメモリ読み書きテストと合わせて、オーバーセリング/厳しい制限があるかどうかを確認する必要があります。

仮想メモリ: SWAP仮想メモリを表示します

ディスク容量: 使用中サイズ/総サイズ のディスク容量を表示します

ブートディスクパス：ブートディスクのパスを表示します

OS: システム名とアーキテクチャを表示します

カーネル: システムカーネルバージョンを表示します

システム稼働時間: ホストマシンが起動してからテスト時までの稼働時間を表示します

タイムゾーン: ホストマシンのシステムタイムゾーンを表示します

負荷: システム負荷を表示します

仮想化アーキテクチャ: ホストマシンがどの仮想化アーキテクチャから来ているかを表示します。一般的に ```Dedicated > KVM > Xen``` 仮想化が推奨されます。他の仮想化はパフォーマンス低下を引き起こし、使用時にパフォーマンス共有/損失が発生しますが、これも確実ではありません。専用サーバーのみが完全に独立したリソース占有を持ち、他の仮想化はほとんどリソース共有があります。これはホストマシンの販売者が良心的かどうかによって異なります。具体的なパフォーマンスの優劣は、後の専門テストを見る必要があります。

NAT種類: NAT種類を表示します。具体的には ```Full Cone > Restricted Cone > Port Restricted Cone > Symmetric``` が推奨されます。検出できない場合は ```Inconclusive``` と表示されます。一般的に特別な用途（特殊なプロキシとリアルタイム通信の要件に関連する）に使用しない限り、この指標を気にする必要はありません。

TCP加速方式：一般的に ```cubic/bbr``` 輻輳制御プロトコルです。一般的にプロキシサーバーとして使用する場合、bbrを使用するとネットワーク速度が改善されますが、通常の用途ではこの指標に注目する必要はありません。

IPV4/IPV6 ASN: ホストマシンのIPが属するASN組織IDと名前を表示します。同じIDCに複数のASNがある可能性があり、1つのASNの下に異なるIPセグメントのサーバーを販売する複数の業者がいる可能性があります。具体的な上流/下流関係は複雑です。bgp.toolを使用してさらに詳しく調べることができます。

IPV4/IPV6 ロケーション: データベース内の対応するプロトコルのIPの地理的位置を表示します。

IPV4 アクティブIP: bgp.tools情報に基づいて、現在のCIDRチャンクのアクティブなネイバー数/総ネイバー数を照会する。

### **CPUテスト**

コマンドラインパラメータを通じて```GeekBench```と```Sysbench```のテストを選択できます：

| 比較項目 | sysbench | geekbench |
|------------------|----------|-----------|
| 適用範囲 | 軽量、ほぼすべてのサーバーで実行可能 | 重量級、小型マシンでは実行不可 |
| テスト要件 | ネットワーク不要、特別なハードウェア要件なし | ネットワーク必要、IPV4環境、最低1Gメモリ |
| オープンソース状況 | LUAベース、オープンソース、各アーキテクチャ版をコンパイル可能（本プロジェクトではGoに再構築して内蔵） | 公式バイナリはクローズドソース、自己コンパイル不可 |
| テスト安定性 | コアテストコンポーネントは10年以上変更なし | 各メジャーバージョンでテスト項目更新、スコアはバージョン間で比較困難（各バージョンは当時最高のCPUを基準） |
| テスト内容 | 計算性能のみテスト、素数計算ベース | 多様な性能テストカバー、スコアは重み付け計算、一部テストは実際にはあまり使用されない |
| 適用シーン | 迅速なテストに適合、計算性能のみテスト | 総合的な全面テストに適合 |

デフォルトでは```Sysbench```を使用してテストを行います。基準は概ね以下の通りです：

CPUテストのシングルコア```Sysbench```スコアが5000以上なら第一ティア、4000〜5000点なら第二ティア、1000点ごとに大体一ランクと考えられます。

AMDの7950xシングルコアのフルパフォーマンススコアは約6500、AMDの5950xシングルコアのフルパフォーマンススコアは約5700、Intelの通常のCPU（E5など）は約1000〜800、500未満のシングルコアCPUはパフォーマンスが比較的低いと言えます。

時々、マルチコアスコアとシングルコアスコアが同じ場合があります。これは販売者がプログラムの並列CPU使用を制限していることを示しています。典型的な例はTencent Cloudです。

Sysbenchのベンチマークは[CPU Performance Ladder For Sysbench](https://sysbench.spiritlhl.net/)のラダーチャートで見ることができる。

```GeekBench```の基準は[公式ウェブサイト](https://browser.geekbench.com/processor-benchmarks/)の階層チャートを参照してください。具体的なスコアは各```GeekBench```バージョンで異なるため、テスト時の```GeekBench```バージョンに注意してください。

補足ですが、```GeekBench```がテストする多くの内容は、サーバー使用過程で実際には必要ないことが多いです。テストは参考程度にしてください。もちろん```Sysbench```は非常に包括的ではありませんが、基本的な計算性能に基づいてCPUのパフォーマンスを大まかに比較できます。

実際にはCPUパフォーマンスは十分であれば良く、科学計算やビデオエンコード以外では、特に高性能CPUを追求する必要はありません。

### **メモリテスト**

一般的に、IO速度が```10240MB/s```未満かどうかを判断するだけで十分です。この値を下回る場合、メモリパフォーマンスが良くなく、オーバーセリング/オーバーコミットの問題がある可能性が非常に高いです。

オーバーコミットの原因は、仮想メモリの使用（ディスクをメモリとして使用）、ZRAM（CPUパフォーマンスを犠牲）、バルーンドライバの使用、KSMメモリマージの使用など、様々な可能性があります。

### **ディスクテスト**

```dd```テストは誤差が大きい可能性がありますが、テスト速度が速くディスクサイズ制限がありません。```fio```テストはより現実的ですが、テスト速度が遅く、ディスクおよびメモリサイズの最低要件があります。

同時に、サーバーには異なるファイルシステムがある可能性があり、特定のファイルシステムのIOエンジンは同じハードウェア条件下でも読み書き速度が速くなる場合があります。これは正常です。プロジェクトはデフォルトで```fio```を使用してテストを行います。テストに使用されるIOエンジンの優先順位は```libaio > posixaio > psync```です。代替オプションの```dd```テストは```fio```テストが利用できない場合に自動的に置き換えられます。

```fio```テスト結果を例に基準は以下の通りです：

| OSタイプ | 主要指標 | 副次指標 |
|---------|-------------------|---------------------|
| Windows/Mac | 4K読み → 64K読み → 書き込みテスト | グラフィカルインターフェースシステムは読み取りパフォーマンスを優先 |
| Linux (GUIなし) | 4K読み + 4K書き + 1M読み書き| 読み/書き値は通常類似 |

以下のディスクタイプの指標値は、通常〜フルパフォーマンス状態を示し、```libaio```をIOテストエンジンとして使用し、```Linux```でテストを実施した場合を指します：

| ドライブタイプ | 4K(IOPS)パフォーマンス | 1M(IOPS)パフォーマンス |
|------------|--------------------------|----------------------|
| NVMe SSD | ≥ 200 MB/s | 5-10 GB/s |
| 標準SSD | 50-100 MB/s | 2-3 GB/s |
| HDD (機械式ハードディスク) | 10-40 MB/s | 500-600 MB/s |
| 性能不良 | < 10 MB/s | < 200 MB/s |

迅速な評価：

1. **主要チェック**: 4K読み(IOPS) 4K書き(IOPS)
    - ほぼ同じで大きな差がない
    - ≥ 200 MB/s = NVMe SSD
    - 50-100 MB/s = 標準SSD
    - 10-40 MB/s = HDD (機械式ハードディスク)
    - < 10 MB/s = 性能不良、オーバーセリング/制限が深刻

2. **副次チェック**: 1M合計(IOPS)
    - プロバイダが設定したIO制限
    - リソースのオーバーコミット状況
    - 値が高いほど良い
    - NVMe SSDは通常4-6 GB/s達成
    - 標準SSDは通常1-2 GB/s達成

NVMe SSDの1M(IOPS)値が< 1GB/sの場合、深刻なリソースオーバーコミットが存在することを示します。

注意：ここでテストされるのは実際のIOであり、本プロジェクトに限定されます。本プロジェクト以外のテストによるIOは基準の普遍性を保証できません。他のテストでは同じパラメータを使用していない可能性があり、IO直接読み書きを設定していない可能性、IOエンジンの設定が一致しない可能性、テスト時間の設定が一致しない可能性があり、これらはすべて基準にズレを生じさせる原因となります。

### **ストリーミングメディアロック解除**

一般的なストリーミングメディアプラットフォームのロック解除を検索します。もちろん、すべてがストリーミングメディアというわけではなく、他の一般的なプラットフォームのロック解除も含まれています。一般的に、IPのロック解除地域は一貫しており、あちこちに変動することはありません。複数のプラットフォームでロック解除地域が一致しない場合、そのIPはIPXOなどのプラットフォームからレンタルされている可能性が高く、各プラットフォームのデータベース識別が遅いため、IP品質も一般的に良くないと考えられます。

### **IP品質検出**

14のデータベースのIP関連情報を検出します。一般的に、使用タイプ、会社タイプ、およびその他のセキュリティ情報の判別を見るだけで十分です。セキュリティスコアは参考程度です。複数のプラットフォームで対応する検出項目が一致している場合、現在のIPが確かにそうであることを証明しています。単一のデータベースソースの情報だけを信頼しないでください。

### **メールポート検出**

- **SMTP（25）**：メールサーバー間でメールを転送するために使用されます（メール送信）。
- **SMTPS（465）**：暗号化されたSMTPメール送信（SSL/TLS方式）に使用されます。
- **SMTP（587）**：クライアントからメールサーバーへのメール送信に使用され、STARTTLS暗号化をサポートします。
- **POP3（110）**：メールクライアントがサーバーからメールをダウンロードするために使用され、暗号化されていません。
- **POP3S（995）**：暗号化されたPOP3用で、安全にメールをダウンロードします（SSL/TLS方式）。
- **IMAP（143）**：メールクライアントがオンラインでメールを管理するために使用されます（メールの閲覧、同期）、暗号化されていません。
- **IMAPS（993）**：暗号化されたIMAP用で、安全にメールを管理します（SSL/TLS方式）。

具体的に現在のホストマシンがメールサーバーとして使用されていない、または電子メールを受信しない場合、この項目の指標は気にする必要はありません。

---
