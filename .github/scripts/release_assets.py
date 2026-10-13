#!/usr/bin/env python3
"""Stage and verify every new asset before replacing any published asset."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import sys
import time
import urllib.error
import urllib.parse
import urllib.request


class Publisher:
    def __init__(self, repository, credential):
        if not re.fullmatch(r'[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+', repository) or not credential:
            raise ValueError('missing publication configuration')
        self.base = '/repos/' + repository
        self.credential = credential
        self.key = os.environ.get('GITHUB_RUN_ID', 'local') + '-' + os.environ.get('GITHUB_RUN_ATTEMPT', '1')

    def request(self, path, method='GET', body=None):
        for attempt in range(3):
            req = urllib.request.Request('https://api.github.com' + path, method=method,
                data=None if body is None else json.dumps(body).encode(), headers={
                    'Authorization': 'Bearer ' + self.credential, 'Accept': 'application/vnd.github+json',
                    'Content-Type': 'application/json', 'X-GitHub-Api-Version': '2022-11-28'})
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

    def assets(self, release_id):
        result = []
        for page in range(1, 21):
            rows = self.request(f'{self.base}/releases/{release_id}/assets?per_page=100&page={page}')
            result.extend(rows)
            if len(rows) < 100:
                return {row['name']: row for row in result}
        raise RuntimeError('asset pagination limit exceeded')

    def ensure_release(self, tag):
        endpoint = self.base + '/releases/tags/' + urllib.parse.quote(tag, safe='')
        try:
            return self.request(endpoint)
        except urllib.error.HTTPError as exc:
            if exc.code != 404:
                raise
        try:
            return self.request(self.base + '/releases', 'POST', {
                'tag_name': tag, 'name': tag, 'body': 'Validated automated binaries',
                'target_commitish': os.environ.get('GITHUB_SHA', 'main')})
        except (urllib.error.URLError, TimeoutError):
            # Creation may have succeeded even if its response was lost.
            return self.request(endpoint)

    def rename(self, asset_id, name):
        endpoint = self.base + '/releases/assets/' + str(asset_id)
        for attempt in range(3):
            try:
                return self.request(endpoint, 'PATCH', {'name': name})
            except (urllib.error.URLError, TimeoutError):
                if self.request(endpoint)['name'] == name:
                    return
                if attempt == 2:
                    raise
                time.sleep(2 ** attempt)

    def delete(self, asset_id):
        endpoint = self.base + '/releases/assets/' + str(asset_id)
        for attempt in range(3):
            try:
                self.request(endpoint, 'DELETE')
                return
            except urllib.error.HTTPError as exc:
                if exc.code == 404:
                    return
                if exc.code not in (408, 429, 500, 502, 503, 504):
                    raise
            except (urllib.error.URLError, TimeoutError):
                pass
            if attempt == 2:
                raise RuntimeError('asset cleanup failed')
            time.sleep(2 ** attempt)

    def stage(self, release_id, path):
        name = path.name + '.ci-stage-' + self.key
        digest = self.digest(path)
        size = path.stat().st_size
        for attempt in range(3):
            existing = self.assets(release_id).get(name)
            if existing:
                if existing.get('state') == 'uploaded' and existing['size'] == size and existing.get('digest') == 'sha256:' + digest:
                    return existing
                self.delete(existing['id'])
            try:
                with path.open('rb') as handle:
                    req = urllib.request.Request(
                        'https://uploads.github.com' + self.base + f'/releases/{release_id}/assets?name=' + urllib.parse.quote(name, safe=''),
                        data=handle, method='POST', headers={'Authorization': 'Bearer ' + self.credential,
                            'Content-Type': 'application/octet-stream', 'Content-Length': str(size)})
                    with urllib.request.urlopen(req, timeout=600) as response:
                        asset = json.load(response)
                if asset.get('state') != 'uploaded' or asset['size'] != size:
                    raise RuntimeError('incomplete asset upload')
                if asset.get('digest') and asset['digest'] != 'sha256:' + digest:
                    raise RuntimeError('asset checksum mismatch')
                return asset
            except urllib.error.HTTPError as exc:
                if exc.code not in (408, 429, 500, 502, 503, 504) or attempt == 2:
                    raise
            except (TimeoutError, urllib.error.URLError):
                if attempt == 2:
                    raise
            time.sleep(2 ** attempt)

    @staticmethod
    def digest(path):
        result = hashlib.sha256()
        with path.open('rb') as handle:
            for block in iter(lambda: handle.read(1024 * 1024), b''):
                result.update(block)
        return result.hexdigest()

    def publish(self, tag, paths):
        if not paths or len({p.name for p in paths}) != len(paths):
            raise ValueError('missing or duplicate publication assets')
        for path in paths:
            if not path.is_file() or path.is_symlink() or path.stat().st_size == 0:
                raise ValueError('invalid publication asset')
        release = self.ensure_release(tag)
        originals = self.assets(release['id'])
        staged = [(path, self.stage(release['id'], path)) for path in paths]
        backups, promoted = [], []
        try:
            for path, asset in staged:
                old = originals.get(path.name)
                if old:
                    self.rename(old['id'], path.name + '.ci-backup-' + self.key)
                    backups.append((path.name, old))
                self.rename(asset['id'], path.name)
                promoted.append((path, asset))
        except Exception:
            # Restore the previous complete snapshot if promotion fails midway.
            for path, asset in reversed(promoted):
                self.rename(asset['id'], path.name + '.ci-stage-' + self.key)
            for name, old in reversed(backups):
                self.rename(old['id'], name)
            raise
        for _, old in backups:
            self.delete(old['id'])


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('tag')
    parser.add_argument('files', nargs='+', type=Path)
    args = parser.parse_args()
    Publisher(os.environ.get('GH_REPO', os.environ.get('GITHUB_REPOSITORY', '')),
              os.environ.get('GH_TOKEN', os.environ.get('GITHUB_TOKEN', ''))).publish(args.tag, args.files)
    print('Validated release assets published.')


if __name__ == '__main__':
    try:
        main()
    except Exception:
        print('Release publication failed.', file=sys.stderr)
        sys.exit(1)
