import { spawn } from 'node:child_process'
import { readdir, readFile, writeFile, open, rm, stat, utimes } from 'node:fs/promises'
import { appendFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { createHash } from 'node:crypto'
import { menuDriver } from './dsh-options.mjs'
import { GoECSError, capture, localFacts } from './dsh-platform.mjs'

const psString = value => "'" + String(value).replace(/'/g, "''") + "'"
const shString = value => "'" + String(value).replace(/'/g, "'\\''") + "'"
const encodedPS = code => ['powershell.exe', '-NoProfile', '-NonInteractive', '-EncodedCommand', Buffer.from(code, 'utf16le').toString('base64')]
const pause = ms => new Promise(resolve => setTimeout(resolve, ms))

export function elevationCommand(planFile, platform = process.platform, executable = process.execPath) {
  const entry = fileURLToPath(import.meta.url)
  if (platform === 'win32') {
    const code = "$env:ELECTRON_RUN_AS_NODE='1';& " + [executable, entry, '--elevated-worker', planFile].map(psString).join(' ') + ';exit $LASTEXITCODE'
    const argument = '-NoProfile -NonInteractive -EncodedCommand ' + Buffer.from(code, 'utf16le').toString('base64')
    return encodedPS('$ErrorActionPreference=\'Stop\';$p=Start-Process -FilePath powershell.exe -Verb RunAs -PassThru -Wait -ArgumentList ' + psString(argument) + ';exit $p.ExitCode')
  }
  const command = [executable, entry, '--elevated-worker', planFile].map(shString).join(' ')
  if (platform === 'darwin') return ['/usr/bin/osascript', '-e', 'do shell script ' + JSON.stringify('ELECTRON_RUN_AS_NODE=1 ' + command) + ' with administrator privileges']
  return ['pkexec', '/usr/bin/env', 'ELECTRON_RUN_AS_NODE=1', executable, entry, '--elevated-worker', planFile]
}

// The rule covers only this engine on currently connected network profiles.
// It is removed when this task finishes; existing allow/block rules are intact.
export function firewallScript(binary, name, remove = false) {
  if (remove) return "$ErrorActionPreference='Stop';if(Get-NetFirewallRule -Name " + psString(name) + ' -ErrorAction SilentlyContinue){Remove-NetFirewallRule -Name ' + psString(name) + ' -ErrorAction Stop}'
  return "$ErrorActionPreference='Stop';$program=" + psString(binary) + ';$name=' + psString(name) + ';' +
    '$blocked=Get-NetFirewallApplicationFilter -Program $program -ErrorAction SilentlyContinue|Get-NetFirewallRule|Where-Object {$_.Enabled -eq \'True\' -and $_.Action -eq \'Block\'};if($blocked){throw \'FIREWALL_BLOCKED: An existing firewall rule blocks this engine. Review it in Windows Security.\'};' +
    "$profiles=@(Get-NetConnectionProfile|ForEach-Object {if($_.NetworkCategory -eq 'DomainAuthenticated'){'Domain'}else{[string]$_.NetworkCategory}}|Select-Object -Unique);if(!$profiles.Count){throw 'No active network profile was found'};" +
    "New-NetFirewallRule -Name $name -DisplayName 'GoECS benchmark (temporary)' -Program $program -Direction Inbound -Action Allow -Profile $profiles -Protocol Any -EdgeTraversalPolicy Block|Out-Null"
}

export async function startElevatedProcess(binary, args, options, { launch = capture, alreadyAdmin = false } = {}) {
  const { cwd, onData, onReady = () => {} } = options
  const planFile = path.join(cwd, 'privileged-plan.json'), eventsFile = path.join(cwd, 'privileged-events.jsonl'), leaseFile = path.join(cwd, 'privileged-lease'), stopFile = path.join(cwd, 'privileged-stop')
  const hash = createHash('sha256').update(await readFile(binary)).digest('hex')
  const rule = 'GoECS-' + path.basename(cwd)
  await writeFile(eventsFile, '', { mode: 0o600, flag: 'wx' })
  await writeFile(leaseFile, String(Date.now()), { mode: 0o600, flag: 'wx' })
  await writeFile(planFile, JSON.stringify({ binary, args, cwd, env: options.env, choice: options.choice, deadlineMs: options.deadlineMs, hash, firewall: process.platform === 'win32', rule }), { mode: 0o600, flag: 'wx' })
  let settled = false, reading = false, position = 0, remainder = '', ready = false, resolveDone, rejectDone
  const events = await open(eventsFile, 'r'), decoder = new TextDecoder()
  const done = new Promise((resolve, reject) => { resolveDone = resolve; rejectDone = reject })
  const fail = error => { if (!settled) { settled = true; void writeFile(stopFile, 'stop', { mode: 0o600 }).catch(() => {}); rejectDone(error) } }
  const stop = async () => { await writeFile(stopFile, 'stop', { mode: 0o600 }); if (!ready) fail(new GoECSError('ELEVATION_CANCELLED', 'Administrator authorization was cancelled before testing')); return done }
  const poll = async () => {
    if (reading || settled) return
    reading = true
    try {
      const now = new Date(); await utimes(leaseFile, now, now)
      let bytesRead
      do {
        const buffer = Buffer.alloc(65536)
        ;({ bytesRead } = await events.read(buffer, 0, buffer.length, position))
        position += bytesRead; remainder += decoder.decode(buffer.subarray(0, bytesRead), { stream: true })
        let end
        while ((end = remainder.indexOf('\n')) >= 0) {
          const event = JSON.parse(remainder.slice(0, end)); remainder = remainder.slice(end + 1)
          if (event.type === 'data') onData(event.value)
          else if (event.type === 'ready') { ready = true; onReady() }
          else if (event.type === 'deadline') options.onDeadline?.()
          else if (event.type === 'done') { if (!settled) { settled = true; resolveDone(event.value) } }
          else if (event.type === 'error') fail(new GoECSError(event.code, event.message))
        }
      } while (bytesRead === 65536)
    } catch (error) { fail(error) } finally { reading = false }
  }
  const command = alreadyAdmin ? [process.execPath, fileURLToPath(import.meta.url), '--elevated-worker', planFile] : elevationCommand(planFile)
  const launched = launch(command, { timeout: (options.deadlineMs || 3600000) + 180000, env: { ...options.env, ELECTRON_RUN_AS_NODE: '1' }, limit: 65536 }).then(async result => {
    // Read the last event after the authorization helper exits.
    while (reading) await pause(20)
    await poll()
    if (!settled) fail(new GoECSError('ELEVATION_FAILED', result.code === 0 ? 'The privileged helper ended without a confirmed result' : 'Administrator authorization was declined or unavailable. Retry with the current account, or allow the operating-system authorization prompt'))
  }, () => fail(new GoECSError('ELEVATION_FAILED', 'Administrator authorization failed. Retry with the current account')))
  const timer = setInterval(() => void poll(), 100)
  const authorizationTimer = setTimeout(() => { if (!ready && !settled) { void writeFile(stopFile, 'stop', { mode: 0o600 }); fail(new GoECSError('ELEVATION_TIMEOUT', 'Administrator authorization timed out. No benchmark result was produced')) } }, 120000)
  const completed = done.finally(async () => {
    clearInterval(timer); clearTimeout(authorizationTimer)
    while (reading) await pause(20)
    await events.close()
    // Keep the stop marker and lease until the native helper has exited.
    void launched.finally(async () => { for (const file of [planFile, leaseFile, stopFile]) await rm(file, { force: true }).catch(() => {}) })
  })
  return { stop, done: completed }
}

export async function runElevatedWorker(planFile, { facts = localFacts, run = startLocalProcess, command = capture } = {}) {
  const plan = JSON.parse(await readFile(planFile, 'utf8'))
  const eventsFile = path.join(plan.cwd, 'privileged-events.jsonl'), leaseFile = path.join(plan.cwd, 'privileged-lease'), stopFile = path.join(plan.cwd, 'privileged-stop')
  const send = value => appendFileSync(eventsFile, JSON.stringify(value) + '\n')
  let child, monitor, ruleCreated = false, cancellation, result, failure
  try {
    if (!(await facts()).admin) throw new GoECSError('ELEVATION_REQUIRED', 'The operating system did not grant administrator privileges')
    if (createHash('sha256').update(await readFile(plan.binary)).digest('hex') !== plan.hash) throw new GoECSError('ENGINE_CHANGED', 'The engine changed while authorization was pending. Review the executable and retry')
    if (Date.now() - (await stat(leaseFile)).mtimeMs > 10000 || await readFile(stopFile).then(() => true, () => false)) throw new GoECSError('ELEVATION_CANCELLED', 'Authorization was cancelled before the benchmark started')
    if (plan.firewall) {
      const rule = await command(encodedPS(firewallScript(plan.binary, plan.rule)))
      if (rule.code !== 0) throw new GoECSError('FIREWALL_PREPARATION_FAILED', 'Windows firewall preparation failed before testing. Review existing rules or retry without elevation')
      ruleCreated = true
    }
    send({ type: 'ready' })
    child = await run(plan.binary, plan.args, { cwd: plan.cwd, env: plan.env, choice: plan.choice, deadlineMs: plan.deadlineMs, onData: value => send({ type: 'data', value }), onDeadline: () => send({ type: 'deadline' }) })
    let checking = false
    monitor = setInterval(async () => {
      if (checking || cancellation) return
      checking = true
      try {
        const stale = Date.now() - (await stat(leaseFile)).mtimeMs > 10000
        if (stale || await readFile(stopFile).then(() => true, () => false)) { cancellation = child.stop(); void cancellation.catch(() => {}) }
      } catch { cancellation = child.stop(); void cancellation.catch(() => {}) }
      finally { checking = false }
    }, 200)
    result = await child.done
    if (cancellation) await cancellation
  } catch (error) { failure = error }
  finally {
    clearInterval(monitor)
    if (ruleCreated) {
      try { if ((await command(encodedPS(firewallScript(plan.binary, plan.rule, true)))).code !== 0) throw new Error(); }
      catch { failure = new GoECSError('FIREWALL_CLEANUP_FAILED', 'Testing ended but the temporary GoECS firewall rule could not be removed: ' + plan.rule) }
    }
    if (failure) send({ type: 'error', code: failure.code || 'ELEVATED_PROCESS_FAILED', message: failure.message })
    else send({ type: 'done', value: result })
  }
}

async function groupActive(pid) {
  if (process.platform === 'linux') {
    for (const name of await readdir('/proc')) {
      if (!/^\d+$/.test(name)) continue
      try {
        const record = await readFile('/proc/' + name + '/stat', 'utf8'), fields = record.slice(record.lastIndexOf(')') + 2).split(' ')
        if (Number(fields[2]) === pid && fields[0] !== 'Z') return true
      } catch (error) { if (!['ENOENT','ESRCH'].includes(error.code)) throw new GoECSError('PROCESS_STOP_UNCONFIRMED', 'Cannot inspect local process group') }
    }
    return false
  }
  const processes = await capture(['ps', '-axo', 'pgid=,stat='], { timeout: 2000 })
  if (processes.code !== 0) throw new GoECSError('PROCESS_STOP_UNCONFIRMED', 'Could not confirm local process termination')
  return processes.stdout.split('\n').some(line => { const [group, state] = line.trim().split(/\s+/); return Number(group) === pid && state && !state.startsWith('Z') })
}

export async function startLocalProcess(binary, args, { cwd, env, choice, onData, deadlineMs = 3600000, onDeadline = () => {} }) {
  let pid, write, terminate, done
  if (choice !== 'custom') {
    let pty
    try { pty = await import('@lydell/node-pty') } catch { throw new GoECSError('PTY_UNAVAILABLE', 'A terminal provider is required to run original GoECS menu presets on this system') }
    const terminal = (pty.spawn || pty.default?.spawn)(binary, args, { name: 'xterm-256color', cols: 140, rows: 48, cwd, env })
    pid = terminal.pid; write = data => terminal.write(data)
    const drive = menuDriver(choice, write)
    const dataListener = terminal.onData(data => {
      // Bubble Tea's terminal queries need a terminal response even in CI.
      if (data.includes('\x1b[6n')) write('\x1b[1;1R')
      if (data.includes('\x1b]11;?')) write('\x1b]11;rgb:1111/1818/2727\x1b\\')
      onData(data); drive(data)
    })
    done = new Promise((resolve, reject) => {
      const exitListener = terminal.onExit(result => {
        dataListener.dispose(); exitListener.dispose()
        // ConPTY keeps its output worker alive until the terminal is released.
        try { if (process.platform === 'win32') terminal.kill(); resolve(result) } catch (error) { reject(error) }
      })
    })
    terminate = () => terminal.kill()
  } else {
    const child = spawn(binary, args, { cwd, env, detached: process.platform !== 'win32', windowsHide: true, stdio: ['pipe', 'pipe', 'pipe'] })
    pid = child.pid; write = data => child.stdin.write(data)
    child.stdout.on('data', data => onData(data.toString())); child.stderr.on('data', data => onData(data.toString()))
    done = new Promise((resolve, reject) => { child.once('error', reject); child.once('close', (exitCode, signal) => resolve({ exitCode, signal })) })
    terminate = () => child.kill('SIGTERM')
  }
  let finished = false, cancellation, rejectStop
  const stopFailure = new Promise((_resolve, reject) => { rejectStop = reject })
  const stop = async () => {
    if (cancellation) return cancellation
    if (finished) return done
    cancellation = (async () => {
      if (process.platform === 'win32') {
        const result = await capture(['taskkill.exe', '/PID', String(pid), '/T', '/F'], { timeout: 10000 }).catch(() => null)
        if (!result || result.code !== 0) throw new GoECSError('PROCESS_STOP_UNCONFIRMED', 'Could not confirm Windows process-tree termination')
      } else {
        try { process.kill(-pid, 'SIGTERM') } catch { terminate() }
        // A terminal parent can exit before a child that ignores SIGTERM.
        // Finish cancelling the whole process group before returning to the UI.
        await new Promise(resolve => setTimeout(resolve, 750))
        try { process.kill(-pid, 'SIGKILL') } catch { if (!finished) terminate() }
        for (let attempt = 0; attempt < 40; attempt++) {
          if (!await groupActive(pid)) break
          if (attempt === 39) throw new GoECSError('PROCESS_STOP_UNCONFIRMED', 'Local child processes did not terminate')
          await new Promise(resolve => setTimeout(resolve, 50))
        }
      }
    })()
    // A failed stop must reach the caller even if the child never exits.
    cancellation.catch(rejectStop)
    await cancellation
    return done
  }
  const timer = setTimeout(() => { onDeadline(); void stop().catch(() => {}) }, deadlineMs)
  done = Promise.race([done, stopFailure]).finally(async () => { finished = true; clearTimeout(timer); if (cancellation) await cancellation })
  return { pid, write, stop, done }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url) && process.argv[2] === '--elevated-worker') {
  await runElevatedWorker(process.argv[3])
}
