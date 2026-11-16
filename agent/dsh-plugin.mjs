import { GoECSService, publicError } from './dsh-service.mjs'
import { GoECSError } from './dsh-platform.mjs'

// Services are optional: headless/CLI profiles do not need the Web connection.
export function apply(ctx, config = {}) {
  if (typeof ctx?.effect !== 'function' || typeof ctx?.inject !== 'function') throw new GoECSError('DSH_INCOMPATIBLE', 'This DSH runtime lacks the plugin lifecycle API. Use a supported DSH version or the standalone goecs-dsh CLI.')
  const service = new GoECSService(config)
  service.host = { tools: false, connection: false, uiOptional: true }
  ctx.effect(() => () => service.dispose())
  ctx.inject(['connection'], scope => {
    if (!scope.connection?.fetch?.register) return
    service.host.connection = true
    scope.connection.fetch.register({ path: '/api/oneclickvirt-ecs', methods: ['POST'], requestBody: 'buffered', async fetch(request) {
      try {
        const text = await request.text()
        if (text.length > 17 * 1024 * 1024) throw new GoECSError('REQUEST_LIMIT', 'Request is too large')
        const input = JSON.parse(text)
        let value
        if (input.action === 'status') value = await service.status()
        else if (input.action === 'settings') value = await service.saveSettings(input.settings)
        else if (input.action === 'saveTarget') value = await service.saveTarget(input.target)
        else if (input.action === 'removeTarget') value = await service.removeTarget(input.targetId)
        else if (input.action === 'inspect') value = await service.inspect(input.targetId, request.signal)
        else if (input.action === 'install') value = await service.install(input.targetId, input.version, request.signal)
        else if (input.action === 'prepare') value = await service.prepare()
        else if (input.action === 'options') value = await service.options(input.targetId)
        else if (input.action === 'start') value = await service.start(input.test || {}, { interactive: true, signal: request.signal })
        else if (input.action === 'continue') value = await service.accept(input.token, input.decisions, { user: true })
        else if (input.action === 'task') value = service.view(service.getTask(input.id), Number.isInteger(input.offset) && input.offset >= 0 ? input.offset : 0)
        else if (input.action === 'stop') value = await service.stop(input.id)
        else if (input.action === 'report') value = await service.report(input.id, input.format)
        else if (input.action === 'import') value = await service.importReport(input.text)
        else if (input.action === 'history') value = await service.history()
        else if (input.action === 'loadHistory') value = await service.loadHistory(input.id)
        else throw new GoECSError('UNKNOWN_ACTION', 'Unknown plugin operation')
        return Response.json({ data: value }, { headers: { 'Cache-Control': 'no-store' } })
      } catch (error) {
        return Response.json({ error: publicError(error) }, { status: error instanceof GoECSError ? 400 : 500, headers: { 'Cache-Control': 'no-store' } })
      }
    } })
  })
  ctx.inject(['tools'], scope => {
    if (typeof scope.tools?.register !== 'function') return
    service.host.tools = true
    registerTools(scope, service)
  })
}

function registerTools(ctx, service) {
  const owner = exec => String(exec.agent?.session?.id || exec.agent?.sessionId || exec.sessionId || 'headless')
  const tools = [
    { name: 'goecs_info', description: 'Inspect the GoECS plugin, original test-menu choices, configured targets and installed engine flags. Never infer that every probe uses an HTTP/SOCKS proxy. No benchmark starts.', parameters: { action: { type: 'string', enum: ['status', 'target', 'options'], required: true }, targetId: { type: 'string' } }, run: async (args, exec) => args.action === 'target' ? service.inspect(args.targetId, exec.signal) : args.action === 'options' ? service.options(args.targetId) : service.status(owner(exec)), safe: true },
    { name: 'goecs_run', description: 'Start a user-requested GoECS test on an explicit target using the original Go menu preset. Menu 2 retains full concurrency. Public report upload is enabled by default; set options.upload=false or options.privacy=true to disable it. A detected proxy requires the user to choose its route. This returns a background task id; inspect with goecs_task and stop explicitly. Privilege reminders, credentials and SSH trust are user-only in the plugin page.', parameters: { targetId: { type: 'string' }, choice: { type: 'string', enum: ['1','2','3','4','5','6','7','8','9','10','11','custom'], required: true }, language: { type: 'string', enum: ['zh','en'] }, networkMode: { type: 'string', enum: ['direct','proxy'] }, options: { type: 'json' } }, run: (args, exec) => service.start(args, { signal: exec.signal, owner: owner(exec) }) },
    { name: 'goecs_task', description: 'Read, export or cancel a GoECS task belonging to this DSH session. Report formats: txt, md, html, json. A failed or partial test is never a successful measurement.', parameters: { id: { type: 'string', required: true }, action: { type: 'string', enum: ['status','report','stop'], required: true }, format: { type: 'string', enum: ['txt','md','html','json'] }, offset: { type: 'integer' } }, run: (args, exec) => args.action === 'stop' ? service.stop(args.id, owner(exec)) : args.action === 'report' ? service.report(args.id, args.format, owner(exec)) : service.view(service.getTask(args.id, owner(exec)), Math.max(0, args.offset || 0)) },
  ]
  for (const tool of tools) {
    // Raw JSON Schema is the common registration contract across supported DSH versions.
    const properties = Object.fromEntries(Object.entries(tool.parameters).map(([name, { required, ...spec }]) => [name, spec.type === 'json' ? { type: 'object', additionalProperties: true } : spec]))
    const parameters = { type: 'object', properties, required: Object.keys(tool.parameters).filter(name => tool.parameters[name].required), additionalProperties: false }
    ctx.tools.register({ name: tool.name, description: tool.description, parameters, output: { schema: {}, render: (_args, value) => [{ type: 'text', text: JSON.stringify(value) }] }, isConcurrencySafe: () => tool.safe === true, execute: async (args, exec = {}) => {
      if (exec.signal?.aborted) throw exec.signal.reason
      if (!args || typeof args !== 'object' || Array.isArray(args) || Object.keys(args).some(name => !Object.hasOwn(properties, name))) throw new GoECSError('INVALID_ARGUMENTS', 'Invalid tool arguments')
      for (const [name, spec] of Object.entries(tool.parameters)) {
        const value = args[name]
        if (value === undefined && !spec.required) continue
        if ((spec.type === 'string' && typeof value !== 'string') || (spec.type === 'integer' && !Number.isInteger(value)) || (spec.type === 'json' && (!value || typeof value !== 'object' || Array.isArray(value))) || (spec.enum && !spec.enum.includes(value))) throw new GoECSError('INVALID_ARGUMENTS', 'Invalid tool argument: ' + name)
      }
      return JSON.parse(JSON.stringify(await tool.run(args, exec)))
    } })
  }
}
