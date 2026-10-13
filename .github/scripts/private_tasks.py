#!/usr/bin/env python3
"""Public launch control; sources, registry updates and diagnostics stay private."""
import argparse
import json
import os
from pathlib import Path
import re
import sys
import urllib.error
import urllib.request
from private_guard import api, clone, execute, git_env, quiet, redact, root, token, verify_assets

CONFIG = {
    'security': ('SecurityVersion', 'network/version.go', './cmd', 'securityCheck', 'securityCheck', ['network/security/data']),
    'privatespeedtest': ('PrivateSpeedTestVersion', 'pst/model.go', '.', 'privatespeedtest', 'cnprivatespeedtest', ['pst/snapshot/servers.json', 'pst/snapshot/manifest.json', 'transfer/snapshot/targets.json', 'transfer/snapshot/manifest.json']),
    'openspeedtest': ('OpenSpeedTestVersion', 'main.go', '.', 'openspeedtest', 'openspeedtest', ['speedtest_servers.json', 'speedtest_origin_servers.json', 'speedtest_servers_stats.json']),
}


def commit_private(repository, source, paths):
    env = git_env()
    quiet(['git', 'config', 'user.name', 'github-actions[bot]'], cwd=source, env=env)
    quiet(['git', 'config', 'user.email', '41898282+github-actions[bot]@users.noreply.github.com'], cwd=source, env=env)
    quiet(['git', 'add', '--', *paths], cwd=source, env=env)
    if quiet(['git', 'diff', '--cached', '--name-only'], cwd=source, env=env):
        quiet(['git', 'commit', '-q', '-m', 'chore: refresh private registry [skip ci]'], cwd=source, env=env)
        # Normal push rejects concurrent source writes; never force source history.
        quiet(['git', 'push', '-q', 'origin', 'HEAD:main'], cwd=source, env=env)


def upload(repository, tag, paths):
    if repository == 'openspeedtest' and not api('/repos/oneclickvirt/openspeedtest')['private']:
        raise RuntimeError('openspeedtest publication must remain private')
    full = 'oneclickvirt/' + repository
    verify_assets([str(p) for p in paths])
    credential = token() if repository == 'openspeedtest' else os.environ['GITHUB_TOKEN']
    try:
        release = api('/repos/' + full + '/releases/tags/' + tag, credential=credential)
    except urllib.error.HTTPError as exc:
        if exc.code != 404:
            raise
        release = api('/repos/' + full + '/releases', 'POST', {'tag_name': tag, 'name': tag, 'body': 'Automated binary build.', 'generate_release_notes': False}, credential=credential)
    for path in paths:
        if not path.is_file() or path.is_symlink() or path.stat().st_size > 512 * 1024 * 1024:
            raise RuntimeError('unapproved release asset')
        for asset in api('/repos/' + full + '/releases/' + str(release['id']) + '/assets?per_page=100', credential=credential):
            if asset['name'] == path.name:
                api('/repos/' + full + '/releases/assets/' + str(asset['id']), 'DELETE', credential=credential)
        req = urllib.request.Request('https://uploads.github.com/repos/' + full + '/releases/' + str(release['id']) + '/assets?name=' + path.name,
            data=path.read_bytes(), method='POST', headers={'Authorization': 'Bearer ' + credential, 'Content-Type': 'application/octet-stream'})
        with urllib.request.urlopen(req, timeout=120) as response:
            response.read()


def refresh_script(repository, transfer=False):
    if repository == 'security':
        return 'go run ./cmd/update-dnsbl-data -allow-stale\n'
    if repository == 'openspeedtest':
        return 'python3 parse_servers.py\n'
    if transfer:
        return 'go run ./cmd/update-transfer-registry -output transfer/snapshot/targets.json -manifest transfer/snapshot/manifest.json -allow-stale\n'
    return 'go run ./cmd/update-registry -source https://paste.spiritlhl.net/speedtest/registry/servers.json -fallback-source .upstream-openspeedtest/speedtest_servers.json -output pst/snapshot/servers.json -manifest pst/snapshot/manifest.json -timeout 90s -allow-stale\n'


