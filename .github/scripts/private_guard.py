#!/usr/bin/env python3
"""Keep private build output out of the Actions log and public caches."""
import argparse
import base64
import datetime
import gzip
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import time
import traceback
import glob
import zipfile
import tarfile
import urllib.error
import urllib.request

LOG_BRANCH = 'ci-private-logs'
PRIVATE = {'security', 'privatespeedtest', 'openspeedtest', 'ICMPTargetSpiders', 'jks', 'worker_traffic'}
SAFE_EXPORT = re.compile(r'^[A-Za-z0-9_./:\\@+~=-]{0,2048}$')


def root():
    key = re.sub(r'[^A-Za-z0-9_.-]', '_', os.environ.get('CI_JOB_KEY', os.environ.get('GITHUB_JOB', 'job')))
    key += '-' + re.sub(r'[^0-9-]', '', os.environ.get('GITHUB_RUN_ID', '0') + '-' + os.environ.get('GITHUB_RUN_ATTEMPT', '1'))
    value = Path(os.environ.get('CI_STATE_ROOT', os.environ.get('RUNNER_TEMP', tempfile.gettempdir()))) / ('private-ci-' + key)
    value.mkdir(parents=True, exist_ok=True)
    return value


def token():
    value = os.environ.get('PRIVATE_CI_TOKEN', '')
    if not value:
        raise RuntimeError('credential missing')
    return value


def api(path, method='GET', body=None, credential=None):
    for attempt in range(3):
        req = urllib.request.Request('https://api.github.com' + path, method=method,
            data=None if body is None else json.dumps(body).encode(),
            headers={'Authorization': 'Bearer ' + (credential or token()),
                     'Accept': 'application/vnd.github+json', 'Content-Type': 'application/json',
                     'X-GitHub-Api-Version': '2022-11-28'})
        try:
            with urllib.request.urlopen(req, timeout=45) as response:
                data = response.read()
            return json.loads(data) if data else None
        except urllib.error.HTTPError as exc:
            if method != 'GET' or exc.code not in (408, 429, 500, 502, 503, 504) or attempt == 2:
                raise
        except (TimeoutError, urllib.error.URLError):
            if method != 'GET' or attempt == 2:
                raise
        time.sleep(2 ** attempt)

def git_env():
    empty_config = root() / 'empty-gitconfig'
    empty_config.write_text('')
    empty_hooks = root() / 'disabled-hooks'
    empty_hooks.mkdir(exist_ok=True)
    env = dict(os.environ)
    env.update({'GIT_TERMINAL_PROMPT': '0', 'GIT_CONFIG_NOSYSTEM': '1',
        'GIT_CONFIG_GLOBAL': empty_config.as_posix(), 'GIT_CONFIG_COUNT': '3',
        'GIT_CONFIG_KEY_0': 'http.https://github.com/.extraheader',
        'GIT_CONFIG_VALUE_0': 'Authorization: Basic ' + base64.b64encode(('x-access-token:' + token()).encode()).decode(),
        'GIT_CONFIG_KEY_1': 'credential.helper', 'GIT_CONFIG_VALUE_1': '',
        'GIT_CONFIG_KEY_2': 'core.hooksPath', 'GIT_CONFIG_VALUE_2': empty_hooks.as_posix()})
    workspace = Path(os.environ.get('GITHUB_WORKSPACE', os.getcwd())).resolve()
    candidates = [workspace] + [item.parent.resolve() for item in (workspace / 'deps').glob('*/.git')]
    directories = [directory for directory in candidates if directory == workspace or workspace in directory.parents]
    for index, directory in enumerate(directories, start=3):
        env['GIT_CONFIG_KEY_' + str(index)] = 'safe.directory'
        env['GIT_CONFIG_VALUE_' + str(index)] = directory.as_posix()
    env['GIT_CONFIG_COUNT'] = str(3 + len(directories))
    return env


def quiet(command, cwd=None, env=None, timeout=3600):
    result = subprocess.run(command, cwd=cwd, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=timeout)
    if result.returncode:
        with (root() / 'control.log').open('ab') as handle:
            handle.write(result.stdout)
        raise RuntimeError('command failed')
    return result.stdout.decode(errors='replace').strip()


