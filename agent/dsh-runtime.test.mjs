import test from 'node:test'
import assert from 'node:assert/strict'
import os from 'node:os'
import path from 'node:path'
import { mkdtemp, readFile, rm } from 'node:fs/promises'
import { GoECSService } from './dsh-service.mjs'
import { parseHelp, buildArguments, menuDriver } from './dsh-options.mjs'
import { processEnvironment, safeProxy, redactSecrets, normalizeTarget, systemProxies, capture, localFacts } from './dsh-platform.mjs'
import { startLocalProcess } from './dsh-process.mjs'
import { checksumFor, selectAsset } from './dsh-engine.mjs'
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

const help = '  -basic\n    Enable basic info (default true)\n  -menu\n    Enable menu\n  -json string\n    JSON path\n  -timeout duration\n    Test deadline\n  -log\n    Logging\n  -upload\n    Upload\n  -ut-http-proxy string\n    Proxy\n'
const facts = { platform: 'linux', arch: 'amd64', home: '/home/example', admin: false, proxies: [] }
async function fixture(t, proxy = false) {
  const dir = await mkdtemp(path.join(os.tmpdir(), 'goecs-dsh-test-'))
  const service = new GoECSService({ stateDir: dir }); await service.ready
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
