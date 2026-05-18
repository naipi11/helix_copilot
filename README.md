# helix_copilot

Helix Copilot 是一个轻量级命令行桥接器，用来把 GitHub Copilot language server 接入 Helix 的 LSP 配置，以便在 Helix 中使用 Copilot 代码补全能力。

当前实现目标是最小可用原型：

- 使用 Go 编写主程序。
- 通过 `npx @github/copilot-language-server` 启动官方 Copilot language server。
- 提供 `login` 命令触发 GitHub Copilot 设备登录流程。
- 提供 `model` 命令保存当前模型选择。
- 提供 `configure-helix` 命令生成 Helix `languages.toml` 片段。
- 提供 `lsp` 命令作为 Helix language server 入口。

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

该命令会代理调用：

```bash
npx --yes @github/copilot-language-server login
```

按终端提示完成 GitHub Copilot 设备授权。

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

- 这是不修改 Helix 源码的桥接方案，核心依赖 Helix 对 LSP completion / inline completion 能力的支持。
- `:/model` 目前是 CLI 命令形式，不是 Helix 内部命令。
- Copilot language server 的登录和补全协议由官方 npm 包负责，本项目只做轻量封装和 Helix 配置入口。

## 开发计划

1. 完善多语言 `languages.toml` 合并逻辑，避免覆盖用户现有配置。
2. 增加安装脚本和包管理器发布配置。
3. 验证 Helix 对 Copilot inline completion 的实际兼容性。
4. 如果必须支持真正的 `:/model`，再评估是否维护 Helix patch 或上游插件接口。