def redact(data):
    text = data.decode(errors='replace') if isinstance(data, bytes) else str(data)
    values = [os.environ.get(k, '') for k in ('PRIVATE_CI_TOKEN', 'GITHUB_TOKEN', 'KEYSTORE_PASSWORD', 'KEY_PASSWORD')]
    for value in values:
        if value:
            for variant in (value, base64.b64encode(value.encode()).decode(), base64.b64encode(('x-access-token:' + value).encode()).decode()):
                text = text.replace(variant, '[REDACTED]')
    text = re.sub(r'(?:gh[pousr]_[A-Za-z0-9]+|github_pat_[A-Za-z0-9_]+)', '[REDACTED]', text)
    text = re.sub(r'https://[^\s/@]+:[^\s/@]+@', 'https://[REDACTED]@', text)
    text = re.sub(r'-----BEGIN [^-]*PRIVATE KEY-----.*?-----END [^-]*PRIVATE KEY-----', '[PRIVATE KEY REDACTED]', text, flags=re.S)
    return text.encode()


def archive(repository, status='completed'):
    if repository not in PRIVATE:
        raise RuntimeError('log destination not allowlisted')
    full = 'oneclickvirt/' + repository
    if not api('/repos/' + full)['private']:
        raise RuntimeError('log destination must be private')
    run_id = os.environ.get('GITHUB_RUN_ID', '0')
    attempt = os.environ.get('GITHUB_RUN_ATTEMPT', '1')
    if not re.fullmatch(r'\d+', run_id) or not re.fullmatch(r'\d+', attempt):
        raise RuntimeError('invalid run identity')
    run_key = run_id + '-' + attempt
    job_key = re.sub(r'[^A-Za-z0-9_.-]', '_', os.environ.get('CI_JOB_KEY', os.environ.get('GITHUB_JOB', 'job')))
    payload = b''
    for path in sorted(root().glob('*.log')):
        payload += ('\n===== ' + path.name + ' =====\n').encode() + path.read_bytes()
    payload = gzip.compress(redact(payload), mtime=0)
    if len(payload) > 40 * 1024 * 1024:
        raise RuntimeError('private log too large')
    for retry in range(8):
        with tempfile.TemporaryDirectory(prefix='private-log-', dir=root()) as temp:
            folder = Path(temp)
            env = git_env()
            quiet(['git', 'init', '-q', str(folder)], env=env)
            quiet(['git', 'remote', 'add', 'origin', 'https://github.com/' + full + '.git'], cwd=folder, env=env)
            remote = quiet(['git', 'ls-remote', '--heads', 'origin', 'refs/heads/' + LOG_BRANCH], cwd=folder, env=env)
            old = ''
            if remote:
                quiet(['git', 'fetch', '-q', '--depth=1', 'origin', 'refs/heads/' + LOG_BRANCH], cwd=folder, env=env)
                old = quiet(['git', 'rev-parse', 'FETCH_HEAD'], cwd=folder, env=env)
                quiet(['git', 'checkout', '-q', '--detach', old], cwd=folder, env=env)
                if not (folder / 'manifest.json').exists():
                    raise RuntimeError('existing log branch has unknown format')
            for path in folder.rglob('*'):
                if '.git' not in path.parts and path.is_symlink():
                    raise RuntimeError('unexpected log symlink')
            target = folder / 'runs' / run_key
            target.mkdir(parents=True, exist_ok=True)
            (target / (job_key + '.log.gz')).write_bytes(payload)
            (target / (job_key + '.json')).write_text(json.dumps({'status': status, 'public_repository': os.environ.get('GITHUB_REPOSITORY', ''), 'run_id': run_id, 'attempt': attempt, 'job': job_key}, indent=2))
            (target / 'run.json').write_text(json.dumps({'updated_at': datetime.datetime.now(datetime.timezone.utc).isoformat()}, indent=2))
            def log_order(name):
                record = folder / 'runs' / name / 'run.json'
                timestamp = json.loads(record.read_text()).get('updated_at', '') if record.exists() else ''
                return timestamp, tuple(int(v) for v in name.split('-'))
            # Retrying an old run is a new execution and must remain traceable.
            # Legacy entries without timestamps keep their original ID order.
            runs = sorted((p.name for p in (folder / 'runs').iterdir() if p.is_dir()), key=log_order, reverse=True)
            for key in runs[5:]:
                shutil.rmtree(folder / 'runs' / key)
            (folder / 'manifest.json').write_text(json.dumps({'format': 'oneclickvirt.private-ci/v1', 'retained_runs': runs[:5], 'retention': 5, 'retention_order': 'last saved execution', 'history': 'one parentless snapshot'}, indent=2))
            env.update({'GIT_AUTHOR_NAME': 'github-actions[bot]', 'GIT_AUTHOR_EMAIL': '41898282+github-actions[bot]@users.noreply.github.com',
                        'GIT_COMMITTER_NAME': 'github-actions[bot]', 'GIT_COMMITTER_EMAIL': '41898282+github-actions[bot]@users.noreply.github.com'})
            quiet(['git', 'read-tree', '--empty'], cwd=folder, env=env)
            quiet(['git', 'add', 'runs', 'manifest.json'], cwd=folder, env=env)
            tree = quiet(['git', 'write-tree'], cwd=folder, env=env)
            commit = quiet(['git', 'commit-tree', tree, '-m', 'Private CI log snapshot [skip ci]'], cwd=folder, env=env)
            try:
                quiet(['git', 'push', '-q', '--force-with-lease=refs/heads/' + LOG_BRANCH + ':' + old,
                       'origin', commit + ':refs/heads/' + LOG_BRANCH], cwd=folder, env=env)
                return
            except RuntimeError:
                if retry == 7:
                    raise
                time.sleep(min(2 ** retry, 8))


