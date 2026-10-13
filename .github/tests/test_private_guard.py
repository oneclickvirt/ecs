import base64
import contextlib
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
import zipfile
import tarfile

spec = importlib.util.spec_from_file_location('guard', '.github/scripts/private_guard.py')
guard = importlib.util.module_from_spec(spec)
spec.loader.exec_module(guard)


class GuardTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.env = patch.dict(os.environ, {'CI_STATE_ROOT': self.temp.name, 'CI_JOB_KEY': 'fixture', 'GITHUB_RUN_ID': '1', 'GITHUB_RUN_ATTEMPT': '1', 'PRIVATE_CI_TOKEN': 'sensitive-fixture-value'}, clear=False)
        self.env.start()

    def tearDown(self):
        self.env.stop()
        self.temp.cleanup()

    def test_stdout_stderr_and_workflow_commands_are_private(self):
        console = io.StringIO()
        with contextlib.redirect_stdout(console), contextlib.redirect_stderr(console):
            with self.assertRaises(RuntimeError):
                guard.execute("printf 'PRIVATE_SOURCE_MARKER\\n::error::PRIVATE_SOURCE_MARKER\\n'\nprintf 'PRIVATE_ERROR_MARKER\\n' >&2\nexit 7\n")
        self.assertNotIn('PRIVATE_', console.getvalue())
        log = next(guard.root().glob('step-*.log')).read_text()
        self.assertIn('PRIVATE_SOURCE_MARKER', log)
        self.assertIn('PRIVATE_ERROR_MARKER', log)

    def test_unapproved_exports_cannot_reach_actions(self):
        actual = Path(self.temp.name) / 'actual-env'
        with patch.dict(os.environ, {'GITHUB_ENV': str(actual), 'CI_ALLOWED_ENV': 'VERSION'}):
            with self.assertRaises(RuntimeError):
                guard.execute('echo "PRIVATE_SOURCE=secret" >> "$GITHUB_ENV"\n')
        self.assertFalse(actual.exists())

    def test_secret_redaction(self):
        value = os.environ['PRIVATE_CI_TOKEN']
        message = value + ' ' + base64.b64encode(value.encode()).decode() + ' github_pat_something_secret_123 ' + '-----BEGIN RSA PRIVATE KEY-----\nprivate-material\n-----END RSA PRIVATE KEY-----'
        result = guard.redact(message)
        self.assertNotIn(value.encode(), result)
        self.assertNotIn(b'private-material', result)
        self.assertNotIn(b'github_pat_', result)

    def test_windows_uses_git_bash_instead_of_wsl_alias(self):
        folder = Path(self.temp.name) / 'Program Files'
        binary = folder / 'Git/bin/bash.exe'
        binary.parent.mkdir(parents=True)
        binary.write_bytes(b'controlled executable fixture')
        with patch.dict(os.environ, {'ProgramFiles': str(folder)}):
            self.assertEqual(guard.bash_path(platform='nt'), str(binary))

    def test_unicode_script_and_output_preserved_privately(self):
        console = io.StringIO()
        with contextlib.redirect_stdout(console), contextlib.redirect_stderr(console):
            guard.execute("printf '中文诊断 ✓\\n'\n")
        self.assertEqual(console.getvalue(), '')
        self.assertIn('中文诊断 ✓', next(guard.root().glob('step-*.log')).read_text(encoding='utf-8'))

    @unittest.skipIf(os.name == 'nt', 'POSIX file creation permissions')
    def test_build_permissions_do_not_inherit_controller_umask(self):
        previous = os.umask(0o077)
        try:
            state = guard.root()
            guard.execute("touch build-fixture\n", cwd=state)
            self.assertEqual((state / 'build-fixture').stat().st_mode & 0o777, 0o644)
            self.assertEqual(state.stat().st_mode & 0o777, 0o700)
            self.assertEqual(next(state.glob('step-*.log')).stat().st_mode & 0o777, 0o600)
        finally:
            os.umask(previous)

    def test_source_and_keystore_archives_rejected(self):
        path = Path(self.temp.name) / 'release.zip'
        for name in ['deps/private/source.go', 'source.ts', 'source.java', 'signing.jks', '../outside', 'build.log']:
            with zipfile.ZipFile(path, 'w') as z:
                z.writestr(name, 'private')
            with self.assertRaises(RuntimeError):
                guard.verify_assets([str(path)])

    def test_linux_xz_package_verified_and_private_members_rejected(self):
        path = Path(self.temp.name) / 'release.tar.xz'
        for name, accepted in [('usr/bin/application', True), ('source.go', False), ('keys/signing.jks', False), ('../outside', False)]:
            with tarfile.open(path, 'w:xz') as archive:
                info = tarfile.TarInfo(name)
                payload = b'\x7fELFcontrolled-fixture'
                info.size = len(payload)
                archive.addfile(info, io.BytesIO(payload))
            if accepted:
                guard.verify_assets([str(path)])
            else:
                with self.assertRaises(RuntimeError):
                    guard.verify_assets([str(path)])

    def test_public_log_destination_rejected_before_git(self):
        with patch.object(guard, 'api', return_value={'private': False}), patch.object(guard, 'quiet') as command:
            with self.assertRaises(RuntimeError):
                guard.archive('security')
            command.assert_not_called()

    def test_five_runs_and_one_commit_with_real_git(self):
        remote = Path(self.temp.name) / 'remote.git'
        subprocess.run(['git', 'init', '-q', '--bare', str(remote)], check=True)
        original = guard.quiet
        def local(command, *args, **kwargs):
            command = [str(remote) if x == 'https://github.com/oneclickvirt/security.git' else x for x in command]
            return original(command, *args, **kwargs)
        with patch.object(guard, 'api', return_value={'private': True}), patch.object(guard, 'quiet', side_effect=local):
            for run in range(1, 7):
                os.environ['GITHUB_RUN_ID'] = str(run)
                (guard.root() / 'fixture.log').write_text('controlled private fixture')
                guard.archive('security', 'success')
        names = subprocess.check_output(['git', '--git-dir', str(remote), 'ls-tree', '-r', '--name-only', 'ci-private-logs'], text=True)
        runs = {x.split('/')[1] for x in names.splitlines() if x.startswith('runs/')}
        self.assertEqual(runs, {'2-1', '3-1', '4-1', '5-1', '6-1'})
        count = subprocess.check_output(['git', '--git-dir', str(remote), 'rev-list', '--count', 'ci-private-logs'], text=True).strip()
        self.assertEqual(count, '1')


if __name__ == '__main__':
    unittest.main(verbosity=2)
