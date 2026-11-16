import { readFile, access } from 'node:fs/promises'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
const root = new URL('../', import.meta.url)
const pkg = JSON.parse(await readFile(new URL('package.json', root), 'utf8'))
if (pkg.name !== 'dsh-plugin-goecs' || pkg.private || !pkg.engines.dsh || pkg.license !== 'GPL-3.0-only') throw new Error('Invalid public DSH package metadata')
const files = pkg.files.filter(file => !file.startsWith('README'))
if (files.some(file => /\*|vendor|\.back|\.go$|test|secret|token/i.test(file))) throw new Error('Package must use a public runtime-file allowlist')
for (const file of files) {
  await access(new URL(file, root))
  if (/\.(?:m?js)$/.test(file)) {
    const result = spawnSync(process.execPath, ['--check', fileURLToPath(new URL(file, root))], { encoding: 'utf8' })
    if (result.status !== 0) throw new Error(file + ': ' + result.stderr)
  }
}
const dictionaries = Object.fromEntries(await Promise.all(['en','zh'].map(async lang => [lang, JSON.parse(await readFile(new URL('skills/oneclickvirt-ecs/' + lang + '.json', root), 'utf8'))])))
const keys = value => Object.keys(value).sort().join('\n')
if (keys(dictionaries.en) !== keys(dictionaries.zh)) throw new Error('English and Chinese UI keys differ')
const client = await readFile(new URL('agent/dsh-client.js', root), 'utf8')
for (const match of client.matchAll(/\bt\('([^']+)'\)/g)) if (!Object.hasOwn(dictionaries.en, match[1])) throw new Error('Missing translation: ' + match[1])
if (!client.includes('const dictionaries = /* GOECS_LOCALES */ ' + JSON.stringify(dictionaries))) throw new Error('Run npm run build to update client translations')
const entry = await readFile(new URL(pkg.exports['.'], root), 'utf8')
if (!/export function apply\(/.test(entry)) throw new Error('Missing native apply(ctx) entry')
const patch = await readFile(new URL(pkg.dsh.bundle.patch, root), 'utf8')
if (!patch.includes('name: ' + pkg.name)) throw new Error('Bundle patch does not load the package')
console.log('DSH manifest, runtime allowlist, syntax, translations and bundle entry verified.')
