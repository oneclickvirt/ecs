## 目录 / Table of Contents / 目次

[![Hits](https://hits.spiritlhl.net/goecs.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false)](https://hits.spiritlhl.net) [![Downloads](https://ghdownload.spiritlhl.net/oneclickvirt/ecs?color=36c600)](https://github.com/oneclickvirt/ecs/releases)

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
- [上游及回程线路检测](#上游及回程线路检测)
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
- [Nearby Speed Testing](#Nearby-Speed-Testing)

## 日本語
- [システム基本情報](#システム基本情報)
- [CPUテスト](#CPUテスト)
- [メモリテスト](#メモリテスト)
- [ディスクテスト](#ディスクテスト)
- [ストリーミングメディアロック解除](#ストリーミングメディアロック解除)
- [IP品質検出](#IP品質検出)
- [メールポート検出](#メールポート検出)
- [近隣スピードテスト](#近隣スピードテスト)

---

## 中文

### **系统基础信息**

依赖项目：[https://github.com/oneclickvirt/basics](https://github.com/oneclickvirt/basics) [https://github.com/oneclickvirt/gostun](https://github.com/oneclickvirt/gostun)

CPU型号: 一般来说，按CPU的发布时间，都是新款则AMD好于Intel，都是旧款则Intel好于AMD，而Apple的M系列芯片则是断层式领先。

CPU数量: 会检测是物理核心还是逻辑核心，优先展示物理核心，查不到物理核心才去展示逻辑核心。在服务器实际使用过程中，程序一般是按逻辑核心分配执行的，非视频转码和科学计算，物理核心一般都是开超线程成逻辑核心用，横向比较的时候，对应类型的核心数量才有比较的意义。当然如果一个是物理核一个是虚拟核，在后面CPU测试得分类似的情况下，肯定是物理核更优，无需担忧CPU性能被共享的问题。

CPU缓存：显示的宿主机的CPU三级缓存信息。对普通应用可能影响不大，但对数据库、编译、大规模并发请求等场景，L3 大小显著影响性能。

AES-NI: AES指令集是加密解密加速用的，对于HTTPS(TLS/SSL)网络请求、VPN代理(配置使用的AES加密的)、磁盘加密这些场景有明显的优化，更快更省资源。

VM-x/AMD-V/Hyper-V: 是当前测试宿主机是否支持嵌套虚拟化的指标，如果测试环境是套在docker里测或者没有root权限，那么这个默认就是检测不到显示不支持嵌套虚拟化。这个指标在你需要在宿主机上开设虚拟机(如 KVM、VirtualBox、VMware)的时候有用，其他用途该指标用处不大。

内存: 显示内存 正在使用的大小/总大小 ，不含虚拟内存。

气球驱动: 显示宿主机是否启用了气球驱动。气球驱动用于宿主机和虚拟机之间动态调节内存分配：宿主机可以通过驱动要求虚拟机“放气”释放部分内存，或“充气”占用更多内存。启用它通常意味着宿主机具备内存超售能力，但是否真的存在超售，需要结合下面的内存读写测试查看是否有超售/严格的限制。

内核页合并：显示宿主机是否启用了内核页合并机制。KSM 会将多个进程中内容相同的内存页合并为一份，以减少物理内存占用。启用它通常意味着宿主机可能在进行内存节省或存在一定程度的内存超售。是否真正造成性能影响或内存紧张，需要结合下面的内存读写测试查看是否有超售/严格的限制。

虚拟内存: swap虚拟内存 是磁盘上划出的虚拟内存空间，用来在物理内存不足时临时存放数据。它能防止内存不足导致程序崩溃，但频繁使用会明显拖慢系统，Linux 官方推荐的 swap 配置如下：

| 物理内存大小             | 推荐 SWAP 大小 |
| ------------------ | ---------- |
| ≤ 2G               | 内存的 2 倍    |
| 2G < 内存 ≤ 8G       | 等于物理内存大小   |
| ≥ 8G               | 约 8G 即可    |
| 需要休眠 (hibernation) | 至少等于物理内存大小 |

硬盘空间: 显示硬盘 正在使用的大小/总大小

启动盘路径：显示启动盘的路径

系统: 显示系统名字和架构

内核: 显示系统内核版本

系统在线时间: 显示宿主机自从开机到测试时的已在线时长

时区: 显示宿主机系统时区

负载: 显示系统负载

虚拟化架构: 显示宿主机来自什么虚拟化架构，一般来说推荐```Dedicated > KVM > Xen```虚拟化，其他虚拟化都会存在性能损耗，导致使用的时候存在性能共享/损耗，但这个也说不准，独立服务器才拥有完全独立的资源占用，其他虚拟化基本都会有资源共享，取决于宿主机的持有者对这个虚拟机是否有良心，具体性能优劣还是得看后面的专项性能测试。

NAT类型: 显示NAT类型，具体推荐```Full Cone > Restricted Cone > Port Restricted Cone > Symmetric```，测不出来或者非正规协议的类型会显示```Inconclusive```，一般来说只有特殊用途，比如有特殊的代理、实时通讯、做FRP内穿端口等需求才需要特别关注，其他一般情况下都不用关注本指标。

TCP加速方式：一般是```cubic/bbr```拥塞控制协议，一般来说做代理服务器用bbr可以改善网速，普通用途不必关注此指标。

IPV4/IPV6 ASN: 显示宿主机IP所属的ASN组织ID和名字，同一个IDC可能会有多个ASN，ASN下可能有多个商家售卖不同段的IP的服务器，具体的上下游关系错综复杂，可使用 bgp.tool 进一步查看。

IPV4/IPV6 Location: 显示对应协议的IP在数据库中的地理位置。

IPV4 Active IPs: 根据 bgp.tools 信息查询当前CIDR分块中 活跃邻居数量/总邻居数量 由于是非实时的，可能存在延迟。

IPV6 子网掩码：根据宿主机信息查询的本机IPV6子网大小

### **CPU测试**

依赖项目：[https://github.com/oneclickvirt/cputest](https://github.com/oneclickvirt/cputest)

支持通过命令行参数选择```GeekBench```和```Sysbench```进行测试：

| 比较项             | sysbench | geekbench |
|------------------|----------|-----------|
| 适用范围         | 轻量级，几乎可在任何服务器上运行 | 重量级，小型机器无法运行 |
| 测试要求         | 无需网络，无特殊硬件需求 | 需联网，IPV4环境，至少1G内存 |
| 开源情况         | 基于LUA，开源，可自行编译各架构版本(本项目有重构为Go版本内置) | 官方二进制闭源代码，不支持自行编译 |
| 测试稳定性       | 核心测试组件10年以上未变 | 每个大版本更新测试项，分数不同版本间难以对比(每个版本对标当前最好的CPU) |
| 测试内容         | 仅测试计算性能，基于素数计算 | 覆盖多种性能测试，分数加权计算，但部分测试实际不常用 |
| 适用场景         | 适合快速测试，仅测试计算性能 | 适合综合全面的测试 |
| 排行榜         | [sysbench.spiritlhl.net](https://sysbench.spiritlhl.net/) | [browser.geekbench.com](https://browser.geekbench.com/) |

默认使用```Sysbench```进行测试，基准大致如下：

CPU测试单核```Sysbench```得分在5000以上的可以算第一梯队，4000到5000分算第二梯队，每1000分大致算一档。

AMD的7950x单核满血性能得分在6500左右，AMD的5950x单核满血性能得分5700左右，Intel普通的CPU(E5之类的)在1000~800左右，低于500的单核CPU可以说是性能比较差的了。

有时候多核得分和单核得分一样，证明商家在限制程序并发使用CPU，典型例子腾讯云。

```Sysbench```的基准可见 [CPU Performance Ladder For Sysbench](https://sysbench.spiritlhl.net/) 天梯图，具体得分不分测试的sysbench的版本。

```GeekBench```的基准可见 [官方网站](https://browser.geekbench.com/processor-benchmarks/) 天梯图，具体得分每个```GeekBench```版本都不一样，注意使用时测试的```GeekBench```版本是什么。

多说一句，```GeekBench```测的很多内容，实际在服务器使用过程中根本用不到，测试仅供参考。当然```Sysbench```非常不全面，但它基于最基础的计算性能可以大致比较CPU的性能。

实际上CPU性能测试够用就行，除非是科学计算以及视频转码，一般不需要特别追求高性能CPU。如果有性能需求，那么需要关注程序本身吃的是多核还是单核，对应看多核还是单核得分。

### **内存测试**

依赖项目：[https://github.com/oneclickvirt/memorytest](https://github.com/oneclickvirt/memorytest)

一般来说，只需要判断 IO 速度是否低于 `10240 MB/s (≈10 GB/s)`，
如果低于这个值，那么证明内存性能不佳，极大概率存在超售超卖问题。

至于超开的原因可能是：

* 开了虚拟内存 (硬盘当内存用)
* 开了 ZRAM (牺牲 CPU 性能)
* 开了气球驱动 (Balloon Driver)
* 开了 KSM 内存融合

原因多种多样。

| 内存类型 | 典型频率 (MHz)   | 单通道带宽                                 | 双通道带宽                                   |
| ---- | ------------ | ------------------------------------- | --------------------------------------- |
| DDR3 | 1333 \~ 2133 | 10 \~ 17 GB/s (≈ 10240 \~ 17408 MB/s) | 20 \~ 34 GB/s (≈ 20480 \~ 34816 MB/s)   |
| DDR4 | 2133 \~ 3200 | 17 \~ 25 GB/s (≈ 17408 \~ 25600 MB/s) | 34 \~ 50 GB/s (≈ 34816 \~ 51200 MB/s)   |
| DDR5 | 4800 \~ 7200 | 38 \~ 57 GB/s (≈ 38912 \~ 58368 MB/s) | 76 \~ 114 GB/s (≈ 77824 \~ 116736 MB/s) |

根据上表内容，本项目测试的粗略判断方法：

* **< 20 GB/s (20480 MB/s)** → 可能是 DDR3（或 DDR4 单通道 / 低频）
* **20 \~ 40 GB/s (20480 \~ 40960 MB/s)** → 大概率 DDR4
* **≈ 50 GB/s (≈ 51200 MB/s)** → 基本就是 DDR5

### **硬盘测试**

依赖项目：[https://github.com/oneclickvirt/disktest](https://github.com/oneclickvirt/disktest)

```dd```测试可能误差偏大但测试速度快无硬盘大小限制，```fio```测试真实一些但测试速度慢有硬盘以及内存大小的最低需求。

同时，服务器可能有不同的文件系统，某些文件系统的IO引擎在同样的硬件条件下测试的读写速度更快，这是正常的。项目默认使用```fio```进行测试，测试使用的IO引擎优先级为```libaio > posixaio > psync```，备选项```dd```测试在```fio```测试不可用时自动替换。

以```fio```测试结果为例基准如下：

| 操作系统类型 | 主要指标 | 次要指标 |
|---------|-------------------|---------------------|
| Windows/MacOS | 4K读 → 64K读 → 写入测试 | 图形界面系统优先考虑读取性能 |
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

依赖项目：[https://github.com/oneclickvirt/CommonMediaTests](https://github.com/oneclickvirt/CommonMediaTests) [https://github.com/oneclickvirt/UnlockTests](https://github.com/oneclickvirt/UnlockTests)

默认只检测跨国流媒体解锁。

一般来说，正常的情况下，一个IP多个流媒体的解锁地区都是一致的不会到处乱飘，如果发现多家平台解锁地区不一致，那么IP大概率来自IPXO等平台租赁或者是刚刚宣告和被使用，未被流媒体普通的数据库所识别修正地域。由于各平台的IP数据库识别速度不一致，所以有时候有的平台解锁区域正常，有的飘到路由上的某个位置，有的飘到IP未被你使用前所在的位置。

| DNS 类型       | 解锁方式判断是否必要 | DNS 对解锁影响 | 说明                                      |
| ------------ | ---------- | --------- | --------------------------------------- |
| 官方主流 DNS     | 否          | 小         | 流媒体解锁主要依赖节点 IP，DNS 解析基本不会干扰解锁。          |
| 非主流 / 自建 DNS | 是          | 大         | 流媒体解锁结果受 DNS 解析影响较大，需要判断是原生解锁还是 DNS 解锁。 |

所以测试过程中，如果宿主机当前使用的是官方主流的DNS，不会进行是否为原生解锁的判断。

### **IP质量检测**

依赖项目：[https://github.com/oneclickvirt/securityCheck](https://github.com/oneclickvirt/securityCheck)

检测14个数据库的IP相关信息，一般来说看使用类型和公司类型还有安全信息的其他判别足矣，安全得分真的图一乐。多个平台比较对应检测项目都为对应值，证明当前IP确实如此，不要仅相信一个数据库源的信息。

* **使用类型 & 公司类型**：显示IP归属和使用场景，例如是否属于家庭用户、企业办公、托管服务或云/数据中心。
* **云提供商 / 数据中心 / 移动设备**：判断IP是否来自云服务、数据中心或移动网络，帮助识别共享或高风险IP。
* **代理 / VPN / Tor / Tor出口**：检测IP是否用于隐藏真实身份或位置，可能涉及匿名访问或滥用行为。
* **网络爬虫 / 机器人**：识别自动化访问或采集程序，对安全风险评估有参考价值。
* **匿名 / 滥用者 / 威胁 / 中继 / Bogon**：显示IP历史行为特征和是否属于保留/未分配IP，辅助判断IP可信度。
* **安全得分、声誉、信任得分、威胁得分、欺诈得分、滥用得分**：各数据库对IP的量化安全评价，仅供参考。
* **社区投票 & 黑名单记录**：展示用户反馈及公共黑名单信息，可快速识别潜在风险。
* **Google搜索可行性**：检测IP访问Google搜索服务的可行性，间接反映网络限制或屏蔽情况。

多平台对比更可靠，不同数据库算法和更新频率不同，单一来源可能存在误判。多个数据库显示相似结果，说明这个结果更可靠。

### **邮件端口检测**

依赖项目：[https://github.com/oneclickvirt/portchecker](https://github.com/oneclickvirt/portchecker)

- **SMTP（25）**：用于邮件服务器之间传输邮件（发送邮件）。
- **SMTPS（465）**：用于加密的 SMTP 发送邮件（SSL/TLS 方式）。
- **SMTP（587）**：用于客户端向邮件服务器发送邮件，支持 STARTTLS 加密。
- **POP3（110）**：用于邮件客户端从服务器下载邮件，不加密。
- **POP3S（995）**：用于加密的 POP3，安全地下载邮件（SSL/TLS 方式）。
- **IMAP（143）**：用于邮件客户端在线管理邮件（查看、同步邮件），不加密。
- **IMAPS（993）**：用于加密的 IMAP，安全地管理邮件（SSL/TLS 方式）。

具体当前宿主机不做邮局，不收发电子邮件，那么该项目指标不需要理会。

### **上游及回程线路检测**

依赖项目：[https://github.com/oneclickvirt/backtrace](https://github.com/oneclickvirt/backtrace)

#### 上游类型与运营商等级说明

- **直接上游（Direct Upstream）**  
  当前运营商直接购买网络服务的上级运营商，通常是 BGP 邻居。

- **间接上游（Indirect Upstream）**  
  直接上游的上级，形成层层向上的关系链。可通过 BGP 路由路径中的多跳信息识别。

| 等级 | 描述 |
|------|------|
| **Tier 1 Global** | 全球顶级运营商（如 AT&T、Verizon、NTT、Telia 等），之间免费互联（Settlement-Free Peering），不依赖他人即可访问全球任意网络。 |
| **Tier 1 Regional** | 区域性顶级运营商，在特定区域具有一级能力，但在全球范围互联性稍弱。 |
| **Tier 1 Indirect** | 间接连接的 Tier 1（非直接购买），通过中间上游间接接入 Tier 1 网络。 |
| **Tier 2** | 需要向 Tier 1 付费购买上网能力的二级运营商，通常是各国主流电信商或 ISP。 |
| **CDN Provider** | 内容分发网络提供商，如 Cloudflare、Akamai、Fastly 等，主要用于内容加速而非传统上游。 |
| **Direct/Indirect Others** | 其他类型的直接或间接连接，如 IX（Internet Exchange）成员、私有对等互联等。 |

上游质量判断：直接接入的高等级上游（特别是 Tier 1 Global）越多，通常网络连通性越好。但实际网络质量也受到以下因素影响：

  - 上下游之间的商业结算关系；
  - 购买的带宽套餐和服务质量；
  - 对等端口（Peering Ports）大小和负载；
  - 网络拥塞、路由策略、延迟路径等。

无法完全从 BGP 路由中判断。

一般来说，**接入高质量上游越多，网络连通性越优**。但由于存在诸多不可见的商业和技术因素，**无法仅凭上游等级准确判断网络质量**，上游检测约等于图一乐，实际得看对应的路由情况和长时间Ping的情况。

然后是检测当前的宿主机的IP地址 到 四个主要POP点城市的三个主要运营商的接入点的IP地址 的线路，具体来说

电信163、联通4837、移动CMI 是常见的线路，移动CMI对两广地区的移动运营商特供延迟低，也能算优质，仅限两广移动。

电信CN2GIA > 电信CN2GT 移动CMIN2 联通9929 算优质的线路

用什么运营商连宿主机的IP就看哪个运营商的线路就行了，具体线路的路由情况，看在下一个检测项看到对应的ICMP检测路由信息。

### **三网回程路由检测**

依赖项目：[https://github.com/oneclickvirt/nt3](https://github.com/oneclickvirt/nt3)

默认检测广州为目的地，实际可使用命令行参数指定目的地，见对应的参数说明。

主要就是看是不是直连，是不是延迟低，是不是没有隐藏路由信息，有没有一些优质线路或IX链接。

如果路由全球跑，延迟起飞，那么线路自然不会好到哪里去。

有时候路由信息完全藏起来了，只知道实际使用的延迟低，实际可能也是优质线路只是查不到信息，这就没办法直接识别了。

### **就近测速**

依赖项目：[https://github.com/oneclickvirt/speedtest](https://github.com/oneclickvirt/speedtest)

先测的官方推荐的测速点，然后测有代表性的国际测速点，最后测国内三大运营商ping值最低的测速点。

境内使用为主就看境内测速即可，境外使用看境外测速，官方测速点可以代表受测的宿主机本地带宽基准。

一般来说中国境外的服务器的带宽100Mbps起步，中国境内的服务器1Mbps带宽起步，具体看线路优劣，带宽特别大有时候未必用得上，够用就行了。

日常我偏向使用1Gbps带宽的服务器，至少下载依赖什么的速度足够快，境内小水管几Mbps真的下半天下不完，恨不得到机房插个U盘转移数据。

---

## English

### Basic System Information

Dependency project: [https://github.com/oneclickvirt/basics](https://github.com/oneclickvirt/basics) [https://github.com/oneclickvirt/gostun](https://github.com/oneclickvirt/gostun)

**CPU Model**: Generally speaking, based on CPU release time, newer models favor AMD over Intel, while older models favor Intel over AMD. Apple's M-series chips are in a league of their own.

**CPU Count**: Detects whether cores are physical or logical, prioritizing physical cores display. Logical cores are shown only when physical cores cannot be detected. In actual server usage, programs are generally allocated based on logical cores. For non-video transcoding and scientific computing, physical cores usually have hyperthreading enabled to function as logical cores. When making comparisons, only corresponding core types are meaningful. Of course, if one is physical and the other is virtual, with similar CPU test scores, physical cores are definitely better without worrying about CPU performance sharing issues.

**CPU Cache**: Displays the host's CPU L3 cache information. While it may not significantly impact regular applications, for databases, compilation, and large-scale concurrent requests, L3 cache size significantly affects performance.

**AES-NI**: AES instruction set is used for encryption/decryption acceleration, providing significant optimization for HTTPS (TLS/SSL) network requests, VPN proxies (configured with AES encryption), and disk encryption scenarios - faster and more resource-efficient.

**VM-x/AMD-V/Hyper-V**: Indicators showing whether the current test host supports nested virtualization. If the test environment runs in Docker or lacks root permissions, this will show as unsupported by default. This indicator is useful when you need to create virtual machines (like KVM, VirtualBox, VMware) on the host; otherwise, it has limited utility.

**Memory**: Shows memory usage as currently used size/total size, excluding virtual memory.

**Balloon Driver**: Shows whether the host has balloon driver enabled. Balloon driver is used for dynamic memory allocation between host and virtual machines: the host can request virtual machines to "deflate" and release some memory, or "inflate" to occupy more memory. Enabling it usually means the host has memory overselling capability, but whether actual overselling exists needs to be checked with memory read/write tests below for overselling/strict limitations.

**Kernel Page Merging**: Shows whether the host has kernel page merging mechanism enabled. KSM merges memory pages with identical content from multiple processes into a single copy to reduce physical memory usage. Enabling it usually means the host may be implementing memory savings or has some degree of memory overselling. Whether it actually causes performance impact or memory pressure needs to be checked with memory read/write tests below for overselling/strict limitations.

**Virtual Memory**: Swap virtual memory is virtual memory space allocated on disk, used to temporarily store data when physical memory is insufficient. It prevents program crashes due to insufficient memory, but frequent use will noticeably slow down the system. Linux officially recommends swap configuration as follows:

| Physical Memory Size | Recommended SWAP Size |
| -------------------- | --------------------- |
| ≤ 2G                | 2x memory size        |
| 2G < memory ≤ 8G    | Equal to physical memory |
| ≥ 8G                | About 8G is sufficient |
| Hibernation needed  | At least equal to physical memory |

**Disk Space**: Shows disk usage as currently used size/total size

**Boot Disk Path**: Shows the boot disk path

**System**: Shows system name and architecture

**Kernel**: Shows system kernel version

**System Uptime**: Shows host uptime from boot to test time

**Time Zone**: Shows host system time zone

**Load**: Shows system load

**Virtualization Architecture**: Shows what virtualization architecture the host comes from. Generally recommended: `Dedicated > KVM > Xen` virtualization. Other virtualization types have performance losses, causing performance sharing/loss during use. However, this isn't definitive - only dedicated servers have completely independent resource usage. Other virtualization basically involves resource sharing, depending on whether the host holder is conscientious about this virtual machine. Actual performance superiority still depends on subsequent specialized performance tests.

**NAT Type**: Shows NAT type. Specifically recommended: `Full Cone > Restricted Cone > Port Restricted Cone > Symmetric`. Undetectable or non-standard protocol types show as `Inconclusive`. Generally, only special purposes like specific proxies, real-time communication, or FRP port forwarding need special attention to this indicator; other general situations don't need to focus on this metric.

**TCP Acceleration Method**: Usually `cubic/bbr` congestion control protocols. Generally speaking, using bbr for proxy servers can improve network speed; regular usage doesn't need to focus on this indicator.

**IPV4/IPV6 ASN**: Shows the ASN organization ID and name that the host IP belongs to. The same IDC may have multiple ASNs, and ASNs may have multiple merchants selling servers with different IP segments. The specific upstream and downstream relationships are complex; use bgp.tool for further investigation.

**IPV4/IPV6 Location**: Shows the geographic location of the corresponding protocol's IP in the database.

**IPV4 Active IPs**: Based on bgp.tools information, queries active neighbor count/total neighbor count in the current CIDR block. Since this is non-real-time, there may be delays.

**IPV6 Subnet Mask**: Queries the local IPV6 subnet size based on host information.

### CPU Testing

Dependency project: [https://github.com/oneclickvirt/cputest](https://github.com/oneclickvirt/cputest)

Supports command-line parameter selection between `GeekBench` and `Sysbench` for testing:

| Comparison Item | sysbench | geekbench |
|----------------|----------|-----------|
| Applicability | Lightweight, runs on almost any server | Heavy, cannot run on small machines |
| Test Requirements | No network needed, no special hardware requirements | Requires internet, IPV4 environment, at least 1G memory |
| Open Source | LUA-based, open source, can compile for various architectures (this project has Go version built-in) | Official binary closed source, doesn't support self-compilation |
| Test Stability | Core test components unchanged for 10+ years | Updates test items with each major version, scores difficult to compare across versions (each version benchmarks against current best CPUs) |
| Test Content | Only tests computational performance, based on prime calculations | Covers various performance tests, weighted score calculation, but some tests aren't commonly used |
| Use Case | Suitable for quick testing, only tests computational performance | Suitable for comprehensive testing |
| Leaderboard | [sysbench.spiritlhl.net](https://sysbench.spiritlhl.net/) | [browser.geekbench.com](https://browser.geekbench.com/) |

Default uses `Sysbench` for testing, with rough benchmarks as follows:

CPU test single-core `Sysbench` scores above 5000 can be considered first-tier, 4000-5000 points second-tier, roughly one tier per 1000 points.

AMD 7950x single-core full performance scores around 6500, AMD 5950x single-core full performance scores around 5700, Intel regular CPUs (E5 series) around 1000-800, single-core CPUs below 500 can be considered poor performance.

Sometimes multi-core and single-core scores are identical, proving the merchant is limiting program concurrent CPU usage, typical example being Tencent Cloud.

`Sysbench` benchmarks can be seen in the [CPU Performance Ladder For Sysbench](https://sysbench.spiritlhl.net/) tier chart. Specific scores depend on the sysbench version tested.

`GeekBench` benchmarks can be seen in the [official website](https://browser.geekbench.com/processor-benchmarks/) tier chart. Specific scores differ for each `GeekBench` version, pay attention to which `GeekBench` version was used during testing.

One more thing: `GeekBench` tests many things that are actually unused in server operations, tests are for reference only. Of course, `Sysbench` is very incomplete, but it can roughly compare CPU performance based on the most basic computational performance.

Actually, CPU performance testing should be sufficient. Unless for scientific computing and video transcoding, generally no need to particularly pursue high-performance CPUs. If there are performance requirements, need to focus on whether the program itself uses multi-core or single-core, and look at multi-core or single-core scores accordingly.

### Memory Testing

Dependency project: [https://github.com/oneclickvirt/memorytest](https://github.com/oneclickvirt/memorytest)

Generally speaking, you only need to determine if IO speed is below `10240 MB/s (≈10 GB/s)`.
If below this value, it proves poor memory performance with high probability of overselling issues.

Possible reasons for overselling:

* Virtual memory enabled (using disk as memory)
* ZRAM enabled (sacrificing CPU performance)
* Balloon driver enabled
* KSM memory fusion enabled

Various reasons exist.

| Memory Type | Typical Frequency (MHz) | Single Channel Bandwidth | Dual Channel Bandwidth |
| ----------- | ----------------------- | ------------------------ | ---------------------- |
| DDR3 | 1333 ~ 2133 | 10 ~ 17 GB/s (≈ 10240 ~ 17408 MB/s) | 20 ~ 34 GB/s (≈ 20480 ~ 34816 MB/s) |
| DDR4 | 2133 ~ 3200 | 17 ~ 25 GB/s (≈ 17408 ~ 25600 MB/s) | 34 ~ 50 GB/s (≈ 34816 ~ 51200 MB/s) |
| DDR5 | 4800 ~ 7200 | 38 ~ 57 GB/s (≈ 38912 ~ 58368 MB/s) | 76 ~ 114 GB/s (≈ 77824 ~ 116736 MB/s) |

Based on the above table, this project's rough judgment method:

* **< 20 GB/s (20480 MB/s)** → Possibly DDR3 (or DDR4 single channel / low frequency)
* **20 ~ 40 GB/s (20480 ~ 40960 MB/s)** → Most likely DDR4
* **≈ 50 GB/s (≈ 51200 MB/s)** → Basically DDR5

### Disk Testing

Dependency project: [https://github.com/oneclickvirt/disktest](https://github.com/oneclickvirt/disktest)

`dd` testing may have larger errors but tests quickly with no disk size limitations. `fio` testing is more realistic but tests slowly with minimum disk and memory size requirements.

Additionally, servers may have different file systems. Certain file systems' IO engines achieve faster read/write speeds under the same hardware conditions, which is normal. The project defaults to using `fio` for testing, with IO engine priority: `libaio > posixaio > psync`. Alternative `dd` testing automatically replaces when `fio` testing is unavailable.

Using `fio` test results as benchmark examples:

| OS Type | Primary Metrics | Secondary Metrics |
|---------|-----------------|-------------------|
| Windows/MacOS | 4K Read → 64K Read → Write Tests | GUI systems prioritize read performance |
| Linux (No GUI) | 4K Read + 4K Write + 1M Read/Write | Read/Write values usually similar |

The following disk types refer to regular~full performance states, using `libaio` as IO test engine, tested under `Linux`:

| Drive Type | 4K (IOPS) Performance | 1M (IOPS) Performance |
|------------|----------------------|----------------------|
| NVMe SSD | ≥ 200 MB/s | 5-10 GB/s |
| Standard SSD | 50-100 MB/s | 2-3 GB/s |
| HDD (Mechanical) | 10-40 MB/s | 500-600 MB/s |
| Poor Performance | < 10 MB/s | < 200 MB/s |

Quick Assessment:

1. **Primary Check**: 4K Read (IOPS) 4K Write (IOPS)
    - Nearly identical with little difference
    - ≥ 200 MB/s = NVMe SSD
    - 50-100 MB/s = Standard SSD
    - 10-40 MB/s = HDD (Mechanical)
    - < 10 MB/s = Poor performance, severe overselling/restrictions

2. **Secondary Check**: 1M Total (IOPS)
    - Provider's IO limitations
    - Resource overselling situation
    - Higher values are better
    - NVMe SSD usually reaches 4-6 GB/s
    - Standard SSD usually reaches 1-2 GB/s

If NVMe SSD's 1M (IOPS) value < 1GB/s indicates serious resource overselling.

Note: This tests real IO, limited to this project. IO tests from other projects don't guarantee universal benchmarks because they may use different parameters, may not set direct IO read/write, may have inconsistent IO engines, or inconsistent test times, all causing benchmark deviations.

### Streaming Media Unlocking

Dependency project: [https://github.com/oneclickvirt/CommonMediaTests](https://github.com/oneclickvirt/CommonMediaTests) [https://github.com/oneclickvirt/UnlockTests](https://github.com/oneclickvirt/UnlockTests)

Default only checks cross-border streaming media unlocking.

Generally speaking, under normal circumstances, multiple streaming services for one IP should have consistent unlock regions without scattered locations. If multiple platforms show inconsistent unlock regions, the IP likely comes from platforms like IPXO rentals or has been recently announced and used, not yet recognized and corrected by streaming media common databases. Due to inconsistent IP database recognition speeds across platforms, sometimes some platforms unlock regions normally, some drift to certain router locations, and some drift to where the IP was before you used it.

| DNS Type | Unlock Method Judgment Necessary | DNS Impact on Unlocking | Description |
| -------- | ------------------------------- | ----------------------- | ----------- |
| Official Mainstream DNS | No | Small | Streaming unlock mainly relies on node IP, DNS resolution basically doesn't interfere with unlocking |
| Non-mainstream / Self-built DNS | Yes | Large | Streaming unlock results greatly affected by DNS resolution, need to judge if it's native unlock or DNS unlock |

So during testing, if the host currently uses official mainstream DNS, no judgment of whether it's native unlocking will be performed.

### IP Quality Detection

Dependency project: [https://github.com/oneclickvirt/securityCheck](https://github.com/oneclickvirt/securityCheck)

Detects IP-related information from 14 databases. Generally speaking, looking at usage type, company type, and security information's other discriminators is sufficient. Security scores are really just for fun. When multiple platforms show corresponding detection items all having corresponding values, it proves the current IP is indeed as such - don't trust information from just one database source.

* **Usage Type & Company Type**: Shows IP attribution and usage scenarios, such as whether it belongs to home users, enterprise office, hosting services, or cloud/data centers.
* **Cloud Provider / Data Center / Mobile Device**: Determines if IP comes from cloud services, data centers, or mobile networks, helping identify shared or high-risk IPs.
* **Proxy / VPN / Tor / Tor Exit**: Detects if IP is used to hide real identity or location, possibly involving anonymous access or abuse behavior.
* **Web Crawler / Bot**: Identifies automated access or collection programs, with reference value for security risk assessment.
* **Anonymous / Abuser / Threat / Relay / Bogon**: Shows IP historical behavior characteristics and whether it belongs to reserved/unallocated IPs, assisting in judging IP credibility.
* **Security Score, Reputation, Trust Score, Threat Score, Fraud Score, Abuse Score**: Various databases' quantified security evaluations of IPs, for reference only.
* **Community Voting & Blacklist Records**: Shows user feedback and public blacklist information, can quickly identify potential risks.
* **Google Search Feasibility**: Tests IP's feasibility for accessing Google search services, indirectly reflecting network restrictions or blocking situations.

Multi-platform comparison is more reliable. Different databases have different algorithms and update frequencies; single sources may misjudge. Similar results from multiple databases indicate higher reliability.

### Email Port Detection

Dependency project: [https://github.com/oneclickvirt/portchecker](https://github.com/oneclickvirt/portchecker)

- **SMTP (25)**: Used for mail transfer between mail servers (sending mail).
- **SMTPS (465)**: Used for encrypted SMTP mail sending (SSL/TLS method).
- **SMTP (587)**: Used for clients sending mail to mail servers, supports STARTTLS encryption.
- **POP3 (110)**: Used for mail clients downloading mail from servers, unencrypted.
- **POP3S (995)**: Used for encrypted POP3, securely downloading mail (SSL/TLS method).
- **IMAP (143)**: Used for mail clients managing mail online (viewing, syncing mail), unencrypted.
- **IMAPS (993)**: Used for encrypted IMAP, securely managing mail (SSL/TLS method).

If the current host doesn't function as a mail server and doesn't send/receive emails, this project indicator can be ignored.

### Nearby Speed Testing

Dependency project: [https://github.com/oneclickvirt/speedtest](https://github.com/oneclickvirt/speedtest)

First test the officially recommended speed test points, then test representative international speed test points.

Official speed test points can represent the local bandwidth baseline of the host machine being tested.

In daily use, I prefer to use servers with 1Gbps bandwidth, at least the speed of downloading dependencies is fast enough.

---

## 日本語

### システム基本情報

依存プロジェクト：[https://github.com/oneclickvirt/basics](https://github.com/oneclickvirt/basics) [https://github.com/oneclickvirt/gostun](https://github.com/oneclickvirt/gostun)

**CPUモデル**: 一般的に、CPUのリリース時期に基づいて、新しいモデルではAMDがIntelより優れており、古いモデルではIntelがAMDより優れています。一方、AppleのMシリーズチップは圧倒的に優位に立っています。

**CPU数量**: 物理コアか論理コアかを検出し、物理コアの表示を優先します。物理コアが検出できない場合のみ論理コアを表示します。実際のサーバー使用において、プログラムは一般的に論理コアベースで実行が割り当てられます。動画変換や科学計算以外では、物理コアは通常ハイパースレッディングを有効にして論理コアとして使用されます。比較する際は、対応するコアタイプの数量のみが比較意義を持ちます。もちろん、一つが物理コア、もう一つが仮想コアで、CPUテストスコアが似ている場合、物理コアの方が明らかに優れており、CPU性能共有の問題を心配する必要はありません。

**CPUキャッシュ**: ホストのCPU L3キャッシュ情報を表示します。一般的なアプリケーションにはあまり影響しないかもしれませんが、データベース、コンパイル、大規模な並行リクエストなどのシナリオでは、L3キャッシュサイズが性能に大きく影響します。

**AES-NI**: AES命令セットは暗号化・復号化の高速化に使用され、HTTPS（TLS/SSL）ネットワークリクエスト、VPNプロキシ（AES暗号化設定使用）、ディスク暗号化などのシナリオで明らかな最適化を提供し、より高速でリソース効率的です。

**VM-x/AMD-V/Hyper-V**: 現在のテストホストがネスト仮想化をサポートしているかどうかの指標です。テスト環境がDockerで実行されているか、root権限がない場合、デフォルトではネスト仮想化不サポートとして表示されます。この指標は、ホスト上で仮想マシン（KVM、VirtualBox、VMwareなど）を作成する必要がある場合に有用で、その他の用途では限定的です。

**メモリ**: メモリ使用量を現在使用中サイズ/総サイズで表示します。仮想メモリは含まれません。

**バルーンドライバー**: ホストでバルーンドライバーが有効になっているかどうかを表示します。バルーンドライバーはホストと仮想マシン間での動的メモリ割り当てに使用され、ホストは仮想マシンに「収縮」してメモリの一部を解放するか、「膨張」してより多くのメモリを占有するよう要求できます。有効化は通常、ホストがメモリオーバーセリング機能を持っていることを意味しますが、実際にオーバーセリングが存在するかどうかは、下記のメモリ読み書きテストでオーバーセリング/厳格な制限を確認する必要があります。

**カーネルページマージ**: ホストでカーネルページマージ機能が有効になっているかどうかを表示します。KSMは複数のプロセスから同一内容のメモリページを1つにマージして物理メモリ使用量を削減します。有効化は通常、ホストがメモリ節約を実施しているか、ある程度のメモリオーバーセリングがあることを意味します。実際に性能影響やメモリ不足を引き起こすかどうかは、下記のメモリ読み書きテストでオーバーセリング/厳格な制限を確認する必要があります。

**仮想メモリ**: スワップ仮想メモリは、ディスク上に割り当てられた仮想メモリ空間で、物理メモリ不足時にデータを一時的に格納するために使用されます。メモリ不足によるプログラムクラッシュを防ぎますが、頻繁な使用はシステムを著しく遅くします。Linuxが公式推奨するスワップ設定は以下の通りです：

| 物理メモリサイズ | 推奨SWAPサイズ |
| ---------------- | -------------- |
| ≤ 2G            | メモリの2倍    |
| 2G < メモリ ≤ 8G | 物理メモリサイズと同等 |
| ≥ 8G            | 約8Gで十分     |
| 休止状態（hibernation）必要 | 最低でも物理メモリサイズと同等 |

**ディスク容量**: ディスク使用量を現在使用中サイズ/総サイズで表示します

**起動ディスクパス**: 起動ディスクのパスを表示します

**システム**: システム名とアーキテクチャを表示します

**カーネル**: システムカーネルバージョンを表示します

**システム稼働時間**: ホストの起動からテスト時点までの稼働時間を表示します

**タイムゾーン**: ホストシステムのタイムゾーンを表示します

**負荷**: システム負荷を表示します

**仮想化アーキテクチャ**: ホストがどの仮想化アーキテクチャから来ているかを表示します。一般的に推奨順序：`Dedicated > KVM > Xen`仮想化。その他の仮想化は性能損失があり、使用時に性能共有/損失が発生しますが、これは断定的ではありません。専用サーバーのみが完全に独立したリソース占有を持ちます。その他の仮想化は基本的にリソース共有があり、ホスト保有者がこの仮想マシンに対して良心的かどうかに依存します。実際の性能優劣は後続の専門性能テストを見る必要があります。

**NATタイプ**: NATタイプを表示します。具体的な推奨順序：`Full Cone > Restricted Cone > Port Restricted Cone > Symmetric`。検出不可能または非標準プロトコルタイプは`Inconclusive`と表示されます。一般的に、特殊な用途、例えば特殊なプロキシ、リアルタイム通信、FRPポート転送などの需要がある場合のみ特別に注意が必要で、その他の一般状況では本指標に注意する必要はありません。

**TCP高速化方式**: 通常`cubic/bbr`輻輳制御プロトコルです。一般的に、プロキシサーバーでbbrを使用すると通信速度を改善できますが、普通の用途では本指標に注意する必要はありません。

**IPV4/IPV6 ASN**: ホストIPが属するASN組織IDと名前を表示します。同じIDCが複数のASNを持つ可能性があり、ASNの下に異なるIPセグメントのサーバーを販売する複数の業者がいる可能性があります。具体的な上下流関係は複雑で、bgp.toolでさらなる調査が可能です。

**IPV4/IPV6 Location**: 対応プロトコルのIPのデータベース内地理位置を表示します。

**IPV4 Active IPs**: bgp.tools情報に基づき、現在のCIDRブロック内のアクティブ近隣数/総近隣数を照会します。非リアルタイムのため遅延がある可能性があります。

**IPV6 サブネットマスク**: ホスト情報に基づいてローカルIPV6サブネットサイズを照会します。

### CPUテスト

依存プロジェクト：[https://github.com/oneclickvirt/cputest](https://github.com/oneclickvirt/cputest)

コマンドラインパラメータで`GeekBench`と`Sysbench`のテスト選択をサポートします：

| 比較項目 | sysbench | geekbench |
|----------|----------|-----------|
| 適用範囲 | 軽量、ほぼすべてのサーバーで実行可能 | 重量級、小型マシンでは実行不可 |
| テスト要件 | ネットワーク不要、特殊ハードウェア不要 | インターネット必要、IPV4環境、最低1Gメモリ |
| オープンソース状況 | LUAベース、オープンソース、各アーキテクチャ版を自分でコンパイル可能（本プロジェクトはGo版を内蔵） | 公式バイナリクローズドソース、自分でコンパイル不可 |
| テスト安定性 | コアテストコンポーネント10年以上変更なし | 各メジャーバージョン更新でテスト項目変更、異なるバージョン間でスコア比較困難（各バージョンは現在最高のCPUを基準） |
| テスト内容 | 計算性能のみテスト、素数計算ベース | 多種性能テストをカバー、スコア加重計算、但し一部テストは実際には非常用 |
| 適用シナリオ | 迅速テストに適し、計算性能のみテスト | 総合的で全面的なテストに適用 |
| ランキング | [sysbench.spiritlhl.net](https://sysbench.spiritlhl.net/) | [browser.geekbench.com](https://browser.geekbench.com/) |

デフォルトで`Sysbench`でテストを行い、基準は大まかに以下の通りです：

CPUテストシングルコア`Sysbench`スコア5000以上は第一階層、4000-5000点は第二階層、1000点毎に大体一階層と考えられます。

AMD 7950xシングルコアフル性能スコアは6500前後、AMD 5950xシングルコアフル性能スコアは5700前後、Intel普通のCPU（E5系など）は1000~800前後、500以下のシングルコアCPUは性能が比較的劣ると言えます。

時々マルチコアスコアとシングルコアスコアが同じになることがあり、これは業者がプログラムの並行CPU使用を制限していることを証明します。典型例はTencent Cloudです。

`Sysbench`の基準は[CPU Performance Ladder For Sysbench](https://sysbench.spiritlhl.net/)階層図で確認可能、具体的なスコアはテストしたsysbenchのバージョンに依存しません。

`GeekBench`の基準は[公式サイト](https://browser.geekbench.com/processor-benchmarks/)階層図で確認可能、具体的なスコアは各`GeekBench`バージョンで異なり、使用時にテストした`GeekBench`バージョンに注意してください。

もう一つ付け加えると、`GeekBench`はテストする多くの内容が実際のサーバー使用過程では全く使われないため、テストは参考のみです。もちろん`Sysbench`は非常に不完全ですが、最も基本的な計算性能に基づいてCPUの性能を大まかに比較できます。

実際にはCPU性能テストは十分であれば良く、科学計算や動画変換以外では、特に高性能CPUを追求する必要は一般的にありません。性能需要がある場合は、プログラム自体がマルチコアかシングルコアを使うかに注意し、対応してマルチコアかシングルコアのスコアを見る必要があります。

### メモリテスト

依存プロジェクト：[https://github.com/oneclickvirt/memorytest](https://github.com/oneclickvirt/memorytest)

一般的に、IO速度が`10240 MB/s (≈10 GB/s)`を下回るかどうかを判断するだけで十分です。
この値を下回る場合、メモリ性能が不良で、オーバーセリング問題が存在する可能性が極めて高いことを証明します。

オーバーセリングの原因は以下が考えられます：

* 仮想メモリの有効化（ディスクをメモリとして使用）
* ZRAMの有効化（CPU性能を犠牲）
* バルーンドライバーの有効化
* KSMメモリ融合の有効化

原因は多種多様です。

| メモリタイプ | 典型的周波数 (MHz) | シングルチャネル帯域幅 | デュアルチャネル帯域幅 |
| ----------- | ------------------ | -------------------- | -------------------- |
| DDR3 | 1333 ~ 2133 | 10 ~ 17 GB/s (≈ 10240 ~ 17408 MB/s) | 20 ~ 34 GB/s (≈ 20480 ~ 34816 MB/s) |
| DDR4 | 2133 ~ 3200 | 17 ~ 25 GB/s (≈ 17408 ~ 25600 MB/s) | 34 ~ 50 GB/s (≈ 34816 ~ 51200 MB/s) |
| DDR5 | 4800 ~ 7200 | 38 ~ 57 GB/s (≈ 38912 ~ 58368 MB/s) | 76 ~ 114 GB/s (≈ 77824 ~ 116736 MB/s) |

上表内容に基づく、本プロジェクトテストの粗略判断方法：

* **< 20 GB/s (20480 MB/s)** → DDR3の可能性（またはDDR4シングルチャネル/低周波数）
* **20 ~ 40 GB/s (20480 ~ 40960 MB/s)** → 高確率でDDR4
* **≈ 50 GB/s (≈ 51200 MB/s)** → 基本的にDDR5

### ディスクテスト

依存プロジェクト：[https://github.com/oneclickvirt/disktest](https://github.com/oneclickvirt/disktest)

`dd`テストは誤差が大きい可能性がありますが、テスト速度が速くディスクサイズ制限がありません。`fio`テストはより現実的ですが、テスト速度が遅く、ディスクおよびメモリサイズの最低要件があります。

同時に、サーバーは異なるファイルシステムを持つ可能性があり、一部のファイルシステムのIOエンジンは同じハードウェア条件下でテストの読み書き速度がより速く、これは正常です。プロジェクトはデフォルトで`fio`でテストを行い、使用するIOエンジンの優先度は`libaio > posixaio > psync`、代替オプション`dd`テストは`fio`テストが使用不可能時に自動置換されます。

`fio`テスト結果を例に基準は以下の通りです：

| OS タイプ | 主要指標 | 次要指標 |
|----------|----------|----------|
| Windows/MacOS | 4K読み取り → 64K読み取り → 書き込みテスト | グラフィカルインターフェースシステムは読み取り性能を優先考慮 |
| Linux（グラフィカルインターフェースなし）| 4K読み取り + 4K書き込み + 1M読み書き | 読み取り/書き込み値は通常類似 |

以下のディスクタイプは通常~フル血状態の性能を指し、`libaio`をIOテストエンジンとし、`Linux`下でテストを行うことを指します：

| ドライブタイプ | 4K (IOPS) 性能 | 1M (IOPS) 性能 |
|---------------|----------------|----------------|
| NVMe SSD | ≥ 200 MB/s | 5-10 GB/s |
| 標準SSD | 50-100 MB/s | 2-3 GB/s |
| HDD（機械式ハードディスク）| 10-40 MB/s | 500-600 MB/s |
| 性能不良 | < 10 MB/s | < 200 MB/s |

迅速評価：

1. **主要チェック**: 4K読み取り (IOPS) 4K書き込み (IOPS)
    - ほぼ同じで差は小さい
    - ≥ 200 MB/s = NVMe SSD
    - 50-100 MB/s = 標準SSD
    - 10-40 MB/s = HDD（機械式ハードディスク）
    - < 10 MB/s = ゴミ性能、オーバーセリング/制限が深刻

2. **次要チェック**: 1M総計 (IOPS)
    - プロバイダーが設定したIO制限
    - リソースオーバーセリング状況
    - 数値が高いほど良い
    - NVMe SSDは通常4-6 GB/sに達する
    - 標準SSDは通常1-2 GB/sに達する

NVMe SSDの1M (IOPS)値 < 1GB/s の場合、深刻なリソースオーバーセリングが存在することを示します。

注意：ここでテストするのは真のIOで、本プロジェクト限定です。本プロジェクト以外でテストしたIOは基準の汎用性を保証しません。彼らがテスト時に同じパラメータを使用していない可能性、IO直接読み書きを設定していない可能性、IOエンジン設定が一致しない可能性、テスト時間設定が一致しない可能性があり、すべて基準の偏差を引き起こします。

### ストリーミングメディアロック解除

依存プロジェクト：[https://github.com/oneclickvirt/CommonMediaTests](https://github.com/oneclickvirt/CommonMediaTests) [https://github.com/lmc999/RegionRestrictionCheck](https://github.com/lmc999/RegionRestrictionCheck)

デフォルトでは国境を越えるストリーミングメディアのロック解除のみをチェックします。

一般的に、正常な状況下では、一つのIPの複数のストリーミングメディアのロック解除地域はすべて一致し、あちこち飛び回ることはありません。複数のプラットフォームでロック解除地域が一致しない場合、IPはIPXOなどのプラットフォームからのレンタルか、最近宣告され使用されたもので、ストリーミングメディアの一般的なデータベースに認識修正されていない可能性が高いです。各プラットフォームのIPデータベース認識速度が一致しないため、時々あるプラットフォームではロック解除地域が正常、あるプラットフォームではルート上のある位置に飛ぶ、あるプラットフォームではIPがあなたによって使用される前にいた位置に飛ぶことがあります。

| DNS タイプ | ロック解除方式判断の必要性 | DNSのロック解除への影響 | 説明 |
| --------- | ------------------------- | ---------------------- | ---- |
| 公式主流DNS | 不要 | 小 | ストリーミングメディアのロック解除は主にノードIPに依存し、DNS解析は基本的にロック解除を干渉しない |
| 非主流/自建DNS | 必要 | 大 | ストリーミングメディアのロック解除結果はDNS解析の影響を大きく受け、ネイティブロック解除かDNSロック解除かを判断する必要がある |

そのため、テスト過程で、ホストが現在使用しているのが公式主流のDNSの場合、ネイティブロック解除かどうかの判断は行われません。

### IP品質検出

依存プロジェクト：[https://github.com/oneclickvirt/securityCheck](https://github.com/oneclickvirt/securityCheck)

14のデータベースのIP関連情報を検出します。一般的に、使用タイプと会社タイプ、そしてセキュリティ情報のその他識別を見れば十分で、セキュリティスコアは本当にお遊びです。複数のプラットフォームで対応する検出項目がすべて対応する値になっている場合、現在のIPが確実にそうであることを証明します。一つのデータベースソースの情報のみを信じてはいけません。

* **使用タイプ & 会社タイプ**: IP帰属と使用シナリオを表示し、例えば家庭ユーザー、企業オフィス、ホスティングサービス、またはクラウド/データセンターに属するかどうか。
* **クラウドプロバイダー / データセンター / モバイルデバイス**: IPがクラウドサービス、データセンター、またはモバイルネットワークから来ているかを判断し、共有または高リスクIPの識別に役立つ。
* **プロキシ / VPN / Tor / Tor出口**: IPが真の身元や位置を隠すために使用されているかを検出し、匿名アクセスや悪用行為に関与している可能性がある。
* **ネットワーククローラー / ロボット**: 自動化されたアクセスまたは収集プログラムを識別し、セキュリティリスク評価に参考価値がある。
* **匿名 / 悪用者 / 脅威 / 中継 / Bogon**: IP履歴行動特徴と予約/未割り当てIPに属するかどうかを表示し、IP信頼度判断を補助。
* **セキュリティスコア、評判、信頼スコア、脅威スコア、詐欺スコア、悪用スコア**: 各データベースのIPに対する定量化されたセキュリティ評価、参考のみ。
* **コミュニティ投票 & ブラックリスト記録**: ユーザーフィードバックと公共ブラックリスト情報を展示し、潜在的リスクを迅速に識別可能。
* **Google検索実行可能性**: IPがGoogle検索サービスにアクセスする実行可能性を検出し、ネットワーク制限やブロック状況を間接的に反映。

マルチプラットフォーム比較がより信頼性が高く、異なるデータベースのアルゴリズムと更新頻度が異なるため、単一ソースは誤判断の可能性があります。複数のデータベースが類似の結果を示す場合、その結果はより信頼性が高いことを説明します。

### メールポート検出

依存プロジェクト：[https://github.com/oneclickvirt/portchecker](https://github.com/oneclickvirt/portchecker)

- **SMTP (25)**: メールサーバー間でのメール転送（メール送信）に使用。
- **SMTPS (465)**: 暗号化されたSMTPメール送信（SSL/TLS方式）に使用。
- **SMTP (587)**: クライアントからメールサーバーへのメール送信、STARTTLS暗号化をサポート。
- **POP3 (110)**: メールクライアントがサーバーからメールをダウンロードするために使用、暗号化なし。
- **POP3S (995)**: 暗号化されたPOP3、安全にメールをダウンロード（SSL/TLS方式）に使用。
- **IMAP (143)**: メールクライアントがオンラインでメール管理（メール閲覧、同期）、暗号化なし。
- **IMAPS (993)**: 暗号化されたIMAP、安全にメール管理（SSL/TLS方式）に使用。

現在のホストがメール局として機能せず、電子メールの送受信を行わない場合、この項目指標は無視して構いません。

### 近隣スピードテスト

依存プロジェクト：[https://github.com/oneclickvirt/speedtest](https://github.com/oneclickvirt/speedtest)

まず公式推奨の測定ポイントをテストし、次に代表的な国際測定ポイントをテストします。

公式測定ポイントは、テスト対象のホストマシンのローカル帯域幅ベースラインを表すことができます。

日常的には1Gbps帯域幅のサーバーを使用することを好みます。少なくとも依存関係のダウンロードなどの速度が十分に速いからです。
