#!/usr/bin/env python3
"""Validate immutable components, bump the consumer, then recover its release."""
import argparse
import base64
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import sys
import traceback
import urllib.error
import urllib.parse
from private_guard import api, clone, git_env, quiet, redact, root

PREFIX = 'github.com/oneclickvirt/'
PRIVATE = PREFIX + 'security,' + PREFIX + 'privatespeedtest'
STATE = '.github/component-state.json'
STABLE = re.compile(r'^v(\d+)\.(\d+)\.(\d+)$')
CONSUMERS = {'ecs': ('master', 'build_binary.yaml'), 'ecs-gui': ('main', 'build.yml')}
CONTROLLERS = {'security': ('securityCheck', 'securityCheck'), 'privatespeedtest': ('cnprivatespeedtest', 'privatespeedtest')}


def version(value):
    match = STABLE.fullmatch(value)
    return tuple(map(int, match.groups())) if match else None


def bump(value):
    parts = version(value)
    if parts is None:
        raise RuntimeError('invalid application version')
    return 'v' + '.'.join(map(str, (parts[0], parts[1], parts[2] + 1)))


def command(args, cwd=None):
    env = git_env()
    env.update({'GOPRIVATE': PRIVATE, 'GONOSUMDB': PRIVATE + ',' + PREFIX + 'ecs,' + PREFIX + 'speedtest',
                'GOPROXY': 'https://proxy.golang.org|direct', 'GOWORK': 'off', 'GIT_TERMINAL_PROMPT': '0'})
    return quiet(args, cwd=cwd, env=env, timeout=3600)


def requirements(folder):
    return {item['Path']: item['Version'] for item in json.loads(command(['go', 'mod', 'edit', '-json'], folder)).get('Require', [])}


def rows(endpoint):
    values = []
    for page in range(1, 21):
        suffix = '&' if '?' in endpoint else '?'
        result = api(endpoint + suffix + f'per_page=100&page={page}')
        values.extend(result)
        if len(result) < 100:
            return values
    raise RuntimeError('component pagination limit exceeded')


def valid_release(release, binary=None):
    if release.get('draft') or release.get('prerelease') or version(release.get('tag_name', '')) is None:
        return False
    assets = [a for a in release.get('assets', []) if '.ci-stage-' not in a['name'] and '.ci-backup-' not in a['name']]
    if binary:
        expected = {binary + '-' + goos + '-' + arch + ('.exe' if goos == 'windows' else '')
                    for goos, arches in [('linux', ['amd64', '386', 'arm64', 'armv7', 'riscv64', 'mips', 'mipsle', 'mips64', 'mips64le']),
                                        ('darwin', ['amd64', 'arm64']), ('freebsd', ['amd64', '386', 'arm64', 'armv7']),
                                        ('windows', ['amd64', '386', 'arm64'])] for arch in arches}
        return expected <= {a['name'] for a in assets if a.get('size', 0) > 0 and a.get('state', 'uploaded') == 'uploaded'}
    # ECS's cross builds publish archives, not the component raw binary names.
    return len([a for a in assets if a.get('size', 0) > 0 and a.get('state', 'uploaded') == 'uploaded']) >= 18


def latest_component(module, current):
    old = version(current)
    if old is None:
        # Legacy timestamp pre-releases remain pinned until an explicit migration.
        return current
    name = module[len(PREFIX):]
    tags = rows('/repos/oneclickvirt/' + name + '/tags')
    candidates = [t['name'] for t in tags if version(t['name']) and version(t['name'])[0] == old[0] and version(t['name']) > old]
    if name in CONTROLLERS and candidates:
        controller, binary = CONTROLLERS[name]
        released = {r['tag_name'] for r in rows('/repos/oneclickvirt/' + controller + '/releases') if valid_release(r, binary)}
        candidates = [v for v in candidates if v in released]
    return max(candidates, key=version) if candidates else current


def application_version(kind, folder=Path('.')):
    if kind == 'ecs':
        match = re.search(r'DefaultVersion\s*=\s*"(v\d+\.\d+\.\d+)"', (folder/'api/api.go').read_text())
        if not match:
            raise RuntimeError('missing application version')
        return match[1]
    value = 'v' + (folder/'VERSION').read_text().strip()
    if version(value) is None:
        raise RuntimeError('missing GUI version')
    return value


