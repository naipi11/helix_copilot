# helix-copilot

`helix-copilot` 为打了补丁的 Helix 编辑器提供 GitHub Copilot 支持，支持原生 ghost text（幽灵文本）/ 内联补全。

本仓库由 `naipi11` GitHub 账号维护。

## 上游 Helix

本项目基于上游 Helix 编辑器：

- 上游仓库：<https://github.com/helix-editor/helix>
- Helix 官网：<https://helix-editor.com/>
- Helix 的许可证和运行时文件保留自上游项目。

`helix/` 目录包含打了补丁的 Helix 源码，用于构建支持原生内联补全渲染和接受行为的 `hx` 编辑器。

## 包含内容

- `hx`：打了补丁的 Helix 编辑器二进制文件，支持原生 Copilot ghost text。
- `helix-copilot`：Go 编写的 CLI 工具和 LSP 桥接器。
- `helix-copilot lsp`：启动 Helix 和 `@github/copilot-language-server` 之间的代理。
- `helix-copilot login`：运行 GitHub Copilot 设备登录流程。
- `helix-copilot configure-helix`：安全地将 Copilot language-server 设置合并到 Helix 的 `languages.toml`。
- `helix-copilot model <name>`：存储选定的 Copilot 模型。

## 系统要求

- 拥有 GitHub Copilot 访问权限的 GitHub 账号。
- Node.js 和 npm 在 `PATH` 中可用。
- 如果从源码构建 `helix-copilot`，需要 Go 1.24+。
- 如果从源码构建打了补丁的 `hx`，需要 Rust 工具链。

### Windows 特别说明

**✅ Windows 支持已完全修复！** 从 v0.2.0 开始，Windows 上的自动补全和 ghost text 可以正常工作。

LSP 桥接器会：
- 自动检测已安装的 `@github/copilot-language-server`（在 `%APPDATA%\npm` 或全局 npm 目录）
- 首次运行时自动安装（如果未找到）
- 直接通过 `node.exe` 启动 language server，绕过 `npx.cmd` 的 `cmd.exe` 包装层
- 将诊断日志写入 `%APPDATA%\helix-copilot\proxy.log`（可通过 `HELIX_COPILOT_LOG=0` 禁用）

如果遇到问题，请查看日志文件：
```powershell
notepad "$env:APPDATA\helix-copilot\proxy.log"
```

## 从 Release 安装

发布版本后，从以下地址下载适合您平台的压缩包：

```text
https://github.com/naipi11/helix_copilot/releases
```

每个 release 压缩包包含：

```text
helix-copilot        # Windows 上是 helix-copilot.exe
hx                  # Windows 上是 hx.exe
runtime/            # Helix 运行时文件
```

### Linux / macOS 安装脚本

```bash
curl -fsSL https://raw.githubusercontent.com/naipi11/helix_copilot/main/scripts/install.sh | bash
```

可选环境变量：

```bash
VERSION=v0.1.0 BIN_DIR="$HOME/.local/bin" bash scripts/install.sh
```

### Windows PowerShell 安装脚本

```powershell
iwr https://raw.githubusercontent.com/naipi11/helix_copilot/main/scripts/install.ps1 -UseBasicParsing | iex
```

本地运行时的可选参数：

```powershell
./scripts/install.ps1 -Version v0.1.0 -BinDir "$HOME\bin"
```

## 从源码构建

### 构建 Go CLI

```bash
go build -o helix-copilot ./cmd/helix-copilot
```

安装到 Go bin 目录：

```bash
go install ./cmd/helix-copilot
```

### 构建打了补丁的 Helix

```bash
cd helix
cargo build --release --locked
```

打了补丁的编辑器二进制文件位于：

```text
helix/target/release/hx
```

Windows 上是 `hx.exe`。

确保 `hx` 和 `helix-copilot` 都在您的 `PATH` 中。

## GitHub Copilot 登录

运行：

```bash
helix-copilot login
```

