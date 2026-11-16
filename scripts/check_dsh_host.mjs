import assert from 'node:assert/strict'
import path from 'node:path'
import os from 'node:os'
import { mkdtemp, rm } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { pathToFileURL } from 'node:url'
import * as Plugin from '../agent/dsh-plugin.mjs'
const root = process.argv[2]
if (!root) throw new Error('Pass an isolated DSH runtime directory containing node_modules')
const require = createRequire(path.join(path.resolve(root), 'package.json'))
const load = name => import(pathToFileURL(require.resolve(name)))
const [{ Context }, prompt, tools, executor] = await Promise.all(['@deepseek-ai/cordis', '@deepseek-ai/dsh-system-prompt', '@deepseek-ai/dsh-tools', '@deepseek-ai/dsh-tool-executor-native'].map(load))
const directory = await mkdtemp(path.join(os.tmpdir(), 'goecs-real-dsh-')), ctx = new Context()
try {
  await ctx.plugin(prompt.default)
  await ctx.plugin(tools.default)
  await ctx.plugin(executor)
  await ctx.plugin(Plugin, { stateDir: directory })
  const schemas = ctx.tools.schemas()
  assert.deepEqual(schemas.map(s => s.name).sort(), ['goecs_info', 'goecs_run', 'goecs_task'])
  const response = await ctx.tools.execute({ calls: [{ type: 'tool-call', id: 'goecs-test-info', name: 'goecs_info', arguments: JSON.stringify({ action: 'status' }) }], signal: new AbortController().signal })
  const result = response.results[0].result
  assert.equal(result.isError, false)
  assert.equal(result.value.uiOptional, true)
  assert.equal(result.value.host.tools, true)
  assert.equal(result.value.host.connection, false)
  assert.equal(result.value.tasks.length, 0)
  assert.equal(ctx.get('connection'), undefined)
  const invalid = await ctx.tools.execute({ calls: [{ type: 'tool-call', id: 'goecs-test-invalid', name: 'goecs_info', arguments: JSON.stringify({ action: 'invalid' }) }], signal: new AbortController().signal })
  assert.equal(invalid.results[0].result.isError, true)
  console.log(JSON.stringify({ runtime: require('@deepseek-ai/dsh-tools/package.json').version, nativeTools: schemas.length, noWebRequired: true, validation: 'passed' }))
} finally {
  await ctx.fiber.dispose()
  await rm(directory, { recursive: true, force: true })
}
