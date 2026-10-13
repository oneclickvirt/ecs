"""Merge Dependabot Action patch/minor updates after all head checks pass."""
import json
import os
import re
import time
import urllib.request

repository = os.environ['GITHUB_REPOSITORY']
event = json.load(open(os.environ['GITHUB_EVENT_PATH']))

def call(path, method='GET', body=None):
    req = urllib.request.Request('https://api.github.com/repos/' + repository + path, method=method,
        data=None if body is None else json.dumps(body).encode(),
        headers={'Authorization': 'Bearer ' + os.environ['GH_TOKEN'], 'Accept': 'application/vnd.github+json', 'Content-Type': 'application/json'})
    with urllib.request.urlopen(req, timeout=40) as response:
        value = response.read()
    return json.loads(value) if value else None

def versioned_changes(files):
    count = 0
    for file in files:
        path = file['filename']
        if not path.startswith('.github/') or not path.endswith(('.yml', '.yaml')) or 'patch' not in file:
            return False
        old, new = {}, {}
        for line in file['patch'].splitlines():
            if not line.startswith(('-', '+')) or line.startswith(('---', '+++')):
                continue
            match = re.match(r'[+-]\s*(?:-\s*)?uses:\s*([^\s@]+)@[0-9a-f]{40}\s+#\s*v?(\d+)(?:\.(\d+))?(?:\.(\d+))?\s*$', line)
            if not match:
                # Updating any workflow behavior or an unversioned action
                # remains a reviewed change.
                return False
            (old if line[0] == '-' else new)[match[1]] = tuple(int(x or 0) for x in match.groups()[1:])
        if set(old) != set(new):
            return False
        for action in old:
            if old[action][0] != new[action][0] or new[action] < old[action]:
                return False
            count += 1
    return count > 0

run = event['workflow_run']
if run['conclusion'] == 'success' and run['event'] == 'pull_request':
    for reference in run.get('pull_requests', []):
        number = reference['number']
        pr = call('/pulls/' + str(number))
        if pr['state'] != 'open' or pr['user']['login'] != 'dependabot[bot]' or pr['head']['repo']['full_name'] != repository:
            continue
        if pr['head']['sha'] != run['head_sha'] or pr.get('draft'):
            continue
        files = call('/pulls/' + str(number) + '/files?per_page=100')
        if pr['changed_files'] > 100 or not versioned_changes(files):
            continue
        sha = pr['head']['sha']
        for attempt in range(90):
            checks = call('/commits/' + sha + '/check-runs?per_page=100')['check_runs']
            statuses = call('/commits/' + sha + '/status')['statuses']
            if any(x['status'] == 'completed' and x['conclusion'] not in ('success', 'neutral', 'skipped') for x in checks) or any(x['state'] in ('failure', 'error') for x in statuses):
                break
            if checks and all(x['status'] == 'completed' for x in checks) and all(x['state'] == 'success' for x in statuses):
                latest = call('/pulls/' + str(number))
                if latest['head']['sha'] == sha and latest['state'] == 'open':
                    call('/pulls/' + str(number) + '/merge', 'PUT', {'sha': sha, 'merge_method': 'squash'})
                break
            time.sleep(20)
