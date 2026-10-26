#!/usr/bin/env python3
"""OneClickVirt repository MCP adapter. Python 3.9+, stdio, no extra packages."""
import argparse
import ipaddress
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import sys
import tempfile
import urllib.error
import urllib.parse
import urllib.request

VERSION = "1.0.0"
PROTOCOLS = ("2024-11-05", "2025-03-26", "2025-06-18", "2025-11-25")
LIMIT = 1024 * 1024


def schema(properties=None, required=()):
    return {"type": "object", "properties": properties or {},
            "required": list(required), "additionalProperties": False}


def string(description="", **extra):
    return dict(type="string", description=description, **extra)


def validate(value, spec, name="arguments"):
    expected = spec.get("type")
    if expected == "object":
        if not isinstance(value, dict):
            raise ValueError(name + " must be an object")
        props = spec.get("properties", {})
        if set(spec.get("required", [])) - value.keys():
            raise ValueError(name + " has missing required fields")
        if not spec.get("additionalProperties", True) and value.keys() - props.keys():
            raise ValueError(name + " has unknown fields")
        for key, data in value.items():
            if key in props:
                validate(data, props[key], key)
    elif expected == "string":
        maximum = min(512, spec.get("maxLength", 512))
        if not isinstance(value, str) or len(value) > maximum:
            raise ValueError(name + " must be a string of at most " + str(maximum) + " characters")
        if "pattern" in spec and not re.fullmatch(spec["pattern"], value):
            raise ValueError(name + " has invalid characters")
    elif expected == "integer":
        if isinstance(value, bool) or not isinstance(value, int):
            raise ValueError(name + " must be an integer")
        if value < spec.get("minimum", value) or value > spec.get("maximum", value):
            raise ValueError(name + " is outside the supported range")
    if "enum" in spec and value not in spec["enum"]:
        raise ValueError(name + " is not a supported choice")