def replace_field(path, field, value):
    text, count = re.subn(r'(\b' + re.escape(field) + r'\s*=\s*")[^"]*(")', lambda m: m[1] + value + m[2], path.read_text())
    if count != 1:
        raise RuntimeError('ambiguous version field')
    path.write_text(text)


def sync_versions(kind, new_version, desired):
    if kind == 'ecs':
        replace_field(Path('api/api.go'), 'DefaultVersion', new_version)
    else:
        Path('VERSION').write_text(new_version[1:] + '\n')
        replace_field(Path('FyneApp.toml'), 'Version', new_version[1:])
        replace_field(Path('internal/appmeta/version.go'), 'Version', new_version[1:])
        replace_field(Path('internal/appmeta/version.go'), 'UpstreamECSVersion', desired[PREFIX+'ecs'])


def workspace():
    folder = Path('deps')
    if folder.exists():
        raise RuntimeError('updater workspace already exists')
    folder.mkdir()
    modules = requirements(Path('.'))
    sources = []
    for module, ref in modules.items():
        if module.startswith(PREFIX):
            destination = folder/module[len(PREFIX):]
            clone('oneclickvirt/' + module[len(PREFIX):], destination, ref)
            sources.append(destination)
    command(['go', 'work', 'init', '.'])
    command(['go', 'work', 'use'] + [str(f) for f in sources])
    command(['go', 'work', 'edit', '-go=' + json.loads(command(['go', 'mod', 'edit', '-json']))['Go']])
    with Path('.git/info/exclude').open('a') as handle:
        handle.write('\n/deps/\n/go.work\n/go.work.sum\n')


def test_consumer(kind):
    env = git_env()
    env.update({'GOPRIVATE': PRIVATE, 'GONOSUMDB': PRIVATE + ',' + PREFIX + 'ecs,' + PREFIX + 'speedtest',
                'GOPROXY': 'https://proxy.golang.org|direct'})
    env.pop('GOWORK', None)
    commands = [['go', 'test', '-race', '-count=1', './...'], ['go', 'vet', './...']]
    if kind == 'ecs-gui':
        commands += [['go', 'test', '-race', '-count=1', '-tags=ecs_structured', './...'],
                     ['go', 'vet', '-tags=ecs_structured', './...']]
    for args in commands:
        quiet(args, env=env, timeout=3600)
    quiet(['go', 'build', '-trimpath', '-buildvcs=false', '-ldflags=-checklinkname=0', '-o', str(root()/'consumer-check'), '.'], env=env, timeout=3600)


def fingerprints(paths):
    return {p: hashlib.sha256(Path(p).read_bytes()).hexdigest() for p in sorted(paths)}


def release_complete(kind, tag):
    try:
        release = api('/repos/oneclickvirt/' + kind + '/releases/tags/' + tag)
    except urllib.error.HTTPError as exc:
        if exc.code == 404:
            return False
        raise
    if kind == 'ecs':
        return valid_release(release)
    return not release['draft'] and len([a for a in release.get('assets', []) if a.get('size', 0) > 0 and '.ci-stage-' not in a['name'] and '.ci-backup-' not in a['name']]) >= 7


def recover_release(kind, tag, state):
    if state.get('version') != tag or state.get('fingerprints') != fingerprints(state.get('fingerprints', {})):
        raise RuntimeError('validated update state no longer matches the source')
    full = 'oneclickvirt/' + kind
    head = command(['git', 'rev-parse', 'HEAD'])
    endpoint = '/repos/' + full + '/git/ref/tags/' + tag
    try:
        ref = api(endpoint)
    except urllib.error.HTTPError as exc:
        if exc.code != 404:
            raise
        # GITHUB_TOKEN tag creation does not recursively trigger a second build.
        api('/repos/' + full + '/git/refs', 'POST', {'ref': 'refs/tags/' + tag, 'sha': head}, credential=os.environ['GITHUB_TOKEN'])
        ref = {'object': {'type': 'commit', 'sha': head}}
    obj = ref['object']
    while obj['type'] == 'tag':
        obj = api('/repos/' + full + '/git/tags/' + obj['sha'])['object']
    # Later CI-only commits may be on the default branch; validate versioned files at the immutable tag.
    for path, digest in state['fingerprints'].items():
        content = api('/repos/' + full + '/contents/' + urllib.parse.quote(path) + '?ref=' + tag)
        if hashlib.sha256(base64.b64decode(content['content'])).hexdigest() != digest:
            raise RuntimeError('release tag does not match validated update')
    if release_complete(kind, tag):
        return
    branch, workflow = CONSUMERS[kind]
    runs = api('/repos/' + full + '/actions/workflows/' + workflow + '/runs?head_sha=' + obj['sha'] + '&per_page=20')['workflow_runs']
    if any(r['status'] != 'completed' for r in runs):
        return
    api('/repos/' + full + '/actions/workflows/' + workflow + '/dispatches', 'POST', {'ref': tag})


