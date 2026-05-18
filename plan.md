#本文件为项目开发计划书

#项目目标
在 Helix 编辑器原生实现 GitHub Copilot ghost text（幽灵字/行内补全），达到 VS Code 级别的体验。
不再使用外挂桥接方案，而是直接修改 Helix 源码，在 LSP 层和 UI 渲染层原生支持 textDocument/inlineCompletion。

#技术栈
- Rust（修改 Helix 主程序）
- Go（继续维护 helix-copilot CLI 工具：登录、模型切换、配置生成）
- 直接修改 helix 源码，patch 官方仓库

#验收标准
1. 在 Helix 中登录 Copilot 账号后，打字时自动出现灰色幽灵字建议（ghost text）
2. 按 Tab 接受建议，按 Esc 拒绝
3. 支持通过 :/model 切换当前模型
4. 支持 Copilot 官方支持的所有语言
5. 流畅度接近原版 Helix，不卡顿
6. 可通过包管理器在 Windows/Linux 安装

#项目资源
**helix原项目地址**
https://github.com/helix-editor/helix.git
**helix软件官网**
https://helix-editor.cn/
**copilot官方文档**
https://docs.github.com/zh/copilot
**本项目仓库地址（初始为空）**
https://github.com/naipi11/helix_copilot.git
#本token期限为30天，仅限于访问repo资源
做好仓库管理，项目完成后编写一份完整清晰的README
