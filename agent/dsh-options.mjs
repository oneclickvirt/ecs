import { stripTerminal } from './dsh-report.mjs'

export const MENU = Object.freeze([
  ['1', '融合怪完全体', 'Full suite'], ['2', '融合怪并发完全体', 'Full concurrent suite'],
  ['3', '极简版', 'Minimal suite'], ['4', '精简版', 'Standard suite'], ['5', '精简网络版', 'Network suite'],
  ['6', '精简解锁版', 'Unlock suite'], ['7', '网络单项', 'Network only'], ['8', '解锁单项', 'Unlock only'],
  ['9', '硬件单项', 'Hardware only'], ['10', 'IP 质量检测', 'IP quality'], ['11', '三网回程线路', 'Three-network route'],
  ['custom', '高级自定义', 'Custom tests'],
])
const managed = new Set(['h', 'help', 'v', 'version', 'menu', 'json', 'l', 'lang', 'ut-http-proxy', 'ut-socks-proxy'])
const aliases = { cpum: 'cpu-method', cput: 'cpu-thread', memorym: 'memory-method', diskm: 'disk-method', nt3loc: 'nt3-location', nt3t: 'nt3-type', analysis: 'analyze' }
const choices = {
  'cpu-method': ['sysbench', 'geekbench', 'winsat'], 'cpu-thread': ['single', 'multi'],
  'memory-method': ['stream', 'sysbench', 'dd', 'winsat', 'auto'], 'disk-method': ['fio', 'dd', 'winsat'],
  'dns-mode': ['auto', 'system', 'doh', 'dot'], 'nt3-location': ['GZ', 'SH', 'BJ', 'CD', 'ALL'], 'nt3-type': ['both', 'ipv4', 'ipv6'],
  'ping-scope': ['auto', 'china', 'international'], 'ping-sort': ['latency', 'name'], 'tcp-format': ['compact', 'full'], 'tcp-sort': ['name', 'latency'],
  utipver: ['auto', 'ipv4', 'ipv6'], utregion: Array.from({ length: 23 }, (_, n) => String(n)),
}
const optionGroup = name => /^(?:cpu|memory|disk|hardware)/.test(name) ? 'hardware' : name.startsWith('deep') ? 'deep' : /^(?:ut|security|email)/.test(name) ? 'unlock' : /^(?:nt3|backtrace|ping|tcp|tgdc|web)/.test(name) ? 'network' : /^(?:speed|spnum)/.test(name) ? 'speed' : 'general'
export function parseHelp(raw) {
  const flags = []
  for (const line of stripTerminal(raw).split('\n')) {
    const m = line.match(/^\s{0,4}--?([\w-]+)(?:\s+(string|int|duration|float|uint))?(?:\s*$|\t(.+)$)/)
    if (m) flags.push({ name: m[1], type: m[2] || 'bool', description: m[3] || '', managed: managed.has(m[1]), group: optionGroup(m[1]), ...(aliases[m[1]] ? { aliasOf: aliases[m[1]] } : {}), ...(choices[m[1]] ? { choices: choices[m[1]] } : {}) })
    else if (flags.length && line.trim()) flags.at(-1).description += (flags.at(-1).description ? ' ' : '') + line.trim()
  }
  return flags.map(flag => ({ ...flag, defaultValue: flag.description.match(/\(default (?:"([^"]*)"|([^)]*))\)/)?.slice(1).find(value => value !== undefined) ?? (flag.type === 'bool' ? 'false' : '') }))
}
export function validateOptions(options, flags) {
  if (!options || typeof options !== 'object' || Array.isArray(options)) throw new Error('Options must be an object')
  if (Object.keys(options).length > 90) throw new Error('Too many options')
  const result = [], catalog = new Map(flags.map(f => [f.name, f]))
  for (const [name, value] of Object.entries(options)) {
    const flag = catalog.get(name)
    if (!flag || managed.has(name)) throw new Error('Unsupported or managed option: ' + name)
    if (flag.type === 'bool' && typeof value !== 'boolean') throw new Error(name + ' must be boolean')
    if (flag.type !== 'bool' && typeof value !== 'string' && typeof value !== 'number') throw new Error(name + ' must be text or a number')
    if (String(value).length > 2048 || /[\x00-\x1f\x7f]/.test(String(value))) throw new Error('Invalid option value: ' + name)
    if (flag.type === 'int' && !/^-?\d+$/.test(String(value))) throw new Error(name + ' must be an integer')
    result.push('-' + name + '=' + value)
  }
  return result
}
export function buildArguments(input, flags, reportPath) {
  const choice = String(input.choice || '1')
  if (!MENU.some(item => item[0] === choice)) throw new Error('Select a test from the original menu')
  const language = input.language === 'en' ? 'en' : 'zh'
  const args = ['-lang=' + language, '-upload=' + (input.options?.upload !== false && input.options?.privacy !== true), '-log=false', '-json=' + reportPath]
  if (choice === 'custom') args.push('-menu=false')
  args.push(...validateOptions(input.options || {}, flags))
  return { args, choice, language }
}

export function menuDriver(choice, write) {
  let selected = choice === 'custom', buffer = '', started = selected
  return data => {
    buffer = (buffer + stripTerminal(data)).slice(-65536)
    if (!selected && /请输入选项|Please enter your choice/.test(buffer)) {
      selected = true; write(choice + '\r'); buffer = ''; return
    }
    if (!selected && /融合怪完全体|Full Suite/.test(buffer) && /高级自定义|Advanced Custom/.test(buffer)) {
      selected = true
      // The original Go TUI owns all presets. Move within its verified menu.
      write('\x1b[B'.repeat(Number(choice) - 1) + '\r'); buffer = ''; return
    }
    if (selected && !started && /硬盘测试|CPU测试|系统基础|Basic System|CPU Test|硬件|IP质量|跨国|测试IP类型/.test(buffer)) started = true
    if (selected && /使用中国专项测试.*?[?？]|Use China-specific.*?[?？]|是否使用中国.*?[?？]|是否使用.*中国专项|Use China.*\(yes\/no\)/is.test(buffer)) {
      write('\r'); buffer = ''
    }
    if (/Press Enter to exit|按.*回车.*退出/i.test(buffer)) { write('\r'); buffer = '' }
  }
}