def clone(repository, destination, ref='main'):
    if not re.fullmatch(r'(?:oneclickvirt|spiritLHLS)/[A-Za-z0-9_.-]+', repository):
        raise RuntimeError('repository not allowlisted')
    if not re.fullmatch(r'[A-Za-z0-9_.-]+', ref):
        raise RuntimeError('invalid source ref')
    dest = Path(destination).absolute()
    if dest.exists() or dest.is_symlink():
        raise RuntimeError('checkout destination already exists')
    dest.parent.mkdir(parents=True, exist_ok=True)
    for attempt in range(3):
        with tempfile.TemporaryDirectory(prefix=dest.name + '-clone-', dir=dest.parent) as temp:
            checkout = Path(temp) / 'source'
            try:
                quiet(['git', 'clone', '-q', '--depth=1', '--branch', ref, '--single-branch',
                       'https://github.com/' + repository + '.git', str(checkout)], env=git_env(), timeout=300)
                if dest.exists() or dest.is_symlink():
                    raise RuntimeError('checkout destination already exists')
                checkout.rename(dest)
                return
            except subprocess.TimeoutExpired:
                retry = True
            except RuntimeError:
                diagnostic = (root() / 'control.log').read_bytes()[-65536:] if (root() / 'control.log').exists() else b''
                retry = bool(re.search(rb'HTTP[^0-9]*(408|429|500|502|503|504)|timed? out|timeout|connection (reset|refused|closed)|temporary failure|TLS handshake|unexpected EOF|could not resolve host|failed to connect|remote end hung up', diagnostic, re.I))
                if not retry:
                    raise
            if attempt == 2:
                raise RuntimeError('source download failed')
        time.sleep(2 ** attempt)

def exports(path, destination, allowed):
    if not path.exists() or not destination:
        return
    lines = []
    for line in path.read_text().splitlines():
        key, separator, value = line.partition('=')
        if not separator or key not in allowed or not SAFE_EXPORT.fullmatch(value):
            raise RuntimeError('unapproved workflow export')
        if redact(value).decode() != value:
            raise RuntimeError('secret in workflow export')
        lines.append(line)
    with open(destination, 'a', encoding='utf-8') as handle:
        handle.write('\n'.join(lines) + ('\n' if lines else ''))


def bash_path(platform=None):
    if (platform or os.name) != 'nt':
        value = shutil.which('bash')
        if value:
            return value
    else:
        # Windows System32/bash.exe starts WSL, not the Git Bash used by Actions.
        for parent in (os.environ.get('ProgramFiles', r'C:\Program Files'), os.environ.get('ProgramW6432', r'C:\Program Files')):
            for suffix in ('Git/bin/bash.exe', 'Git/usr/bin/bash.exe'):
                candidate = Path(parent) / suffix
                if candidate.is_file():
                    return str(candidate)
    raise RuntimeError('trusted Bash installation unavailable')


