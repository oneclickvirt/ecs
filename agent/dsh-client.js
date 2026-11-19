window.__ModuleLoader__.load({
  id: 'dsh-plugin-goecs',
  factory(require) {
    const React = require('react'), h = React.createElement
    const dictionaries = /* GOECS_LOCALES */ {"en":{"meta":{"title":"GoECS","description":"Test this host or an SSH host and read faithful TXT, HTML, Markdown and JSON reports."},"panel":"GoECS","subtitle":"A clear view of your machine","tests":"Tests","results":"Reports","settings":"Settings","target":"Test target","local":"This DSH host","remote":"SSH host","inspect":"Check connection","install":"Download / update engine","version":"GoECS version","latest":"Latest stable","ready":"Ready","running":"Running","completed":"Completed","failed":"Failed","cancelled":"Cancelled","stopping":"Stopping","unknown":"Unconfirmed","imported":"Imported","start":"Start test","stop":"Stop test","chooseTest":"Choose a test","originalMenu":"Original GoECS menu","concurrentHint":"Menu 2 runs all tests concurrently. Performance measurements can affect one another.","sequentialHint":"The original program controls execution and section-by-section output.","advanced":"Advanced options","default":"Engine default","language":"Report language","optionsHint":"Options come from the selected engine. Empty values keep its defaults. Presets still use the original menu; explicit values are passed to the engine. Deep stress and device probes require explicit configuration.","loading":"Loading…","noEngine":"Download the official engine, or set its absolute path in Settings.","live":"Live output","empty":"Start a test or open an existing result.","viewMode":"Switch reading mode","txt":"Report TXT","html":"Classified report","md":"Markdown","json":"Structured JSON","openPastebin":"Open in Pastebin","pastebinLoading":"Loading the shared report…","pastebinError":"The shared report could not be loaded. Retry, open it in Pastebin, or select Report TXT.","retryPastebin":"Retry","copy":"Copy","download":"Download","copied":"Copied","copyFailed":"Clipboard is unavailable. Download the report instead.","source":"Source","preview":"Preview","import":"Open TXT / JSON","history":"History","refresh":"Refresh","save":"Save","saved":"Saved","sidebar":"Show GoECS in the sidebar","sidebarHint":"When hidden, reopen it from the sidebar More menu.","admin":"Request administrator authorization before testing","adminHint":"Only the test helper and engine request elevation; DSH stays under your account. Turn this off to run the complete suite without elevation. No prompt appears in CLI/CI.","network":"Network route","ask":"Ask when a proxy is detected","direct":"Direct / current OS route","proxy":"Use detected proxy","proxyHint":"HTTP/SOCKS does not cover every native TCP, ICMP or route probe. VPN/TUN routes remain controlled by the OS.","binary":"Local GoECS executable","binaryHint":"On first open, reuse a local executable or verified cache; otherwise download and fill its path automatically. You can also select an existing executable.","timeout":"Plugin task deadline (seconds)","name":"Name","host":"Hostname / IP","port":"SSH port","username":"Username","keyFile":"Private key file on the DSH host","password":"Password (memory only)","passphrase":"Key passphrase (memory only)","fingerprint":"SSH host fingerprint","fingerprintHint":"Verify this fingerprint through a trusted channel. A changed key is never accepted automatically.","remoteBinary":"Remote executable path (optional)","addHost":"Save SSH host","removeHost":"Remove host","trust":"Verify SSH identity","trustAccept":"I verified this fingerprint","continue":"Continue complete test","cancel":"Cancel","review":"Before starting","adminNotice":"The operating system will ask permission to run only the local test helper and GoECS as root / Administrator. DSH does not need to restart. You can also run the complete suite with the current account.","proxyNotice":"A proxy is configured on the selected host. Choose the route you intend to test.","elapsed":"Elapsed","seconds":"s","restore":"Open GoECS","privacy":"Reports are saved in the plugin data directory. Public upload and share links are enabled by default; turn upload off or enable private mode to disable them.","hostSide":"Web tests run on the connected DSH host, not inside the browser.","noPermission":"Current account","administrator":"Administrator","noTasks":"No saved reports yet","nativeStatus":"Native report status","error":"Operation failed","installing":"Downloading and verifying…","close":"Close","chooseProxy":"Detected proxy","headless":"CLI / CI uses the same backend without loading this page. Select the network route explicitly when a proxy is present.","extraProxy":"Explicit proxy URL (optional)","durationHint":"The plugin deadline controls task cleanup. The engine's timeout remains an independent advanced option.","customHint":"Custom tests use the original CLI parameters. Preset modes use the original interactive menu.","more":"More","noOutput":"Waiting for GoECS output…","searchOptions":"Filter options","engineInfo":"Engine and target","back":"Back to tests","preparing":"Preparing engine and options…","retry":"Retry engine preparation","publicUpload":"Public upload and share links","share":"Share links","enabled":"On","disabled":"Off","group_general":"General and reports","group_hardware":"CPU, memory and disk","group_deep":"Deep tests","group_unlock":"Unlock and IP quality","group_network":"Routes and latency","group_speed":"Speed tests","optionLabels":{"analyze":"Summary analysis","backtrace":"Upstream and backhaul","basic":"System information","cpu":"CPU test","cpu-method":"CPU method","cpu-thread":"CPU threads","data-offline":"Embedded data sources","deep":"Deep tests","deep-burn-duration":"CPU burn duration","deep-disk-paths":"Deep disk directories","deep-gpu-device":"GPU device","deep-smart-devices":"SMART devices","disk":"Disk test","disk-method":"Disk method","diskmc":"Multiple disks","diskp":"Disk directory","dns-mode":"DNS mode","email":"Email ports","hardware-budget":"Hardware deadline","log":"Engine log file","memory":"Memory test","memory-method":"Memory method","nt3":"Detailed routes","nt3-location":"Route location","nt3-type":"Route IP family","ping":"Latency tests","ping-scope":"Latency scope","ping-sort":"Latency sorting","privacy":"Private mode (disables public upload)","security":"IP quality","speed":"Speed tests","spnum":"Speed nodes per operator","tcp":"TCP handshake probes","tcp-format":"TCP output format","tcp-sort":"TCP platform sorting","tgdc":"Telegram DC latency","timeout":"Engine deadline","upload":"Public upload and share links","ut":"Media unlock","ut-concurrency":"Unlock concurrency","ut-dns":"Unlock DNS servers","ut-interface":"Unlock interface / IP","utipver":"Unlock IP family","utregion":"Unlock region","utshowip":"Show unlock IP (compatibility)","web":"Website latency"},"regions":{"0":"Global","1":"Global + Taiwan","2":"Global + Hong Kong","3":"Global + Japan","4":"Global + Korea","5":"Global + North America","6":"Global + South America","7":"Global + Europe","8":"Global + Africa","9":"Global + Oceania","10":"Taiwan only","11":"Hong Kong only","12":"Japan only","13":"Korea only","14":"North America only","15":"South America only","16":"Europe only","17":"Africa only","18":"Oceania only","19":"Sports only","20":"All","21":"AI only","22":"Southeast Asia only"},"adminUnavailable":"This target cannot show a local authorization dialog. The complete selected suite can run with its current account; individual probes may require root / Administrator.","authorize":"Authorize and start test","authorizing":"Waiting for operating-system authorization and network preparation…","firewallNotice":"Before testing, allow this GoECS executable through Windows Firewall on the currently connected network profiles. The rule applies only to this engine, is removed after the task, and does not override existing block rules.","firewallCurrent":"Windows Firewall can still ask permission when a probe first listens for connections. Allow GoECS in that system dialog. Without administrator authorization the plugin cannot prepare the firewall rule in advance.","historyHint":"Reports stay in this plugin data directory. Add a remark, open a report, or remove one or several records.","selectAll":"Select all","clearSelection":"Clear selection","deleteSelected":"Delete selected","remark":"Remark","saveRemark":"Save remark","open":"Open report","delete":"Delete","deleted":"Deleted","untitled":"Untitled report","noShare":"No public share link","administratorStart":"Administrator / root","administratorFirewallStart":"Administrator / root + firewall","currentStart":"Continue with current account"},"zh":{"meta":{"title":"Go 融合怪","description":"测试本机或 SSH 主机，阅读、复制和导出原文、分类报告、Markdown 与 JSON。"},"panel":"Go 融合怪","subtitle":"看清每一项测试结果","tests":"开始测试","results":"测试报告","settings":"插件设置","target":"测试主机","local":"运行 DSH 的本机","remote":"SSH 远程主机","inspect":"检测连接","install":"下载／更新本体","version":"融合怪版本","latest":"最新正式版","ready":"准备就绪","running":"测试中","completed":"已完成","failed":"失败","cancelled":"已取消","stopping":"正在停止","unknown":"状态待确认","imported":"已导入","start":"开始测试","stop":"停止测试","chooseTest":"选择测试","originalMenu":"融合怪原有菜单","concurrentHint":"选项 2 保持所有测试并发；性能测试可能互相影响。","sequentialHint":"由原程序控制执行，保留逐项输出的行为。","advanced":"高级参数","default":"使用本体默认值","language":"报告语言","optionsHint":"参数读取自当前本体，留空沿用默认值。预设仍由原菜单执行，显式填写的参数传给本体；深度压力、设备检测仅在主动配置后运行。","loading":"正在读取…","noEngine":"先下载官方本体，或在设置中填写已有程序的绝对路径。","live":"实时输出","empty":"开始测试，或打开已有结果。","viewMode":"切换阅读模式","txt":"报告 TXT","html":"分类报告","md":"美化排版","json":"结构化 JSON","openPastebin":"在 Pastebin 打开","pastebinLoading":"正在加载分享报告…","pastebinError":"暂时无法加载分享报告。可重试、在 Pastebin 打开或切换到报告 TXT。","retryPastebin":"重试","copy":"复制","download":"下载","copied":"已复制","copyFailed":"剪贴板不可用，请下载报告。","source":"源码","preview":"预览","import":"打开 TXT／JSON","history":"历史记录","refresh":"刷新","save":"保存","saved":"已保存","sidebar":"在左侧栏显示融合怪入口","sidebarHint":"隐藏后可从左侧栏的“更多”重新打开。","admin":"测试前请求管理员授权","adminHint":"仅为测试助手和本体请求提权，DSH 保持普通用户运行。关闭后以当前账号提交完整测试；CLI／CI 不弹出提权窗口。","network":"测试线路","ask":"检测到代理时询问","direct":"直连／当前系统路由","proxy":"使用检测到的代理","proxyHint":"HTTP／SOCKS 无法覆盖所有原生 TCP、ICMP 和路由探测；VPN／TUN 线路由系统控制。","binary":"本机融合怪程序路径","binaryHint":"首次打开自动复用本机程序或已校验缓存；未找到时自动下载并填写路径，也可指定已有程序。","timeout":"插件任务最长时间（秒）","name":"主机备注","host":"主机名／IP","port":"SSH 端口","username":"用户名","keyFile":"DSH 主机上的私钥文件路径","password":"密码（仅本次运行保存在内存）","passphrase":"私钥口令（仅保存在内存）","fingerprint":"SSH 主机指纹","fingerprintHint":"请通过可信渠道核对指纹。主机密钥变化后需要重新核对。","remoteBinary":"远程融合怪路径（可选）","addHost":"保存远程主机","removeHost":"移除主机","trust":"核对 SSH 主机身份","trustAccept":"已核对，使用此指纹","continue":"以当前权限继续完整测试","cancel":"取消","review":"开始测试前","adminNotice":"系统将弹窗请求仅以 root／管理员身份运行本机测试助手和融合怪，无需重启 DSH。也可选择当前账号继续完整测试。","proxyNotice":"当前测试主机配置了代理，请选择本次想检测的线路。","elapsed":"已用时","seconds":"秒","restore":"打开 Go 融合怪","privacy":"报告保存在插件独立的数据目录；默认公开上传并生成分享链接，可关闭上传或启用隐私模式。","hostSide":"Web 版检测的是连接的 DSH 主机，浏览器仅负责操作与展示。","noPermission":"普通权限","administrator":"管理员权限","noTasks":"还没有保存的报告","nativeStatus":"本体报告状态","error":"操作失败","installing":"正在下载并校验…","close":"关闭","chooseProxy":"选择检测到的代理","headless":"CLI／CI 使用相同后端，无需加载本页；存在代理时应显式指定测试线路。","extraProxy":"指定代理地址（可选）","durationHint":"插件时限负责终止任务和清理；本体的 timeout 仍可在高级参数中单独配置。","customHint":"自定义使用原有命令行参数；预设测试通过原有交互菜单执行。","more":"更多","noOutput":"正在等待融合怪输出…","searchOptions":"筛选参数","engineInfo":"本体与测试环境","back":"返回测试","preparing":"正在准备本体并读取参数…","retry":"重试准备本体","publicUpload":"公开上传并生成分享链接","share":"分享链接","enabled":"开启","disabled":"关闭","group_general":"常规与报告","group_hardware":"CPU、内存与磁盘","group_deep":"深度测试","group_unlock":"解锁与 IP 质量","group_network":"路由与延迟","group_speed":"网络测速","optionLabels":{"analyze":"测试摘要分析","backtrace":"上游及回程线路","basic":"系统基础信息","cpu":"CPU 测试","cpu-method":"CPU 测试方法","cpu-thread":"CPU 线程模式","data-offline":"使用内置数据源","deep":"深度测试","deep-burn-duration":"CPU 压力测试时长","deep-disk-paths":"深度测试磁盘目录","deep-gpu-device":"GPU 设备","deep-smart-devices":"SMART 检测设备","disk":"磁盘测试","disk-method":"磁盘测试方法","diskmc":"多磁盘检测","diskp":"磁盘测试目录","dns-mode":"DNS 模式","email":"邮件端口检测","hardware-budget":"硬件测试时限","log":"保存本体日志","memory":"内存测试","memory-method":"内存测试方法","nt3":"三网详细路由","nt3-location":"三网路由地区","nt3-type":"三网路由 IP 类型","ping":"延迟测试","ping-scope":"延迟测试范围","ping-sort":"延迟结果排序","privacy":"隐私模式（关闭公开上传）","security":"IP 质量检测","speed":"网络测速","spnum":"每个运营商测速节点数","tcp":"TCP 握手探测","tcp-format":"TCP 输出格式","tcp-sort":"TCP 平台排序","tgdc":"Telegram DC 延迟","timeout":"本体测试总时限","upload":"公开上传并生成分享链接","ut":"跨国平台解锁","ut-concurrency":"解锁测试并发数","ut-dns":"解锁测试 DNS","ut-interface":"解锁测试网卡或 IP","utipver":"解锁测试 IP 类型","utregion":"解锁测试区域","utshowip":"显示解锁 IP（兼容参数）","web":"常见网站延迟"},"regions":{"0":"全球","1":"全球及台湾","2":"全球及香港","3":"全球及日本","4":"全球及韩国","5":"全球及北美","6":"全球及南美","7":"全球及欧洲","8":"全球及非洲","9":"全球及大洋洲","10":"仅台湾","11":"仅香港","12":"仅日本","13":"仅韩国","14":"仅北美","15":"仅南美","16":"仅欧洲","17":"仅非洲","18":"仅大洋洲","19":"仅体育","20":"全部","21":"仅 AI","22":"仅东南亚"},"adminUnavailable":"此目标无法显示本机授权弹窗，可用当前账号继续完整所选测试；个别项目可能需要 root／管理员权限。","authorize":"授权并开始测试","authorizing":"正在等待系统授权并准备网络权限…","firewallNotice":"正式测试前，为此融合怪本体设置 Windows 防火墙放行，仅限当前连接的网络配置。规则仅适用于此程序，任务结束后删除，不覆盖已有阻止规则。","firewallCurrent":"Windows 防火墙仍可能在项目首次监听连接时弹窗，请在系统窗口允许融合怪联网。未获管理员授权时，插件无法提前设置防火墙规则。","historyHint":"记录保存在插件独立数据目录，可修改备注、打开报告，或删除单条和多条记录。","selectAll":"全选","clearSelection":"清除选择","deleteSelected":"删除所选","remark":"备注","saveRemark":"保存备注","open":"打开报告","delete":"删除","deleted":"已删除","untitled":"未命名报告","noShare":"没有公开分享链接","administratorStart":"使用 root／管理员权限开始测试","administratorFirewallStart":"使用 root／管理员权限并启用防火墙","currentStart":"普通权限继续测试"}}
    const key = 'oneclickvirt.ecs.reader.v1'
    // DSH's locale registry addresses flat keys, including dotted key names.
    const flattenLocale = (value, prefix = '') => Object.fromEntries(Object.entries(value).flatMap(([name, entry]) => typeof entry === 'string' ? [[prefix + name, entry]] : Object.entries(flattenLocale(entry, prefix + name + '.'))))
    const clientLocales = Object.fromEntries(Object.entries(dictionaries).map(([language, value]) => [language, flattenLocale(value)]))
    const saved = { tab: 'tests', target: 'local', choice: '1', language: 'zh', taskId: '', options: {}, version: 'latest' }
    try { const value = JSON.parse(localStorage.getItem(key) || '{}'); for (const name of Object.keys(saved)) if (typeof value[name] === typeof saved[name]) saved[name] = value[name] } catch {}
    const persist = () => { try { localStorage.setItem(key, JSON.stringify(saved)) } catch {} }
    let updateSidebar = () => {}, disposed = false
    async function api(input, signal) {
      const response = await fetch('/api/oneclickvirt-ecs', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(input), signal })
      const result = await response.json()
      if (!response.ok) { const error = new Error(result.error?.message || 'GoECS request failed'); error.code = result.error?.code; error.details = result.error?.details; throw error }
      return result.data
    }
    const css = `.ge{height:100%;overflow:auto;color:var(--dsw-alias-label-primary,#1e293b);background:var(--dsw-alias-bg-base,#f8fafc);font:14px/1.6 system-ui,sans-serif;padding:calc(28px + var(--dsh-frame-top-clearance,0px)) 30px 36px;box-sizing:border-box}.ge *{box-sizing:border-box}.ge-inner{max-width:1180px;margin:auto}.ge-head{display:flex;justify-content:space-between;gap:16px;align-items:center;margin-bottom:25px}.ge-brand{display:flex;gap:13px;align-items:center}.ge-mark{width:44px;height:44px;background:#0f766e;color:#f0fdfa;border-radius:14px;display:grid;place-items:center}.ge h1{font-size:25px;letter-spacing:-.6px;line-height:1.3;margin:0}.ge small,.ge-muted{color:var(--dsw-alias-label-secondary,#64748b)}.ge-nav{display:flex;gap:4px;padding:4px;border-radius:12px;border:1px solid var(--dsw-alias-border-l2,#e2e8f0);width:max-content;margin-bottom:24px}.ge button,.ge input,.ge select,.ge textarea{font:inherit;color:inherit;border:1px solid var(--dsw-alias-border-l2,#dbe3eb);border-radius:9px;background:var(--dsw-alias-bg-layer-1,#fff);padding:9px 12px;max-width:100%}.ge button{cursor:pointer;line-height:1.4}.ge button:hover{background:var(--dsw-alias-interactive-bg-hover,#edf5f5)}.ge button:disabled{opacity:.45;cursor:default}.ge button.ge-primary,.ge button.ge-selected{background:#0f766e;color:white;border-color:#0f766e}.ge button.ge-danger{color:#dc2626}.ge :focus-visible{outline:2px solid #14b8a6;outline-offset:3px}.ge-grid{display:grid;grid-template-columns:minmax(0,1.1fr) minmax(0,.9fr);gap:18px}.ge-card{background:var(--dsw-alias-bg-layer-1,#fff);border:1px solid var(--dsw-alias-border-l2,#e2e8f0);border-radius:16px;padding:21px;margin-bottom:18px}.ge-card h2{font-size:16px;margin:0 0 17px}.ge label{display:flex;flex-direction:column;gap:6px;font-size:13px}.ge label input,.ge label select{width:100%}.ge-row{display:flex;align-items:center;gap:10px;flex-wrap:wrap}.ge-row.between{justify-content:space-between}.ge-stack{display:flex;flex-direction:column;gap:14px}.ge-actions{display:flex;gap:10px;flex-wrap:wrap;margin-top:18px}.ge-choices{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:9px}.ge-choice{text-align:left;min-height:56px;display:flex;gap:10px;align-items:center}.ge-choice strong{font-size:12px;border-radius:7px;min-width:25px;padding:2px 5px;background:#0f766e15;text-align:center}.ge-choice.ge-selected strong{background:#ffffff25}.ge-help{font-size:12px;line-height:1.7;margin:9px 0 0;color:var(--dsw-alias-label-secondary,#64748b)}.ge-status{font-size:12px;padding:5px 10px;border-radius:20px;background:#0f766e18;color:#0f766e}.ge-error,.ge-warning{padding:12px 15px;margin:12px 0;border-radius:10px;background:#ef444410;color:var(--dsw-alias-state-error-primary,#b91c1c);white-space:pre-wrap;overflow-wrap:anywhere}.ge-warning{background:#d9770610;color:var(--dsw-alias-label-secondary,#92400e)}.ge-terminal{white-space:pre-wrap;overflow:auto;overflow-wrap:anywhere;max-height:440px;min-height:140px;background:#111827;color:#d1fae5;padding:18px;border-radius:12px;font:12px/1.8 ui-monospace,monospace}.ge-log-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:12px}.ge-report{margin-top:17px;overflow:auto}.ge-report section{margin:22px 0}.ge-report h2{font-size:18px;border-bottom:1px solid var(--dsw-alias-border-l2,#e2e8f0);padding-bottom:12px}.ge-report dl>div{display:flex;gap:20px;padding:8px 0;border-bottom:1px solid var(--dsw-alias-border-l2,#edf2f7)}.ge-report dt{min-width:150px;color:var(--dsw-alias-label-secondary,#64748b)}.ge-report dd{margin:0;overflow-wrap:anywhere}.ge-report pre{white-space:pre-wrap;overflow-wrap:anywhere;padding:16px;border-radius:10px;background:var(--dsw-alias-bg-base,#f8fafc);font:12px/1.8 ui-monospace,monospace}.ge-report table{border-collapse:collapse;width:100%;font-size:13px}.ge-report th,.ge-report td{padding:9px;border:1px solid var(--dsw-alias-border-l2,#e2e8f0);text-align:left;white-space:pre-wrap}.ge-table{overflow:auto}.ge-modal-backdrop{position:fixed;inset:0;background:#0f172a66;z-index:200;display:grid;place-items:center;padding:24px}.ge-modal{background:var(--dsw-alias-bg-layer-1,#fff);border:1px solid var(--dsw-alias-border-l2,#e2e8f0);color:var(--dsw-alias-label-primary,#1e293b);max-width:600px;width:100%;max-height:85vh;overflow:auto;border-radius:18px;padding:26px;box-shadow:0 24px 90px #0003}.ge-modal h2{margin-top:0}.ge code{font:12px/1.6 ui-monospace,monospace;overflow-wrap:anywhere}.ge-option-group{border:0;padding:18px 0 0;margin:12px 0;border-top:1px solid var(--dsw-alias-border-l2,#e2e8f0)}.ge-option-group legend{font-weight:600;padding-right:12px}.ge a{color:#0f766e;overflow-wrap:anywhere}.ge-flags{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px;margin-top:16px}.ge-toggle{flex-direction:row!important;justify-content:space-between;align-items:center;gap:18px!important;padding:10px 0}.ge-toggle input{width:auto!important;accent-color:#0f766e}.ge details summary{cursor:pointer;font-weight:600}.ge-menu-item{display:block;width:100%;padding:8px 12px;border:0;background:transparent;color:inherit;cursor:pointer}.ge-history{display:flex;flex-direction:column;gap:8px}.ge-history-item{border:1px solid var(--dsw-alias-border-l2,#e2e8f0);border-radius:12px;padding:14px;display:flex;flex-direction:column;gap:10px}.ge-history-select{flex-direction:row!important;align-items:center;gap:8px!important}.ge-history-select input{width:auto!important}.ge-history-item> .ge-row:last-child input{min-width:180px;flex:1}.ge-history button{text-align:left}.ge-toast{margin:12px 0;color:#0f766e}.ge-screenreader{position:absolute;width:1px;height:1px;overflow:hidden;clip:rect(0,0,0,0)}@media(max-width:850px){.ge{padding:calc(18px + var(--dsh-frame-top-clearance,0px)) 16px 24px}.ge-grid{grid-template-columns:1fr}.ge-head{align-items:flex-start}.ge-flags{grid-template-columns:1fr}.ge-report dl>div{display:block}.ge-report dd{margin-top:4px}}@media(prefers-reduced-motion:no-preference){.ge button{transition:background .15s,border-color .15s}}`
    function Icon() { return h('svg', { viewBox: '0 0 24 24', width: 20, height: 20, fill: 'none', stroke: 'currentColor', strokeWidth: 1.7, 'aria-hidden': true }, h('path', { d: 'M2 14h5l3-9 5 15 3-9h4' })) }
    let seat = null
    const observers = new Set()
    const notifySeat = value => { seat = value; for (const notify of observers) notify() }
    const subscribeSeat = notify => { observers.add(notify); return () => observers.delete(notify) }
    function Seat() {
      const node = React.useRef(null)
      React.useLayoutEffect(() => {
        const measure = () => {
          const rect = node.current?.getBoundingClientRect()
          if (rect) notifySeat({ top: rect.top, left: rect.left, width: rect.width, height: rect.height })
        }
        const resize = new ResizeObserver(measure)
        resize.observe(node.current); window.addEventListener('resize', measure); measure()
        return () => { resize.disconnect(); window.removeEventListener('resize', measure); notifySeat(null) }
      }, [])
      return h('div', { ref: node, style: { height: '100%', minHeight: 0 } })
    }
    function Keeper({ t }) {
      const rect = React.useSyncExternalStore(subscribeSeat, () => seat)
      const [started, setStarted] = React.useState(Boolean(rect))
      React.useEffect(() => { if (rect) setStarted(true) }, [Boolean(rect)])
      if (!started && !rect) return null
      return h('div', { 'data-goecs-keeper': true, hidden: !rect, style: { position: 'fixed', ...(rect || {}), display: rect ? 'block' : 'none', zIndex: 0, pointerEvents: rect ? 'auto' : 'none' } }, h(Panel, { t }))
    }
    function Report({ document, markdown = false }) {
      if (!document) return null
      return h('div', { className: 'ge-report' }, document.sections.map((section, i) => h('section', { key: i }, section.title && h('h2', null, section.title), section.blocks.map((block, j) => {
        if (block.kind === 'fields') return markdown ? h('ul', { key: j }, block.fields.map((field, n) => h('li', { key: n }, h('strong', null, field.name + ': '), field.value))) : h('dl', { key: j }, block.fields.map((field, n) => h('div', { key: n }, h('dt', null, field.name), h('dd', null, field.value))))
        if (block.kind === 'table') return h('div', { className: 'ge-table', key: j }, h('table', null, h('thead', null, h('tr', null, block.columns.map((cell, n) => h('th', { key: n }, cell)))), h('tbody', null, block.rows.map((row, n) => h('tr', { key: n }, row.values.map((cell, z) => h('td', { key: z }, cell)))))))
        return h('pre', { key: j }, block.lines.join('\n'))
      }))))
    }
    const pastebinOrigin = 'https://paste.spiritlhl.net'
    function PastebinReport({ url, language, t }) {
      const frame = React.useRef(null)
      const [attempt, setAttempt] = React.useState(0)
      const [state, setState] = React.useState('loading')
      const code = url.split('/').at(-1)
      React.useEffect(() => {
        const timer = setTimeout(() => setState('error'), 20000)
        const receive = event => {
          if (event.origin !== pastebinOrigin || event.source !== frame.current?.contentWindow) return
          const { type, id, state } = event.data || {}
          if (type !== 'goecs.pastebin.report' || id !== code || !['ready', 'error'].includes(state)) return
          clearTimeout(timer)
          setState(state)
        }
        window.addEventListener('message', receive)
        return () => { clearTimeout(timer); window.removeEventListener('message', receive) }
      }, [code, attempt, language])
      const retry = () => { setState('loading'); setAttempt(previous => previous + 1) }
      return h('div', { className: 'ge-pastebin' },
        state === 'loading' && h('p', { className: 'ge-help', role: 'status' }, t('pastebinLoading')),
        state === 'error' && h('div', { className: 'ge-row', role: 'status' },
          h('p', { className: 'ge-help' }, t('pastebinError')),
          h('button', { type: 'button', onClick: retry }, t('retryPastebin'))),
        h('iframe', {
          key: attempt, ref: frame, hidden: state === 'error', title: t('html'),
          src: url + '?embed=goecs&lang=' + (language === 'en' ? 'en' : 'zh'),
          sandbox: 'allow-scripts allow-same-origin allow-downloads',
          allow: 'clipboard-write https://paste.spiritlhl.net', referrerPolicy: 'no-referrer',
          style: { width: '100%', height: 'min(72vh, 850px)', minHeight: '420px', border: 0 },
        }))
    }
    function Panel({ t }) {
      const [tab, setTab] = React.useState(saved.tab), [status, setStatus] = React.useState(null), [target, setTarget] = React.useState(saved.target), [choice, setChoice] = React.useState(saved.choice), [language, setLanguage] = React.useState(saved.language)
      const [options, setOptions] = React.useState(saved.options), [catalog, setCatalog] = React.useState([]), [search, setSearch] = React.useState(''), [version, setVersion] = React.useState(saved.version)
      const [preparing, setPreparing] = React.useState(true), [catalogError, setCatalogError] = React.useState(''), [catalogRetry, setCatalogRetry] = React.useState(0), [advancedOpen, setAdvancedOpen] = React.useState(false)
      const [settings, setSettings] = React.useState(null), [remote, setRemote] = React.useState({ name: '', host: '', port: 22, username: '', keyFile: '', password: '', passphrase: '', hostKey: '', binaryPath: '' })
      const [facts, setFacts] = React.useState(null), [task, setTask] = React.useState(null), [taskId, setTaskId] = React.useState(saved.taskId), [raw, setRaw] = React.useState(''), [selectedMode, setMode] = React.useState(''), [result, setResult] = React.useState(null), [source, setSource] = React.useState(false)
      const [busy, setBusy] = React.useState(false), [error, setError] = React.useState(''), [toast, setToast] = React.useState(''), [decision, setDecision] = React.useState(null), [route, setRoute] = React.useState('direct'), [proxyIndex, setProxyIndex] = React.useState(0), [privilege, setPrivilege] = React.useState('current'), [trust, setTrust] = React.useState(null), [history, setHistory] = React.useState([]), [historyDrafts, setHistoryDrafts] = React.useState({}), [historySelected, setHistorySelected] = React.useState([]), [now, setNow] = React.useState(Date.now())
      const lifetime = React.useRef(null), offset = React.useRef(0), importing = React.useRef(null), terminal = React.useRef(null)
      const active = ['running', 'stopping'].includes(task?.status)
      const ready = task?.id === taskId && task.reportReady && !active
      const shareURL = ready ? task.pastebinShareURL : ''
      const mode = selectedMode && (selectedMode !== 'html' || shareURL) ? selectedMode : shareURL ? 'html' : 'txt'
      const format = mode === 'md' && !source ? 'json' : mode
      const currentResult = result?.id === taskId && result.format === format ? result : null
      const refresh = async signal => { const data = await api({ action: 'status' }, signal); setStatus(data); setSettings(data.settings); updateSidebar(data.settings.showSidebar); return data }
      const execute = async work => { setBusy(true); setError(''); setToast(''); try { return await work() } catch (e) { if (e.code === 'HOST_TRUST_REQUIRED' || e.code === 'HOST_KEY_CHANGED') setTrust({ target, fingerprint: e.details?.fingerprint, changed: e.code === 'HOST_KEY_CHANGED' }); setError(e.message) } finally { setBusy(false) } }
      React.useEffect(() => {
        const controller = new AbortController(); lifetime.current = controller
        refresh(controller.signal).then(data => {
          if (!controller.signal.aborted && !saved.taskId) {
            const running = data.tasks.find(item => ['running', 'stopping'].includes(item.status))
            if (running) { setTaskId(running.id); setTab('results') }
          }
        }).catch(e => { if (!controller.signal.aborted) { setError(e.message); setPreparing(false) } })
        const tick = setInterval(() => setNow(Date.now()), 1000)
        return () => { controller.abort(); clearInterval(tick) }
      }, [])
      React.useEffect(() => { Object.assign(saved, { tab, target, choice, language, taskId, options, version }); persist() }, [tab, target, choice, language, taskId, options, version])
      React.useEffect(() => {
        if (!status) return
        const controller = new AbortController(); setCatalog([]); setFacts(null); setCatalogError(''); setPreparing(true)
        const load = async () => {
          const value = await api({ action: 'prepare', targetId: target }, controller.signal)
          if (controller.signal.aborted) return
          setCatalog(value.flags)
          if (value.binaryPath) {
            if (target === 'local') {
              if (status.settings.binaryPath !== value.binaryPath) setStatus(previous => ({ ...previous, settings: { ...previous.settings, binaryPath: value.binaryPath } }))
              setSettings(previous => ({ ...previous, binaryPath: value.binaryPath }))
            } else setStatus(previous => ({ ...previous, targets: previous.targets.map(item => item.id === target ? { ...item, binaryPath: value.binaryPath, engineTag: value.tag || item.engineTag } : item) }))
          }
        }
        load().catch(e => {
          if (controller.signal.aborted) return
          setCatalogError(e.message)
          if (e.code === 'HOST_TRUST_REQUIRED' || e.code === 'HOST_KEY_CHANGED') setTrust({ target, fingerprint: e.details?.fingerprint, changed: e.code === 'HOST_KEY_CHANGED' })
        }).finally(() => { if (!controller.signal.aborted) setPreparing(false) })
        return () => controller.abort()
      }, [target, Boolean(status), status?.settings.binaryPath, status?.targets.find(item => item.id === target)?.binaryPath, catalogRetry])
      React.useEffect(() => {
        if (!taskId) return
        const controller = new AbortController(); let running = false
        offset.current = 0; setRaw('')
        let restored = false
        const update = async () => {
          if (running) return
          running = true
          try {
            const value = await api({ action: 'task', id: taskId, offset: offset.current, live: true }, controller.signal)
            if (controller.signal.aborted) return
            if (value.nextOffset !== offset.current) setRaw(value.output)
            offset.current = value.nextOffset
            setTask(value)
            if (value.reportReady && !['running', 'stopping'].includes(value.status)) clearInterval(timer)
          } catch (e) {
            if (!controller.signal.aborted && e.code === 'TASK_NOT_FOUND' && !restored) {
              restored = true
              try { await api({ action: 'loadHistory', id: taskId }, controller.signal) } catch (failure) { if (!controller.signal.aborted) { setTaskId(''); setTask(null); setError(failure.message) } }
            } else if (!controller.signal.aborted) setError(e.message)
          } finally { running = false }
        }
        const timer = setInterval(update, 1000); void update()
        return () => { controller.abort(); clearInterval(timer) }
      }, [taskId])
      React.useEffect(() => { if (terminal.current && active) terminal.current.scrollTop = terminal.current.scrollHeight }, [raw])
      React.useEffect(() => {
        if (!ready || tab !== 'results' || mode === 'html') return
        const controller = new AbortController()
        api({ action: 'report', id: taskId, format }, controller.signal).then(value => {
          if (!controller.signal.aborted) setResult({ ...value, format, id: taskId })
        }).catch(e => { if (!controller.signal.aborted) setError(e.message) })
        return () => controller.abort()
      }, [taskId, ready, mode, format, tab])
      React.useEffect(() => { if (tab === 'results' || tab === 'history') api({ action: 'history' }).then(value => { setHistory(value); setHistoryDrafts(previous => Object.fromEntries(value.map(item => [item.id, previous[item.id] ?? item.remark ?? '']))); setHistorySelected(previous => previous.filter(id => value.some(item => item.id === id))) }).catch(() => {}) }, [tab, task?.status])
      const selectTask = value => {
        if (value.id !== taskId) { setRaw(''); setMode(''); setResult(null); setSource(false) }
        setTask(value); setTaskId(value.id); setTab('results')
      }
      const changeOption = (name, value) => setOptions(previous => {
        const next = { ...previous }
        if (value === '') delete next[name]
        else next[name] = value
        // Upload and privacy are inverse controls. Keeping the public-upload
        // switch usable makes a saved privacy choice reversible without
        // forcing the operator to find a hidden advanced field first.
        if (name === 'privacy' && value === true) delete next.upload
        if (name === 'upload' && value === true) delete next.privacy
        return next
      })
      const start = () => execute(async () => {
        const value = await api({ action: 'start', test: { targetId: target, choice, language, options } })
        if (value.decision) { setDecision(value.decision); setRoute(value.decision.needsProxy ? '' : (status.settings.networkMode === 'proxy' ? 'proxy' : 'direct')); setProxyIndex(0); setPrivilege(value.decision.canElevate || value.decision.needsFirewall ? 'elevate' : 'current') }
        else selectTask(value.task)
      })
      const continueTest = (privilegeChoice = privilege) => execute(async () => { const value = await api({ action: 'continue', token: decision.token, decisions: { networkMode: route, proxyIndex, privilege: privilegeChoice, fromUI: true } }); setDecision(null); selectTask(value.task) })
      const savePreferences = () => execute(async () => { const value = await api({ action: 'settings', settings }); updateSidebar(value.showSidebar); await refresh(); setToast(t('saved')) })
      const install = () => execute(async () => { await api({ action: 'install', targetId: target, version }); await refresh(); const value = await api({ action: 'options', targetId: target }); setCatalog(value.flags); setToast(t('saved')) })
      const inspect = () => execute(async () => { setFacts(await api({ action: 'inspect', targetId: target })); setToast(t('ready')) })
      const saveRemote = () => execute(async () => { const value = await api({ action: 'saveTarget', target: remote }); setTarget(value.id); setRemote({ ...value, password: '', passphrase: '' }); await refresh(); setToast(t('saved')) })
      const acceptTrust = () => execute(async () => {
        const value = status.targets.find(item => item.id === trust.target)
        if (!value || value.id === 'local') return
        await api({ action: 'saveTarget', target: { ...value, hostKey: trust.fingerprint } }); setTrust(null); await refresh(); setToast(t('saved'))
      })
      const copy = () => execute(async () => {
        const value = await api({ action: 'report', id: taskId, format: mode })
        if (navigator.clipboard?.writeText) await navigator.clipboard.writeText(value.content)
        else throw new Error(t('copyFailed'))
        setToast(t('copied'))
      })
      const download = () => execute(async () => { const value = await api({ action: 'report', id: taskId, format: mode }); const url = URL.createObjectURL(new Blob([value.content], { type: ({ json: 'application/json' })[mode] || 'text/plain;charset=utf-8' })); const link = document.createElement('a'); link.href = url; link.download = 'goecs-' + taskId + '.' + mode; link.click(); setTimeout(() => URL.revokeObjectURL(url), 1000) })
      const importFile = event => execute(async () => {
        const file = event.target.files?.[0]; if (!file) return
        if (file.size > 16 * 1024 * 1024) throw new Error('Report exceeds 16 MiB')
        const text = await file.text()
        selectTask(await api({ action: 'import', text })); event.target.value = ''
      })
      const openHistory = id => execute(async () => { selectTask(await api({ action: 'loadHistory', id })) })
      const saveHistoryRemark = id => execute(async () => { const remark = historyDrafts[id] ?? ''; await api({ action: 'updateHistory', id, remark }); setHistory(previous => previous.map(item => item.id === id ? { ...item, remark } : item)); setToast(t('saved')) })
      const deleteHistory = ids => execute(async () => { await api({ action: 'deleteHistory', ids }); setHistory(previous => previous.filter(item => !ids.includes(item.id))); setHistorySelected(previous => previous.filter(id => !ids.includes(id))); if (taskId && ids.includes(taskId)) { setTaskId(''); setTask(null); setRaw(''); setMode(''); setResult(null) } setToast(t('deleted')) })
      const button = (label, onClick, props = {}) => h('button', { type: 'button', onClick, disabled: busy, ...props }, label)
      const field = (label, value, change, props = {}) => h('label', null, label, h('input', { value: value ?? '', onChange: e => change(e.target.value), ...props }))
      const targetSelect = h('label', null, t('target'), h('select', { 'data-goecs-target': true, value: target, onChange: e => { setTarget(e.currentTarget.value); setOptions({}) } }, (status?.targets || [{ id: 'local', name: t('local') }]).map(item => h('option', { key: item.id, value: item.id }, item.id === 'local' ? t('local') : item.name))))
      function Advanced() {
        const available = catalog.filter(item => !item.managed && !item.aliasOf && (!search || (item.name + item.description + (dictionaries[language]?.optionLabels?.[item.name] || '')).toLowerCase().includes(search.toLowerCase())))
        const change = changeOption
        const control = item => {
          const label = dictionaries.en.optionLabels[item.name] ? t('optionLabels.' + item.name) : item.name
          const props = { 'aria-label': label + ' (-' + item.name + ')', 'data-goecs-option': item.name, value: options[item.name] === undefined ? '' : String(options[item.name]), onChange: e => change(item.name, item.type === 'bool' && e.target.value !== '' ? e.target.value === 'true' : e.target.value) }
          const initial = t('default') + (item.defaultValue !== '' ? ' · ' + item.defaultValue : '')
          const select = item.type === 'bool' || item.choices
          return h('label', { key: item.name }, h('span', null, label, h('code', null, ' -' + item.name)), select ? h('select', props, h('option', { value: '' }, initial), ...(item.type === 'bool' ? ['true', 'false'] : item.choices).map(value => h('option', { key: value, value }, item.type === 'bool' ? t(value === 'true' ? 'enabled' : 'disabled') : item.name === 'utregion' ? value + ' · ' + t('regions.' + value) : value))) : h('input', { ...props, placeholder: initial, type: item.type === 'int' ? 'number' : 'text' }), h('small', null, item.description))
        }
        return h('details', { className: 'ge-card', open: advancedOpen, onToggle: e => setAdvancedOpen(e.currentTarget.open) }, h('summary', null, t('advanced') + (available.length ? ' · ' + available.length : '')), h('p', { className: 'ge-help' }, t('optionsHint')),
          preparing && h('p', { role: 'status' }, t('preparing')),
          catalogError && h('div', { className: 'ge-error', role: 'alert' }, catalogError, h('div', { className: 'ge-actions' }, button(t('retry'), () => setCatalogRetry(value => value + 1)))),
          !!catalog.length && h(React.Fragment, null, h('input', { type: 'search', 'aria-label': t('searchOptions'), placeholder: t('searchOptions'), value: search, onChange: e => setSearch(e.target.value) }),
            ...['general', 'hardware', 'deep', 'unlock', 'network', 'speed'].map(group => {
              const items = available.filter(item => item.group === group)
              return items.length ? h('fieldset', { key: group, className: 'ge-option-group' }, h('legend', null, t('group_' + group)), h('div', { className: 'ge-flags' }, items.map(control))) : null
            })))
      }
      function HistoryPage() {
        const allSelected = history.length > 0 && historySelected.length === history.length
        const rows = history.map(item => {
          const links = item.shareURLs?.length ? h('div', { className: 'ge-row' }, ...item.shareURLs.map(url => h('a', { key: url, href: url, target: '_blank', rel: 'noreferrer noopener' }, url))) : h('small', { className: 'ge-muted' }, t('noShare'))
          return h('article', { className: 'ge-history-item', key: item.id },
            h('div', { className: 'ge-row between' }, h('label', { className: 'ge-history-select' }, h('input', { type: 'checkbox', checked: historySelected.includes(item.id), disabled: busy, onChange: e => setHistorySelected(previous => e.target.checked ? [...new Set([...previous, item.id])] : previous.filter(id => id !== item.id)) }), h('strong', null, item.remark || item.target || t('untitled'))), h('small', null, (item.startedAt || '') + ' · ' + (item.target || '') + ' · ' + (item.choice || '') + ' · ' + t(item.status || 'completed'))),
            links,
            h('div', { className: 'ge-row' }, h('input', { 'aria-label': t('remark'), maxLength: 128, value: historyDrafts[item.id] ?? '', onChange: e => setHistoryDrafts(previous => ({ ...previous, [item.id]: e.target.value })) }), button(t('saveRemark'), () => saveHistoryRemark(item.id)), button(t('open'), () => openHistory(item.id)), button(t('delete'), () => deleteHistory([item.id]), { className: 'ge-danger' })))
        })
        return h('section', { className: 'ge-card ge-stack' },
          h('div', { className: 'ge-row between' }, h('h2', null, t('history')), h('div', { className: 'ge-row' }, button(t('import'), () => importing.current?.click()), button(allSelected ? t('clearSelection') : t('selectAll'), () => setHistorySelected(allSelected ? [] : history.map(item => item.id))), button(t('deleteSelected'), () => deleteHistory(historySelected), { className: 'ge-danger', disabled: !historySelected.length }))),
          h('p', { className: 'ge-help' }, t('historyHint')),
          h('div', { className: 'ge-history' }, rows.length ? rows : h('small', null, t('noTasks')))
        )
      }
      const selectedRemote = status?.targets.find(item => item.id === target)
      return h('div', { className: 'ge', 'data-goecs-panel': true }, h('style', null, css), h('div', { className: 'ge-inner' },
        h('header', { className: 'ge-head' }, h('div', { className: 'ge-brand' }, h('div', { className: 'ge-mark' }, h(Icon)), h('div', null, h('h1', null, t('panel')), h('small', null, t('subtitle')))), h('div', { className: 'ge-row' }, h('span', { className: 'ge-status' }, task ? t(task.status) : preparing ? t('preparing') : catalogError ? t('failed') : t('ready')), h('small', null, status?.version))),
        h('nav', { className: 'ge-nav', 'aria-label': t('panel') }, ['tests','results','history','settings'].map(name => button(t(name), () => setTab(name), { key: name, className: tab === name ? 'ge-selected' : '', 'aria-current': tab === name ? 'page' : undefined }))),
        error && h('div', { className: 'ge-error', role: 'alert' }, error), toast && h('div', { className: 'ge-toast', role: 'status' }, toast), status?.warning && h('div', { className: 'ge-warning' }, status.warning), h('input', { ref: importing, className: 'ge-screenreader', type: 'file', accept: '.txt,.log,.md,.json', onChange: importFile, tabIndex: -1 }),
        tab === 'tests' && h(React.Fragment, null, h('div', { className: 'ge-grid' }, h('section', { className: 'ge-card' }, h('h2', null, t('chooseTest')), h('div', { className: 'ge-choices' }, (status?.menus || []).map(item => button(h(React.Fragment, null, h('strong', null, item[0] === 'custom' ? '+' : item[0]), language === 'en' ? item[2] : item[1]), () => setChoice(item[0]), { key: item[0], className: 'ge-choice' + (choice === item[0] ? ' ge-selected' : ''), 'aria-pressed': choice === item[0] }))), h('p', { className: 'ge-help' }, t(choice === '2' ? 'concurrentHint' : choice === 'custom' ? 'customHint' : 'sequentialHint'))),
          h('div', null, h('section', { className: 'ge-card ge-stack' }, h('h2', null, t('engineInfo')), targetSelect, h('p', { className: 'ge-help' }, t('hostSide')), h('div', { className: 'ge-row' }, button(t('inspect'), inspect), button(t('settings'), () => setTab('settings'))), facts && h('div', { className: 'ge-help' }, facts.platform + ' / ' + facts.arch + ' · ' + t(facts.admin ? 'administrator' : 'noPermission')), h('label', null, t('language'), h('select', { value: language, onChange: e => setLanguage(e.target.value) }, h('option', { value: 'zh' }, '中文'), h('option', { value: 'en' }, 'English'))), field(t('version'), version, setVersion, { placeholder: status?.version ? 'latest / v' + status.version : 'latest' }), button(preparing ? t('preparing') : busy ? t('loading') : t('install'), install, { disabled: busy || preparing || active }), (settings?.binaryPath && target === 'local' || selectedRemote?.binaryPath && target !== 'local') && h('code', { 'data-goecs-binary': true }, target === 'local' ? settings.binaryPath : selectedRemote.binaryPath), !catalog.length && !preparing && h('p', { className: 'ge-help' }, t('noEngine'))), h('section', { className: 'ge-card' }, button(t('start'), start, { className: 'ge-primary', disabled: busy || preparing || !catalog.length || active }), h('label', { className: 'ge-toggle' }, h('span', null, t('publicUpload')), h('input', { type: 'checkbox', checked: options.upload !== false && options.privacy !== true, onChange: e => changeOption('upload', e.currentTarget.checked) })), h('p', { className: 'ge-help' }, t('privacy'))))), Advanced()),
        tab === 'results' && h(React.Fragment, null,
          h('section', { className: 'ge-card' }, h('div', { className: 'ge-log-header' }, h('h2', null, t('live')), h('div', { className: 'ge-row' }, task?.startedAt && h('small', null, t('elapsed') + ' ' + Math.max(0, Math.floor(((task.finishedAt ? Date.parse(task.finishedAt) : now) - Date.parse(task.startedAt)) / 1000)) + ' ' + t('seconds')), active && button(t('stop'), () => execute(async () => setTask(await api({ action: 'stop', id: taskId }))), { className: 'ge-danger' }))),
            h('pre', { className: 'ge-terminal', ref: terminal, 'aria-live': 'off' }, raw || t('noOutput')), task?.phase === 'permissions' && h('p', { className: 'ge-help', role: 'status' }, t('authorizing')), task?.nativeStatus && h('p', { className: 'ge-help' }, t('nativeStatus') + ': ' + task.nativeStatus), task?.warnings?.map((warning, i) => h('p', { className: 'ge-help', key: i }, warning)), task?.error && h('div', { className: 'ge-error' }, task.error.message)),
          ready && h('section', { className: 'ge-card', 'data-goecs-final-report': true },
            h('div', { className: 'ge-row between' },
              h('label', null, t('viewMode'), h('select', {
                value: mode, onChange: e => setMode(e.target.value),
              }, ['txt', 'html', 'md', 'json'].map(name => h('option', {
                key: name, value: name, disabled: name === 'html' && !shareURL,
              }, t(name))))),
              h('div', { className: 'ge-row' },
                shareURL && h('a', { href: shareURL, target: '_blank', rel: 'noreferrer noopener' }, t('openPastebin')),
                mode === 'md' && button(t(source ? 'preview' : 'source'), () => setSource(!source)),
                mode !== 'html' && button(t('copy'), copy, { disabled: busy || !currentResult }),
                mode !== 'html' && button(t('download'), download, { disabled: busy || !currentResult }))),
            mode === 'html'
              ? h(PastebinReport, { key: taskId + ':' + shareURL + ':' + task.language, url: shareURL, language: task.language || language, t })
              : mode === 'md' && !source
                ? h(Report, { document: currentResult?.document, markdown: true })
                : h('div', { className: 'ge-report' }, h('pre', null, currentResult?.content))),
          !taskId && h('p', { className: 'ge-help' }, t('empty')),
          null),
        tab === 'history' && HistoryPage(),
        tab === 'settings' && settings && h('div', { className: 'ge-grid' }, h('section', { className: 'ge-card ge-stack' }, h('h2', null, t('settings')),
          h('label', { className: 'ge-toggle' }, h('span', null, t('sidebar')), h('input', { type: 'checkbox', checked: settings.showSidebar, onChange: e => setSettings({ ...settings, showSidebar: e.target.checked }) })), h('p', { className: 'ge-help' }, t('sidebarHint')),
          h('label', { className: 'ge-toggle' }, h('span', null, t('admin')), h('input', { type: 'checkbox', checked: settings.requestAdmin, onChange: e => setSettings({ ...settings, requestAdmin: e.target.checked }) })), h('p', { className: 'ge-help' }, t('adminHint')),
          h('label', null, t('network'), h('select', { value: settings.networkMode, onChange: e => setSettings({ ...settings, networkMode: e.target.value }) }, ['ask','direct','proxy'].map(value => h('option', { key: value, value }, t(value))))), h('p', { className: 'ge-help' }, t('proxyHint')), field(t('binary'), settings.binaryPath, value => setSettings({ ...settings, binaryPath: value })), h('p', { className: 'ge-help' }, t('binaryHint')), field(t('timeout'), settings.maxTaskSeconds, value => setSettings({ ...settings, maxTaskSeconds: Number(value) }), { type: 'number', min: 10, max: 86400 }), h('p', { className: 'ge-help' }, t('durationHint')), button(t('save'), savePreferences, { className: 'ge-primary' }), h('p', { className: 'ge-help' }, t('headless'))),
          h('section', { className: 'ge-card ge-stack' }, h('h2', null, t('remote')), targetSelect, target !== 'local' && h('div', { className: 'ge-row' }, button(t('refresh'), () => setRemote({ ...selectedRemote, password: '', passphrase: '' })), button(t('removeHost'), () => execute(async () => { await api({ action: 'removeTarget', targetId: target }); setTarget('local'); await refresh() }), { className: 'ge-danger' })),
            ...['name','host','port','username','keyFile','password','passphrase','hostKey','binaryPath'].map(name => field(t(({ hostKey: 'fingerprint', binaryPath: 'remoteBinary' })[name] || name), remote[name], value => setRemote({ ...remote, [name]: value }), { key: name, type: name === 'password' || name === 'passphrase' ? 'password' : name === 'port' ? 'number' : 'text', autoComplete: 'off' })), h('p', { className: 'ge-help' }, t('fingerprintHint')), button(t('addHost'), saveRemote, { className: 'ge-primary' }))),
        decision && h('div', { className: 'ge-modal-backdrop' }, h('section', { className: 'ge-modal ge-stack', role: 'dialog', 'aria-modal': true, 'aria-label': t('review') }, h('h2', null, t('review')), decision.needsAdmin && h('p', null, t(decision.canElevate ? 'adminNotice' : 'adminUnavailable')), decision.needsProxy && h(React.Fragment, null, h('p', null, t('proxyNotice')), h('div', { className: 'ge-row' }, button(t('direct'), () => setRoute('direct'), { className: route === 'direct' ? 'ge-selected' : '' }), button(t('proxy'), () => setRoute('proxy'), { className: route === 'proxy' ? 'ge-selected' : '' }))), route === 'proxy' && h('label', null, t('chooseProxy'), h('select', { value: proxyIndex, onChange: e => setProxyIndex(Number(e.target.value)) }, decision.facts.proxies.map((proxy, i) => h('option', { key: i, value: i }, proxy.source + ' · ' + proxy.url)))), decision.needsFirewall && h('p', { className: 'ge-help' }, t(decision.canElevate ? 'firewallNotice' : 'firewallCurrent')), h('p', { className: 'ge-help' }, t('proxyHint')), h('div', { className: 'ge-actions' }, decision.canElevate && button(t(decision.needsFirewall ? 'administratorFirewallStart' : 'administratorStart'), () => continueTest('elevate'), { className: 'ge-primary', disabled: busy || !route }), button(t('currentStart'), () => continueTest('current'), { disabled: busy || !route }), button(t('cancel'), () => setDecision(null))))),
        trust && h('div', { className: 'ge-modal-backdrop' }, h('section', { className: 'ge-modal', role: 'dialog', 'aria-modal': true, 'aria-label': t('trust') }, h('h2', null, t('trust')), h('p', null, t('fingerprintHint')), h('code', null, trust.fingerprint), h('div', { className: 'ge-actions' }, button(t('trustAccept'), acceptTrust, { className: 'ge-primary' }), button(t('cancel'), () => setTrust(null)))))))
    }
    return { inject: ['slots', 'locale', 'layout'], apply(ctx) {
      disposed = false
      ctx.effect(() => ctx.locale.register('oneclickvirt-ecs', clientLocales))
      const t = ctx.locale.bind('oneclickvirt-ecs')
      let sidebarDispose = null, visible = null
      updateSidebar = show => {
        if (disposed || visible === show) return
        visible = show
        if (typeof sidebarDispose === 'function') sidebarDispose()
        sidebarDispose = show ? ctx.slots.inject('sidebar.panellist', () => ctx.slots.register({ name: 'sidebar.panellist', id: 'oneclickvirt-ecs', order: 25, label: () => t('panel'), locale: 'oneclickvirt-ecs' }, Icon)) : null
      }
      ctx.slots.inject('main', () => ctx.slots.register({ name: 'main', key: 'oneclickvirt-ecs', locale: 'oneclickvirt-ecs' }, Seat))
      ctx.slots.inject('shell.overlay', () => ctx.slots.register({ name: 'shell.overlay', id: 'oneclickvirt-ecs-keeper', locale: 'oneclickvirt-ecs' }, Keeper))
      ctx.slots.inject('sidebar.panellist.actions', () => ctx.slots.register({ name: 'sidebar.panellist.actions', id: 'oneclickvirt-ecs-open', order: 80 }, ({ onDismiss }) => h('button', { type: 'button', className: 'ge-menu-item', onClick: () => { ctx.layout.selectPanel('oneclickvirt-ecs'); onDismiss?.() } }, t('restore'))))
      updateSidebar(true)
      const controller = new AbortController()
      api({ action: 'status' }, controller.signal).then(value => updateSidebar(value.settings.showSidebar)).catch(() => {})
      ctx.effect(() => () => { disposed = true; controller.abort(); updateSidebar = () => {}; if (typeof sidebarDispose === 'function') sidebarDispose() })
    } }
  },
})
