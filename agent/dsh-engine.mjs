import { mkdir, readFile, writeFile, rename, chmod, stat, rm, readdir } from 'node:fs/promises'
import path from 'node:path'
import os from 'node:os'
import { setTimeout as delay } from 'node:timers/promises'
import { createHash, randomUUID } from 'node:crypto'
import { unzipSync } from 'fflate'
import https from 'node:https'
import { HttpsProxyAgent } from 'https-proxy-agent'
import { SocksProxyAgent } from 'socks-proxy-agent'
import { GoECSError, localFacts } from './dsh-platform.mjs'

export const sha256 = bytes => createHash('sha256').update(bytes).digest('hex')
const tagPattern = /^v\d+\.\d+\.\d+(?:[-+][\w.-]+)?$/
export function selectAsset(release, { platform, arch }) {
  if (!release || release.draft || release.prerelease || !tagPattern.test(release.tag_name || '')) throw new GoECSError('INVALID_RELEASE', 'Select a published stable GoECS release')
  const name = 'goecs_' + (platform === 'win32' ? 'windows' : platform) + '_' + arch + '.zip'
  const asset = release.assets?.find(a => a.name === name), checksums = release.assets?.find(a => a.name === 'checksums.txt')
  if (!asset || !checksums) throw new GoECSError('ASSET_UNAVAILABLE', 'This release does not publish ' + name)
  for (const file of [asset, checksums]) {
    const url = new URL(file.browser_download_url)
    if (url.origin !== 'https://github.com' || !url.pathname.startsWith('/oneclickvirt/ecs/releases/download/' + release.tag_name + '/')) throw new GoECSError('INVALID_ASSET', 'Unexpected release asset origin')
  }
  return { tag: release.tag_name, asset, checksums }
}
export function checksumFor(text, filename) {
  const matches = text.split(/\r?\n/).map(line => line.match(/^([a-fA-F0-9]{64})\s+\*?(.+?)\s*$/)).filter(m => m?.[2] === filename)
  if (matches.length !== 1) throw new GoECSError('CHECKSUM_MISSING', 'Release checksum is missing or ambiguous: ' + filename)
  return matches[0][1].toLowerCase()
}
function requestBytes(url, limit, signal, agent, redirects = 0) {
  const parsed = new URL(url)
  if (parsed.protocol !== 'https:' || !/^(?:github\.com|api\.github\.com|[\w.-]+\.githubusercontent\.com|githubapi\.spiritlhl\.workers\.dev|githubapi\.spiritlhl\.top|cdn0\.spiritlhl\.top)$/.test(parsed.hostname) || redirects > 5) throw new GoECSError('INVALID_ASSET', 'Unexpected release download redirect')
  return new Promise((resolve, reject) => {
    const request = https.get(parsed, { agent, signal, headers: { 'User-Agent': 'dsh-plugin-goecs/0.2.22' } }, response => {
      if ([301, 302, 303, 307, 308].includes(response.statusCode) && response.headers.location) {
        response.resume()
        try { resolve(requestBytes(new URL(response.headers.location, parsed), limit, signal, agent, redirects + 1)) } catch (error) { reject(error) }
        return
      }
      if (response.statusCode !== 200) { response.resume(); reject(new GoECSError('DOWNLOAD_HTTP', 'Download returned HTTP ' + response.statusCode)); return }
      const chunks = []; let size = 0
      response.on('data', part => {
        size += part.length
        if (size > limit) response.destroy(new GoECSError('DOWNLOAD_TOO_LARGE', 'Download exceeds its size limit'))
        else chunks.push(part)
      })
      response.once('aborted', () => reject(Object.assign(new Error('Download connection reset'), { code: 'ECONNRESET' })))
      response.once('error', reject); response.once('end', () => resolve(Buffer.concat(chunks)))
    })
    request.once('error', reject)
  })
}
export function releaseSources(url) {
  if (url.startsWith('https://api.github.com/repos/oneclickvirt/ecs/releases/')) return [url, url.replace('api.github.com', 'githubapi.spiritlhl.workers.dev'), url.replace('api.github.com', 'githubapi.spiritlhl.top')]
  if (url.startsWith('https://github.com/oneclickvirt/ecs/releases/download/')) return [url, 'https://cdn0.spiritlhl.top/' + url]
  throw new GoECSError('INVALID_ASSET', 'Unexpected release source')
}
export async function boundedFetch(url, limit, signal, proxies = [], { request = requestBytes, retryDelay = 250 } = {}) {
  let failure
  // Download routing is independent from the route selected for measurements.
  for (let attempt = 0; attempt < 2; attempt++) {
    for (const source of releaseSources(url)) for (const proxy of [null, ...proxies]) {
      signal?.throwIfAborted()
      const timeout = AbortSignal.timeout(source.includes('/releases/download/') ? 45000 : 12000), combined = signal ? AbortSignal.any([signal, timeout]) : timeout
      const agent = proxy ? (/^socks/.test(proxy.url) ? new SocksProxyAgent(proxy.url) : new HttpsProxyAgent(proxy.url)) : undefined
      try { return await request(source, limit, combined, agent) }
      catch (error) {
        if (signal?.aborted) throw signal.reason
        if (error instanceof GoECSError && error.code !== 'DOWNLOAD_HTTP') throw error
        failure = error
      } finally { agent?.destroy() }
    }
    if (attempt === 0) await delay(retryDelay, undefined, { signal })
  }
  throw new GoECSError('DOWNLOAD_FAILED', 'Official release and fallback downloads failed (' + (failure?.code || 'NETWORK_ERROR') + '). Retry or select an existing GoECS executable in Settings; any existing engine was preserved')
}
export async function resolveRelease(tag = 'latest', signal, proxies = [], fetchBytes = boundedFetch) {
  if (tag !== 'latest' && !tagPattern.test(tag)) throw new GoECSError('INVALID_VERSION', 'Use latest or an exact stable vX.Y.Z tag')
  const endpoint = 'https://api.github.com/repos/oneclickvirt/ecs/releases/' + (tag === 'latest' ? 'latest' : 'tags/' + tag)
  return JSON.parse((await fetchBytes(endpoint, 2 * 1024 * 1024, signal, proxies)).toString())
}
export async function installEngine(cacheRoot, target, tag = 'latest', signal, { fetchBytes = boundedFetch } = {}) {
  const { proxies } = await localFacts()
  const selected = selectAsset(await resolveRelease(tag, signal, proxies, fetchBytes), target)
  const checksums = (await fetchBytes(selected.checksums.browser_download_url, 1024 * 1024, signal, proxies)).toString()
  const archiveSHA = checksumFor(checksums, selected.asset.name)
  if (selected.asset.digest?.startsWith('sha256:') && selected.asset.digest.slice(7) !== archiveSHA) throw new GoECSError('CHECKSUM_MISMATCH', 'Release metadata and checksum disagree; existing engine was preserved')
  const directory = path.join(cacheRoot, selected.tag, target.platform + '-' + target.arch)
  const binaryPath = path.join(directory, target.platform === 'win32' ? 'goecs.exe' : 'goecs')
  try {
    const record = JSON.parse(await readFile(binaryPath + '.json', 'utf8')), bytes = await readFile(binaryPath)
    if (record.archiveSHA === archiveSHA && record.binarySHA === sha256(bytes) && record.tag === selected.tag) return { ...record, binaryPath }
  } catch {}
  const archive = await fetchBytes(selected.asset.browser_download_url, 256 * 1024 * 1024, signal, proxies)
  if (sha256(archive) !== archiveSHA) throw new GoECSError('CHECKSUM_MISMATCH', 'GoECS archive failed SHA-256 verification; existing engine was preserved')
  const allowed = new Set(['goecs', 'goecs.exe', selected.asset.name.slice(0, -4), selected.asset.name.slice(0, -4) + '.exe'])
  const files = unzipSync(archive, { filter: file => {
    if (file.name.includes('..') || file.name.startsWith('/') || file.name.includes('\\')) throw new GoECSError('INVALID_ARCHIVE', 'Unsafe archive entry')
    if (!allowed.has(path.posix.basename(file.name))) return false
    if (file.originalSize > 512 * 1024 * 1024) throw new GoECSError('INVALID_ARCHIVE', 'Executable exceeds its size limit')
    return true
  } })
  const entries = Object.entries(files)
  if (entries.length !== 1) throw new GoECSError('INVALID_ARCHIVE', 'Expected exactly one GoECS executable')
  const bytes = Buffer.from(entries[0][1]), record = { tag: selected.tag, asset: selected.asset.name, archiveSHA, binarySHA: sha256(bytes) }
  await mkdir(directory, { recursive: true, mode: 0o700 })
  const temp = binaryPath + '.' + randomUUID() + '.tmp'
  try {
    await writeFile(temp, bytes, { mode: 0o700, flag: 'wx' }); await chmod(temp, 0o700); await rename(temp, binaryPath)
    await writeFile(binaryPath + '.json', JSON.stringify(record), { mode: 0o600 })
  } finally { await rm(temp, { force: true }) }
  return { ...record, binaryPath }
}
export async function verifyExecutable(binaryPath) {
  if (!path.isAbsolute(binaryPath) || /[\x00-\x1f]/.test(binaryPath)) throw new GoECSError('INVALID_BINARY', 'Select an absolute executable path')
  const value = await stat(binaryPath)
  if (!value.isFile()) throw new GoECSError('INVALID_BINARY', 'GoECS executable is not a regular file')
  const binarySHA = sha256(await readFile(binaryPath))
  let tag = 'configured'
  try {
    const record = JSON.parse(await readFile(binaryPath + '.json', 'utf8'))
    if (record.binarySHA === binarySHA && tagPattern.test(record.tag)) tag = record.tag
  } catch {}
  return { binaryPath, binarySHA, tag }
}

// Reuse verified downloads offline before looking for a user's PATH installation.
export async function discoverEngines(cacheRoot, target, env = process.env) {
  const candidates = [], executable = target.platform === 'win32' ? 'goecs.exe' : 'goecs'
  const tags = (await readdir(cacheRoot).catch(() => [])).filter(tag => tagPattern.test(tag)).sort((a, b) => b.localeCompare(a, 'en', { numeric: true }))
  for (const tag of tags) {
    const binaryPath = path.join(cacheRoot, tag, target.platform + '-' + target.arch, executable)
    try {
      const descriptor = await verifyExecutable(binaryPath)
      if (descriptor.tag === tag) candidates.push(binaryPath)
    } catch {}
  }
  const dirs = String(env.PATH || env.Path || '').split(path.delimiter)
  if (target.platform !== 'win32') dirs.push(path.join(os.homedir(), '.local', 'bin'), '/usr/local/bin', '/usr/bin')
  for (const dir of dirs) {
    if (!path.isAbsolute(dir)) continue
    const binaryPath = path.join(dir, executable)
    try { if ((await stat(binaryPath)).isFile()) candidates.push(binaryPath) } catch {}
  }
  return [...new Set(candidates)]
}