def execute(script, cwd=None, container=False, shell='bash'):
    state = root()
    index = len(list(state.glob('step-*.log')))
    if shell not in ('bash', 'pwsh') or (container and shell != 'bash'):
        raise RuntimeError('unsupported private shell')
    script_path = state / ('step-' + str(index) + ('.ps1' if shell == 'pwsh' else '.sh'))
    if shell == 'pwsh':
        script = "$ErrorActionPreference='Stop'\n$ProgressPreference='SilentlyContinue'\n$PSNativeCommandUseErrorActionPreference=$true\n" + script
    else:
        # The controller uses 077 for diagnostics. Give native build fixtures
        # their usual creation permissions; the enclosing state stays private.
        script = 'umask 022\n' + script
    script_path.write_text(script, encoding='utf-8', newline='\n')
    os.chmod(script_path, 0o600)
    log = state / ('step-' + str(index).zfill(3) + '.log')
    env_file = state / ('env-' + str(index))
    output_file = state / ('output-' + str(index))
    path_file = state / ('path-' + str(index))
    env = {k: v for k, v in os.environ.items() if not k.startswith(('ACTIONS_', 'GITHUB_STATE', 'GITHUB_STEP_SUMMARY')) and k != 'PRIVATE_CI_TOKEN'}
    if os.environ.get('PRIVATE_CI_TOKEN') and not container:
        env.update({k: v for k, v in git_env().items() if k.startswith('GIT_')})
    env['GITHUB_ENV'], env['GITHUB_OUTPUT'], env['GITHUB_PATH'] = env_file.as_posix(), output_file.as_posix(), path_file.as_posix()
    command = [bash_path(), '--noprofile', '--norc', '-e', '-o', 'pipefail', script_path.as_posix()]
    if shell == 'pwsh':
        executable = shutil.which('pwsh')
        if not executable:
            raise RuntimeError('PowerShell unavailable')
        command = [executable, '-NoLogo', '-NoProfile', '-NonInteractive', '-ExecutionPolicy', 'Bypass', '-File', str(script_path)]
    if container:
        context = state / 'image'
        context.mkdir(exist_ok=True)
        (context / 'Dockerfile').write_text('FROM golang:1.26-bookworm\nRUN apt-get update && apt-get install -y --no-install-recommends python3 python3-venv dnsutils whois traceroute iputils-ping ca-certificates && rm -rf /var/lib/apt/lists/*\n')
        quiet(['docker', 'build', '-q', '-t', 'private-ci-build:local', str(context)], timeout=1200)
        out = state / 'out'
        out.mkdir(exist_ok=True)
        source = Path(cwd).resolve()
        command = ['docker', 'run', '--rm', '--log-driver=none', '--cap-drop=ALL', '--security-opt=no-new-privileges',
            '--pids-limit=512', '--memory=6g', '--user', str(os.getuid()) + ':' + str(os.getgid()),
            '--tmpfs', '/tmp:rw,exec,nosuid,size=2g', '--mount', 'type=bind,src=' + str(source) + ',dst=/src',
            '--mount', 'type=bind,src=' + str(out) + ',dst=/out', '--mount', 'type=bind,src=' + str(script_path) + ',dst=/command.sh,readonly',
            '-w', '/src', '-e', 'HOME=/tmp/home', '-e', 'GOCACHE=/tmp/go-build', '-e', 'GOMODCACHE=/tmp/go-mod',
            '-e', 'GOPRIVATE=github.com/oneclickvirt/openspeedtest,github.com/oneclickvirt/security,github.com/oneclickvirt/privatespeedtest',
            '-e', 'GOWORK=off', 'private-ci-build:local', 'bash', '-e', '-o', 'pipefail', '/command.sh']
    try:
        with log.open('wb') as handle:
            result = subprocess.run(command, cwd=cwd, env=env, stdin=subprocess.DEVNULL, stdout=handle, stderr=subprocess.STDOUT, timeout=int(os.environ.get('CI_STEP_TIMEOUT', '14400')))
        if not container and result.returncode == 0:
            exports(env_file, os.environ.get('GITHUB_ENV'), set(os.environ.get('CI_ALLOWED_ENV', '').split(',')))
            exports(output_file, os.environ.get('GITHUB_OUTPUT'), set(os.environ.get('CI_ALLOWED_OUTPUT', '').split(',')))
            if path_file.exists() and os.environ.get('GITHUB_PATH'):
                for value in path_file.read_text().splitlines():
                    if not SAFE_EXPORT.fullmatch(value) or redact(value).decode() != value:
                        raise RuntimeError('unapproved workflow path')
                with open(os.environ['GITHUB_PATH'], 'a') as handle:
                    handle.write(path_file.read_text())
        if result.returncode:
            raise RuntimeError('private build failed')
    finally:
        script_path.unlink(missing_ok=True)


