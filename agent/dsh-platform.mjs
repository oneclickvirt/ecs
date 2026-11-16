import os from 'node:os'
import { spawn } from 'node:child_process'

export class GoECSError extends Error {
  constructor(code, message, details) { super(message); this.name = 'GoECSError'; this.code = code; this.details = details }
}
export const PROXY_KEYS = ['http_proxy', 'https_proxy', 'all_proxy', 'no_proxy', 'HTTP_PROXY', 'HTTPS_PROXY', 'ALL_PROXY', 'NO_PROXY']
export const ARCHES = { x64: 'amd64', arm64: 'arm64', ia32: '386', arm: 'arm', amd64: 'amd64', aarch64: 'arm64', x86_64: 'amd64', i686: '386', armv7l: 'arm', riscv64: 'riscv64', s390x: 's390x', ppc64le: 'ppc64le' }
export function normalizeTarget(platform, arch) {
  platform = ({ Linux: 'linux', Darwin: 'darwin', Windows_NT: 'win32', windows: 'win32' })[platform] || platform
  arch = ARCHES[arch] || arch
  if (!['linux', 'darwin', 'win32'].includes(platform) || !/^(?:amd64|arm64|386|arm|mips(?:64)?(?:le)?|ppc64(?:le)?|riscv64|s390x)$/.test(arch)) throw new GoECSError('UNSUPPORTED_TARGET', 'No supported system/architecture was detected')
  return { platform, arch }
}
export function processEnvironment(base = process.env, mode = 'direct', proxy) {
  const env = {}
  for (const [key, value] of Object.entries(base)) {
    if (value !== undefined && !/(?:token|secret|password|passwd|api[_-]?key|private[_-]?key|^DSH_)/i.test(key) && !/^(?:http_proxy|https_proxy|all_proxy|no_proxy)$/i.test(key)) env[key] = String(value)
  }
  if (mode === 'proxy') {
    if (!proxy?.url) throw new GoECSError('PROXY_UNAVAILABLE', 'Select an available proxy')
    for (const key of ['HTTP_PROXY', 'HTTPS_PROXY', 'ALL_PROXY', 'http_proxy', 'https_proxy', 'all_proxy']) env[key] = proxy.url
    env.NO_PROXY = env.no_proxy = 'localhost,127.0.0.1,::1'
  }
  env.TERM = 'xterm-256color'; env.NONINTERACTIVE = 'true'
  return env
}
export function proxyURL(value) {
  const url = new URL(value)
  if (!['http:', 'https:', 'socks:', 'socks5:', 'socks5h:'].includes(url.protocol) || !url.hostname || url.hash || url.search || (url.pathname && url.pathname !== '/')) throw new GoECSError('INVALID_PROXY', 'Use an HTTP(S) or SOCKS5 proxy origin')
  return url.toString().replace(/\/$/, '')
}
export function safeProxy(proxy) {
  const url = new URL(proxy.url); url.username = ''; url.password = ''
  return { source: proxy.source, url: url.toString().replace(/\/$/, ''), authenticated: Boolean(new URL(proxy.url).username || new URL(proxy.url).password) }
}
export function systemProxies(text, platform) {
  const proxies = [], add = (value, source) => { try { const url = proxyURL(value); if (!proxies.some(p => p.url === url)) proxies.push({ url, source }) } catch {} }
  if (platform === 'darwin') {
    for (const [name, protocol] of [['HTTP', 'http'], ['HTTPS', 'http'], ['SOCKS', 'socks5']]) {
      if (!new RegExp(name + 'Enable\\s*:\\s*1').test(text)) continue
      const host = text.match(new RegExp(name + 'Proxy\\s*:\\s*(\\S+)'))?.[1], port = text.match(new RegExp(name + 'Port\\s*:\\s*(\\d+)'))?.[1]
      if (host && port) add(protocol + '://' + (host.includes(':') && !host.startsWith('[') ? '[' + host + ']' : host) + ':' + port, 'macOS ' + name)
    }
  } else if (platform === 'win32') {
    for (const item of text.trim().split(';')) {
      if (!item) continue
      const [kind, endpoint] = item.includes('=') ? item.split('=', 2) : ['http', item]
      add((kind === 'socks' ? 'socks5' : 'http') + '://' + endpoint, 'Windows system')
    }
  }
  return proxies
}
export function redactSecrets(value, secrets = []) {
  let text = String(value).replace(/((?:https?|socks5?h?):\/\/)[^\s/@]+(?::[^\s/@]*)?@/gi, '$1[redacted]@')
  for (const secret of secrets) if (typeof secret === 'string' && secret.length >= 4) text = text.split(secret).join('[redacted]')
  return text
}
export function capture(argv, { env = processEnvironment(), cwd, timeout = 10000, signal, limit = 1024 * 1024 } = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(argv[0], argv.slice(1), { cwd, env, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'], signal })
    let out = '', err = '', error
    const timer = setTimeout(() => { error = new GoECSError('COMMAND_TIMEOUT', 'Command timed out'); child.kill() }, timeout)
    const append = (which, chunk) => {
      if (which === 'out') out += chunk; else err += chunk
      if (Buffer.byteLength(out) + Buffer.byteLength(err) > limit) { error = new GoECSError('OUTPUT_LIMIT', 'Command output limit exceeded'); child.kill() }
    }
    child.stdout.on('data', chunk => append('out', chunk.toString())); child.stderr.on('data', chunk => append('err', chunk.toString()))
    child.once('error', e => { clearTimeout(timer); reject(e) })
    child.once('close', code => { clearTimeout(timer); if (error) reject(error); else resolve({ code, stdout: out, stderr: err }) })
  })
}
export async function localFacts({ env = process.env, platform = process.platform, arch = process.arch } = {}) {
  const target = normalizeTarget(platform, arch)
  let admin = typeof process.getuid === 'function' && process.getuid() === 0
  if (platform === 'win32') {
    const value = await capture(['powershell.exe', '-NoProfile', '-NonInteractive', '-Command', '([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)']).catch(() => null)
    admin = value?.stdout.trim().toLowerCase() === 'true'
  }
  const proxies = []
  for (const [key, value] of Object.entries(env)) {
    if (!/^(?:http_proxy|https_proxy|all_proxy)$/i.test(key) || !value) continue
    try { const url = proxyURL(value); if (!proxies.some(p => p.url === url)) proxies.push({ url, source: key }) } catch { /* Invalid ambient values are not executed. */ }
  }
  if (platform === 'darwin') {
    const output = await capture(['/usr/sbin/scutil', '--proxy']).catch(() => null)
    for (const proxy of systemProxies(output?.stdout || '', platform)) if (!proxies.some(p => p.url === proxy.url)) proxies.push(proxy)
  }
  if (platform === 'win32') {
    const output = await capture(['powershell.exe', '-NoProfile', '-NonInteractive', '-Command', "$p=Get-ItemProperty 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings'; if($p.ProxyEnable -eq 1){$p.ProxyServer}"]).catch(() => null)
    for (const proxy of systemProxies(output?.stdout || '', platform)) if (!proxies.some(p => p.url === proxy.url)) proxies.push(proxy)
  }
  return { ...target, hostname: os.hostname(), admin, proxies, proxyNotice: 'HTTP/SOCKS selection applies to supported requests. Native TCP/ICMP/routes and transparent VPN/TUN routes may remain direct or use the OS route.' }
}
