import os from 'node:os'
import path from 'node:path'
import { mkdir, readFile, writeFile, rename, readdir } from 'node:fs/promises'
import { randomUUID } from 'node:crypto'
import { localFacts, capture, GoECSError, processEnvironment, safeProxy, redactSecrets, proxyURL } from './dsh-platform.mjs'
import { installEngine, verifyExecutable } from './dsh-engine.mjs'
import { parseHelp, buildArguments, MENU } from './dsh-options.mjs'
import { startLocalProcess } from './dsh-process.mjs'
import { connectSSH, remoteFacts, remotePath, remoteMkdir, uploadEngine, sshCapture, quotePOSIX, powershell, startRemoteProcess, readRemoteFile, validateSSH } from './dsh-ssh.mjs'
import { parseReport, exportReport, stripTerminal } from './dsh-report.mjs'

const defaults = { showSidebar: true, requestAdmin: true, networkMode: 'ask', binaryPath: '', maxTaskSeconds: 3600 }
const publicTarget = t => ({ id: t.id, name: t.name, ...validateSSH(t), binaryPath: t.binaryPath || '', engineTag: t.engineTag || '' })
const publicError = error => ({ code: error.code || 'OPERATION_FAILED', message: redactSecrets(error.message || 'Operation failed'), ...(error.details ? { details: error.details } : {}) })
const ps = value => "'" + String(value).replace(/'/g, "''") + "'"
export class GoECSService {
  constructor(config = {}) {
    const profile = (process.env.DSH_PROFILE || process.env.DSH_USER_AGENT_SUFFIX?.match(/dsh-profile\/([\w.-]+)/)?.[1] || 'standalone').replace(/[^\w.-]/g, '_')
    this.directory = config.stateDir || process.env.GOECS_DSH_STATE_DIR || path.join(process.env.DSH_HOME || path.join(os.homedir(), '.dsh'), 'storages', 'oneclickvirt-ecs', profile)
    this.config = config; this.settings = { ...defaults }; this.targets = new Map(); this.secrets = new Map(); this.tasks = new Map(); this.pending = new Map(); this.busy = new Set(); this.revision = 0; this.closed = false; this.writes = Promise.resolve()
    this.ready = this.init()
  }
  async init() {
    await mkdir(this.directory, { recursive: true, mode: 0o700 })
    try {
      const data = JSON.parse(await readFile(path.join(this.directory, 'settings.json'), 'utf8'))
      this.settings = this.validateSettings(data.settings || {})
      for (const t of data.targets || []) if (/^[\w-]{1,64}$/.test(t.id) && t.id !== 'local') this.targets.set(t.id, publicTarget(t))
    } catch (error) { if (error.code !== 'ENOENT') this.loadWarning = 'Saved preferences could not be read; default preferences are active. The existing file has been preserved.' }
  }
  validateSettings(input) {
    const value = { ...this.settings, ...input }
    if (typeof value.showSidebar !== 'boolean' || typeof value.requestAdmin !== 'boolean' || !['ask', 'direct', 'proxy'].includes(value.networkMode) || typeof value.binaryPath !== 'string' || value.binaryPath.length > 4096 || !Number.isInteger(value.maxTaskSeconds) || value.maxTaskSeconds < 10 || value.maxTaskSeconds > 86400) throw new GoECSError('INVALID_SETTINGS', 'Invalid sidebar, privilege, network or timeout preference')
    return Object.fromEntries(Object.keys(defaults).map(key => [key, value[key]]))
  }
  async persist() {
    if (this.loadWarning) throw new GoECSError('SETTINGS_UNREADABLE', this.loadWarning)
    const text = JSON.stringify({ settings: this.settings, targets: [...this.targets.values()].map(publicTarget) }, null, 2)
    const work = async () => {
      const destination = path.join(this.directory, 'settings.json'), temporary = destination + '.' + randomUUID() + '.tmp'
      await writeFile(temporary, text, { mode: 0o600, flag: 'wx' }); await rename(temporary, destination)
    }
    this.writes = this.writes.catch(() => {}).then(work); await this.writes
  }
  async status(owner) {
    await this.ready
    return { version: '0.2.21', settings: this.settings, warning: this.loadWarning || '', menus: MENU, targets: [{ id: 'local', name: 'Local / 本机' }, ...this.targets.values()].map(t => t.id === 'local' ? t : publicTarget(t)), tasks: [...this.tasks.values()].filter(t => owner === undefined || t.owner === owner).map(t => this.view(t)), uiOptional: true, host: this.host || { standalone: true, uiOptional: true } }
  }
  async saveSettings(input) {
    await this.ready; const previous = this.settings; this.settings = this.validateSettings(input)
    try { await this.persist(); this.revision++; return this.settings } catch (error) { this.settings = previous; throw error }
  }
  async saveTarget(input) {
    await this.ready
    const id = input.id || randomUUID()
    if (!/^[\w-]{1,64}$/.test(id) || id === 'local' || typeof input.name !== 'string' || !input.name.trim() || input.name.length > 128) throw new GoECSError('INVALID_TARGET', 'Enter a target name')
    if (this.busy.has(id)) throw new GoECSError('TARGET_BUSY', 'Wait until the target task or installation finishes before changing it')
    const value = publicTarget({ ...input, id, name: input.name.trim() })
    const previous = this.targets.get(id)
    this.targets.set(id, value)
    try { await this.persist() } catch (error) { if (previous) this.targets.set(id, previous); else this.targets.delete(id); throw error }
    if (typeof input.password === 'string' || typeof input.passphrase === 'string') this.secrets.set(id, { password: input.password || '', passphrase: input.passphrase || '' })
    this.revision++; return value
  }
  async removeTarget(id) {
    await this.ready
    if (this.busy.has(id)) throw new GoECSError('TARGET_BUSY', 'Stop the target task first')
    const previous = this.targets.get(id); this.targets.delete(id)
    try { await this.persist() } catch (error) { if (previous) this.targets.set(id, previous); throw error }
    this.secrets.delete(id); this.revision++; return { removed: true }
  }
  target(id = 'local') {
    if (id === 'local') return { id, name: 'Local / 本机', binaryPath: this.settings.binaryPath }
    const value = this.targets.get(id)
    if (!value) throw new GoECSError('TARGET_NOT_FOUND', 'Select a saved target')
    return value
  }
  async environment(targetId = 'local', signal) {
    await this.ready; const target = this.target(targetId)
    if (target.id === 'local') return { target, facts: await localFacts(), client: null }
    const client = await connectSSH(target, this.secrets.get(target.id), signal)
    try { return { target, client, facts: await remoteFacts(client) } } catch (error) { client.end(); throw error }
  }
  async inspect(targetId, signal) {
    const value = await this.environment(targetId, signal)
    try { return { target: value.target.id, ...value.facts, proxies: value.facts.proxies.map(safeProxy), binaryPath: value.target.binaryPath } } finally { value.client?.end() }
  }
  async install(targetId = 'local', version = 'latest', signal) {
    await this.ready
    if (this.busy.has(targetId)) throw new GoECSError('TARGET_BUSY', 'Wait until the active task finishes before upgrading')
    this.busy.add(targetId)
    let environment
    try {
      environment = await this.environment(targetId, signal)
      const { facts, target, client } = environment
      const engine = await installEngine(path.join(this.directory, 'engines'), facts, version, signal)
      if (!client) await this.saveSettings({ binaryPath: engine.binaryPath })
      else {
        const directory = remotePath(facts.home, '.cache', 'goecs-dsh', 'engines', engine.tag, facts.platform + '-' + facts.arch)
        const binaryPath = remotePath(directory, facts.platform === 'win32' ? 'goecs.exe' : 'goecs')
        await remoteMkdir(client, directory, facts.platform); await uploadEngine(client, engine.binaryPath, binaryPath, facts.platform, engine.binarySHA)
        this.targets.set(targetId, { ...target, binaryPath, engineTag: engine.tag }); await this.persist(); this.revision++
      }
      return { ...engine, binaryPath: this.target(targetId).binaryPath, target: targetId }
    } finally { environment?.client?.end(); this.busy.delete(targetId) }
  }
  async inspectEngine(environment) {
    const { target, client, facts } = environment
    if (!target.binaryPath) throw new GoECSError('INSTALL_REQUIRED', 'Download and verify GoECS, or configure its absolute executable path')
    let descriptor, result
    if (!client) {
      descriptor = await verifyExecutable(target.binaryPath)
      result = await capture([target.binaryPath, '-h'], { timeout: 15000 })
    } else {
      if (!(facts.platform === 'win32' ? /^[A-Za-z]:[\\/]/ : /^\//).test(target.binaryPath) || /[\x00-\x1f]/.test(target.binaryPath)) throw new GoECSError('INVALID_BINARY', 'Select an absolute remote executable path')
      descriptor = { binaryPath: target.binaryPath, tag: target.engineTag || 'configured' }
      result = await sshCapture(client, facts.platform === 'win32' ? powershell('& ' + ps(target.binaryPath) + ' -h;exit $LASTEXITCODE') : quotePOSIX(target.binaryPath) + ' -h', { timeout: 15000 })
    }
    if (result.code !== 0) throw new GoECSError('ENGINE_HELP_FAILED', 'Configured GoECS did not return its option list')
    const flags = parseHelp(result.stdout + '\n' + result.stderr)
    if (!flags.some(f => f.name === 'json') || !flags.some(f => f.name === 'menu')) throw new GoECSError('ENGINE_INCOMPATIBLE', 'This executable does not expose the required GoECS menu and JSON-file interfaces')
    return { ...descriptor, flags }
  }
  async options(targetId = 'local') {
    const environment = await this.environment(targetId)
    try { return await this.inspectEngine(environment) } finally { environment.client?.end() }
  }
  async preflight(input, { interactive = false, signal } = {}) {
    await this.ready
    const revision = this.revision
    const targetId = input.targetId || 'local', environment = await this.environment(targetId, signal)
    try {
      const engine = await this.inspectEngine(environment), { facts } = environment
      buildArguments(input, engine.flags, 'report.json')
      const mode = input.networkMode || this.settings.networkMode
      if (!['ask', 'direct', 'proxy'].includes(mode)) throw new GoECSError('INVALID_NETWORK', 'Choose direct or proxy')
      const needsProxy = mode === 'ask' && facts.proxies.length > 0
      const needsAdmin = interactive && this.settings.requestAdmin && !facts.admin
      const warnings = []
      if (!facts.admin) warnings.push('Current user is not root/Administrator. Start DSH as an administrator for permission-dependent probes, or continue the complete suite with the current account. Individual failures remain visible.')
      if (mode === 'proxy' || needsProxy) warnings.push(facts.proxyNotice || 'HTTP/SOCKS does not proxy every native TCP/ICMP/route probe. Results retain the target operating-system route for those protocols.')
      if (needsProxy && !interactive) throw new GoECSError('PROXY_CHOICE_REQUIRED', 'A proxy was detected. In CLI/CI explicitly choose --network direct or --network proxy before starting a test')
      if (mode === 'proxy' && !facts.proxies.length && !input.proxyURL) throw new GoECSError('PROXY_UNAVAILABLE', 'No proxy was detected; configure an explicit proxy first')
      if (this.closed) throw new GoECSError('PLUGIN_STOPPING', 'Plugin is stopping')
      if (revision !== this.revision) throw new GoECSError('PLAN_EXPIRED', 'Settings or target changed. Review the test again')
      const token = randomUUID(), expires = Date.now() + 5 * 60 * 1000
      this.pending.set(token, { input: { ...input, networkMode: mode }, revision: this.revision, expires })
      for (const [key, value] of this.pending) if (value.expires < Date.now()) this.pending.delete(key)
      return { token, needsProxy, needsAdmin, warnings, facts: { ...facts, proxies: facts.proxies.map(safeProxy) }, engine: { ...engine, flags: undefined }, expires }
    } finally { environment.client?.end() }
  }
  async start(input, { interactive = false, signal, owner = '' } = {}) {
    if (this.closed) throw new GoECSError('PLUGIN_STOPPING', 'Plugin is stopping')
    const plan = await this.preflight(input, { interactive, signal })
    if (interactive && (plan.needsAdmin || plan.needsProxy)) return { decision: plan }
    return this.accept(plan.token, { networkMode: input.networkMode || (plan.facts.proxies.length ? this.settings.networkMode : 'direct') }, { owner, user: false })
  }
  async accept(token, decisions = {}, { owner = '', user = false } = {}) {
    if (this.closed) throw new GoECSError('PLUGIN_STOPPING', 'Plugin is stopping')
    const pending = this.pending.get(token)
    if (!pending || pending.expires < Date.now() || pending.revision !== this.revision) throw new GoECSError('PLAN_EXPIRED', 'Settings or target changed. Review the test again')
    if (user !== true && decisions.fromUI) throw new GoECSError('USER_ONLY', 'The permission and route decision belongs to the user')
    const input = { ...pending.input }
    const mode = decisions.networkMode || input.networkMode
    if (!['direct', 'proxy'].includes(mode)) throw new GoECSError('PROXY_CHOICE_REQUIRED', 'Select direct or proxy')
    input.networkMode = mode
    if (Number.isInteger(decisions.proxyIndex) && decisions.proxyIndex >= 0) input.proxyIndex = decisions.proxyIndex
    const targetId = input.targetId || 'local'
    if (this.busy.has(targetId)) throw new GoECSError('TARGET_BUSY', 'Only one benchmark or installation can run on this target at a time')
    this.pending.delete(token); this.busy.add(targetId)
    const task = { id: randomUUID(), targetId, choice: input.choice || '1', language: input.language || 'zh', networkMode: mode, status: 'running', startedAt: new Date().toISOString(), raw: '', owner, controller: new AbortController(), error: null }
    this.tasks.set(task.id, task)
    task.promise = this.runTask(task, input).catch(error => {
      task.status = ['SSH_DISCONNECTED', 'REMOTE_STOP_UNCONFIRMED', 'PROCESS_STOP_UNCONFIRMED'].includes(error.code) ? 'unknown' : task.stopReason || 'failed'
      if (task.status !== 'cancelled') task.error = publicError(error)
    }).finally(async () => {
      task.finishedAt = new Date().toISOString(); this.busy.delete(targetId)
      task.report ||= parseReport(stripTerminal(task.raw), this.metadata(task))
      task.report.metadata = { ...task.report.metadata, ...this.metadata(task) }
      const filename = path.join(this.directory, 'report-' + task.id + '.json')
      await writeFile(filename, JSON.stringify(task.report), { mode: 0o600 }).catch(() => { task.saveWarning = 'Report remains available in memory; saving it failed' })
    })
    return { task: this.view(task) }
  }
  metadata(task) { return { target: task.targetId, choice: task.choice, language: task.language, networkMode: task.networkMode, admin: task.facts?.admin || false, status: task.status, startedAt: task.startedAt, finishedAt: task.finishedAt || '', engine: task.engine || null, warnings: task.warnings || [] } }
  async runTask(task, input) {
    const environment = await this.environment(task.targetId, task.controller.signal)
    const { client, target, facts } = environment
    task.facts = { ...facts, proxies: facts.proxies.map(safeProxy) }
    const secret = this.secrets.get(task.targetId) || {}, secrets = [secret.password, secret.passphrase]
    try {
      const engine = await this.inspectEngine(environment); task.engine = { tag: engine.tag, binarySHA: engine.binarySHA || '', platform: facts.platform, arch: facts.arch }
      const directory = client ? remotePath(facts.home, '.cache', 'goecs-dsh', 'tasks', task.id) : path.join(this.directory, 'task-' + task.id)
      if (client) await remoteMkdir(client, directory, facts.platform); else await mkdir(directory, { mode: 0o700 })
      const filename = client ? remotePath(directory, 'native-report.json') : path.join(directory, 'native-report.json')
      const built = buildArguments(input, engine.flags, filename)
      const proxy = input.proxyURL ? { url: proxyURL(input.proxyURL), source: 'explicit' } : facts.proxies[input.proxyIndex || 0]
      if (input.networkMode === 'proxy' && !proxy) throw new GoECSError('PROXY_UNAVAILABLE', 'The selected proxy is no longer available')
      const env = processEnvironment(client ? {} : process.env, input.networkMode, proxy)
      if (input.networkMode === 'proxy') {
        const protocol = new URL(proxy.url).protocol
        const key = protocol.startsWith('socks') ? 'ut-socks-proxy' : 'ut-http-proxy'
        if (engine.flags.some(f => f.name === key)) built.args.push('-' + key + '=' + proxy.url)
        task.warnings = ['Proxy applies to supported HTTP/SOCKS requests. Native TCP/ICMP/routes can use the OS route; a VPN/TUN can affect that route.']
      }
      if (!facts.admin) (task.warnings ||= []).push('Complete suite is running without root/Administrator; permission-dependent results can be unavailable or fail individually.')
      const onData = value => {
        const data = redactSecrets(value, secrets)
        if (Buffer.byteLength(task.raw) + Buffer.byteLength(data) > 16 * 1024 * 1024) { task.error = { code: 'OUTPUT_LIMIT', message: 'Terminal output exceeded 16 MiB; task was stopped' }; task.stopReason = 'failed'; void task.process?.stop().catch(() => {}); return }
        task.raw += data
      }
      const options = { cwd: directory, env, choice: built.choice, onData, deadlineMs: this.settings.maxTaskSeconds * 1000, onDeadline: () => { task.stopReason = 'failed'; task.error = { code: 'TASK_TIMEOUT', message: 'Plugin task deadline reached; available results were preserved' } }, platform: facts.platform }
      task.process = client ? await startRemoteProcess(client, target.binaryPath, built.args, options) : await startLocalProcess(target.binaryPath, built.args, options)
      if (task.controller.signal.aborted) await task.process.stop()
      const outcome = await task.process.done
      task.status = task.stopReason || (outcome.exitCode === 0 ? 'completed' : 'failed')
      task.exitCode = outcome.exitCode
      let native
      try { native = JSON.parse(client ? await readRemoteFile(client, filename) : await readFile(filename, 'utf8')) } catch {}
      const reportText = typeof native?.text === 'string' && native.text ? redactSecrets(native.text, secrets) : stripTerminal(task.raw)
      task.report = parseReport(reportText, this.metadata(task))
      if (native) {
        const clean = value => typeof value === 'string' ? redactSecrets(value, secrets) : Array.isArray(value) ? value.map(clean) : value && typeof value === 'object' ? Object.fromEntries(Object.entries(value).map(([k, v]) => [k, clean(v)])) : value
        task.report.nativeReport = clean(native)
        task.nativeStatus = typeof native.status === 'string' ? native.status : ''
      }
      if (!native) task.report.warnings.push('The native JSON report is unavailable. Available terminal output is preserved; no successful measurements are inferred.')
      if (task.status === 'failed' && !task.error) task.error = { code: 'ENGINE_EXIT', message: 'GoECS exited with code ' + outcome.exitCode }
    } finally { client?.end() }
  }
  view(task, offset = 0) {
    return { id: task.id, targetId: task.targetId, choice: task.choice, status: task.status, nativeStatus: task.nativeStatus || task.report?.nativeReport?.status || '', startedAt: task.startedAt, finishedAt: task.finishedAt || '', networkMode: task.networkMode, admin: task.facts?.admin, output: task.raw.slice(offset, offset + 65536), nextOffset: Math.min(task.raw.length, offset + 65536), length: task.raw.length, error: task.error, warnings: task.warnings || [], saveWarning: task.saveWarning || '' }
  }
  getTask(id, owner) {
    const task = this.tasks.get(id)
    if (!task || (owner !== undefined && task.owner !== owner)) throw new GoECSError('TASK_NOT_FOUND', 'Task is unavailable in this session')
    return task
  }
  async stop(id, owner) {
    const task = this.getTask(id, owner)
    if (!['running', 'stopping'].includes(task.status)) return this.view(task)
    task.status = 'stopping'; task.stopReason = 'cancelled'; task.controller.abort()
    if (task.process) await task.process.stop()
    await task.promise; return this.view(task)
  }
  async report(id, format = 'json', owner) {
    const task = this.getTask(id, owner)
    const document = task.report || parseReport(stripTerminal(task.raw), this.metadata(task))
    return { format, content: exportReport(document, format), document: format === 'json' ? document : undefined }
  }
  async importReport(text) {
    if (typeof text !== 'string' || Buffer.byteLength(text) > 16 * 1024 * 1024) throw new GoECSError('REPORT_LIMIT', 'Report must be text of at most 16 MiB')
    let native
    if (text.trimStart().startsWith('{')) {
      try { native = JSON.parse(text) } catch { /* Preserve ordinary text that starts with a brace. */ }
      if (native) {
        const original = native.rawText ?? native.text
        if (typeof original !== 'string') throw new GoECSError('INVALID_REPORT', 'JSON report has no original text')
        text = original
      }
    }
    const importedAt = new Date().toISOString()
    const task = { id: randomUUID(), targetId: 'import', choice: 'import', status: 'imported', startedAt: importedAt, finishedAt: importedAt, raw: text, report: parseReport(text), owner: '' }
    if (native?.schema_version) task.report.nativeReport = native
    else if (native?.nativeReport) task.report.nativeReport = native.nativeReport
    task.report.metadata = this.metadata(task)
    this.tasks.set(task.id, task); return this.view(task)
  }
  async history() {
    await this.ready
    const files = (await readdir(this.directory)).filter(name => /^report-[\w-]+\.json$/.test(name))
    const output = []
    for (const filename of files) { try { const doc = JSON.parse(await readFile(path.join(this.directory, filename), 'utf8')); output.push({ id: filename.slice(7, -5), ...doc.metadata }) } catch {} }
    return output.sort((a, b) => String(b.finishedAt || b.startedAt).localeCompare(String(a.finishedAt || a.startedAt))).slice(0, 30)
  }
  async loadHistory(id) {
    if (!/^[\w-]{1,64}$/.test(id)) throw new GoECSError('INVALID_REPORT', 'Invalid report id')
    if (!this.tasks.has(id)) {
      const report = JSON.parse(await readFile(path.join(this.directory, 'report-' + id + '.json'), 'utf8'))
      this.tasks.set(id, { id, ...report.metadata, targetId: report.metadata.target, raw: report.rawText, report, owner: '' })
    }
    return this.view(this.tasks.get(id))
  }
  async dispose() {
    this.closed = true
    await Promise.allSettled([...this.tasks.values()].filter(t => ['running', 'stopping'].includes(t.status)).map(t => this.stop(t.id)))
    this.secrets.clear(); this.pending.clear(); await this.writes.catch(() => {})
  }
}
export { publicError }