该命令会启动 Copilot language server，请求设备登录码，并提示您在浏览器中完成授权。

## 配置 Helix

运行：

```bash
helix-copilot configure-helix
```

默认情况下会更新：

```text
~/.config/helix/languages.toml        # Linux/macOS
%APPDATA%\helix\languages.toml        # Windows
```

先测试合并输出：

```bash
helix-copilot configure-helix --output ./languages.test.toml
```

该命令会合并而不是盲目覆盖：

- 添加或更新 `[language-server.copilot]`。
- 将 `copilot` 添加到现有的 `language-servers` 数组中，不会重复。
- 保留其他语言设置，如 `auto-format`、`indent`、调试器模板和语法条目。
- 添加 Python 的 `pylsp + copilot` 设置，并禁用 `pylsp` 的样式诊断。

## 使用方法

使用打了补丁的 `hx` 二进制文件打开支持的源文件。在插入模式下，Copilot 建议会以 ghost text 形式出现（如果可用）。

关键行为：

- `Tab`：接受可见的 ghost text；当没有 ghost text 时回退到正常的智能 tab。
- `Esc`：拒绝 ghost text 并返回正常模式。
- `:model <name>`：在打了补丁的 Helix 构建中，调用 `helix-copilot model <name>` 来存储选定的模型。更改模型后需要重启 language server 或编辑器。

您也可以在 Helix 外部设置模型：

```bash
helix-copilot model gpt-5.4-mini
```

## 高级配置

### 手动指定 Language Server 路径

如果自动检测失败，可以在配置文件中手动指定：

**配置文件位置：**
- Linux/macOS: `~/.config/helix-copilot/config.json`
- Windows: `%APPDATA%\helix-copilot\config.json`

**示例配置：**
```json
{
  "model": "gpt-5.4-mini",
  "languageServerPackage": "@github/copilot-language-server",
  "languageServerJSPath": "C:\\path\\to\\copilot-language-server\\dist\\language-server.js"
}
```

### 环境变量

- `HELIX_COPILOT_LOG=0`：禁用诊断日志
- `HELIX_COPILOT_LS_PATH`：覆盖 language server 入口脚本路径

## 包管理器模板

本仓库包含启动打包文件：

- `.goreleaser.yaml`
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `packaging/scoop/helix-copilot.json`
- `packaging/homebrew/helix-copilot.rb`

Scoop 和 Homebrew 文件是模板。发布真实 release 资产后替换 `TODO` 校验和。

## 开发检查

运行项目自有包的 Go 测试：

```bash
go test ./cmd/helix-copilot ./internal/config ./internal/helixconfig ./internal/login ./internal/lsp ./internal/proxylog
```

运行打了补丁的 Helix 检查：

```bash
cd helix
cargo check -p helix-term
```

避免在仓库根目录运行 `go test ./...`，因为它会进入 vendored/上游 Helix tree-sitter 语法绑定，这些不是 Go CLI 模块的一部分。

## 故障排除

### Windows 上没有补全提示

1. 检查日志文件：`%APPDATA%\helix-copilot\proxy.log`
2. 确认 Copilot LS 已安装：`npm ls -g @github/copilot-language-server`
3. 手动安装：`npm install -g @github/copilot-language-server`
4. 检查 Helix 配置：`%APPDATA%\helix\languages.toml` 中是否有 `[language-server.copilot]`

### Ghost text 颜色太浅

这是 Helix 主题配置问题。编辑 `%APPDATA%\helix\config.toml`（Windows）或 `~/.config/helix/config.toml`（Linux/macOS）调整主题，或换一个对比度更高的主题。

### 补全语言不是中文

Copilot 的补全语言由代码上下文决定。在代码中多写中文注释，Copilot 会学习上下文并倾向于生成中文补全。

## 许可证

本仓库包含原创的 `helix-copilot` 代码以及打了补丁的 Helix 源码。详见相关源文件和上游 Helix 许可证。