def execute(argv, timeout=30, cwd=None, parse_json=True):
    """Bound output, kill process groups on timeout, never invoke a shell."""
    if not argv or not shutil.which(argv[0]):
        raise ValueError("Required local command is missing: " + (argv[0] if argv else ""))
    with tempfile.TemporaryFile() as out, tempfile.TemporaryFile() as err:
        proc = subprocess.Popen(argv, stdin=subprocess.DEVNULL, stdout=out, stderr=err,
                                start_new_session=os.name != "nt", cwd=cwd)
        try:
            proc.wait(timeout=timeout)
        except subprocess.TimeoutExpired:
            if os.name != "nt":
                os.killpg(proc.pid, signal.SIGKILL)
            else:
                subprocess.run(["taskkill", "/PID", str(proc.pid), "/T", "/F"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=False)
                if proc.poll() is None:
                    proc.kill()
            proc.wait()
            raise ValueError("Command exceeded its time limit; no result is claimed")
        out.seek(0)
        err.seek(0)
        stdout, stderr = out.read(LIMIT + 1), err.read(16385)
        if proc.returncode:
            raise ValueError("Command failed (exit %s): %s" % (
                proc.returncode, (stderr[:16384] or stdout[:16384]).decode("utf-8", "replace")))
        if len(stdout) > LIMIT:
            raise ValueError("Command result exceeds 1 MiB; narrow the request")
        text = stdout.decode("utf-8", "replace").strip()
        if not parse_json:
            return text
        try:
            return json.loads(text)
        except ValueError:
            return text


def github_json(path):
    headers = {"Accept": "application/vnd.github+json", "User-Agent": "oneclickvirt-mcp/" + VERSION}
    token = os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN")
    if token:
        headers["Authorization"] = "Bearer " + token
    class SafeRedirect(urllib.request.HTTPRedirectHandler):
        def redirect_request(self, req, fp, code, msg, response_headers, newurl):
            redirected = super().redirect_request(req, fp, code, msg, response_headers, newurl)
            if redirected is not None and urllib.parse.urlsplit(newurl).netloc != "api.github.com":
                redirected.remove_header("Authorization")
            return redirected
    req = urllib.request.Request("https://api.github.com" + path, headers=headers)
    try:
        with urllib.request.build_opener(SafeRedirect()).open(req, timeout=25) as response:
            data = response.read(8 * LIMIT + 1)
    except urllib.error.HTTPError as exc:
        raise ValueError("GitHub API returned %s; no release or asset is assumed" % exc.code) from None
    if len(data) > 8 * LIMIT:
        raise ValueError("GitHub response is too large")
    return json.loads(data)


class Adapter:
    def __init__(self, profile_path):
        self.path = Path(profile_path).resolve()
        self.profile = json.loads(self.path.read_text(encoding="utf-8"))
        self.repo = self.profile["repository"]
        if not re.fullmatch(r"oneclickvirt/[A-Za-z0-9_.-]+", self.repo):
            raise ValueError("Invalid repository profile")

    def binary(self):
        return os.environ.get("ONECLICKVIRT_MCP_BINARY") or self.profile.get("binary", "")

    def tools(self):
        result = []
        def add(name, description, spec=None, write=False, external=False):
            result.append({"name": name, "description": description, "inputSchema": spec or schema(),
                           "annotations": {"readOnlyHint": not write, "destructiveHint": write,
                                           "idempotentHint": not write, "openWorldHint": external}})
        add("integration_info", "Show repository, supported operations, prerequisites and documentation.")
        if self.profile.get("releases"):
            add("release_assets", "Resolve a current published release and list its actual assets. No assumed architectures or tags.",
                schema({"version": string("Exact release tag, or latest", pattern=r"[A-Za-z0-9_.+/-]{1,128}"),
                        "filter": string("Optional case-insensitive substring in asset names", maxLength=128)}), external=True)
        backend = self.profile.get("runtime")
        if backend:
            add("list_instances", "List instances on this local runtime using its existing CLI.")
            add("instance_status", "Read one instance's state without returning environment variables or cloud-init secrets.",
                schema({"id": string(pattern=r"[A-Za-z0-9][A-Za-z0-9_.-]{0,127}")}, ["id"]))
            if os.environ.get("ONECLICKVIRT_MCP_ALLOW_WRITE") == "1":
                add("instance_action", "Start, gracefully stop or restart an explicitly named instance. No delete or reinstall operation.",
                    schema({"id": string(pattern=r"[A-Za-z0-9][A-Za-z0-9_.-]{0,127}"),
                            "action": string(enum=["start", "stop", "restart"])}, ["id", "action"]), write=True)
        if self.profile.get("presets"):
            add("probe_help", "Read the installed probe's actual help output.")
            add("run_probe", "Run one bounded preset with the existing local binary. Check integration_info for its resource and network effects.",
                schema({"preset": string(enum=list(self.profile["presets"])),
                        "target": string("Only for a target-aware preset", pattern=r"[A-Za-z0-9][A-Za-z0-9.:_-]{0,252}")}, ["preset"]),
                write=self.profile.get("benchmark", False), external=True)
        if self.profile.get("network"):
            add("network_status", "Read Linux IPv6 addresses and routes. Does not change network configuration.")
        if self.profile.get("counter"):
            add("counter_monthly", "Read an existing counter's monthly statistics without incrementing it.",
                schema({"counter": string(pattern=r"[A-Za-z0-9_.-]{1,128}")}, ["counter"]), external=True)
        return result

    def instances(self):
        backend = self.profile["runtime"]
        binary = self.binary()
        if backend in ("docker", "podman", "containerd"):
            raw = execute([binary, "ps", "-a", "--format", "{{json .}}"])
            if isinstance(raw, dict):
                rows = [raw]
            elif not raw:
                rows = []
            else:
                rows = [json.loads(line) for line in raw.splitlines()]
            result = []
            for row in rows:
                item = {key: row.get(key) for key in ("ID", "Names", "Image", "State", "Status")}
                if isinstance(item["Names"], list):
                    item["Names"] = ",".join(item["Names"])
                result.append(item)
            return result
        if backend in ("incus", "lxd"):
            rows = execute([binary, "list", "--format=json"])
            if not isinstance(rows, list):
                raise ValueError("Runtime did not return an instance array")
            return [{k: row.get(k) for k in ("name", "status", "type", "location")} for row in rows]
        if backend == "pve":
            return execute([binary, "get", "/cluster/resources", "--type", "vm", "--output-format", "json"])
        if backend == "qemu":
            uri = os.environ.get("ONECLICKVIRT_LIBVIRT_URI", "qemu:///system")
            if not re.fullmatch(r"qemu:///(system|session)", uri):
                raise ValueError("Only local libvirt system/session URIs are supported")
            names = execute([binary, "-c", uri, "list", "--all", "--name"], parse_json=False)
            return [{"name": name} for name in str(names).splitlines() if name]
        if backend == "kubevirt":
            namespace = os.environ.get("ONECLICKVIRT_KUBE_NAMESPACE", "default")
            if not re.fullmatch(r"[a-z0-9][a-z0-9-]{0,62}", namespace):
                raise ValueError("Invalid Kubernetes namespace")
            data = execute([binary, "get", "virtualmachines", "--namespace", namespace, "-o", "json"])
            return [{"name": row["metadata"]["name"], "namespace": namespace,
                     "status": row.get("status", {})} for row in data["items"]]
        raise ValueError("Unsupported runtime")

    def find_instance(self, identifier):
        rows = self.instances()
        for row in rows:
            names = str(row.get("Names", "")).split(",")
            if identifier in names + [str(row.get(k, "")) for k in ("name", "vmid", "ID")]:
                return row
        raise ValueError("Instance not found on the configured runtime")

    def call(self, name, args):
        available = {t["name"]: t for t in self.tools()}
        if name not in available:
            raise ValueError("Unknown or disabled tool")
        validate(args, available[name]["inputSchema"])
        if name == "integration_info":
            return self.profile
        if name == "release_assets":
            repo = self.profile.get("release_repository", self.repo)
            version = args.get("version", "latest")
            suffix = "latest" if version == "latest" else "tags/" + urllib.parse.quote(version, safe="")
            value = github_json("/repos/" + repo + "/releases/" + suffix)
            if value.get("draft") or (version == "latest" and value.get("prerelease")):
                raise ValueError("Requested release is not a stable published release")
            match = args.get("filter", "").lower()
            return {"repository": repo, "version": value["tag_name"], "url": value["html_url"],
                    "published_at": value["published_at"], "assets": [
                        {k: a.get(k) for k in ("name", "size", "digest", "browser_download_url")}
                        for a in value["assets"] if match in a["name"].lower()]}
        if name == "list_instances":
            return self.instances()
        if name == "instance_status":
            row = self.find_instance(args["id"])
            if self.profile["runtime"] == "qemu":
                uri = os.environ.get("ONECLICKVIRT_LIBVIRT_URI", "qemu:///system")
                row["state"] = execute([self.binary(), "-c", uri, "domstate", args["id"]])
            return row
        if name == "instance_action":
            row = self.find_instance(args["id"])
            backend, action, identifier = self.profile["runtime"], args["action"], args["id"]
            binary = self.binary()
            if backend in ("docker", "containerd", "podman", "incus", "lxd"):
                argv = [binary, action, identifier]
            elif backend == "qemu":
                argv = [binary, "-c", os.environ.get("ONECLICKVIRT_LIBVIRT_URI", "qemu:///system"),
                        {"start": "start", "stop": "shutdown", "restart": "reboot"}[action], identifier]
            elif backend == "pve":
                node = row.get("node", "")
                kind = row.get("type", "")
                if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.-]*", node) or kind not in ("qemu", "lxc") or not identifier.isdecimal():
                    raise ValueError("Invalid PVE resource identity")
                endpoint = "/nodes/%s/%s/%s/status/%s" % (node, kind, identifier, {"stop": "shutdown", "restart": "reboot"}.get(action, action))
                argv = [binary, "create", endpoint]
            else:
                namespace = row["namespace"]
                argv = ["virtctl", action, identifier, "--namespace", namespace]
            return {"id": identifier, "action": action, "output": execute(argv, timeout=90), "accepted": True}
        if name == "probe_help":
            return execute([self.binary()] + self.profile.get("help_args", ["-h"]), timeout=15)
        if name == "run_probe":
            preset = self.profile["presets"][args["preset"]]
            if self.profile.get("benchmark") and os.environ.get("ONECLICKVIRT_MCP_ALLOW_BENCHMARKS") != "1":
                raise ValueError("Benchmark execution is disabled; set ONECLICKVIRT_MCP_ALLOW_BENCHMARKS=1 on an intended test host")
            if bool(preset.get("target")) != bool(args.get("target")):
                raise ValueError("This preset %s a target" % ("requires" if preset.get("target") else "does not accept"))
            argv = [self.binary()] + preset["args"]
            if preset.get("target"):
                if self.repo == "oneclickvirt/backtrace":
                    ipaddress.ip_address(args["target"])
                argv += [preset["target"], args["target"]]
            with tempfile.TemporaryDirectory(prefix="oneclickvirt-probe-") as directory:
                argv = [a.replace("{scratch}", directory) for a in argv]
                result = execute(argv, timeout=preset.get("timeout", 45), cwd=directory)
            return {"preset": args["preset"], "result": result}
        if name == "network_status":
            return {"addresses": execute(["ip", "-j", "-6", "addr", "show"]),
                    "routes": execute(["ip", "-j", "-6", "route", "show"])}
        if name == "counter_monthly":
            base = os.environ.get("ONECLICKVIRT_COUNTER_URL", "https://hits.spiritlhl.net").rstrip("/")
            parsed = urllib.parse.urlsplit(base)
            if parsed.scheme != "https" or not parsed.netloc or parsed.username or parsed.query or parsed.fragment:
                raise ValueError("Counter endpoint must be an HTTPS base URL")
            url = base + "/api/monthly?" + urllib.parse.urlencode({"counter": args["counter"]})
            request = urllib.request.Request(url, headers={"Accept": "application/json", "User-Agent": "oneclickvirt-mcp/" + VERSION})
            with urllib.request.urlopen(request, timeout=20) as response:
                content = response.read(LIMIT + 1)
                if len(content) > LIMIT:
                    raise ValueError("Counter response exceeds 1 MiB")
                return json.loads(content)
        raise ValueError("Unsupported operation")


