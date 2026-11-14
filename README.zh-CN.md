<p align="center">
  <img src="docs/logo-ghost-wave-effect.svg" alt="Reasonix" width="360"/>
</p>

> [!NOTE]
> **本仓库是 Reasonix 的社区维护分支。** 上游（[esengine/DeepSeek-Reasonix](https://github.com/esengine/DeepSeek-Reasonix)）已于 2026-11-12 以 v2.61.0 停止维护，因此本分支继续为个人与社区维护该项目。目前仅提供 Windows 预编译版本；macOS 与 Linux 请从源码构建，详见下方[本仓库说明](#本仓库说明)。

<p align="center">
  <a href="./README.md">English</a>
  &nbsp;·&nbsp;
  <strong>简体中文</strong>
  &nbsp;·&nbsp;
  <a href="./docs/GUIDE.zh-CN.md">指南</a>
  &nbsp;·&nbsp;
  <a href="./docs/ACP.md">ACP</a>
  &nbsp;·&nbsp;
  <a href="./docs/EXTENSIONS.md">扩展开发</a>
  &nbsp;·&nbsp;
  <a href="./docs/SPEC.md">规格</a>
  &nbsp;·&nbsp;
  <a href="https://esengine.github.io/DeepSeek-Reasonix/">官方网站</a>
</p>

<p align="center">
  <a href="https://github.com/Hamza-ali1223/DeepSeek-Reasonix/releases"><img src="https://img.shields.io/github/v/release/Hamza-ali1223/DeepSeek-Reasonix.svg?filter=fork-v*&include_prereleases&style=flat-square&color=8250df&labelColor=161b22&label=fork%20release" alt="Reasonix fork release"/></a>
  <a href="https://www.npmjs.com/package/reasonix"><img src="https://img.shields.io/npm/v/reasonix.svg?style=flat-square&color=cb3837&labelColor=161b22&logo=npm&logoColor=white" alt="npm version"/></a>
  <a href="https://github.com/Hamza-ali1223/DeepSeek-Reasonix/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/Hamza-ali1223/DeepSeek-Reasonix/ci.yml?style=flat-square&label=ci&labelColor=161b22&logo=githubactions&logoColor=white" alt="CI"/></a>
  <a href="./LICENSE"><img src="https://img.shields.io/npm/l/reasonix.svg?style=flat-square&color=8b949e&labelColor=161b22" alt="license"/></a>
  <a href="https://www.npmjs.com/package/reasonix"><img src="https://img.shields.io/npm/dm/reasonix.svg?style=flat-square&color=3fb950&labelColor=161b22&label=downloads" alt="downloads"/></a>
  <a href="https://github.com/Hamza-ali1223/DeepSeek-Reasonix/stargazers"><img src="https://img.shields.io/github/stars/Hamza-ali1223/DeepSeek-Reasonix.svg?style=flat-square&color=dbab09&labelColor=161b22&logo=github&logoColor=white" alt="GitHub stars"/></a>
  <a href="https://atomgit.com/esengine/DeepSeek-Reasonix"><img src="https://atomgit.com/esengine/DeepSeek-Reasonix/star/badge.svg" alt="AtomGit stars"/></a>
  <a href="https://github.com/esengine/DeepSeek-Reasonix/graphs/contributors"><img src="https://img.shields.io/github/contributors/esengine/DeepSeek-Reasonix.svg?style=flat-square&color=bc8cff&labelColor=161b22&logo=github&logoColor=white" alt="contributors"/></a>
</p>

<p align="center">
  <a href="https://trendshift.io/repositories/27020?utm_source=trendshift-badge&amp;utm_medium=badge&amp;utm_campaign=badge-trendshift-27020" target="_blank" rel="noopener noreferrer"><img src="https://trendshift.io/api/badge/trendshift/repositories/27020/monthly?language=Go" alt="esengine/DeepSeek-Reasonix | Trendshift" width="250" height="55"/></a>
  <a href="https://trendshift.io/repositories/27020?utm_source=repository-badge&amp;utm_medium=badge&amp;utm_campaign=badge-repository-27020" target="_blank" rel="noopener noreferrer"><img src="https://trendshift.io/api/badge/repositories/27020" alt="esengine/DeepSeek-Reasonix | Trendshift" width="250" height="55"/></a>
</p>

<br/>

<p align="center"><strong>开源 · MIT · 程序跑在你电脑上，模型自选</strong></p>
<h3 align="center">终端、桌面、浏览器、编辑器里都能用的 AI 编程助手。</h3>
<p align="center">它会读你的项目、改文件、跑命令和测试。每一步先问你，或在你划定的范围内自己做，由你选；它用编辑工具改过的文件可以退回去。第一次用，从桌面应用 Studio 开始。</p>
<p align="center">维护者：<strong>Huahui Yu</strong>。</p>

## 它是什么

Reasonix 是一个开源的 AI 编程助手。程序跑在你的电脑上；AI 模型是你自己配置的服务，它读到的内容会发给这个服务。你用大白话描述任务，它在你选的项目文件夹里读代码、改文件、跑命令和测试，每一步都摆给你看。

模型你自己选：内置 DeepSeek 预设，其他兼容 OpenAI 接口的服务填一条配置就行，费用直接付给模型服务商。

## 本仓库说明

本仓库由 [@Hamza-ali1223](https://github.com/Hamza-ali1223) 维护，是 Reasonix 的延续，供个人使用，也供社区中仍需要该项目的人使用。上游作者已于 2026-11-12 在 v2.61.0 之后突然停止维护；上游 Issue 不再处理，也不会再有上游版本发布。

- **代码沿用上游原样。** 这里的改动是在其之上做修复与新增，而非重写，因此现有指南与文档仍适用于你运行的程序——第一项改动是在 Studio 中渲染 `svg` 围栏。
- **目前提供 Windows 版本。** 每次合并到 `studio` 都会发布一个 [`fork-v*` 版本](https://github.com/Hamza-ali1223/DeepSeek-Reasonix/releases)，包含 CLI（`windows/amd64` 与 `windows/arm64`）、Studio 安装包与便携版 zip，并附带 `SHA256SUMS`。
- **macOS 与 Linux 请从源码构建。** 这两个平台暂无预编译版本，步骤见[从源码构建](#从源码构建)；Linux 的 webview 依赖与构建标签见 [Studio 构建指南](./desktop/README.md#prerequisites)。
- **版本号独立。** 分支版本为 `fork-vX.Y.Z`，不会与上游的 `studio-v*` 冲突；较大的版本跳跃会在合并前的 Pull Request 中标注。
- **贡献者分开记录。** 致谢表中的名字组成并塑造了上游的 Reasonix；本分支产生的工作单独列在其下方。
- **欢迎提交 Issue 与 Pull Request，** 但不承诺响应时间。

Reasonix 采用 MIT 协议，本分支继续沿用该协议。

## 你能得到什么

- **读、改、跑**：读你的项目、改文件、跑测试和命令，都在你选的项目文件夹里。
- **权限分档**：每次都问你，或让日常操作自动放行；你设的禁止规则始终有效。想先看方案，就开计划模式，批准后它才动手。
- **可回退**：改文件前先存快照，按对话轮次回退，不碰 git 历史。只覆盖编辑工具改的文件，命令行命令造成的改动不在内。

## 30 秒上手

下载、填模型密钥、打开项目。第一次用，从桌面应用 Studio 开始。

1. 从[分支发布页](https://github.com/Hamza-ali1223/DeepSeek-Reasonix/releases)下载并安装桌面应用 Reasonix Studio。
2. 打开应用，按提示填入模型服务商和密钥。
3. 打开你的项目文件夹，写下想做的事，例如“给 http 客户端加上失败重试”。
4. 按你选的档位，该问你的操作它会先停下来问；同意就继续，不同意就拒绝。

更喜欢终端的话，下载 `reasonix` 命令行后运行 `reasonix setup` 配好模型，再运行 `reasonix` 即可。详见[安装](#安装)和[快速开始](#快速开始)。

## 四种用法

- **桌面应用（Studio）**：图形界面，适合第一次使用。
- **终端**：全屏文字界面，也能用 `reasonix run "任务"` 在脚本里运行。
- **浏览器**：`reasonix web` 在本机打开网页界面。
- **编辑器**：VS Code 安装扩展即可；其他支持 ACP（Agent Client Protocol，编辑器接入某个 AI 助手的通用约定）的编辑器，可连接 `reasonix acp`。

四种用法背后是同一套程序。下面的内容写给进阶用户和贡献者。

## 在 DeepSeek Harness 里使用

DeepSeek Harness（`dsh`）可以通过 ACP 把任务交给 Reasonix 作为子智能体去做。本仓库为此发布了一个插件索引：

1. 在运行 `dsh` 的机器上先运行 `reasonix setup`。在 DSH 网页界面打开“自定义插件市场”（插件，更多），再到 Discover 页添加这个索引：`https://raw.githubusercontent.com/esengine/DeepSeek-Reasonix/studio/integrations/dsh-plugin/market-index.json`
2. 从中安装“Reasonix subagent”。
3. 打开这个插件的页面，把它的开关打开。关着的时候模型没有 `reasonix` 工具，可能改在自己的 shell 里直接运行 `reasonix -p`，看起来像委派，其实不是。

配置细节、权限策略和限制见 [在 DeepSeek Harness 中使用 Reasonix](./docs/DSH.md)。

## 版本

Reasonix 分为两条版本线，调整的原因见[版本路线公告](https://github.com/esengine/DeepSeek-Reasonix/discussions/10748)。

| 版本线 | 分支 | 状态 | 获取方式 |
| --- | --- | --- | --- |
| **Reasonix 2.x** | `studio`（当前分支） | 活跃开发（本分支） | [分支发布页](https://github.com/Hamza-ali1223/DeepSeek-Reasonix/releases) |
| **Reasonix 1.x** | [`main-v2`](https://github.com/esengine/DeepSeek-Reasonix/tree/main-v2) | 维护 / 稳定 | `npm i -g reasonix` · [桌面端下载](https://reasonix.io/?download=desktop#start) |

- **想要稳定**：继续使用 1.x。它会持续收到 Bug 修复、Provider / API 兼容、
  更新器和安全修复，但不再扩展核心架构。
- **想体验新架构**：使用 2.x Reasonix Studio。它仍在快速迭代，欢迎反馈问题。

1.x 中有价值的修复、测试和行为会逐项 review，适合新架构的会迁移或重新实现到 2.x。

- [从 1.x 迁移到 2.x](./docs/MIGRATING.md)（英文）：两条线共用哪些数据，以及如何在同一台
  机器上同时使用而不丢会话。
- [2.x 路线图](./docs/ROADMAP.md)（英文）：2.x 还要交付什么、如何发布、哪些决定尚未做出。

## 特性

以下是写给进阶用户的技术要点。

- **配置驱动**：provider、agent、启用的工具、插件全部在 `reasonix.toml` 中声明，
  内核无硬编码模型。
- **多模型 · 可组合**：DeepSeek 作为预设内置；任何 OpenAI 兼容
  端点都只是一条配置。可选让两个模型协同（执行器 + 规划器），各自独立、缓存稳定的 session。
- **插件驱动**：MCP server 提供工具、提示词和资源；Extension Protocol v1
  Sidecar 还可以拦截运行时事件、提供 Provider 与结构化 UI，并通过版本化插件包分发。
- **缓存友好的上下文维护**：启动时注入稳定的环境摘要；旧工具输出会先 snip/prune，
  再进入摘要 compaction；内置工具 schema 合约有文档和回归测试保护。
- **零摩擦分发**：`CGO_ENABLED=0` 单二进制；一条命令交叉编译到六个目标平台。
  产物是完全自包含的静态二进制——目标机器上除二进制本身外无需安装任何东西。

## 安装

当前分支是 **Reasonix 2.x**，桌面端是 Reasonix Studio；CLI、Studio 和编辑器集成
都运行同一套本地 Reasonix 引擎。

### Reasonix Studio（2.x）

从最新的 [分支发布](https://github.com/Hamza-ali1223/DeepSeek-Reasonix/releases) 下载对应平台的安装包（标签为 `fork-v*`）：

> [!NOTE]
> **本分支只提供 Windows 安装包。** macOS 与 Linux 请[从源码构建](#从源码构建)；表格列出同一份代码可产出的平台。

| 平台 | 安装包 | 架构 |
| --- | --- | --- |
| macOS | 从源码构建 | Apple Silicon（`arm64`）/ Intel（`amd64`） |
| Windows | 安装器 `-installer.exe` 或便携 `.zip` | x64 |
| Linux | 从源码构建 | x64（`amd64`）/ ARM64（`arm64`，aarch64） |

分支发布包未签名，附有 `SHA256SUMS`。应用内更新仍指向上游的更新源，因此不会提供分支版本：
请手动下载下一个 `fork-v*` 版本安装。

分支发布同时提供 2.x 的 `reasonix` CLI，仅限 `windows × amd64|arm64`。

### Reasonix 1.x（稳定版）

1.x CLI 在任意支持的平台上都可以通过 npm 安装，macOS 也可以使用 Homebrew：

```sh
npm i -g reasonix                  # 任意系统;自动拉取对应平台的原生二进制
brew install esengine/reasonix/reasonix   # macOS
```

1.x 桌面端请前往[官方下载页](https://reasonix.io/?download=desktop#start)。
Windows 安装器使用 Certum 代码签名证书签名。

### VS Code 扩展

扩展不内置 CLI，而是启动本机的 `reasonix acp` 后端，并提供原生聊天、编辑器
上下文、工具调用审批、模型选择和工作区会话。请先安装 1.x CLI。

- **VS Code：** [从 Visual Studio Marketplace 安装](https://marketplace.visualstudio.com/items?itemName=SivanLiu.reasonix-agent)
- **VSCodium / Eclipse Theia：** [从 Open VSX Registry 安装](https://open-vsx.org/extension/SivanLiu/reasonix-agent)
- **扩展 ID：** `SivanLiu.reasonix-agent` · [源码与使用说明](https://github.com/SivanCola/reasonix-vscode)

### 从源码构建

克隆仓库；`studio` 构建 2.x，`main-v2` 构建 1.x：

```sh
git clone https://github.com/Hamza-ali1223/DeepSeek-Reasonix.git
cd DeepSeek-Reasonix
git switch studio
```

#### CLI

CLI 构建需要 **Go 1.25+**。模块固定了 `toolchain`，保持 `GOTOOLCHAIN=auto`
即可让 Go 自动下载对应工具链，也可以手动安装。

```sh
make build      # -> bin/reasonix(.exe)
make cross      # -> dist/（darwin|linux|windows × amd64|arm64）
```

#### Studio

Studio 的前端还需要 **Node 24+ 与 pnpm 10**（`npm install -g pnpm@10`）。

```sh
make studio
```

各平台的 webview 依赖和 Linux 构建标签见 [Studio 构建指南](desktop/README.md#prerequisites)。

## 快速开始

### Reasonix Studio

安装并启动 Studio，然后在应用内连接 provider 和模型即可使用，无需配置 CLI。

### CLI / TUI

```sh
reasonix setup                      # 配置 provider 和模型
reasonix                            # 启动交互式会话
reasonix run "把 main.go 里的 TODO 实现掉"
```

需要项目指令时，可在交互式会话中运行 `/init`。

CLI 进阶用法和详细配置见 **[CLI 命令参考](./docs/CLI.zh-CN.md)**、
**[指南](./docs/GUIDE.zh-CN.md)** 和
**[配置路径](./docs/CONFIG_PATHS.md)**。

## 文档

- **开始使用：** [指南](./docs/GUIDE.zh-CN.md) ·
  [CLI 命令参考](./docs/CLI.zh-CN.md) · [配置路径](./docs/CONFIG_PATHS.md) ·
  [ACP 编辑器接入](./docs/ACP.md) · [在 DeepSeek Harness 中使用](./docs/DSH.md)
- **功能与排障：** [子智能体 Profile](./docs/SUBAGENT_PROFILES.md) ·
  [Context Engine v2](./docs/SESSION_MEMORY_RETRIEVAL.md) ·
  [能力诊断](./docs/CAPABILITY_DIAGNOSTICS.md) ·
  [恢复与安全模式](./docs/RECOVERY.md) ·
  [Checkpoints 与 rewind](./docs/CHECKPOINTS.md)
- **Agent 预设：** [会话级组合](./docs/AGENT_COMPOSITIONS.zh.md)
- **工程与迁移：** [规格](./docs/SPEC.md) ·
  [任务合约与暂停策略](./docs/TASK_CONTRACT.md) ·
  [工具合约](./docs/TOOL_CONTRACT.md) ·
  [从 1.x 迁移到 2.x](./docs/MIGRATING.md) ·
  [2.x 路线图](./docs/ROADMAP.md)
- **扩展开发：** [扩展概览](./docs/EXTENSIONS.md) ·
  [插件包与 Manifest v1](./docs/PLUGIN_PACKAGES.md) ·
  [Extension Protocol](./docs/EXTENSION_PROTOCOL.md) ·
  [Go SDK 与 starter](./sdk/go/README.md)

## Star 趋势

<a href="https://www.star-history.com/?repos=esengine%2FDeepSeek-Reasonix&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/esengine/DeepSeek-Reasonix/star-history/assets/star-history/star-history-dark.svg" />
   <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/esengine/DeepSeek-Reasonix/star-history/assets/star-history/star-history-light.svg" />
   <img alt="Star History Chart" src="https://raw.githubusercontent.com/esengine/DeepSeek-Reasonix/star-history/assets/star-history/star-history-light.svg" />
 </picture>
</a>

<br/>

## 致谢

下面这些朋友的工作塑造了 Reasonix 今天的样子 —— 按 commit 数统计的前 20 名贡献者，统计截至上游最后一个版本。
**他们是上游的原始贡献者：**在上游停止维护之前共同构建了 Reasonix。本节末尾的贡献者头像墙覆盖整个仓库——
同一个人群，再加上此后在本仓库落地过工作的人。
完整贡献者列表在
[GitHub](https://github.com/esengine/DeepSeek-Reasonix/graphs/contributors?all=1)。

<!-- reasonix-top-contributors:start -->
| Contributor | Contributor | Contributor | Contributor |
| --- | --- | --- | --- |
| [**SivanCola**](https://github.com/SivanCola) | [**KHG420**](https://github.com/KHG420) | [**ttmouse**](https://github.com/ttmouse) | [**chuiwenwei**](https://github.com/chuiwenwei) |
| [**lifu963**](https://github.com/lifu963) | [**XTLine**](https://github.com/XTLine) | [**HUQIANTAO**](https://github.com/HUQIANTAO) | [**xxoingr**](https://github.com/xxoingr) |
| [**Harbor404**](https://github.com/Harbor404) | [**GTC2080**](https://github.com/GTC2080) | [**mchenziyi**](https://github.com/mchenziyi) | [**Li-Charles-One**](https://github.com/Li-Charles-One) |
| [**light-front-theory**](https://github.com/light-front-theory) | [**imcreazy**](https://github.com/imcreazy) | [**eghrhegpe**](https://github.com/eghrhegpe) | [**noreply**](https://github.com/noreply) |
| [**par73e**](https://github.com/par73e) | [**Bernardxu123**](https://github.com/Bernardxu123) | [**lancy69**](https://github.com/lancy69) | [**Guan0923**](https://github.com/Guan0923) |
<!-- reasonix-top-contributors:end -->

特别感谢 [**Bernardxu123**](https://github.com/Bernardxu123) 设计的项目 logo和开场视频。

<p align="center">
  <a href="https://github.com/Hamza-ali1223/DeepSeek-Reasonix/graphs/contributors">
    <img src="https://contrib.rocks/image?repo=Hamza-ali1223/DeepSeek-Reasonix&max=100&columns=12" alt="Hamza-ali1223/DeepSeek-Reasonix 贡献者，含上游历史" width="860"/>
  </a>
</p>

<br/>

---

<p align="center">
  <sub>MIT —— 见 <a href="./LICENSE">LICENSE</a></sub>
  <br/>
  <sub>由 <a href="https://github.com/esengine/DeepSeek-Reasonix/graphs/contributors">esengine/DeepSeek-Reasonix</a> 社区共建</sub>
</p>
