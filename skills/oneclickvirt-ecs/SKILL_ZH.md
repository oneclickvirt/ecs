---
name: oneclickvirt-ecs
description: "执行有明确范围的 Go 融合怪检测，解读结果并选择后续测试。"
metadata:
  repository: oneclickvirt/ecs
  integration-version: "1.0.0"
---

# ecs

## 使用范围

执行有明确范围的 Go 融合怪检测，解读结果并选择后续测试。

- 使用当前维护的 Go ECS 程序，区分已停止新功能开发的 Shell 版和 GUI 版。
- system 预设关闭性能基准、流媒体、IP 质量、测速、日志和公开上传。ECS 仍会执行程序自身的基础网络预检和 NAT/STUN 检查，整体限时 25 秒；保留 partial、unavailable 和 timeout 结果，扩展测试前明确范围。
- 同一次安装只解析一次正式版本，并校验标签对应的 SHA-256。即使 output 标签不变，缓存也必须按摘要校验。

## 安装和更新

```bash
curl -fL --retry 3 https://raw.githubusercontent.com/oneclickvirt/ecs/master/scripts/install_agent.py -o install_agent.py
python3 install_agent.py --client codex
```

可将下载地址改为 `https://cnb.cool/oneclickvirt/ecs/-/git/raw/master/scripts/install_agent.py`，并使用 `--source cnb` 从 CNB 的固定提交安装；此模式另需 Git。

用 `--client claude` 或 `--client openclaw` 选择其他客户端。重新下载并执行可更新；`--ref` 固定源码标签或提交。安装前校验全部 SHA-256，保留额外本地文件及已有客户端设置，`install.json` 记录来源提交。

## MCP

运行需要 Python 3.9 以上版本；MCP 使用本机已有程序及其权限，不自动安装虚拟化环境。将安装目录内对应客户端的 `mcp.codex.toml`、`mcp.claude.json` 或 `mcp.openclaw.json` 合并到已有配置。远端宿主机应通过无交互 SSH 运行该主机上的适配器。

```bash
python3 install_agent.py --client codex --install-binary
```

`--version` 可指定检测程序的正式 Release 标签，`--ref` 则指定 Skill/MCP 源码版本，两者独立。默认从 GitHub 解析一次最新正式版，选择本机系统/架构的真实产物并校验 SHA-256；没有匹配产物时停止。已有程序可用 `--binary /完整路径/程序` 复用。

先调用 `integration_info` 查看能力和运行要求。

可用检测预设：

- `system`: `goecs -menu=false -upload=false -log=false -basic=true -cpu=false -memory=false -disk=false -ut=false -security=false -email=false -backtrace=false -nt3=false -speed=false -ping=false -tgdc=false -web=false -json - -dns-mode system -timeout 25s -privacy=true`

`release_assets` 返回实际版本、文件名、大小、摘要及下载地址；需要固定版本时传完整标签。更新产物后应重新安装并验证帮助和实际调用，不能只因文件存在便判定可用。

## 对应源码与说明

- [`go.mod`](https://github.com/oneclickvirt/ecs/blob/master/go.mod)
- [`goecs.sh`](https://github.com/oneclickvirt/ecs/blob/master/goecs.sh)
- [`scripts/embed_geekbench.sh`](https://github.com/oneclickvirt/ecs/blob/master/scripts/embed_geekbench.sh)

- [集成安装说明 / Integration guide](https://www.spiritlhl.net/guide/agent_integrations.html)

给出操作命令前核对这些源码文件及当前文档，保留现有版本固定、数据路径和宿主机标识。报告测试时区分真实程序执行、隔离测试及未验证的环境。
