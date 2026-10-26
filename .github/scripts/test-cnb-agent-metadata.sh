#!/usr/bin/env bash
set -euo pipefail
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
test_tmp=$(mktemp -d)
trap 'rm -rf "$test_tmp"' EXIT
printf 'existing release binary\n' > "$test_tmp/goecs_linux_amd64"
printf 'existing release documentation\n' > "$test_tmp/README.md"
bash "$script_dir/copy-cnb-agent-metadata.sh" HEAD "$test_tmp"
[[ "$(cat "$test_tmp/goecs_linux_amd64")" == 'existing release binary' ]]
[[ "$(cat "$test_tmp/README.md")" == 'existing release documentation' ]]
while IFS= read -r path; do
  cmp <(git show "HEAD:$path") "$test_tmp/$path"
done < <(find "$test_tmp" -type f | sed "s|^$test_tmp/||" | sort | sed '/^README.md$/d; /^goecs_linux_amd64$/d')
[[ "$(find "$test_tmp" -type f | wc -l | tr -d ' ')" == 12 ]]
echo 'Public integration metadata copied; existing release assets and README preserved.'