def finalize(repositories, status='completed'):
    failed = False
    for repository in repositories.split(','):
        try:
            archive(repository, status)
        except BaseException:
            failed = True
    if failed:
        raise RuntimeError('private log persistence failed')


def verify_assets(patterns):
    forbidden = re.compile(r'(?:^|/)(?:\.git|\.env|[^/]+\.(?:go|rs|c|cpp|h|hpp|java|kt|kts|js|jsx|ts|tsx|py|sh|yml|yaml|jks|keystore|pem|key|log))(?:$|/)', re.I)
    paths = [Path(p) for pattern in patterns for p in glob.glob(pattern)]
    if not paths:
        raise RuntimeError('release assets missing')
    for path in paths:
        if not path.is_file() or path.is_symlink() or forbidden.search(path.as_posix()):
            raise RuntimeError('unapproved release asset')
        if zipfile.is_zipfile(path):
            with zipfile.ZipFile(path) as archive:
                for member in archive.infolist():
                    if member.filename.startswith('/') or '..' in Path(member.filename).parts or forbidden.search(member.filename) or (member.external_attr >> 16) & 0o170000 == 0o120000:
                        raise RuntimeError('private material in archive')
        elif path.name.endswith(('.tar.gz', '.tgz', '.tar.xz', '.tar.bz2', '.tar')):
            with tarfile.open(path) as archive:
                for member in archive:
                    if member.name.startswith('/') or '..' in Path(member.name).parts or forbidden.search(member.name) or member.issym() or member.islnk():
                        raise RuntimeError('private material in archive')
        elif path.suffix == '.txt':
            if path.name != 'checksums.txt' or any(not re.fullmatch(r'[0-9a-f]{64}\s+[^\s]+', line) for line in path.read_text().splitlines()):
                raise RuntimeError('unapproved text asset')
        else:
            magic = path.open('rb').read(4)
            if not (magic.startswith((b'\x7fELF', b'MZ')) or magic in (b'\xcf\xfa\xed\xfe', b'\xfe\xed\xfa\xcf', b'\xca\xfe\xba\xbe', b'\xfe\xed\xfa\xce', b'\xce\xfa\xed\xfe')):
                raise RuntimeError('release asset is not a binary')


def main():
    os.umask(0o077)
    parser = argparse.ArgumentParser()
    parser.add_argument('operation', choices=['run', 'clone', 'archive', 'clean', 'verify-assets'])
    parser.add_argument('--repository')
    parser.add_argument('--destination')
    parser.add_argument('--ref', default='main')
    parser.add_argument('--repositories', default='')
    parser.add_argument('--cwd')
    parser.add_argument('--container', action='store_true')
    parser.add_argument('--shell', choices=['bash', 'pwsh'], default='bash')
    parser.add_argument('--status', default='completed')
    parser.add_argument('--assets', nargs='+')
    args = parser.parse_args()
    try:
        if args.operation == 'clone':
            clone(args.repository, args.destination, args.ref)
        elif args.operation == 'run':
            execute(sys.stdin.buffer.read().decode('utf-8'), args.cwd, args.container, args.shell)
        elif args.operation == 'archive':
            finalize(args.repositories, args.status)
        elif args.operation == 'clean':
            shutil.rmtree(root(), ignore_errors=True)
        elif args.operation == 'verify-assets':
            verify_assets(args.assets or [])
    except BaseException:
        # Never print exception text, child output, API bodies or tracebacks.
        try:
            with (root() / 'control.log').open('ab') as handle:
                handle.write(redact(traceback.format_exc()))
        except BaseException:
            pass
        print('Private CI operation failed. Details are restricted to the private ci-private-logs branch.', file=sys.stderr)
        return 1
    return 0


if __name__ == '__main__':
    sys.exit(main())
