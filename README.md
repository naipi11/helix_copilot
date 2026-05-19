# helix_copilot

Helix Copilot 是一个轻量级命令行桥接器，用来把 GitHub Copilot language server 接入 Helix 的 LSP 配置，以便在 Helix 中使用 Copilot 代码补全能力。

## 架构

```
Helix (hx)  ── stdin/stdout ──►  helix-copilot lsp (LSP 代理)
                                         │
                                   转译 completion ↔ inlineCompletion
                                         │
                                         ▼
                              npx @github/copilot-language-server (--stdio)
```

Helix 发送 `textDocument/completion` 请求，桥接层自动转译为 Copilot 的 `textDocument/inlineCompletion`，并将结果转回标准 LSP CompletionItem 格式返回。

当前实现功能：

- 使用 Go 编写的 LSP 代理桥接程序。
- 通过 `npx @github/copilot-language-server` 启动官方 Copilot language server。
- 提供 `login` 命令触发 GitHub Copilot 设备登录流程。
- 提供 `model` 命令保存当前模型选择。
- 提供 `configure-helix` 命令生成 Helix `languages.toml` 片段。
- 提供 `lsp` 命令作为 Helix language server 入口。
- 支持 Copilot 官方支持的所有编程语言。


## 构建

本机需要 Go 1.24 或更新版本。

```bash
go build ./cmd/helix-copilot
```

生成二进制：

```bash
./helix-copilot --help
```

## 安装

```bash
go install github.com/naipi11/helix_copilot/cmd/helix-copilot@latest
```

当前仓库本地开发时：

```bash
go install ./cmd/helix-copilot
```

同时需要 Node.js / npm，因为当前版本通过 `npx` 调用官方 Copilot language server：

```bash
npm view @github/copilot-language-server version
```

## 登录 Copilot

```bash
helix-copilot login
```

该命令会通过 LSP 调用 Copilot language server 的设备登录流程：

1. 启动 `npx --yes @github/copilot-language-server --stdio`
2. 发送 `signInInitiate` 请求
3. 显示 GitHub 设备登录地址和设备码
4. 网页授权后按 Enter，程序发送 `workspace/executeCommand` 执行 `github.copilot.finishDeviceFlow`

## 配置 Helix

`configure-helix` 会**合并**到目标 `languages.toml`，不会再直接覆盖用户已有配置：

```bash
helix-copilot configure-helix
```

也可以输出到指定文件用于检查：

```bash
helix-copilot configure-helix --output ./languages.copilot.toml
```

合并规则：

- 添加或更新 `[language-server.copilot]`
- 为已有 `[[language]]` 的 `language-servers` 追加 `"copilot"`，并去重
- 保留已有语言条目的其他字段，例如 `auto-format`、`indent`、`debugger` 等
- 目标文件不存在时创建一个可用的最小配置
- Python 默认使用 `pylsp + copilot`，并关闭 `pylsp` 的风格诊断，避免 `W292` 这类警告干扰 ghost text 测试

示例：

```toml
[language-server.copilot]
command = "helix-copilot"
args = ["lsp"]

[[language]]
name = "go"
language-servers = ["gopls", "copilot"]
```

注意：TOML 会被规范化写回；请先用 `--output` 在副本上试跑，确认合并结果。

## 切换模型

项目计划要求在 Helix 中通过 `:/model` 切换模型。当前最小原型先提供外部命令：

```bash
helix-copilot model gpt-4o-copilot
```

配置保存到：

```text
$XDG_CONFIG_HOME/helix-copilot/config.json
```

如果未设置 `XDG_CONFIG_HOME`，则使用：

```text
~/.config/helix-copilot/config.json
```

后续如需真正的 Helix 内部 `:/model` 命令，有两条路线：

1. **维护 Helix patch**：在 `helix-term/src/commands/typed.rs` 注册 `model` 命令，并通过外部 CLI 或配置文件更新当前 Copilot 模型。优点是体验接近原生；缺点是需要长期跟随 Helix 上游命令系统变化。
2. **等待/利用上游插件接口**：保持 `helix-copilot model <name>` 作为稳定外部命令，等 Helix 提供更正式的外部命令或插件入口后再接入。优点是维护成本低；缺点是当前不能在 `:` 命令行内完成。

当前建议：先保持外部 CLI 模式，除非验收必须要求内部 `:/model`，否则不要扩大 Helix patch 面。

## 安装脚本与发布

仓库提供最小安装脚本和发布模板：

```bash
# Linux/macOS，从 GitHub Release 下载
curl -fsSL https://raw.githubusercontent.com/naipi11/helix_copilot/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/naipi11/helix_copilot/main/scripts/install.ps1 | iex
```

发布配置：

- `.goreleaser.yaml`：构建 `helix-copilot` 的 Linux / Windows / macOS amd64/arm64 归档
- `.github/workflows/ci.yml`：Go 测试
- `.github/workflows/release.yml`：tag `v*.*.*` 触发 GoReleaser draft release
- `packaging/scoop/helix-copilot.json`：Scoop manifest 模板
- `packaging/homebrew/helix-copilot.rb`：Homebrew formula 模板

注意：当前 GoReleaser 先发布 Go CLI；patched `hx` 多平台构建仍需后续接入 Helix Rust release matrix。

## 运行测试

```bash
go test ./...
```

## 当前限制

- Copilot language server 使用 `textDocument/inlineCompletion`（LSP 3.18 提议特性），而 Helix 使用 `textDocument/completion`。本项目通过内置 LSP 代理进行双向协议转译，可能存在部分兼容性问题。
- `:/model` 目前是 CLI 命令形式，不是 Helix 内部命令。
- Copilot language server 的登录和补全协议由官方 npm 包负责，本项目只做轻量封装和协议转译。

## 开发计划

1. 完善 `languages.toml` 合并逻辑，避免覆盖用户现有配置。
2. 增加安装脚本和包管理器发布配置（Windows/Linux）。
3. 优化 LSP 代理的兼容性和错误处理。
4. 如果必须支持真正的 `:/model`，再评估是否维护 Helix patch 或上游插件接口。