def serve(adapter, input_stream=sys.stdin, output_stream=sys.stdout):
    for line in input_stream:
        if len(line) > LIMIT:
            error = {"jsonrpc": "2.0", "id": None, "error": {"code": -32600, "message": "Message exceeds 1 MiB"}}
            output_stream.write(json.dumps(error) + "\n")
            output_stream.flush()
            continue
        try:
            message = json.loads(line)
        except ValueError:
            output_stream.write(json.dumps({"jsonrpc": "2.0", "id": None, "error": {"code": -32700, "message": "Parse error"}}) + "\n")
            output_stream.flush()
            continue
        if not isinstance(message, dict) or message.get("jsonrpc") != "2.0" or not isinstance(message.get("method"), str):
            output_stream.write(json.dumps({"jsonrpc": "2.0", "id": message.get("id") if isinstance(message, dict) else None,
                                           "error": {"code": -32600, "message": "Invalid request"}}) + "\n")
            output_stream.flush()
            continue
        if "id" not in message:
            continue
        reply = {"jsonrpc": "2.0", "id": message["id"]}
        method = message["method"]
        params = message.get("params", {})
        try:
            if not isinstance(params, dict):
                raise ValueError("params must be an object")
            if method == "initialize":
                requested = params.get("protocolVersion")
                reply["result"] = {"protocolVersion": requested if requested in PROTOCOLS else PROTOCOLS[-1],
                                   "capabilities": {"tools": {}},
                                   "serverInfo": {"name": adapter.profile["skill_name"], "version": VERSION}}
            elif method == "ping":
                reply["result"] = {}
            elif method == "tools/list":
                reply["result"] = {"tools": adapter.tools()}
            elif method == "tools/call":
                try:
                    value = adapter.call(params.get("name"), params.get("arguments", {}))
                    reply["result"] = {"content": [{"type": "text", "text": json.dumps(value, ensure_ascii=False)}], "isError": False}
                except (ValueError, OSError, KeyError, TypeError, urllib.error.URLError) as exc:
                    reply["result"] = {"content": [{"type": "text", "text": str(exc)}], "isError": True}
            else:
                reply["error"] = {"code": -32601, "message": "Method not found"}
        except (ValueError, TypeError) as exc:
            reply["error"] = {"code": -32602, "message": str(exc)}
        output_stream.write(json.dumps(reply, ensure_ascii=True) + "\n")
        output_stream.flush()


def main():
    for stream in (sys.stdin, sys.stdout):
        if hasattr(stream, "reconfigure"):
            stream.reconfigure(encoding="utf-8")
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--profile", type=Path, default=Path(__file__).with_name("profile.json"))
    parser.add_argument("--version", action="version", version=VERSION)
    args = parser.parse_args()
    serve(Adapter(args.profile))


if __name__ == "__main__":
    main()
