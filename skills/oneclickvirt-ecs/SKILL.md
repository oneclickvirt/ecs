---
name: oneclickvirt-ecs
description: "Collect a bounded VPS Fusion Monster system report, interpret results and select explicit follow-up tests."
metadata:
  repository: oneclickvirt/ecs
  integration-version: "1.0.0"
---

# ecs

## Scope

Collect a bounded VPS Fusion Monster system report, interpret results and select explicit follow-up tests.

- Use the maintained Go ECS binary, not the retired Shell edition or GUI executable.
- The system preset disables benchmarks, media/IP checks, network speed tests, logs and public uploads. Additional tests require choosing their resource/traffic scope.
- Resolve the published ECS release once; retain its exact tag and SHA-256. A cached file under a mutable output tag must still be verified.

## Install and update

```bash
curl -fL --retry 3 https://raw.githubusercontent.com/oneclickvirt/ecs/master/scripts/install_agent.py -o install_agent.py
python3 install_agent.py --client codex
```

The installer is also mirrored at `https://cnb.cool/oneclickvirt/ecs/-/git/raw/master/scripts/install_agent.py`; add `--source cnb` to install from an immutable CNB snapshot (requires Git).

Use `--client claude` or `--client openclaw` for another client. Download and run again to update; `--ref` pins a source tag/commit. The installer verifies all SHA-256 checksums before replacing its files, preserves extra local files/client settings, and records the source commit in `install.json`.

## MCP

Requires Python 3.9+. The adapter uses existing local commands and their existing permissions; it does not install a virtualization runtime. Merge the installed `mcp.example.json` into the client configuration. For a remote host, run its adapter over noninteractive SSH.

```bash
python3 install_agent.py --client codex --install-binary
```

`--version` selects the probe release tag; `--ref` selects Skill/MCP source independently. The default resolves the latest stable GitHub release once, selects a real asset for this OS/architecture, and verifies SHA-256. Missing assets fail without guessing. Reuse an existing executable with `--binary /absolute/path/to/program`.

Call `integration_info` first for capabilities and prerequisites.

Available probe presets:

- `system`: `goecs -menu=false -upload=false -log=false -basic=true -cpu=false -memory=false -disk=false -ut=false -security=false -email=false -backtrace=false -nt3=false -speed=false -ping=false -tgdc=false -web=false`

`release_assets` returns the actual version, filenames, sizes, digests and download URLs. Pass an exact tag for a pinned release. After a product update, verify help and an actual call rather than treating file presence as a successful test.

## Source and documentation

- [`go.mod`](https://github.com/oneclickvirt/ecs/blob/master/go.mod)
- [`goecs.sh`](https://github.com/oneclickvirt/ecs/blob/master/goecs.sh)
- [`scripts/embed_geekbench.sh`](https://github.com/oneclickvirt/ecs/blob/master/scripts/embed_geekbench.sh)

- [集成安装说明 / Integration guide](https://www.spiritlhl.net/guide/agent_integrations.html)

Check the relevant source and current documentation before giving commands. Preserve requested version pins, data paths and native host identifiers. Distinguish actual program execution, isolated tests and unverified environments.