def go_task(repository, task, tag):
    if repository not in CONFIG:
        raise RuntimeError('unknown source repository')
    field, version_path, package, binary, target, changed = CONFIG[repository]
    source = root() / 'source'
    ref = tag if task == 'release' and tag else 'main'
    if tag and not re.fullmatch(r'v\d+\.\d+\.\d+', tag):
        raise RuntimeError('invalid semantic version')
    if task == 'release' and not tag:
        raise RuntimeError('release requires an existing semantic version tag')
    clone('oneclickvirt/' + repository, source, ref)
    version_match = re.search(r'\b' + field + r'\s*=\s*"(v\d+\.\d+\.\d+)"', (source / version_path).read_text())
    if tag and (not version_match or version_match[1] != tag):
        raise RuntimeError('release tag does not match source version')
    if task == 'refresh' and repository in ('privatespeedtest', 'openspeedtest') and os.environ.get('GITHUB_EVENT_NAME') == 'schedule':
        import datetime
        days = (datetime.datetime.now(datetime.timezone.utc).date() - datetime.date(2026, 9, 3)).days
        if days < 0 or days % 3:
            return
    if task in ('refresh', 'refresh-transfer', 'prepare'):
        if repository == 'privatespeedtest' and task != 'refresh-transfer':
            clone('oneclickvirt/openspeedtest', source / '.upstream-openspeedtest')
        execute('mkdir -p /tmp/home\n' + refresh_script(repository, task == 'refresh-transfer'), cwd=source, container=True)
    validation = 'mkdir -p /tmp/home\ngo mod verify\ngo test -race -count=1 ./...\ngo vet ./...\n'
    if repository == 'privatespeedtest':
        validation += 'go mod tidy -diff\n'
    if repository == 'openspeedtest':
        validation += 'python3 -m unittest -v parse_servers_test.py\n'
    validation += 'go build -trimpath -buildvcs=false -o /out/check ' + package + '\ngit diff --check\n'
    execute(validation, cwd=source, container=True)
    if task in ('refresh', 'refresh-transfer', 'prepare'):
        commit_private(repository, source, changed)
    if task == 'prepare':
        versions = list(source.glob('**/*version*.go')) + [source / version_path, source / 'main.go']
        if not tag:
            for path in versions:
                if path.is_file():
                    match = re.search(r'\b' + field + r'\s*=\s*"(v\d+\.\d+\.\d+)"', path.read_text())
                    if match:
                        tag = match[1]
                        break
        if not re.fullmatch(r'v\d+\.\d+\.\d+', tag):
            raise RuntimeError('source version unavailable')
        quiet(['git', 'tag', '-a', tag, '-m', 'Release ' + tag], cwd=source, env=git_env())
        quiet(['git', 'push', '-q', 'origin', 'refs/tags/' + tag], cwd=source, env=git_env())
    if task not in ('build', 'release', 'prepare'):
        return
    assets = []
    script = 'mkdir -p /tmp/home\n'
    for goos, goarch, goarm in [('linux', 'amd64', ''), ('linux', '386', ''), ('linux', 'arm64', ''), ('linux', 'arm', '7'), ('linux', 'riscv64', ''),
                               ('linux', 'mips', ''), ('linux', 'mipsle', ''), ('linux', 'mips64', ''), ('linux', 'mips64le', ''),
                               ('darwin', 'amd64', ''), ('darwin', 'arm64', ''), ('freebsd', 'amd64', ''), ('freebsd', '386', ''), ('freebsd', 'arm64', ''),
                               ('freebsd', 'arm', '7'),
                               ('windows', 'amd64', ''), ('windows', '386', ''), ('windows', 'arm64', '')]:
        label = 'armv7' if goarm else goarch
        name = binary + '-' + goos + '-' + label + ('.exe' if goos == 'windows' else '')
        script += f'CGO_ENABLED=0 GOOS={goos} GOARCH={goarch} GOARM={goarm} go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/{name} {package}\n'
        assets.append(root() / 'out' / name)
    execute(script, cwd=source, container=True)
    upload(target, tag or 'output', assets)


def icmp_task(task):
    source = root() / 'source'
    clone('oneclickvirt/ICMPTargetSpiders', source)
    dependencies = 'mkdir -p /tmp/home\npython3 -m venv /tmp/venv\n/tmp/venv/bin/pip install --disable-pip-version-check -r requirements.txt\n'
    if task in ('analyze', 'ci'):
        script = dependencies
        if task == 'analyze':
            script += '/tmp/venv/bin/python build.py\n/tmp/venv/bin/python domain_resolver.py\n/tmp/venv/bin/python ip_finder.py\n'
        else:
            script += '/tmp/venv/bin/python -m unittest discover -p "test*.py"\n'
        execute(script, cwd=source, container=True)
        if task == 'analyze':
            commit_private('ICMPTargetSpiders', source, ['cloud_vpc_analysis.csv', 'cdn_domain_analysis.csv'])
        return
    if task != 'publish-nodes':
        raise RuntimeError('invalid ICMP task')
    script = dependencies + '''/tmp/venv/bin/python - <<'PY'
import json
import pandas as pd
from csv_schema import dataframe_for_nodes_json
data = dataframe_for_nodes_json(pd.read_csv('cdn_domain_analysis.csv')).to_dict(orient='records')
with open('/out/nodes.json', 'w') as handle:
    json.dump(data, handle, ensure_ascii=False, indent=2)
PY
'''
    execute(script, cwd=source, container=True)
    path = root() / 'out' / 'nodes.json'
    fields = {'province', 'isp_code', 'isp', 'ip_version', 'ips', 'latency_ms', 'quality_score', 'ips_latency_ms'}
    if path.is_symlink() or path.stat().st_size > 50 * 1024 * 1024:
        raise RuntimeError('invalid public data export')
    data = json.loads(path.read_text())
    if not isinstance(data, list) or any(not isinstance(row, dict) or set(row) != fields for row in data):
        raise RuntimeError('public export schema mismatch')
    public = root() / 'published-nodes'
    clone('spiritLHLS/icmp_targets', public)
    (public / 'nodes.json').write_bytes(path.read_bytes())
    commit_private('ICMPTargetSpiders', public, ['nodes.json'])


def main():
    os.umask(0o077)
    parser = argparse.ArgumentParser()
    parser.add_argument('--repository', required=True)
    parser.add_argument('--task', default='ci')
    parser.add_argument('--release-tag', default='')
    args = parser.parse_args()
    try:
        if args.repository == 'ICMPTargetSpiders':
            icmp_task(args.task)
        else:
            go_task(args.repository, args.task, args.release_tag)
    except BaseException:
        import traceback
        with (root() / 'control.log').open('ab') as handle:
            handle.write(redact(traceback.format_exc()))
        print('Private task failed. Diagnostic output is restricted to the private log branch.', file=sys.stderr)
        return 1
    return 0


if __name__ == '__main__':
    sys.exit(main())
