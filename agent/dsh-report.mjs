import { createHash } from 'node:crypto'

export const REPORT_SCHEMA = 'goecs.reader/1'
export const stripTerminal = value => String(value).replace(/\x1b\][\s\S]*?(?:\x07|\x1b\\)/g, '').replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, '').replace(/\r\n/g, '\n').replace(/\r/g, '\n').replace(/[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]/g, '')
export const escapeHTML = value => String(value).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c])
const categoryRules = [
  ['system', /系统|基础信息|system|basic info|virtualization/i], ['cpu', /CPU|geekbench/i],
  ['memory', /内存|memory|\bstream\b/i], ['disk', /硬盘|磁盘|disk|fio|I\/O/i],
  ['unlock', /解锁|跨国|流媒体|unlock|media|multination/i], ['ip', /IP质量|IP 质量|IP quality|security|欺诈|风险/i],
  ['mail', /邮件|email|SMTP/i], ['trace', /详细路由|回程路由|NT3|traceroute|nexttrace|route trace/i],
  ['backroute', /回程|上游|backtrace|backhaul|upstream/i], ['speed', /测速|speed|iperf|带宽/i],
  ['latency', /延迟|latency|ping|握手|TCP|TGDC/i],
]
const category = title => categoryRules.find(([, re]) => re.test(title))?.[0] || 'other'
export const shareURLs = text => [...new Set((stripTerminal(text).match(/https:\/\/paste\.spiritlhl\.net\/[^\s<>"'\]\)]+/g) || []).filter(value => !value.includes('/api/')))]
const decoration = /^[\s─━═=*_·.\-+|┌┐└┘├┤┬┴┼╭╮╰╯]{3,}$/u
function heading(line) {
  const text = line.trim()
  if (decoration.test(text)) return null
  const match = text.match(/^(?:[-─━═=*_]{3,}\s*)(.+?)(?:\s*[-─━═=*_]{3,})$/u)
  if (match) return match[1].trim()
  return text.match(/^#{1,4}\s+(.+)$/)?.[1] || null
}
function fields(line, number) {
  // Split explicit blacklist counters, never IPv6, URLs or nested provider data.
  const counters = [...line.matchAll(/(无害记录数|恶意记录数|可疑记录数|无记录数|harmless records|malicious records|suspicious records|undetected records)\s*[:：]\s*(\d+(?:\s*\[[^\]]+\])?)/gi)]
  if (counters.length > 1 && !line.replace(/(无害记录数|恶意记录数|可疑记录数|无记录数|harmless records|malicious records|suspicious records|undetected records)\s*[:：]\s*\d+(?:\s*\[[^\]]+\])?/gi, '').trim()) {
    return counters.map(m => ({ name: m[1], value: m[2], sourceLine: number }))
  }
  if (/^(?:https?:\/\/|(?:\d+\s+)?(?:[0-9a-f]{0,4}:){2,})/i.test(line.trim())) return null
  const match = line.match(/^\s*([^:：\n]{1,90}?)\s*[:：]\s*(\S.*?)\s*$/)
  return match ? [{ name: match[1].trim(), value: match[2], sourceLine: number }] : null
}
const cells = line => line.trim().replace(/^\||\|$/g, '').split(/\s*\|\s*|\t+| {2,}/).map(s => s.trim())
const tableDivider = line => /^\s*\|?\s*:?-{2,}:?\s*(?:\|\s*:?-{2,}:?\s*)+\|?\s*$/.test(line)

export function parseReport(rawText, metadata = {}) {
  if (typeof rawText !== 'string' || Buffer.byteLength(rawText) > 16 * 1024 * 1024) throw new Error('Report must be text of at most 16 MiB')
  const lines = stripTerminal(rawText).split('\n')
  const result = { schemaVersion: REPORT_SCHEMA, source: { sha256: createHash('sha256').update(rawText).digest('hex'), bytes: Buffer.byteLength(rawText), lines: lines.length }, metadata: { ...metadata }, shareURLs: shareURLs(rawText), sections: [], warnings: [], rawText }
  // PTY presets include the interactive menu before the actual report. Keep
  // the transcript byte-exact in TXT/JSON, but start the reader at its report.
  const reportStart = lines.findIndex(line => /^(?:VPS融合怪测试|VPS Fusion Monster Test)$/i.test(heading(line) || ''))
  const firstLine = reportStart > 0 && lines.slice(0, reportStart).some(line => /请选择测试方案|融合怪完全体|Full Suite/i.test(line)) ? reportStart : 0
  let section, block
  const ensureSection = (line, title = '') => {
    if (!section || title) {
      section = { title, category: category(title), sourceStart: line, sourceEnd: line, blocks: [] }
      result.sections.push(section); block = null
    }
    return section
  }
  for (let i = firstLine; i < lines.length; i++) {
    const line = lines[i], n = i + 1, trimmed = line.trim()
    if (!trimmed || decoration.test(trimmed)) { block = null; continue }
    const title = heading(line)
    if (title) { ensureSection(n, title); continue }
    ensureSection(n); section.sourceEnd = n
    const version = trimmed.match(/(?:融合怪.*?版本|Fusion Monster.*?Version|goecs(?: version)?|版本|Version)\s*[:：]\s*(v?\d[\w.+-]*)/i)
    if (version && !result.metadata.engineVersion) result.metadata.engineVersion = version[1]
    if (/^(?:\[?IPv[46]\]?|测试IP类型\s*[:：]\s*IPv[46])\s*[:：]?$/i.test(trimmed)) {
      block = { kind: 'text', sourceStart: n, sourceEnd: n, lines: [line] }; section.blocks.push(block); continue
    }
    const next = lines[i + 1] || '', row = cells(line)
    const hasTable = row.length > 1 && (line.includes('|') || line.includes('\t') || / {2,}/.test(line)) && (tableDivider(next) || (cells(next).length === row.length && !fields(line, n) && !fields(next, n + 1)))
    if (hasTable) {
      block = { kind: 'table', sourceStart: n, sourceEnd: n, columns: row, rows: [] }; section.blocks.push(block)
      if (tableDivider(next)) i++
      while (i + 1 < lines.length && lines[i + 1].trim() && !heading(lines[i + 1]) && !decoration.test(lines[i + 1].trim())) {
        const values = cells(lines[i + 1])
        if (values.length !== row.length) break
        block.rows.push({ values, sourceLine: i + 2 }); block.sourceEnd = i + 2; section.sourceEnd = i + 2; i++
      }
      block = null; continue
    }
    const extracted = fields(line, n)
    if (extracted) {
      block = { kind: 'fields', sourceStart: n, sourceEnd: n, fields: extracted }; section.blocks.push(block); continue
    }
    if (!block || block.kind !== 'text') {
      block = { kind: 'text', sourceStart: n, sourceEnd: n, lines: [] }; section.blocks.push(block)
    }
    block.lines.push(line); block.sourceEnd = n
  }
  if (!result.sections.length) result.warnings.push('No recognizable report content; original text is preserved.')
  return result
}
const md = s => String(s).replace(/\\/g, '\\\\').replace(/[|`*_<>\[\]]/g, '\\$&').replace(/\n/g, ' ')
function fence(text) {
  const longest = Math.max(2, ...[...text.matchAll(/`+/g)].map(m => m[0].length))
  const delimiter = '`'.repeat(longest + 1)
  return delimiter + 'text\n' + text + '\n' + delimiter
}
export function reportMarkdown(report) {
  const out = ['# GoECS', '']
  for (const url of report.shareURLs || []) out.push('[GoECS](' + url + ')', '')
  for (const section of report.sections) {
    if (section.title) out.push('## ' + md(section.title), '')
    for (const b of section.blocks) {
      if (b.kind === 'fields') out.push(...b.fields.map(f => '- **' + md(f.name) + ':** ' + md(f.value)), '')
      else if (b.kind === 'table') out.push('| ' + b.columns.map(md).join(' | ') + ' |', '| ' + b.columns.map(() => '---').join(' | ') + ' |', ...b.rows.map(r => '| ' + r.values.map(md).join(' | ') + ' |'), '')
      else out.push(fence(b.lines.join('\n')), '')
    }
  }
  return out.join('\n')
}
export function reportHTML(report) {
  const body = report.sections.map(section => '<section>' + (section.title ? '<h2>' + escapeHTML(section.title) + '</h2>' : '') + section.blocks.map(b => {
    if (b.kind === 'fields') return '<dl>' + b.fields.map(f => '<div><dt>' + escapeHTML(f.name) + '</dt><dd>' + escapeHTML(f.value) + '</dd></div>').join('') + '</dl>'
    if (b.kind === 'table') return '<div class="table"><table><thead><tr>' + b.columns.map(c => '<th>' + escapeHTML(c) + '</th>').join('') + '</tr></thead><tbody>' + b.rows.map(r => '<tr>' + r.values.map(c => '<td>' + escapeHTML(c) + '</td>').join('') + '</tr>').join('') + '</tbody></table></div>'
    return '<pre>' + escapeHTML(b.lines.join('\n')) + '</pre>'
  }).join('') + '</section>').join('')
  const links = (report.shareURLs || []).map(url => '<p><a href="' + escapeHTML(url) + '" rel="noreferrer noopener">' + escapeHTML(url) + '</a></p>').join('')
  return '<!doctype html><html lang="' + (report.metadata.language === 'en' ? 'en' : 'zh') + '"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta http-equiv="Content-Security-Policy" content="default-src &#39;none&#39;; style-src &#39;unsafe-inline&#39;"><title>GoECS</title><style>body{color:#1e293b;background:#fff;font:15px/1.7 system-ui,sans-serif;max-width:1080px;margin:auto;padding:28px}section{margin:22px 0}h1,h2{line-height:1.35}h2{font-size:20px;border-bottom:1px solid #e2e8f0;padding-bottom:12px}dl>div{display:flex;gap:20px;border-bottom:1px solid #f1f5f9;padding:8px 0}dt{min-width:140px;color:#64748b}dd{margin:0;overflow-wrap:anywhere}pre{white-space:pre-wrap;overflow-wrap:anywhere;background:#f8fafc;padding:16px;border-radius:10px}.table{overflow:auto}table{border-collapse:collapse;width:100%}th,td{border:1px solid #e2e8f0;text-align:left;padding:9px;white-space:pre-wrap}th{background:#f8fafc}@media(prefers-color-scheme:dark){body{background:#101827;color:#e2e8f0}pre,th{background:#172334}dt{color:#a6b6cb}h2,dl>div,th,td{border-color:#334155}}</style></head><body><h1>GoECS</h1>' + links + body + '</body></html>'
}
export function exportReport(report, format) {
  if (format === 'txt') return report.rawText
  if (format === 'json') return JSON.stringify(report, null, 2)
  if (format === 'md') return reportMarkdown(report)
  if (format === 'html') return reportHTML(report)
  throw new Error('Choose txt, md, html or json')
}