def update(kind):
    branch, _ = CONSUMERS[kind]
    current = requirements(Path('.'))
    state = json.loads(Path(STATE).read_text()) if Path(STATE).exists() else {}
    old_version = application_version(kind)
    # Resolve an interrupted publication before starting another version bump.
    if state.get('version') == old_version and not release_complete(kind, old_version):
        recover_release(kind, old_version, state)
        return 'Release recovery checked.'
    if kind == 'ecs':
        desired = {m: latest_component(m, v) for m, v in current.items() if m.startswith(PREFIX)}
    else:
        releases = [r for r in rows('/repos/oneclickvirt/ecs/releases') if valid_release(r)]
        if not releases:
            raise RuntimeError('no verified ECS release available')
        ecs = max(releases, key=lambda r: version(r['tag_name']))['tag_name']
        if version(ecs) <= version(current[PREFIX+'ecs']):
            return 'No component update available.'
        source = root()/'released-ecs'
        clone('oneclickvirt/ecs', source, ecs)
        if application_version('ecs', source) != ecs:
            raise RuntimeError('upstream release version mismatch')
        desired = {m: v for m, v in requirements(source).items() if m.startswith(PREFIX)}
        desired[PREFIX+'ecs'] = ecs
    changed = {m: v for m, v in desired.items() if current.get(m) != v}
    if not changed:
        return 'No component update available.'
    command(['go', 'mod', 'edit'] + ['-require=' + m + '@' + v for m, v in changed.items()])
    command(['go', 'mod', 'tidy'])
    actual = requirements(Path('.'))
    if any(actual.get(m) != v for m,v in desired.items()):
        raise RuntimeError('dependency graph did not retain pinned component versions')
    new_version = bump(old_version)
    sync_versions(kind, new_version, desired)
    workspace()
    test_consumer(kind)
    paths = ['go.mod', 'go.sum', 'api/api.go'] if kind == 'ecs' else ['go.mod', 'go.sum', 'VERSION', 'FyneApp.toml', 'internal/appmeta/version.go']
    edited = command(['git', 'diff', '--name-only']).splitlines()
    if set(edited) - set(paths):
        raise RuntimeError('update modified product layout or unrelated source')
    state = {'format': 1, 'version': new_version, 'components': desired, 'fingerprints': fingerprints(paths)}
    Path(STATE).write_text(json.dumps(state, indent=2, sort_keys=True)+'\n')
    command(['git', 'diff', '--check'])
    command(['git', 'config', 'user.name', 'github-actions[bot]'])
    command(['git', 'config', 'user.email', '41898282+github-actions[bot]@users.noreply.github.com'])
    command(['git', 'add', '--', *paths, STATE])
    command(['git', 'commit', '-q', '-m', 'chore: validate component updates for ' + new_version])
    command(['git', 'push', '-q', 'origin', 'HEAD:refs/heads/' + branch])
    recover_release(kind, new_version, state)
    return 'Validated component update committed; release dispatched.'


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--kind', choices=CONSUMERS, required=True)
    args = parser.parse_args()
    os.umask(0o077)
    try:
        print(update(args.kind))
    except Exception:
        with (root()/'component-update.log').open('ab') as handle:
            handle.write(redact(traceback.format_exc()))
        print('Component update failed; diagnostics are stored privately.', file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
