import { readFile, writeFile } from 'node:fs/promises'
const root = new URL('../', import.meta.url)
const dictionaries = Object.fromEntries(await Promise.all(['en', 'zh'].map(async language => [language, JSON.parse(await readFile(new URL('skills/oneclickvirt-ecs/' + language + '.json', root), 'utf8'))])))
const location = new URL('agent/dsh-client.js', root), original = await readFile(location, 'utf8')
const updated = original.replace(/const dictionaries = \/\* GOECS_LOCALES \*\/ .*/, 'const dictionaries = /* GOECS_LOCALES */ ' + JSON.stringify(dictionaries))
if (process.argv.includes('--check')) {
  if (updated !== original) throw new Error('Client translations are stale; run npm run build')
} else if (updated !== original) await writeFile(location, updated)
