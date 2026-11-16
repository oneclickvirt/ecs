import test from 'node:test'
import assert from 'node:assert/strict'
import os from 'node:os'
import path from 'node:path'
import { mkdtemp, readFile, rm, mkdir } from 'node:fs/promises'
import { zipSync } from 'fflate'
import { GoECSService } from './dsh-service.mjs'
import { parseHelp, buildArguments, menuDriver } from './dsh-options.mjs'
import { processEnvironment, safeProxy, redactSecrets, normalizeTarget, systemProxies, capture, localFacts } from './dsh-platform.mjs'
import { startLocalProcess, startElevatedProcess, runElevatedWorker, elevationCommand, firewallScript } from './dsh-process.mjs'
import { checksumFor, selectAsset, boundedFetch, installEngine, sha256, discoverEngines, chinaDownload, releaseSources } from './dsh-engine.mjs'
import { GoECSError } from './dsh-platform.mjs'
import { quotePOSIX, powershell, hostFingerprint, validateSSH } from './dsh-ssh.mjs'
import { apply } from './dsh-plugin.mjs'
import { writeFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'

test('cancel during preflight remains cancelled', { timeout: 10000 }, async t => {
  t.diagnostic('Creating isolated service')
  const service = await fixture(t)
  t.diagnostic('Starting cancellable preflight')
  service.runTask = async task => new Promise((resolve, reject) => task.controller.signal.addEventListener('abort', () => reject(task.controller.signal.reason), { once: true }))
  const { task } = await service.start({ choice: '2', networkMode: 'direct' })
  t.diagnostic('Cancelling task')
  const stopped = await service.stop(task.id)
  t.diagnostic('Task cancellation completed')
  assert.equal(stopped.status, 'cancelled'); assert.equal(stopped.error, null)
})
test('headless sessions cannot enumerate another session tasks', async t => {
  const service = await fixture(t); service.runTask = async () => {}
  const { task } = await service.start({ choice: '1', networkMode: 'direct' }, { owner: 'session-a' })
  await service.getTask(task.id).promise
  assert.equal((await service.status('session-b')).tasks.length, 0)
  assert.throws(() => service.getTask(task.id, 'session-b'), e => e.code === 'TASK_NOT_FOUND')
  assert.equal((await service.status('session-a')).tasks.length, 1)
})
test('JSON imports preserve original text and native partial status', async t => {
  const service = await fixture(t)
  const native = { schema_version: 'goecs.report/v1', status: 'partial', text: 'CPU: unavailable\n' }
  const task = await service.importReport(JSON.stringify(native)), report = await service.report(task.id)
  assert.equal(report.document.rawText, native.text); assert.equal(report.document.nativeReport.status, 'partial')
  await assert.rejects(service.importReport('{"status":"ok"}'), e => e.code === 'INVALID_REPORT')
})
test('history sorts timestamps instead of UUID filenames', async t => {
  const service = await fixture(t)
  for (const [id, date] of [['zzzz', '2020-01-01'], ['aaaa', '2026-11-16']]) await writeFile(path.join(service.directory, 'report-' + id + '.json'), JSON.stringify({ metadata: { startedAt: date } }))
  assert.deepEqual((await service.history()).map(r => r.id), ['aaaa', 'zzzz'])
})
test('history reports persist remarks and support single or batch deletion', async t => {
  const service = await fixture(t)
  const imported = await service.importReport('----VPS融合怪测试----\n分享: https://paste.spiritlhl.net/history\n----系统基础信息----\nCPU: example\n')
  const second = await service.importReport('----VPS融合怪测试----\n----系统基础信息----\nCPU: second\n')
  assert.equal((await service.history()).find(item => item.id === imported.id).shareURLs[0], 'https://paste.spiritlhl.net/history')
  await service.updateRemark(imported.id, 'production host')
  const persisted = JSON.parse(await readFile(path.join(service.directory, 'report-' + imported.id + '.json'), 'utf8'))
  assert.equal(persisted.metadata.remark, 'production host')
  assert.equal((await service.loadHistory(imported.id)).remark, 'production host')
  await service.deleteHistory([imported.id, second.id])
  assert.deepEqual(await service.history(), [])
  await assert.rejects(service.loadHistory(imported.id), error => error.code === 'ENOENT')
})
test('remote preparation discovers or uploads an engine before option discovery', async t => {
  const service = await fixture(t)
  const target = await service.saveTarget({ name: 'Remote', host: 'example.test', username: 'user' })
  const remote = { ...service.targets.get(target.id), binaryPath: '' }
  let installs = 0
  service.environment = async targetId => ({ target: { ...service.targets.get(targetId), id: targetId }, facts: { ...facts, platform: 'linux', home: '/home/user' }, client: { end() {} } })
  service.findRemoteBinary = async () => ''
  service.install = async targetId => { installs++; service.targets.set(targetId, { ...remote, binaryPath: '/home/user/.cache/goecs-dsh/engines/latest/linux-amd64/goecs', engineTag: 'v0.2.26' }); return { binaryPath: service.targets.get(targetId).binaryPath, tag: 'v0.2.26' } }
  service.inspectEngine = async ({ target: value }) => ({ binaryPath: value.binaryPath, tag: value.engineTag || 'configured', flags: parseHelp(help) })
  const prepared = await service.prepare(target.id)
  assert.equal(installs, 1); assert.equal(prepared.binaryPath, '/home/user/.cache/goecs-dsh/engines/latest/linux-amd64/goecs')
})

const help = '  -basic\n    Enable basic info (default true)\n  -menu\n    Enable menu\n  -json string\n    JSON path\n  -timeout duration\n    Test deadline\n  -log\n    Logging\n  -upload\n    Upload\n  -ut-http-proxy string\n    Proxy\n'
const facts = { platform: 'linux', arch: 'amd64', home: '/home/example', admin: false, proxies: [] }
async function fixture(t, proxy = false) {
  const dir = await mkdtemp(path.join(os.tmpdir(), 'goecs-dsh-test-'))
  const service = new GoECSService({ stateDir: dir }); await service.ready
  service.settings.binaryPath = '/example/goecs'
  service.environment = async () => ({ target: { id: 'local', binaryPath: '/example/goecs' }, facts: { ...facts, proxies: proxy ? [{ source: 'test', url: 'http://127.0.0.1:10808' }] : [] }, client: null })
  service.inspectEngine = async () => ({ flags: parseHelp(help), tag: 'fixture' })
  t.after(async () => { await service.dispose(); await rm(dir, { recursive: true, force: true }) })
  return service
}
test('runtime option discovery forwards selected values but protects managed paths and menu selection', () => {
  const flags = parseHelp(help)
  const option = buildArguments({ choice: '2', options: { timeout: '20s', upload: false } }, flags, '/tmp/test result.json')
  assert.equal(option.choice, '2'); assert.ok(option.args.includes('-json=/tmp/test result.json'))
  assert.ok(!option.args.includes('-menu=false'))
  assert.ok(buildArguments({ choice: 'custom', options: { basic: false } }, flags, '/tmp/report').args.includes('-menu=false'))
  assert.throws(() => buildArguments({ options: { json: '/etc/config' } }, flags, '/tmp/report'))
  assert.throws(() => buildArguments({ options: { basic: 'false' } }, flags, '/tmp/report'))
  assert.throws(() => buildArguments({ choice: '12' }, flags, '/tmp/report'))
})
test('public sharing defaults on and respects explicit opt-out and privacy', () => {
  const flags = parseHelp(help + '  -privacy\n    Private mode\n')
  assert.ok(buildArguments({}, flags, 'report').args.includes('-upload=true'))
  assert.ok(!buildArguments({ options: { upload: false } }, flags, 'report').args.includes('-upload=true'))
  assert.ok(buildArguments({ options: { privacy: true } }, flags, 'report').args.includes('-upload=false'))
})
test('advanced controls expose defaults, choices and canonical aliases from engine help', () => {
  const flags = parseHelp('  -h\tHelp\n  -cpu-method string\n    Method (default "sysbench")\n  -cpum string\n    Method\n  -nt3-location string\n    Location (default "GZ")\n  -deep\n    Deep probes\n')
  assert.equal(flags[0].managed, true)
  assert.equal(flags[1].defaultValue, 'sysbench'); assert.equal(flags[1].group, 'hardware')
  assert.ok(flags[1].choices.includes('winsat')); assert.equal(flags[2].aliasOf, 'cpu-method')
  assert.equal(flags[3].defaultValue, 'GZ'); assert.equal(flags[4].group, 'deep')
})
test('download retries connection resets and HTTP failures through pinned fallback sources', async () => {
  const url = 'https://github.com/oneclickvirt/ecs/releases/download/v0.2.21/checksums.txt', calls = []
  const bytes = await boundedFetch(url, 100, undefined, [], { retryDelay: 0, request: async source => {
    calls.push(source)
    if (calls.length === 1) throw Object.assign(new Error('reset'), { code: 'ECONNRESET' })
    if (calls.length === 2) throw new GoECSError('DOWNLOAD_HTTP', 'HTTP 503')
    return Buffer.from('verified response')
  } })
  assert.equal(bytes.toString(), 'verified response'); assert.equal(calls.length, 3)
  assert.ok(calls[1].endsWith('/v0.2.21/checksums.txt'))
  await assert.rejects(boundedFetch(url, 100, AbortSignal.abort(new Error('cancelled')), [], { request: () => assert.fail('cancel must not download') }), /cancelled/)
})
test('a bad archive preserves the old engine; successful install is discoverable offline', async t => {
  const dir = await mkdtemp(path.join(os.tmpdir(), 'goecs-engine-test-')), target = { platform: 'linux', arch: 'amd64' }
  t.after(() => rm(dir, { recursive: true, force: true }))
  const directory = path.join(dir, 'v0.2.21', 'linux-amd64'), filename = 'goecs_linux_amd64.zip'
  await mkdir(directory, { recursive: true }); await writeFile(path.join(directory, 'goecs'), 'old engine')
  const archive = Buffer.from(zipSync({ goecs: Buffer.from('new engine') })), hash = sha256(archive)
  const asset = name => ({ name, browser_download_url: 'https://github.com/oneclickvirt/ecs/releases/download/v0.2.21/' + name })
  const release = { tag_name: 'v0.2.21', assets: [asset(filename), asset('checksums.txt')] }
  let damaged = true
  const fetchBytes = async url => url.includes('api.github.com') ? Buffer.from(JSON.stringify(release)) : url.endsWith('checksums.txt') ? Buffer.from(hash + '  ' + filename) : damaged ? Buffer.from('broken archive') : archive
  await assert.rejects(installEngine(dir, target, 'v0.2.21', undefined, { fetchBytes, detectChina: async () => false }), error => error.code === 'CHECKSUM_MISMATCH')
  assert.equal(await readFile(path.join(directory, 'goecs'), 'utf8'), 'old engine')
  damaged = false
  const installed = await installEngine(dir, target, 'v0.2.21', undefined, { fetchBytes, detectChina: async () => false })
  assert.equal((await discoverEngines(dir, target, { PATH: '' }))[0], installed.binaryPath)
  await writeFile(installed.binaryPath, 'tampered')
  assert.ok(!(await discoverEngines(dir, target, { PATH: '' })).includes(installed.binaryPath))
})
test('download geography honors the shell override and uses a bounded fallback without retaining addresses', async () => {
  const request = async () => assert.fail('An explicit override must not make a geography request')
  for (const CN of ['true', '1', 'TRUE']) assert.equal(await chinaDownload(undefined, { env: { CN }, request }), true)
  for (const CN of ['false', '0']) assert.equal(await chinaDownload(undefined, { env: { CN }, request }), false)
  let calls = 0
  assert.equal(await chinaDownload(undefined, { env: {}, request: async () => { calls++; return Buffer.from(calls === 1 ? '{"error":true}' : '{"country":"CN"}') } }), true)
  assert.equal(calls, 2)
  assert.equal(await chinaDownload(undefined, { env: {}, request: async () => Buffer.from('当前 IP：[omitted] 来自于：中国 广东 广州') }), true)
  assert.equal(await chinaDownload(undefined, { env: {}, request: async () => Buffer.from('当前 IP：[omitted] 来自于：美国 加利福尼亚') }), false)
  assert.equal(await chinaDownload(undefined, { env: {}, request: async () => Buffer.from('{"country_code":"US"}') }), false)
  assert.equal(await chinaDownload(undefined, { env: {}, request: async () => { throw new Error('offline') } }), false)
  await assert.rejects(chinaDownload(AbortSignal.abort(new Error('cancel geography')), { env: {} }), /cancel geography/)
  // Keep this mock's timer alive just as a real HTTPS request would be.
  assert.equal(await chinaDownload(undefined, { env: {}, timeoutMs: 10, request: (_url, _limit, signal) => new Promise((resolve, reject) => {
    const timer = setTimeout(() => resolve(Buffer.from('{}')), 100)
    signal.addEventListener('abort', () => { clearTimeout(timer); reject(signal.reason) }, { once: true })
  }) }), false)
})
test('China downloads try CDN first and fall back immediately; other networks keep GitHub first', async () => {
  const url = 'https://github.com/oneclickvirt/ecs/releases/download/v0.2.24/checksums.txt', api = 'https://api.github.com/repos/oneclickvirt/ecs/releases/latest'
  assert.equal(releaseSources(url)[0], url); assert.equal(releaseSources(api)[0], api)
  assert.ok(releaseSources(api, true)[0].startsWith('https://githubapi.spiritlhl.top/'))
  const calls = []
  assert.equal((await boundedFetch(url, 100, undefined, [], { china: true, retryDelay: 0, request: async source => { calls.push(source); if (calls.length === 1) throw new Error('CDN offline'); return Buffer.from('fallback') } })).toString(), 'fallback')
  assert.equal(calls[0], 'https://cdn0.spiritlhl.top/' + url); assert.equal(calls[1], url)
})
test('one geography decision applies to release metadata, checksums and archive verification', async t => {
  const dir = await mkdtemp(path.join(os.tmpdir(), 'goecs-china-install-'))
  t.after(() => rm(dir, { recursive: true, force: true }))
  const archive = Buffer.from(zipSync({ goecs: Buffer.from('verified engine') })), filename = 'goecs_linux_amd64.zip', options = []
  const asset = name => ({ name, browser_download_url: 'https://github.com/oneclickvirt/ecs/releases/download/v0.2.24/' + name })
  let detected = 0
  const installed = await installEngine(dir, { platform: 'linux', arch: 'amd64' }, 'v0.2.24', undefined, { detectChina: async () => { detected++; return true }, fetchBytes: async (url, _limit, _signal, _proxies, policy) => {
    options.push(policy)
    return url.includes('api.github.com') ? Buffer.from(JSON.stringify({ tag_name: 'v0.2.24', assets: [asset(filename), asset('checksums.txt')] })) : url.endsWith('checksums.txt') ? Buffer.from(sha256(archive) + '  ' + filename) : archive
  } })
  assert.equal(detected, 1); assert.deepEqual(options, [{ china: true }, { china: true }, { china: true }])
  assert.equal(await readFile(installed.binaryPath, 'utf8'), 'verified engine')
})
test('elevation is a local user decision; agents and remote targets cannot request it', async t => {
  const service = await fixture(t)
  service.environment = async () => ({ target: service.target(), facts: { ...facts, platform: 'win32', elevation: 'windows-uac' }, client: null })
  service.runTask = async () => {}
  const { decision } = await service.start({ choice: '2', networkMode: 'direct' }, { interactive: true })
  assert.equal(decision.canElevate, true); assert.equal(decision.needsFirewall, true)
  await assert.rejects(service.accept(decision.token, { privilege: 'elevate', networkMode: 'direct' }), e => e.code === 'USER_ONLY')
  const { task } = await service.accept(decision.token, { privilege: 'current', networkMode: 'direct' }, { user: true })
  await service.getTask(task.id).promise
  await service.saveSettings({ requestAdmin: false })
  const disabled = await service.preflight({ choice: '1' }, { interactive: true })
  assert.equal(disabled.canElevate, false); assert.equal(disabled.needsFirewall, false); assert.equal(disabled.needsAdmin, false)
  const headless = await service.preflight({ choice: '1', networkMode: 'direct' })
  await assert.rejects(service.accept(headless.token, { privilege: 'elevate', networkMode: 'direct' }, { user: true }), e => e.code === 'USER_ONLY')
})
test('privileged worker streams output before exit and removes only its own firewall rule', { timeout: 15000 }, async t => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), 'goecs-elevation-')), output = [], commands = []
  t.after(() => rm(cwd, { recursive: true, force: true }))
  let ready = false
  const helper = await startElevatedProcess(process.execPath, ['-e', 'console.log("platform result 中文");setTimeout(()=>console.log("finished"),500)'], { cwd, env: processEnvironment(), choice: 'custom', onData: value => output.push(value), onReady: () => { ready = true }, deadlineMs: 5000 }, { launch: async () => {
    const planFile = path.join(cwd, 'privileged-plan.json'), plan = JSON.parse(await readFile(planFile, 'utf8'))
    plan.firewall = true; await writeFile(planFile, JSON.stringify(plan))
    // Heartbeats use file timestamps: an empty lease must not look expired.
    await writeFile(path.join(cwd, 'privileged-lease'), '')
    await runElevatedWorker(planFile, { facts: async () => ({ admin: true }), command: async args => { commands.push(Buffer.from(args.at(-1), 'base64').toString('utf16le')); return { code: 0 } } })
    return { code: 0 }
  } })
  let finished = false; const done = helper.done.then(result => { finished = true; return result })
  for (let i = 0; i < 80 && !output.length; i++) await new Promise(resolve => setTimeout(resolve, 30))
  assert.equal(ready, true); assert.equal(finished, false); assert.match(output.join(''), /platform result 中文/)
  assert.equal((await done).exitCode, 0); assert.match(output.join(''), /finished/)
  assert.equal(commands.length, 2); assert.match(commands[0], /New-NetFirewallRule/); assert.match(commands[1], /Remove-NetFirewallRule -Name 'GoECS-/)
  assert.ok(!commands.join('').includes('Set-NetFirewallProfile'))
})
test('privileged worker checks the executable again and fails before running when it changes', { timeout: 15000 }, async t => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), 'goecs-elevation-changed-')), engine = path.join(cwd, 'engine')
  t.after(() => rm(cwd, { recursive: true, force: true })); await writeFile(engine, 'original')
  const helper = await startElevatedProcess(engine, [], { cwd, env: {}, choice: 'custom', onData: () => assert.fail('No output expected'), deadlineMs: 5000 }, { launch: async () => {
    await writeFile(engine, 'changed')
    await runElevatedWorker(path.join(cwd, 'privileged-plan.json'), { facts: async () => ({ admin: true }), run: () => assert.fail('Changed executable must not run') })
    return { code: 0 }
  } })
  await assert.rejects(helper.done, e => e.code === 'ENGINE_CHANGED')
})
test('cancelling an elevated task terminates its real subprocess and removes the temporary rule', { timeout: 15000 }, async t => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), 'goecs-elevation-stop-')), commands = []
  t.after(() => rm(cwd, { recursive: true, force: true }))
  let started = false
  const helper = await startElevatedProcess(process.execPath, ['-e', 'console.log("started");setInterval(()=>{},1000)'], { cwd, env: processEnvironment(), choice: 'custom', onData: () => { started = true }, deadlineMs: 5000 }, { launch: async () => {
    const planFile = path.join(cwd, 'privileged-plan.json'), plan = JSON.parse(await readFile(planFile, 'utf8'))
    plan.firewall = true; await writeFile(planFile, JSON.stringify(plan))
    await runElevatedWorker(planFile, { facts: async () => ({ admin: true }), command: async args => { commands.push(Buffer.from(args.at(-1), 'base64').toString('utf16le')); return { code: 0 } } })
    return { code: 0 }
  } })
  for (let i = 0; i < 80 && !started; i++) await new Promise(resolve => setTimeout(resolve, 30))
  assert.equal(started, true)
  await helper.stop(); await helper.done
  assert.equal(commands.length, 2); assert.match(commands[1], /Remove-NetFirewallRule/)
})
test('network preparation failure prevents every benchmark and authorization decline never starts a process', { timeout: 15000 }, async t => {
  for (const declined of [false, true]) {
    const cwd = await mkdtemp(path.join(os.tmpdir(), 'goecs-elevation-denied-'))
    t.after(() => rm(cwd, { recursive: true, force: true }))
    const helper = await startElevatedProcess(process.execPath, [], { cwd, env: processEnvironment(), choice: 'custom', onData: () => assert.fail('No benchmark may start'), deadlineMs: 5000 }, { launch: async () => {
      if (declined) return { code: 1 }
      const planFile = path.join(cwd, 'privileged-plan.json'), plan = JSON.parse(await readFile(planFile, 'utf8')); plan.firewall = true; await writeFile(planFile, JSON.stringify(plan))
      await runElevatedWorker(planFile, { facts: async () => ({ admin: true }), command: async () => ({ code: 1 }), run: () => assert.fail('No benchmark may start') })
      return { code: 0 }
    } })
    await assert.rejects(helper.done, e => e.code === (declined ? 'ELEVATION_FAILED' : 'FIREWALL_PREPARATION_FAILED'))
  }
})
test('native Windows firewall and elevated process integration', { skip: process.platform !== 'win32' || process.env.GOECS_FIREWALL_SMOKE !== '1', timeout: 30000 }, async t => {
  assert.equal((await localFacts()).admin, true, 'This explicit smoke test requires the CI runner administrator account')
  const cwd = await mkdtemp(path.join(os.tmpdir(), 'goecs-firewall-smoke-'))
  t.after(() => rm(cwd, { recursive: true, force: true }))
  let output = '', ready = false
  const helper = await startElevatedProcess(process.execPath, ['-e', 'console.log("native-firewall-smoke 中文")'], { cwd, env: processEnvironment(), choice: 'custom', onData: value => { output += value }, onReady: () => { ready = true }, deadlineMs: 5000 }, { alreadyAdmin: true })
  assert.equal((await helper.done).exitCode, 0); assert.equal(ready, true); assert.match(output, /native-firewall-smoke 中文/)
  const check = "if(Get-NetFirewallRule -Name 'GoECS-" + path.basename(cwd) + "' -ErrorAction SilentlyContinue){exit 1}"
  assert.equal((await capture(['powershell.exe', '-NoProfile', '-NonInteractive', '-EncodedCommand', Buffer.from(check, 'utf16le').toString('base64')])).code, 0, 'The temporary firewall rule must be gone after completion')
})
test('authorization commands quote paths; firewall scripts parse on Windows without changing firewall state', async () => {
  const file = "C:/folder with spaces/中文's plan.json"
  const win = elevationCommand(file, 'win32', 'C:/Program Files/node.exe')
  assert.equal(win[0], 'powershell.exe')
  const outer = Buffer.from(win.at(-1), 'base64').toString('utf16le'), inner = Buffer.from(outer.match(/-EncodedCommand ([A-Za-z0-9+/=]+)/)[1], 'base64').toString('utf16le')
  assert.ok(inner.includes("中文''s plan.json")); assert.match(outer, /-Verb RunAs/)
  assert.equal(elevationCommand('/tmp/a b', 'darwin')[0], '/usr/bin/osascript')
  assert.equal(elevationCommand('/tmp/a b', 'linux')[0], 'pkexec')
  if (process.platform === 'win32') for (const script of [outer, inner, firewallScript(file, 'GoECS-test'), firewallScript(file, 'GoECS-test', true)]) {
    const parse = '$errors=$null;$tokens=$null;[System.Management.Automation.Language.Parser]::ParseInput(' + "'" + script.replace(/'/g, "''") + "'" + ',[ref]$tokens,[ref]$errors)|Out-Null;if($errors.Count){exit 1}'
    assert.equal((await capture(['powershell.exe', '-NoProfile', '-NonInteractive', '-EncodedCommand', Buffer.from(parse, 'utf16le').toString('base64')])).code, 0)
  }
})
test('first-open preparation downloads once, fills the path and reuses it without network', async t => {
  const service = await fixture(t); service.settings.binaryPath = ''; service.findEngines = async () => []
  service.environment = async () => ({ target: service.target(), facts, client: null })
  let downloads = 0
  service.downloadEngine = async () => { downloads++; await new Promise(resolve => setTimeout(resolve, 30)); return { binaryPath: path.join(service.directory, 'goecs'), tag: 'v0.2.21' } }
  const [a, b] = await Promise.all([service.prepare(), service.prepare()])
  assert.equal(downloads, 1); assert.equal(a.tag, b.tag); assert.ok(service.settings.binaryPath)
  const persisted = JSON.parse(await readFile(path.join(service.directory, 'settings.json'), 'utf8'))
  assert.equal(persisted.settings.binaryPath, service.settings.binaryPath)
  await service.prepare(); assert.equal(downloads, 1)
})
test('live reports publish arrived sections in source order while the task is still running', async t => {
  const service = await fixture(t)
  const task = { id: 'live-output', targetId: 'local', status: 'running', raw: '----系统基础信息----\nCPU: example\n', startedAt: new Date().toISOString() }
  service.tasks.set(task.id, task)
  assert.equal((await service.report(task.id)).document.sections[0].title, '系统基础信息')
  assert.equal(service.view(task).reportReady, false)
  task.raw += '----CPU测试----\n分数: 100\n'
  const doc = (await service.report(task.id)).document
  assert.equal(doc.metadata.status, 'running'); assert.deepEqual(doc.sections.map(s => s.title), ['系统基础信息', 'CPU测试'])
  task.status = 'completed'; task.report = doc
  assert.equal(service.view(task).reportReady, true)
})
test('each original preset is selected in the original terminal, without duplicated test presets', () => {
  for (let choice = 1; choice <= 11; choice++) {
    const writes = [], driver = menuDriver(String(choice), text => writes.push(text))
    driver('Full Su'); driver('ite\nAdvanced Custom\n')
    assert.equal(writes[0], '\x1b[B'.repeat(choice - 1) + '\r')
    driver('Advanced Custom\n'); assert.equal(writes.length, 1)
  }
  const writes = []; menuDriver('11', text => writes.push(text))('请输入选项: ')
  assert.deepEqual(writes, ['11\r'])
})
test('privilege reminder is optional and never removes tests; noninteractive non-root runs do not wait', async t => {
  const service = await fixture(t)
  assert.equal((await service.preflight({ choice: '1' }, { interactive: true })).needsAdmin, true)
  assert.equal((await service.preflight({ choice: '1' })).needsAdmin, false)
  await service.saveSettings({ requestAdmin: false })
  const plan = await service.preflight({ choice: '2' }, { interactive: true })
  assert.equal(plan.needsAdmin, false); assert.equal(service.pending.get(plan.token).input.choice, '2')
})
test('detected proxy requires a choice in UI and an explicit policy in headless CI', async t => {
  const service = await fixture(t, true)
  const plan = await service.preflight({ choice: '1' }, { interactive: true })
  assert.equal(plan.needsProxy, true)
  await assert.rejects(service.preflight({ choice: '1' }), error => error.code === 'PROXY_CHOICE_REQUIRED')
  const direct = await service.preflight({ choice: '1', networkMode: 'direct' })
  assert.equal(direct.needsProxy, false)
})
test('settings changes invalidate pending decisions and target credentials never enter saved JSON', async t => {
  const service = await fixture(t)
  const plan = await service.preflight({ choice: '1' }, { interactive: true })
  await service.saveSettings({ showSidebar: false })
  await assert.rejects(service.accept(plan.token, { networkMode: 'direct' }, { user: true }), e => e.code === 'PLAN_EXPIRED')
  const target = await service.saveTarget({ name: 'Example', host: 'example.test', username: 'user', password: 'NOT-A-REAL-PASSWORD', passphrase: 'NOT-A-REAL-PASSPHRASE' })
  const stored = await readFile(path.join(service.directory, 'settings.json'), 'utf8')
  assert.ok(!stored.includes('NOT-A-REAL')); assert.ok(!JSON.stringify(await service.status()).includes('NOT-A-REAL'))
  assert.equal(service.secrets.get(target.id).password, 'NOT-A-REAL-PASSWORD')
})
test('direct environment removes proxy and Harness secrets; proxy mode forwards only its selected proxy', async () => {
  const env = { PATH: '/bin', HOME: '/home/example', HTTPS_PROXY: 'http://old:1', Https_Proxy: 'http://mixed:1', No_Proxy: '*', DSH_PROFILE: 'desktop', GH_TOKEN: 'example-private-token', API_KEY: 'example-key' }
  const direct = processEnvironment(env, 'direct')
  for (const key of ['HTTPS_PROXY', 'Https_Proxy', 'No_Proxy', 'GH_TOKEN', 'API_KEY', 'DSH_PROFILE']) assert.ok(!(key in direct))
  assert.equal(direct.HOME, '/home/example')
  const proxy = processEnvironment(env, 'proxy', { url: 'http://new:2' }); assert.equal(proxy.HTTPS_PROXY, 'http://new:2')
  const facts = await localFacts({ env: { Https_Proxy: 'http://mixed:1' }, platform: 'linux', arch: 'x64' })
  assert.deepEqual(facts.proxies, [{ source: 'Https_Proxy', url: 'http://mixed:1' }])
  assert.deepEqual(safeProxy({ source: 'test', url: 'http://u:p@localhost:2' }), { source: 'test', url: 'http://localhost:2', authenticated: true })
  assert.ok(!redactSecrets('http://user:secret@localhost:2 secret', ['secret']).includes('secret'))
})
test('release architectures come from actual assets and ambiguous/missing checksums are rejected', () => {
  const asset = name => ({ name, browser_download_url: 'https://github.com/oneclickvirt/ecs/releases/download/v0.2.18/' + name })
  const release = { tag_name: 'v0.2.18', assets: [asset('checksums.txt'), asset('goecs_windows_arm64.zip')] }
  assert.equal(selectAsset(release, { platform: 'win32', arch: 'arm64' }).asset.name, 'goecs_windows_arm64.zip')
  assert.throws(() => selectAsset(release, { platform: 'darwin', arch: 'arm64' }))
  const sha = 'a'.repeat(64); assert.equal(checksumFor(sha + '  a.zip\n', 'a.zip'), sha)
  assert.throws(() => checksumFor(sha + '  a.zip\n' + sha + '  a.zip', 'a.zip'))
})
test('SSH validates host identity and quoting; Windows commands use Unicode-safe encoded PowerShell', () => {
  assert.equal(hostFingerprint(Buffer.from('example')).slice(0, 7), 'SHA256:')
  assert.throws(() => validateSSH({ host: 'host; bad', username: 'user' }))
  assert.throws(() => validateSSH({ host: 'example.test', username: 'user', hostKey: 'anything' }))
  assert.equal(quotePOSIX("a'b $(echo bad)"), "'a'\\''b $(echo bad)'")
  const encoded = powershell("Write-Output '中文路径'").split(' ').at(-1)
  assert.equal(Buffer.from(encoded, 'base64').toString('utf16le'), "Write-Output '中文路径'")
  for (const [platform, arch] of [['win32','x64'],['darwin','arm64'],['Linux','aarch64']]) assert.ok(normalizeTarget(platform, arch).arch)
})
test('headless plugin entry loads without Web services and does not start tests during registration', async t => {
  const directory = await mkdtemp(path.join(os.tmpdir(), 'goecs-dsh-host-')), effects = [], registered = [], injections = []
  const context = { effect: fn => effects.push(fn()), inject: (names, callback) => { injections.push(names); if (names.includes('tools')) callback({ tools: { register: tool => registered.push(tool) } }) } }
  apply(context, { stateDir: directory })
  assert.equal(registered.length, 3); assert.ok(injections.some(names => names.includes('connection')))
  const status = await registered[0].execute({ action: 'status' }, { signal: new AbortController().signal })
  assert.equal(status.tasks.length, 0); assert.equal(status.uiOptional, true)
  t.after(async () => { for (const cleanup of effects) await cleanup(); await rm(directory, { recursive: true, force: true }) })
})
test('system proxy parsing is shared by local and SSH targets', () => {
  assert.deepEqual(systemProxies('HTTPEnable : 1\nHTTPProxy : 127.0.0.1\nHTTPPort : 7890\nSOCKSEnable : 0', 'darwin'), [{ url: 'http://127.0.0.1:7890', source: 'macOS HTTP' }])
  assert.equal(systemProxies('http=localhost:7890;https=localhost:7890;socks=localhost:7891', 'win32').length, 2)
  assert.deepEqual(systemProxies('invalid value', 'win32'), [])
})
test('incompatible lifecycle fails clearly and optional services remain optional', async t => {
  assert.throws(() => apply({}), e => e.code === 'DSH_INCOMPATIBLE')
  const directory = await mkdtemp(path.join(os.tmpdir(), 'goecs-no-host-')), cleanups = []
  apply({ effect: fn => cleanups.push(fn()), inject: (_names, callback) => callback({}) }, { stateDir: directory })
  t.after(async () => { for (const cleanup of cleanups) await cleanup(); await rm(directory, { recursive: true, force: true }) })
})
test('real subprocess cancellation kills descendants even when the parent exits first', { timeout: 15000 }, async t => {
  let output = ''
  const descendant = 'process.on("SIGTERM",()=>{});setInterval(()=>{},1000)'
  const program = 'const {spawn}=require("node:child_process");const p=spawn(process.execPath,["-e",'+JSON.stringify(descendant)+'],{stdio:"ignore"});console.log(p.pid);setInterval(()=>{},1000)'
  const child = await startLocalProcess(process.execPath, ['-e', program], { cwd: os.tmpdir(), env: processEnvironment(), choice: 'custom', onData: text => { output += text }, deadlineMs: 8000 })
  t.after(() => child.stop())
  for (let i = 0; i < 100 && !/\d+/.test(output); i++) await new Promise(resolve => setTimeout(resolve, 20))
  const pid = Number(output.trim()); assert.ok(pid > 0)
  await child.stop(); await child.done
  for (let i = 0; i < 40; i++) { try { process.kill(pid, 0) } catch { break } await new Promise(resolve => setTimeout(resolve, 50)) }
  if (process.platform === 'linux') {
    const record = await readFile('/proc/' + pid + '/stat', 'utf8').catch(error => { if (error.code !== 'ENOENT') throw error; return '' })
    assert.ok(!record || record.slice(record.lastIndexOf(')') + 2).startsWith('Z '))
  } else assert.throws(() => process.kill(pid, 0), error => error.code === 'ESRCH')
})
test('real headless deadline terminates a long-running child', { timeout: 10000 }, async () => {
  let deadline = false
  const child = await startLocalProcess(process.execPath, ['-e', 'setInterval(()=>{},1000)'], { cwd: os.tmpdir(), env: processEnvironment(), choice: 'custom', onData: () => {}, deadlineMs: 150, onDeadline: () => { deadline = true } })
  await child.done; assert.equal(deadline, true)
})
test('real PTY drives the concurrent menu without requiring a display', { timeout: 10000 }, async t => {
  let output = ''
  const program = 'process.stdin.setRawMode(true);process.stdin.resume();process.stdout.write("Full Suite\\nAdvanced Custom\\n");let input="";process.stdin.on("data",b=>{input+=b.toString();if(input.includes("\\r")){console.log("selected="+Buffer.from(input).toString("hex"));process.exit(0)}})'
  const child = await startLocalProcess(process.execPath, ['-e', program], { cwd: os.tmpdir(), env: processEnvironment(), choice: '2', onData: value => { output += value }, deadlineMs: 5000 })
  t.after(() => child.stop())
  assert.equal((await child.done).exitCode, 0)
  assert.ok(output.includes('selected=1b5b420d'), output)
})
test('natural PTY completion releases resources so its host exits without force', { timeout: 15000 }, async () => {
  const moduleURL = new URL('./dsh-process.mjs', import.meta.url).href
  const program = `import { startLocalProcess, startElevatedProcess, runElevatedWorker, elevationCommand, firewallScript } from ${JSON.stringify(moduleURL)};
    for (let run = 0; run < 2; run++) {
      let output = '';
      const child = await startLocalProcess(process.execPath, ['-e', 'console.log("pty-finished")'], {
        cwd: process.cwd(), env: process.env, choice: '1', onData: value => { output += value }, deadlineMs: 5000
      });
      if ((await child.done).exitCode !== 0 || !output.includes('pty-finished')) throw new Error('PTY did not finish');
    }
    console.log('host-finished');`
  const result = await capture([process.execPath, '--input-type=module', '-e', program], { timeout: 12000 })
  assert.equal(result.code, 0, JSON.stringify(result))
  assert.ok(result.stdout.includes('host-finished'), JSON.stringify(result))
})
test('CLI preserves partial-report exit policy and rejects misspelled switches', { skip: process.platform === 'win32', timeout: 10000 }, async t => {
  const dir = await mkdtemp(path.join(os.tmpdir(), 'goecs-cli-exit-')), binary = path.join(dir, 'goecs')
  const mock = '#!/bin/sh\nif [ "$1" = "-h" ]; then printf "  -menu\\n    Menu\\n  -json string\\n    JSON\\n"; exit 0; fi\nfor arg in "$@"; do case "$arg" in -json=*) printf \'{"status":"partial","text":"Example: unavailable"}\' > "' + '$' + '{arg#-json=}";; esac; done\n'
  await writeFile(binary, mock, { mode: 0o700 })
  const config = path.join(dir, 'config.json'); await writeFile(config, JSON.stringify({ settings: { binaryPath: binary } }))
  const cli = fileURLToPath(new URL('../scripts/goecs_dsh.mjs', import.meta.url)), args = [process.execPath, cli, 'run', '--config', config, '--state', path.join(dir, 'state'), '--choice', 'custom', '--network', 'direct']
  t.after(() => rm(dir, { recursive: true, force: true }))
  assert.equal((await capture(args)).code, 1)
  assert.equal((await capture([...args, '--allow-partial', 'true'])).code, 0)
  assert.equal((await capture([process.execPath, cli, 'status', '--netwrok', 'direct'])).code, 1)
})
