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

生成配置片段：

```bash
helix-copilot configure-helix --output ./languages.copilot.toml
```

将生成内容合并进 Helix 的 `languages.toml`。示例：

```toml
[language-server.copilot]
command = "helix-copilot"
args = ["lsp"]

[[language]]
name = "go"
language-servers = ["gopls", "copilot"]
auto-format = true
```

注意：Helix 的 `languages.toml` 需要按具体语言合并配置，不能盲目覆盖已有配置。

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

后续要实现真正的 Helix 内部 `:/model` 命令，需要修改 Helix 命令系统或等待 Helix 支持外部插件命令入口。

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
