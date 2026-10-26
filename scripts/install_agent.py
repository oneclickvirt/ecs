#!/usr/bin/env python3
"""Install/update this repository's Skill and optional Python stdio MCP adapter."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import subprocess
import sys
import tempfile
import io
import zipfile
import urllib.parse
import urllib.request

REPOSITORY = "oneclickvirt/ecs"
DEFAULT_REF = "master"
SKILL_NAME = "oneclickvirt-ecs"
MAX_FILE = 2 * 1024 * 1024


class SafeRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        redirected = super().redirect_request(req, fp, code, msg, headers, newurl)
        if redirected is not None and urllib.parse.urlsplit(req.full_url).netloc != urllib.parse.urlsplit(newurl).netloc:
            redirected.remove_header("Authorization")
        return redirected


def fetch(url, limit=MAX_FILE):
    headers = {"User-Agent": "oneclickvirt-agent-installer/1.0"}
    token = os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN")
    if token and urllib.parse.urlsplit(url).netloc == "api.github.com":
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(url, headers=headers)
    with urllib.request.build_opener(SafeRedirect()).open(req, timeout=40) as response:
        data = response.read(limit + 1)
    if len(data) > limit:
        raise ValueError("Integration file exceeds the size limit")
    return data


def release_binary(profile, version):
    if not profile.get("asset_prefix"):
        raise ValueError("This integration does not distribute a standalone binary")
    if not re.fullmatch(r"[A-Za-z0-9_.+/-]{1,128}", version):
        raise ValueError("Invalid release tag")
    repo = profile.get("release_repository", REPOSITORY)
    suffix = "latest" if version == "latest" else "tags/" + urllib.parse.quote(version, safe="")
    release = json.loads(fetch("https://api.github.com/repos/" + repo + "/releases/" + suffix))
    if release.get("draft") or (version == "latest" and release.get("prerelease")):
        raise ValueError("Release is not a stable published release")
    system = platform.system().lower()
    machine = platform.machine().lower()
    arch = {"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "arm64": "arm64",
            "i386": "386", "i686": "386", "armv7l": "arm"}.get(machine, machine)
    separator = profile.get("asset_separator", "-")
    stem = separator.join([profile["asset_prefix"], system, arch])
    candidates = [a for a in release["assets"] if a["name"] in (stem, stem + ".exe", stem + ".zip")]
    if len(candidates) != 1:
        raise ValueError("Release has no unique binary for " + system + "/" + arch + "; use --binary with an existing build")
    asset = candidates[0]
    url = asset["browser_download_url"]
    parsed = urllib.parse.urlsplit(url)
    expected_prefix = "https://github.com/" + repo + "/releases/download/"
    if not url.startswith(expected_prefix) or parsed.scheme != "https":
        raise ValueError("Unexpected binary download URL")
    content = fetch(url, limit=256 * 1024 * 1024)
    digest = asset.get("digest") or ""
    checksum = hashlib.sha256(content).hexdigest()
    if digest and digest != "sha256:" + checksum:
        raise ValueError("Published binary digest does not match the download")
    if not digest:
        checksum_assets = [a for a in release["assets"] if a["name"] in ("checksums.txt", "SHA256SUMS")]
        verified = False
        for entry in checksum_assets:
            checksum_url = entry["browser_download_url"]
            if not checksum_url.startswith(expected_prefix):
                raise ValueError("Unexpected checksum download URL")
            for line in fetch(checksum_url).decode().splitlines():
                match = re.fullmatch(r"([a-fA-F0-9]{64})\s+\*?(.+)", line)
                if match and match[2] == asset["name"] and match[1].lower() == checksum:
                    verified = True
        if not verified:
            raise ValueError("Binary has no matching SHA-256 digest/checksum; existing installations were not changed")
    if asset["name"].endswith(".zip"):
        with zipfile.ZipFile(io.BytesIO(content)) as archive:
            entries = [f for f in archive.infolist() if Path(f.filename).name in (profile["binary"], profile["binary"] + ".exe") and not f.is_dir()]
            if len(entries) != 1 or entries[0].file_size > 256 * 1024 * 1024:
                raise ValueError("Archive does not contain one expected executable")
            content = archive.read(entries[0])
    return content, {"version": release["tag_name"], "asset": asset["name"], "archive_sha256": checksum,
                     "binary_sha256": hashlib.sha256(content).hexdigest()}


def resolve_source(source, ref):
    if not re.fullmatch(r"[A-Za-z0-9_.+/-]{1,160}", ref) or ".." in ref:
        raise ValueError("Invalid source ref")
    if source == "github":
        info = json.loads(fetch("https://api.github.com/repos/" + REPOSITORY + "/commits/" + urllib.parse.quote(ref, safe="")))
        commit = info["sha"]
        base = "https://raw.githubusercontent.com/" + REPOSITORY + "/" + commit + "/"
    else:
        # Public Git refs require no credentials. Pin the CNB snapshot itself,
        # whose commit differs from the history-free GitHub source mirror.
        result = subprocess.check_output(["git", "-c", "credential.helper=", "ls-remote", "https://cnb.cool/" + REPOSITORY,
                                          ref if re.fullmatch(r"[a-f0-9]{40}", ref) else "refs/heads/" + ref,
                                          "refs/tags/" + ref, "refs/tags/" + ref + "^{}"], text=True, timeout=45)
        rows = [line.split() for line in result.splitlines()]
        if re.fullmatch(r"[a-f0-9]{40}", ref):
            commit = ref
        elif rows:
            peeled = [row[0] for row in rows if row[1].endswith("^{}")]
            commit = (peeled or [row[0] for row in rows])[0]
        else:
            raise ValueError("CNB source ref does not exist")
        base = "https://cnb.cool/" + REPOSITORY + "/-/git/raw/" + commit + "/"
    if not re.fullmatch(r"[a-f0-9]{40}", commit):
        raise ValueError("Source did not resolve to an immutable commit")
    return commit, lambda path: fetch(base + path)


def load_payload(reader):
    manifest = json.loads(reader("agent/manifest.json"))
    if manifest.get("repository") != REPOSITORY or manifest.get("skill_name") != SKILL_NAME:
        raise ValueError("Manifest does not identify the requested integration")
    expected = {"agent/profile.json", "skills/" + SKILL_NAME + "/SKILL.md", "skills/" + SKILL_NAME + "/SKILL_ZH.md"}
    if manifest.get("mcp"):
        expected.add("agent/server.py")
    if set(manifest.get("files", {})) != expected:
        raise ValueError("Manifest contains missing or unexpected file paths")
    data = {}
    for path, digest in manifest["files"].items():
        content = reader(path)
        if not re.fullmatch(r"[0-9a-f]{64}", digest) or hashlib.sha256(content).hexdigest() != digest:
            raise ValueError("SHA-256 verification failed for " + path)
        data[path] = content
    profile = json.loads(data["agent/profile.json"])
    if profile.get("repository") != REPOSITORY or profile.get("skill_name") != SKILL_NAME:
        raise ValueError("Profile does not match the requested repository")
    return manifest, data


def atomic_write(target, content):
    target.parent.mkdir(parents=True, exist_ok=True)
    if target.is_symlink():
        raise ValueError("Refusing to replace a symlink: " + str(target))
    handle, temporary = tempfile.mkstemp(prefix=".oneclickvirt-", dir=target.parent)
    try:
        with os.fdopen(handle, "wb") as stream:
            stream.write(content)
        os.replace(temporary, target)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def install(manifest, payload, skill_dir, mcp_dir, record, skill_only=False, binary=None, binary_content=None):
    skill_dir = Path(skill_dir).expanduser().absolute()
    mcp_dir = Path(mcp_dir).expanduser().absolute()
    if skill_dir.is_symlink() or mcp_dir.is_symlink():
        raise ValueError("The integration destination itself must not be a symlink")
    has_mcp = manifest.get("mcp") and not skill_only
    writes = {}
    if binary_content is not None:
        writes[Path(binary)] = binary_content
    for filename in ("SKILL.md", "SKILL_ZH.md"):
        writes[skill_dir / filename] = payload["skills/" + SKILL_NAME + "/" + filename]
    if has_mcp:
        writes[mcp_dir / "server.py"] = payload["agent/server.py"]
        writes[mcp_dir / "profile.json"] = payload["agent/profile.json"]
        env = {}
        if binary:
            env["ONECLICKVIRT_MCP_BINARY"] = str(Path(binary).expanduser().resolve())
        config = {"mcpServers": {SKILL_NAME: {"command": sys.executable,
                    "args": [str(mcp_dir / "server.py")], "env": env}}}
        writes[skill_dir / "mcp.example.json"] = (json.dumps(config, ensure_ascii=False, indent=2) + "\n").encode()
        writes[skill_dir / "mcp.claude.json"] = writes[skill_dir / "mcp.example.json"]
        server = dict(config["mcpServers"][SKILL_NAME], requestTimeoutMs=180000)
        writes[skill_dir / "mcp.openclaw.json"] = (json.dumps({"mcp": {"servers": {SKILL_NAME: server}}}, ensure_ascii=False, indent=2) + "\n").encode()
        toml = "[mcp_servers." + SKILL_NAME + "]\ncommand = " + json.dumps(sys.executable, ensure_ascii=False)
        toml += "\nargs = " + json.dumps([str(mcp_dir / "server.py")], ensure_ascii=False) + "\ntool_timeout_sec = 180\n"
        if env:
            toml += "\n[mcp_servers." + SKILL_NAME + ".env]\n"
            toml += "".join(key + " = " + json.dumps(value, ensure_ascii=False) + "\n" for key, value in env.items())
        writes[skill_dir / "mcp.codex.toml"] = toml.encode("utf-8")
    record = dict(record, repository=REPOSITORY, skill_name=SKILL_NAME, mcp_installed=bool(has_mcp),
                  files=manifest["files"])
    writes[skill_dir / "install.json"] = (json.dumps(record, indent=2) + "\n").encode()
    # Verify the entire payload and every destination before replacing anything.
    for path in writes:
        if path.is_symlink():
            raise ValueError("Symlink installation file is not supported: " + str(path))
        if path.exists() and not path.is_file():
            raise ValueError("Destination is not a file: " + str(path))
    for path, content in writes.items():
        atomic_write(path, content)
        if binary_content is not None and path == Path(binary) and os.name != "nt":
            path.chmod(0o755)
    return {"skill_dir": str(skill_dir), "mcp_dir": str(mcp_dir) if has_mcp else None,
            "source_commit": record["commit"], "mcp_installed": bool(has_mcp)}


def main():
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8")
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--client", choices=["claude", "codex", "openclaw"], default="claude")
    parser.add_argument("--source", choices=["github", "cnb"], default="github")
    parser.add_argument("--ref", default=DEFAULT_REF, help="Source branch, tag or immutable commit")
    parser.add_argument("--skill-dir", type=Path)
    parser.add_argument("--mcp-dir", type=Path, default=Path.home() / ".local/share/oneclickvirt-mcp" / SKILL_NAME)
    parser.add_argument("--skill-only", action="store_true")
    parser.add_argument("--binary", help="Absolute path to an already installed runtime/probe binary")
    parser.add_argument("--install-binary", action="store_true", help="Install the verified published probe alongside the adapter")
    parser.add_argument("--version", default="latest", help="Probe release tag; independent of the integration source --ref")
    parser.add_argument("--bin-dir", type=Path, default=Path.home() / ".local/bin")
    parser.add_argument("--source-dir", type=Path, help="Install from a verified local source checkout")
    args = parser.parse_args()
    if args.source_dir:
        root = args.source_dir.resolve()
        commit = subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()
        reader = lambda path: (root / path).read_bytes()
    else:
        commit, reader = resolve_source(args.source, args.ref)
    manifest, payload = load_payload(reader)
    if args.binary and args.install_binary:
        raise ValueError("Choose --binary or --install-binary")
    binary_content = None
    binary_record = {}
    binary = args.binary
    if binary:
        candidate = Path(binary).expanduser()
        if not candidate.is_absolute() or not candidate.is_file() or (os.name != "nt" and not os.access(candidate, os.X_OK)):
            raise ValueError("--binary must name an existing executable by its absolute path")
    if args.install_binary:
        if args.skill_only or not manifest.get("mcp"):
            raise ValueError("Binary installation requires an MCP-enabled integration")
        profile = json.loads(payload["agent/profile.json"])
        binary_content, binary_record = release_binary(profile, args.version)
        binary = str(args.bin_dir.expanduser().absolute() / (profile["binary"] + (".exe" if os.name == "nt" else "")))
    parents = {"claude": Path.home() / ".claude/skills", "codex": Path.home() / ".agents/skills",
               "openclaw": Path.home() / ".openclaw/skills"}
    result = install(manifest, payload, args.skill_dir or parents[args.client] / SKILL_NAME,
                     args.mcp_dir, {"source": args.source, "ref": args.ref, "commit": commit, "binary": binary_record},
                     args.skill_only, binary, binary_content)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    if result["mcp_installed"]:
        filename = {"codex": "mcp.codex.toml", "claude": "mcp.claude.json", "openclaw": "mcp.openclaw.json"}[args.client]
        print("Merge " + filename + " into your client's MCP settings. Existing settings and tokens were not modified.")
    else:
        print("Installed Skill files. This repository does not install an MCP server in this mode.")


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, subprocess.SubprocessError) as exc:
        print("Installation failed: " + str(exc), file=sys.stderr)
        sys.exit(1)
