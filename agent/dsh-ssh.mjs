import ssh2 from 'ssh2'
import { readFile } from 'node:fs/promises'
import { createHash } from 'node:crypto'
import { GoECSError, normalizeTarget, proxyURL, systemProxies } from './dsh-platform.mjs'
import { menuDriver } from './dsh-options.mjs'

const { Client } = ssh2
export const quotePOSIX = value => "'" + String(value).replace(/'/g, "'\\''") + "'"
export const powershell = script => 'powershell.exe -NoLogo -NoProfile -NonInteractive -EncodedCommand ' + Buffer.from(script, 'utf16le').toString('base64')
const psValue = value => "'" + String(value).replace(/'/g, "''") + "'"
export const hostFingerprint = key => 'SHA256:' + createHash('sha256').update(key).digest('base64').replace(/=+$/, '')
export function validateSSH(input) {
  if (!input || typeof input.host !== 'string' || !input.host || input.host.length > 253 || /[\s\x00-\x1f/\\]/.test(input.host)) throw new GoECSError('INVALID_HOST', 'Enter an SSH hostname or IP address')
  if (!Number.isInteger(Number(input.port || 22)) || Number(input.port || 22) < 1 || Number(input.port || 22) > 65535) throw new GoECSError('INVALID_PORT', 'Invalid SSH port')
  if (typeof input.username !== 'string' || !/^[^\s\x00-\x1f]{1,128}$/.test(input.username)) throw new GoECSError('INVALID_USER', 'Enter the remote username')
  if (input.hostKey && !/^SHA256:[A-Za-z0-9+/]{43}$/.test(input.hostKey)) throw new GoECSError('INVALID_HOST_KEY', 'Use the SHA-256 SSH host-key fingerprint')
  return { host: input.host, port: Number(input.port || 22), username: input.username, hostKey: input.hostKey || '', keyFile: input.keyFile || '' }
}
export async function connectSSH(target, secret = {}, signal) {
  const options = validateSSH(target), client = new Client()
  if (signal?.aborted) throw signal.reason
  const privateKey = options.keyFile ? await readFile(options.keyFile) : undefined
  let fingerprint
  await new Promise((resolve, reject) => {
    const abort = () => { client.destroy(); reject(signal.reason) }
    const cleanup = () => signal?.removeEventListener('abort', abort)
    signal?.addEventListener('abort', abort, { once: true })
    client.once('ready', () => { cleanup(); resolve() })
    client.once('error', error => {
      cleanup()
      if (fingerprint && fingerprint !== options.hostKey) reject(new GoECSError(options.hostKey ? 'HOST_KEY_CHANGED' : 'HOST_TRUST_REQUIRED', options.hostKey ? 'SSH host identity changed; verify it before reconnecting' : 'Verify the SSH host fingerprint before connecting', { fingerprint }))
      else reject(new GoECSError('SSH_FAILED', 'SSH authentication or connection failed: ' + error.level))
    })
    client.connect({ ...options, privateKey, password: secret.password || undefined, passphrase: secret.passphrase || undefined, agent: !privateKey && !secret.password ? (process.env.SSH_AUTH_SOCK || (process.platform === 'win32' ? '\\\\.\\pipe\\openssh-ssh-agent' : undefined)) : undefined,
      readyTimeout: 15000, keepaliveInterval: 10000, keepaliveCountMax: 3,
      hostVerifier: key => { fingerprint = hostFingerprint(key); return fingerprint === options.hostKey },
    })
  })
  return client
}
export function sshCapture(client, command, { timeout = 12000, limit = 1024 * 1024 } = {}) {
  return new Promise((resolve, reject) => {
    let channel, out = '', err = '', settled = false
    const finish = (error, code) => { if (settled) return; settled = true; clearTimeout(timer); client.removeListener('close', disconnected); if (error) reject(error); else resolve({ code, stdout: out, stderr: err }) }
    const disconnected = () => finish(new GoECSError('SSH_DISCONNECTED', 'SSH disconnected; the remote result is unconfirmed'))
    client.once('close', disconnected)
    const timer = setTimeout(() => { channel?.close(); finish(new GoECSError('SSH_TIMEOUT', 'Remote command timed out')) }, timeout)
    client.exec(command, (error, stream) => {
      if (error) return finish(error)
      channel = stream
      stream.on('data', bytes => { out += bytes; if (Buffer.byteLength(out) + Buffer.byteLength(err) > limit) { stream.close(); finish(new GoECSError('OUTPUT_LIMIT', 'Remote command output limit exceeded')) } })
      stream.stderr.on('data', bytes => { err += bytes })
      stream.once('error', e => finish(e)); stream.once('close', code => finish(null, code))
    })
  })
}
export async function remoteFacts(client) {
  const unix = await sshCapture(client, 'uname -s; uname -m; id -u; printf "%s\\n" "$HOME"; env | sed -n -E \'/^(http_proxy|https_proxy|all_proxy|HTTP_PROXY|HTTPS_PROXY|ALL_PROXY)=/p\'').catch(() => null)
  if (unix?.code === 0 && /^(Linux|Darwin)\r?\n/.test(unix.stdout)) {
    const [platform, arch, uid, home, ...proxyLines] = unix.stdout.trim().split(/\r?\n/)
    const proxies = []
    for (const line of proxyLines) { const i = line.indexOf('='); try { const url = proxyURL(line.slice(i + 1)); if (!proxies.some(p => p.url === url)) proxies.push({ source: line.slice(0, i), url }) } catch {} }
    if (platform === 'Darwin') {
      const system = await sshCapture(client, '/usr/sbin/scutil --proxy').catch(() => null)
      for (const proxy of systemProxies(system?.stdout || '', 'darwin')) if (!proxies.some(p => p.url === proxy.url)) proxies.push(proxy)
    }
    return { ...normalizeTarget(platform, arch), admin: uid === '0', home, proxies }
  }
  const win = await sshCapture(client, powershell("$p=([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator); $e=@{};Get-ChildItem Env:|Where-Object {$_.Name -match '^(HTTP_PROXY|HTTPS_PROXY|ALL_PROXY)$'}|ForEach-Object {$e[$_.Name]=$_.Value};@{arch=$env:PROCESSOR_ARCHITECTURE;admin=$p;home=$env:USERPROFILE;proxies=$e}|ConvertTo-Json -Compress"))
  if (win.code !== 0) throw new GoECSError('REMOTE_OS_UNKNOWN', 'Remote host must provide a Linux/macOS shell or Windows PowerShell with OpenSSH')
  const info = JSON.parse(win.stdout.trim()), proxies = []
  for (const [source, value] of Object.entries(info.proxies || {})) { try { const url = proxyURL(value); if (!proxies.some(p => p.url === url)) proxies.push({ source, url }) } catch {} }
  const system = await sshCapture(client, powershell("$p=Get-ItemProperty 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings' -ErrorAction SilentlyContinue;if($p.ProxyEnable -eq 1){$p.ProxyServer}")).catch(() => null)
  for (const proxy of systemProxies(system?.stdout || '', 'win32')) if (!proxies.some(p => p.url === proxy.url)) proxies.push(proxy)
  return { ...normalizeTarget('win32', String(info.arch).toLowerCase().replace('amd64', 'amd64')), admin: info.admin === true, home: info.home.replace(/\\/g, '/'), proxies }
}
export const remotePath = (...parts) => parts.join('/').replace(/\/+/g, '/')
export async function remoteMkdir(client, directory, platform) {
  const result = await sshCapture(client, platform === 'win32' ? powershell('[IO.Directory]::CreateDirectory(' + psValue(directory) + ')|Out-Null') : 'umask 077; mkdir -p ' + quotePOSIX(directory))
  if (result.code !== 0) throw new GoECSError('REMOTE_DIRECTORY', 'Unable to create the private GoECS cache directory')
}
export async function uploadEngine(client, localFile, destination, platform, expectedSHA) {
  const sftp = await new Promise((resolve, reject) => client.sftp((e, s) => e ? reject(e) : resolve(s)))
  try {
    await new Promise((resolve, reject) => sftp.fastPut(localFile, destination + '.tmp', { mode: 0o700 }, e => e ? reject(e) : resolve()))
    const command = platform === 'win32' ? powershell('(Get-FileHash -Algorithm SHA256 -LiteralPath ' + psValue(destination + '.tmp') + ').Hash') : (platform === 'darwin' ? 'shasum -a 256 ' : 'sha256sum ') + quotePOSIX(destination + '.tmp')
    const result = await sshCapture(client, command)
    if (result.code !== 0 || result.stdout.trim().split(/\s/)[0].toLowerCase() !== expectedSHA) throw new GoECSError('REMOTE_CHECKSUM', 'Remote executable failed SHA-256 verification; existing executable was preserved')
    const move = platform === 'win32' ? powershell('Move-Item -Force -LiteralPath ' + psValue(destination + '.tmp') + ' -Destination ' + psValue(destination)) : 'chmod 700 ' + quotePOSIX(destination + '.tmp') + ' && mv -f ' + quotePOSIX(destination + '.tmp') + ' ' + quotePOSIX(destination)
    if ((await sshCapture(client, move)).code !== 0) throw new GoECSError('REMOTE_INSTALL', 'Remote executable could not be replaced')
  } finally { sftp.end() }
}
export async function readRemoteFile(client, filename, limit = 16 * 1024 * 1024) {
  const sftp = await new Promise((resolve, reject) => client.sftp((e, s) => e ? reject(e) : resolve(s)))
  try {
    return await new Promise((resolve, reject) => {
      const stream = sftp.createReadStream(filename); const chunks = []; let size = 0
      stream.on('data', data => { size += data.length; if (size > limit) { stream.destroy(); reject(new GoECSError('REPORT_LIMIT', 'Remote report exceeds 16 MiB')) } else chunks.push(data) })
      stream.once('error', reject); stream.once('end', () => resolve(Buffer.concat(chunks).toString('utf8')))
    })
  } finally { sftp.end() }
}
function windowsArgument(value) { return '"' + String(value).replace(/(\\*)"/g, '$1$1\\"').replace(/\\+$/, '$&$&') + '"' }
export async function startRemoteProcess(client, binary, args, { cwd, env, choice, onData, deadlineMs = 3600000, onDeadline = () => {}, platform }) {
  const pidFile = remotePath(cwd, 'process.json')
  let command
  if (platform === 'win32') {
    const script = "[Console]::OutputEncoding=[Text.Encoding]::UTF8;Get-ChildItem Env:|Where-Object {$_.Name -match '(?i)token|secret|password|passwd|api[_-]?key|private[_-]?key|^DSH_|^(http_proxy|https_proxy|all_proxy|no_proxy)$'}|ForEach-Object {Remove-Item -LiteralPath ('Env:'+$_.Name)};" + Object.entries(env).filter(([k]) => /^(?:HTTP_PROXY|HTTPS_PROXY|ALL_PROXY|NO_PROXY|NONINTERACTIVE|TERM)$/i.test(k)).map(([k, v]) => '$env:' + k + '=' + psValue(v) + ';').join('') +
      'Set-Location -LiteralPath ' + psValue(cwd) + ';$p=Start-Process -NoNewWindow -PassThru -FilePath ' + psValue(binary) + ' -ArgumentList ' + psValue(args.map(windowsArgument).join(' ')) + ';@{pid=$p.Id;start=$p.StartTime.ToUniversalTime().Ticks}|ConvertTo-Json -Compress|Set-Content -LiteralPath ' + psValue(pidFile) + ';$p.WaitForExit();exit $p.ExitCode'
    command = powershell(script)
  } else {
    const setup = 'unset http_proxy https_proxy all_proxy no_proxy HTTP_PROXY HTTPS_PROXY ALL_PROXY NO_PROXY; '
    const allowedEnv = Object.entries(env).filter(([k]) => /^(?:http_proxy|https_proxy|all_proxy|no_proxy|NONINTERACTIVE|TERM)$/i.test(k)).map(([k, v]) => k + '=' + quotePOSIX(v)).join(' ')
    // Keep the TUI in the foreground. A background shell job can lose stdin.
    command = 'umask 077; cd ' + quotePOSIX(cwd) + ' || exit 1; ' + setup + 'printf "%s\\n" "$$" > ' + quotePOSIX(pidFile) + '; ps -o lstart= -p "$$" >> ' + quotePOSIX(pidFile) + '; exec env -i HOME="$HOME" PATH="$PATH" USER="$USER" LANG="${LANG:-C.UTF-8}" ' + allowedEnv + ' ' + quotePOSIX(binary) + ' ' + args.map(quotePOSIX).join(' ')
  }
  let channel, settled = false, cancellation, disconnect
  const done = new Promise((resolve, reject) => {
    disconnect = () => { if (!settled) { settled = true; reject(new GoECSError('SSH_DISCONNECTED', 'SSH disconnected; remote process termination is unconfirmed')) } }
    client.once('close', disconnect)
    client.exec(command, choice === 'custom' ? {} : { pty: { term: 'xterm-256color', cols: 140, rows: 48 } }, (error, stream) => {
      if (error) { settled = true; reject(error); return }
      channel = stream
      const drive = menuDriver(choice, data => stream.write(data))
      const receive = bytes => { const data = bytes.toString(); if (data.includes('\x1b[6n')) stream.write('\x1b[1;1R'); if (data.includes('\x1b]11;?')) stream.write('\x1b]11;rgb:1111/1818/2727\x1b\\'); onData(data); drive(data) }
      stream.on('data', receive); stream.stderr.on('data', receive)
      stream.once('error', error => { if (!settled) { settled = true; reject(error) } })
      stream.once('close', (exitCode, signal) => { if (!settled) { settled = true; resolve({ exitCode, signal }) } })
    })
  })
  const stop = async () => {
    if (cancellation) { await cancellation; return completed }
    if (settled) return completed
    const stopCommand = platform === 'win32' ? powershell('$v=Get-Content -Raw -LiteralPath ' + psValue(pidFile) + '|ConvertFrom-Json;$p=Get-Process -Id $v.pid -ErrorAction SilentlyContinue;if($p){if($p.StartTime.ToUniversalTime().Ticks -ne $v.start){exit 2};& taskkill.exe /PID $v.pid /T /F|Out-Null;Start-Sleep -Milliseconds 500;if(Get-Process -Id $v.pid -ErrorAction SilentlyContinue){exit 3}}') :
      'n=0; while [ ! -s ' + quotePOSIX(pidFile) + ' ] && [ "$n" -lt 20 ]; do sleep 0.1; n=$((n+1)); done; p=$(sed -n "1p" ' + quotePOSIX(pidFile) + '); case "$p" in ""|*[!0-9]*) exit 2;; esac; ' +
      'if ! kill -0 "$p" 2>/dev/null; then exit 0; fi; expected=$(sed -n "2p" ' + quotePOSIX(pidFile) + '); actual=$(ps -o lstart= -p "$p"); [ "$expected" = "$actual" ] || exit 3; ' +
      'group=$(ps -o pgid= -p "$p" | tr -d " "); ' +
      'children=$(ps -e -o pid= -o ppid= | awk -v root="$p" \'{parent[$1]=$2} END {for(pid in parent){p=pid;for(n=0;n<100 && p>1;n++){if(p==root){print pid;break}p=parent[p]}}}\'); ' +
      'if [ "$group" = "$p" ]; then kill -TERM -"$p" 2>/dev/null || true; else for c in $children; do kill -TERM "$c" 2>/dev/null || true; done; fi; sleep 1; ' +
      'for c in $children; do kill -KILL "$c" 2>/dev/null || true; done; sleep 0.2; for c in $children; do state=$(ps -o stat= -p "$c"); case "$state" in ""|Z*) ;; *) exit 4;; esac; done'
    cancellation = (async () => {
      try {
        const stopped = await sshCapture(client, stopCommand, { timeout: 8000 })
        if (stopped.code !== 0) throw new GoECSError('REMOTE_STOP_UNCONFIRMED', 'Could not confirm remote termination')
      } catch (error) {
        throw new GoECSError('REMOTE_STOP_UNCONFIRMED', 'Could not confirm remote termination: ' + (error.code || 'SSH_FAILED'))
      } finally { channel?.close() }
    })()
    await cancellation
    return completed
  }
  const timer = setTimeout(() => { onDeadline(); void stop().catch(() => { channel?.close() }) }, deadlineMs)
  const completed = done.then(async value => { if (cancellation) await cancellation; return value }).finally(() => { clearTimeout(timer); client.removeListener('close', disconnect) })
  return { write: data => channel?.write(data), stop, done: completed }
}
