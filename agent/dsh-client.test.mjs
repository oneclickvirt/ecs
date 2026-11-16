import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import vm from 'node:vm'
import React, { act } from 'react'
import { JSDOM } from 'jsdom'
import { parseHelp, MENU } from './dsh-options.mjs'
import { parseReport, exportReport } from './dsh-report.mjs'

const pause = ms => new Promise(resolve => setTimeout(resolve, ms))
test('page preparation, usable parameters, incremental sections and navigation retention', { timeout: 20000 }, async t => {
  const dom = new JSDOM('<div id="root"></div>', { url: 'http://localhost/', runScripts: 'outside-only' }), { window } = dom
  const oldWindow = globalThis.window, oldDocument = globalThis.document
  globalThis.window = window; globalThis.document = window.document; globalThis.IS_REACT_ACT_ENVIRONMENT = true
  const { createRoot } = await import('react-dom/client')
  const h = React.createElement, slots = new Map(), cleanups = []
  let plugin, dictionaries, selected = true, task, submitted, enginePath = '', preparationCalls = 0
  const settings = { showSidebar: true, requestAdmin: false, networkMode: 'direct', binaryPath: '', maxTaskSeconds: 3600 }
  const flags = parseHelp('  -menu\n    Menu\n  -json string\n    JSON\n  -cpu-method string\n    CPU method (default "sysbench")\n  -nt3-location string\n    Location (default "GZ")\n  -upload\n    Upload (default true)\n  -privacy\n    Private mode\n')
  window.ResizeObserver = class { observe() {} disconnect() {} }
  window.HTMLElement.prototype.getBoundingClientRect = () => ({ top: 40, left: 220, width: 1000, height: 850 })
  window.__ModuleLoader__ = { load: module => { plugin = module.factory(() => React) } }
  window.fetch = async (_url, init) => {
    const input = JSON.parse(init.body); let data
    if (input.action === 'status') data = { settings: { ...settings, binaryPath: enginePath }, targets: [{ id: 'local' }], tasks: task ? [task] : [], menus: MENU }
    else if (input.action === 'prepare') { preparationCalls++; if (preparationCalls === 1) return Response.json({ error: { code: 'DOWNLOAD_FAILED', message: 'Simulated ECONNRESET' } }, { status: 400 }); enginePath = '/verified/goecs'; data = { binaryPath: enginePath, flags } }
    else if (input.action === 'start') { submitted = input.test; task = { id: 'live-task', choice: submitted.choice, status: 'running', startedAt: new Date().toISOString(), raw: '----系统基础信息----\nCPU: example\n' }; data = { task } }
    else if (input.action === 'task') data = { ...task, raw: undefined, output: task.raw.slice(input.offset), nextOffset: task.raw.length, length: task.raw.length, reportReady: task.status === 'completed' }
    else if (input.action === 'report') { const document = parseReport(task.raw, { status: task.status }); data = { document: input.format === 'json' ? document : undefined, content: exportReport(document, input.format) } }
    else if (input.action === 'history') data = []
    else throw new Error('Unexpected action ' + input.action)
    return Response.json({ data })
  }
  vm.runInContext(await readFile(new URL('./dsh-client.js', import.meta.url), 'utf8'), dom.getInternalVMContext())
  plugin.apply({ effect: fn => cleanups.push(fn()), locale: { register: (_key, value) => { dictionaries = value; return () => {} }, bind: () => key => key.split('.').reduce((value, part) => value?.[part], dictionaries.zh) || key }, slots: { inject: (_slot, fn) => fn(), register: (options, component) => { slots.set(options.name, component); return () => {} } }, layout: { selectPanel: () => {} } })
  const root = createRoot(window.document.getElementById('root'))
  const render = () => { const Seat = slots.get('main'), Keeper = slots.get('shell.overlay'); root.render(h(React.Fragment, null, selected ? h(Seat) : h('div', null, 'Other DSH page'), h(Keeper, { t: key => key.split('.').reduce((value, part) => value?.[part], dictionaries.zh) || key }))) }
  t.after(async () => { await act(async () => root.unmount()); for (const cleanup of cleanups) if (typeof cleanup === 'function') cleanup(); window.close(); globalThis.window = oldWindow; globalThis.document = oldDocument; delete globalThis.IS_REACT_ACT_ENVIRONMENT })
  const until = async predicate => { for (let i = 0; i < 100; i++) { await act(async () => { await pause(30) }); if (predicate()) return } assert.fail('UI condition did not become true') }
  const button = text => [...window.document.querySelectorAll('button')].find(node => node.textContent === text)
  await act(async () => render())
  await until(() => button('重试准备本体'))
  assert.equal(window.document.querySelector('button.ge-primary').disabled, true)
  assert.equal(submitted, undefined)
  await act(async () => button('重试准备本体').click())
  await until(() => window.document.querySelector('[data-goecs-option="cpu-method"]'))
  assert.ok(preparationCalls >= 1)
  assert.equal(window.document.querySelector('[data-goecs-binary]').textContent, enginePath)
  const cpu = window.document.querySelector('[data-goecs-option="cpu-method"]')
  assert.equal(cpu.tagName, 'SELECT'); assert.ok([...cpu.options].some(item => item.value === 'winsat'))
  assert.ok([...window.document.querySelectorAll('label')].find(node => node.textContent === '公开上传并生成分享链接').querySelector('input').checked)
  await act(async () => { cpu.value = 'winsat'; cpu.dispatchEvent(new window.Event('change', { bubbles: true })) })
  await act(async () => window.document.querySelector('button.ge-primary').click())
  await until(() => window.document.querySelector('.ge-report h2'))
  assert.equal(submitted.choice, '1'); assert.equal(submitted.options['cpu-method'], 'winsat')
  assert.equal(task.status, 'running'); assert.equal(window.document.querySelector('.ge-report h2').textContent, '系统基础信息')
  assert.equal(window.document.querySelectorAll('.ge-report h2').length, 1)
  const retained = window.document.querySelector('[data-goecs-panel]')
  await act(async () => { selected = false; render() })
  assert.equal(window.document.querySelector('[data-goecs-keeper]').hidden, true)
  task.raw += '----CPU测试----\n分数: 123\n'
  await until(() => window.document.querySelectorAll('.ge-report h2').length === 2)
  await act(async () => { selected = true; render() })
  assert.equal(window.document.querySelector('[data-goecs-panel]'), retained)
  assert.deepEqual([...window.document.querySelectorAll('.ge-report h2')].map(node => node.textContent), ['系统基础信息', 'CPU测试'])
  assert.equal(task.status, 'running')
  await act(async () => button('开始测试').click())
  assert.equal(window.document.querySelector('[data-goecs-option="cpu-method"]').value, 'winsat')
  await act(async () => button('测试报告').click())
  task.raw += '分享: https://paste.spiritlhl.net/example\n'; task.status = 'completed'
  await until(() => window.document.querySelector('a[href="https://paste.spiritlhl.net/example"]'))
  assert.ok(window.document.querySelector('.ge-terminal').textContent.includes('分数: 123'))
})
