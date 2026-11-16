import test from 'node:test'
import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { parseReport, reportHTML, reportMarkdown, exportReport, stripTerminal } from './dsh-report.mjs'

test('source remains byte-exact across four exports; metrics retain units, zero and unknown values', () => {
  const raw = '\x1b[32m----系统基础信息----\x1b[0m\r\nCPU: Example: cache 64 MiB\r\nIPv6: 2001:db8::1\r\n----CPU测试--通过sysbench测试----\n单线程得分: 0\n多线程得分: unavailable\n----硬盘测试-通过fio测试----\n读速度: 921.00 MiB/s\n未知记录 保留: 原始值\n'
  const doc = parseReport(raw)
  assert.equal(exportReport(doc, 'txt'), raw)
  assert.equal(doc.source.sha256, createHash('sha256').update(raw).digest('hex'))
  assert.equal(JSON.parse(exportReport(doc, 'json')).rawText, raw)
  const output = reportMarkdown(doc)
  for (const value of ['cache 64 MiB', '2001:db8::1', '**单线程得分:** 0', 'unavailable', '921.00 MiB/s', '未知记录']) assert.ok(output.includes(value), value)
  assert.equal(doc.sections[1].title, 'CPU测试--通过sysbench测试')
  assert.ok(doc.sections[1].blocks.every(b => b.sourceStart >= 5))
})
test('decorative lines disappear only from derived views; repeated actual measurements stay distinct', () => {
  const raw = '------\n----内存测试----\n读取: 123 MB/s\n读取: 456 MB/s\n======\n----内存测试----\n读取: 789 MB/s'
  const doc = parseReport(raw), md = reportMarkdown(doc)
  assert.equal(doc.sections.length, 2)
  assert.equal((md.match(/读取/g) || []).length, 3)
  assert.ok(!md.includes('======'))
  assert.equal(exportReport(doc, 'txt'), raw)
})
test('old Shell tables, unparseable rows and IPv4/IPv6 context remain associated', () => {
  const raw = '----三网回程路由----\nIPv4\n1  192.0.2.1  1.00 ms\n2  *  timeout\nIPv6\n1  2001:db8::1  unavailable\n----网络测速----\n| 节点 | 下载 | 上传 |\n| --- | --- | --- |\n| Example | 0 Mbps | failed |\nmalformed row stays here\n'
  const doc = parseReport(raw), out = reportMarkdown(doc)
  for (const text of ['IPv4', 'IPv6', '2001:db8::1', 'timeout', 'malformed row stays here', '0 Mbps', 'failed']) assert.ok(out.includes(text), text)
  const table = doc.sections[1].blocks.find(b => b.kind === 'table')
  assert.deepEqual(table.rows[0].values, ['Example', '0 Mbps', 'failed'])
})
test('inline counters split safely and nested URLs, IPv6 and annotations retain values', () => {
  const doc = parseReport('----IP质量----\n恶意记录数: 0 [1] 无害记录数: 72 [1]\n链接: https://example.test/a:b\n2001:db8::2\nNetflix: Yes [US: IPv6]')
  const fields = doc.sections[0].blocks.flatMap(b => b.fields || [])
  assert.equal(fields[0].value, '0 [1]'); assert.equal(fields[1].value, '72 [1]')
  assert.equal(fields[2].value, 'https://example.test/a:b')
  assert.equal(fields.at(-1).value, 'Yes [US: IPv6]')
})
test('HTML has no active imported content, tracking resources or script execution', () => {
  const html = reportHTML(parseReport('----系统----\n<script>alert(1)</script>\n<img src=https://example.test/pixel onerror=alert(1)>\nName: <iframe src=x>'))
  assert.ok(!html.includes('<script>')); assert.ok(!html.includes('<img ')); assert.ok(!html.includes('<iframe '))
  assert.ok(html.includes('&lt;script&gt;')); assert.ok(html.includes('Content-Security-Policy'))
})
test('partial output and unrecognized versions remain readable without invented completion', () => {
  const doc = parseReport('VPS融合怪版本: v0.2.18\n----硬盘测试----\n测试仍在运行…', { status: 'running' })
  assert.equal(doc.metadata.engineVersion, 'v0.2.18'); assert.equal(doc.metadata.status, 'running')
  assert.ok(reportMarkdown(doc).includes('测试仍在运行'))
  assert.equal(stripTerminal('\x1b]11;rgb:fff\x1b\\abc\r\n'), 'abc\n')
  assert.throws(() => exportReport(doc, 'pdf'))
})
test('upstream routes and TCP throughput are not misclassified as STREAM memory or latency', () => {
  const doc = parseReport('----Upstream routes----\nExample\n----STREAM memory----\nExample\n----TCP speed test----\nExample')
  assert.deepEqual(doc.sections.map(s => s.category), ['backroute', 'memory', 'speed'])
})
test('live PTY readers start at the actual report while retaining the complete transcript', () => {
  const raw = '^[]11;rgb:1111/1818/2727^[\\\n请选择测试方案:\n1. 融合怪完全体\n----VPS融合怪测试----\n版本: v0.2.23\n----系统基础信息----\nCPU: example\n'
  const doc = parseReport(raw, { status: 'running' })
  assert.equal(exportReport(doc, 'txt'), raw)
  assert.deepEqual(doc.sections.map(s => s.title), ['VPS融合怪测试', '系统基础信息'])
  assert.equal(doc.sections[0].sourceStart, 4)
  assert.ok(!reportHTML(doc).includes('请选择测试方案'))
  assert.ok(!reportMarkdown(doc).includes('^[]11;rgb'))
})
