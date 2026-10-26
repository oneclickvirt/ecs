#!/usr/bin/env bash
# Copy only public integration metadata; release assets and README stay intact.
set -euo pipefail
source_ref=${1:?Source Git revision is required}
destination=${2:?Destination checkout is required}
source_commit=$(git rev-parse --verify "${source_ref}^{commit}")
paths=(
  SKILL.md SKILL_ZH.md .cnb.yml .cnb/settings.yml
  agent/profile.json agent/manifest.json agent/server.py
  scripts/install_agent.py
  skills/oneclickvirt-ecs/SKILL.md skills/oneclickvirt-ecs/SKILL_ZH.md
)
for path in "${paths[@]}"; do
  git cat-file -e "${source_commit}:${path}"
done
for path in "${paths[@]}"; do
  mkdir -p "$destination/$(dirname "$path")"
  git show "${source_commit}:${path}" > "$destination/$path"
done
