# Helix 主题

本目录包含为 helix-copilot 优化的 Helix 编辑器主题。

## copilot-green.toml

**黑底白字绿色 Ghost Text 主题**，专为 Copilot 内联补全优化。

### 特点

- **背景：** 纯黑色 `#000000`
- **普通文字：** 纯白色 `#FFFFFF`
- **Ghost Text（Copilot 补全）：** 亮绿色 `#00FF00` + 斜体
- **语法高亮：** 彩色（绿色字符串、黄色函数、蓝色类型、品红关键字等）

### 安装方法

#### 方法 1：手动复制（推荐）

```bash
# Linux/macOS
cp themes/copilot-green.toml ~/.config/helix/themes/

# Windows PowerShell
Copy-Item themes\copilot-green.toml "$env:APPDATA\helix\themes\"
```

#### 方法 2：符号链接

```bash
# Linux/macOS
mkdir -p ~/.config/helix/themes
ln -s "$(pwd)/themes/copilot-green.toml" ~/.config/helix/themes/

# Windows PowerShell (需要管理员权限)
New-Item -ItemType SymbolicLink -Path "$env:APPDATA\helix\themes\copilot-green.toml" -Target "$(Get-Location)\themes\copilot-green.toml"
```

### 启用主题

编辑 Helix 配置文件：

**Linux/macOS:** `~/.config/helix/config.toml`  
**Windows:** `%APPDATA%\helix\config.toml`

在文件开头添加：

```toml
theme = "copilot-green"

[editor]
# 你的其他配置...
```

保存后重启 Helix，或在 Helix 中运行 `:reload-config`。

### 自定义颜色

如果想调整 Ghost Text 颜色，编辑主题文件中的 `[palette]` 部分：

```toml
[palette]
ghost_text = "#00FF00"      # 改成其他颜色，比如 "#00FF88"（青绿）
```

常用颜色参考：
- `#00FF00` - 亮绿色（默认）
- `#00FF88` - 青绿色
- `#88FF00` - 黄绿色
- `#00FFFF` - 青色
- `#FFFF00` - 黄色

保存后运行 `:reload-config` 即可生效。

## 创建自己的主题

参考 [Helix 主题文档](https://docs.helix-editor.com/themes.html) 和 `copilot-green.toml` 的注释。

关键配置项：
- `ui.virtual.inlay-hint` - 控制 Ghost Text（内联提示）的颜色和样式
- `ui.background` - 背景色
- `ui.text` - 普通文字颜色
