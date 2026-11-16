#!/usr/bin/env node
import { GoECSService, publicError } from '../agent/dsh-service.mjs'
import { writeFile, readFile } from 'node:fs/promises'

function argumentsFor(argv) {
  const args = { options: {} }; let command = 'status'
  const known = new Set(['option', 'state', 'config', 'target', 'choice', 'language', 'network', 'proxy', 'format', 'out', 'version', 'sidebar', 'admin-reminder', 'binary', 'deadline', 'name', 'id', 'host', 'port', 'username', 'key-file', 'fingerprint', 'file', 'allow-partial'])
  if (argv[0] && !argv[0].startsWith('-')) command = argv.shift()
  while (argv.length) {
    const name = argv.shift()
    if (!name.startsWith('--') || !known.has(name.slice(2)) || !argv.length) throw new Error('Unknown option or missing value: ' + name)
    const value = argv.shift()
    if (name === '--option') {
      const index = value.indexOf('='); if (index < 1) throw new Error('Use --option flag=value')
      const raw = value.slice(index + 1); args.options[value.slice(0, index)] = raw === 'true' ? true : raw === 'false' ? false : raw
    } else {
      if (['sidebar', 'admin-reminder', 'allow-partial'].includes(name.slice(2)) && !['true', 'false'].includes(value)) throw new Error(name + ' must be true or false')
      args[name.slice(2)] = value
    }
  }
  return { command, args }
}
const help = `GoECS DSH backend (no Desktop or display required)
  goecs-dsh status
  goecs-dsh install --version latest --target local
  goecs-dsh run --choice 1 --network direct
  goecs-dsh run --choice 2 --network proxy
  goecs-dsh report --id TASK_ID --format md --out report.md
  goecs-dsh settings --sidebar false --admin-reminder false
  goecs-dsh target --name server --host host --username user --key-file /absolute/key --fingerprint SHA256:...
  goecs-dsh inspect --target TARGET_ID
  goecs-dsh import --file report.txt --format json
  --state /absolute/directory isolates preferences, verified engines and reports.
  --config /absolute/config.json supplies operator-owned settings and targets.
  --allow-partial true keeps exit code 0 for a completed but partial native report.
Proxy choices are explicit in automation. Non-root runs keep the complete suite.
SSH passwords/key passphrases can be supplied through GOECS_SSH_PASSWORD and
GOECS_SSH_PASSPHRASE; they are retained in process memory only.
`
if (process.argv.includes('--help') || process.argv.includes('-h')) { console.log(help); process.exit(0) }
let service
try {
  const { command, args } = argumentsFor(process.argv.slice(2))
  service = new GoECSService({ stateDir: args.state }); await service.ready
  if (args.config) {
    const configuration = JSON.parse(await readFile(args.config, 'utf8'))
    if (configuration.settings) service.settings = service.validateSettings(configuration.settings)
    for (const target of configuration.targets || []) await service.saveTarget(target)
  }
  const targetId = args.target || 'local'
  if (targetId !== 'local' && (process.env.GOECS_SSH_PASSWORD || process.env.GOECS_SSH_PASSPHRASE)) service.secrets.set(targetId, { password: process.env.GOECS_SSH_PASSWORD || '', passphrase: process.env.GOECS_SSH_PASSPHRASE || '' })
  let result
  if (command === 'status') result = await service.status()
  else if (command === 'inspect') result = await service.inspect(targetId)
  else if (command === 'install') result = await service.install(targetId, args.version || 'latest')
  else if (command === 'settings') {
    const update = {}
    if (args.sidebar !== undefined) update.showSidebar = args.sidebar === 'true'
    if (args['admin-reminder'] !== undefined) update.requestAdmin = args['admin-reminder'] === 'true'
    if (args.network !== undefined) update.networkMode = args.network
    if (args.binary !== undefined) update.binaryPath = args.binary
    if (args.deadline !== undefined) update.maxTaskSeconds = Number(args.deadline)
    result = await service.saveSettings(update)
  } else if (command === 'target') result = await service.saveTarget({ id: args.id, name: args.name, host: args.host, port: args.port, username: args.username, keyFile: args['key-file'], hostKey: args.fingerprint, binaryPath: args.binary, password: process.env.GOECS_SSH_PASSWORD, passphrase: process.env.GOECS_SSH_PASSPHRASE })
  else if (command === 'run') {
    const { task } = await service.start({ targetId, choice: args.choice || '1', language: args.language || 'zh', networkMode: args.network, options: args.options, proxyURL: args.proxy })
    const live = service.getTask(task.id); let offset = 0
    const print = () => { if (live.raw.length > offset) { process.stderr.write(live.raw.slice(offset)); offset = live.raw.length } }
    const timer = setInterval(print, 200), stop = () => { void service.stop(task.id).catch(() => {}) }
    process.once('SIGINT', stop); process.once('SIGTERM', stop)
    try {
      await live.promise; print(); result = (await service.report(task.id, args.format || 'json')).content
      if (live.status !== 'completed' || (args['allow-partial'] !== 'true' && !['ok', 'skipped'].includes(live.nativeStatus))) process.exitCode = 1
    }
    finally { clearInterval(timer); process.removeListener('SIGINT', stop); process.removeListener('SIGTERM', stop) }
  } else if (command === 'report') { await service.loadHistory(args.id); result = (await service.report(args.id, args.format || 'json')).content }
  else if (command === 'import') { const task = await service.importReport(await readFile(args.file, 'utf8')); result = (await service.report(task.id, args.format || 'json')).content }
  else throw new Error('Unknown command. Use --help')
  const content = typeof result === 'string' ? result : JSON.stringify(result, null, 2)
  if (args.out) await writeFile(args.out, content, { mode: 0o600 }); else console.log(content)
} catch (error) { console.error(JSON.stringify({ error: publicError(error) })); process.exitCode = 1 }
finally { await service?.dispose() }
