#本文件用来规划项目具体实现细节
1. 本项目最终是要建立一个可通过包管理器在 Windows 和 Linux 进行安装的软件。
2. 直接修改 Helix 源码，在 LSP 层实现 textDocument/inlineCompletion 请求/响应支持。
3. 在 Helix UI 渲染层实现 ghost text（灰色幽灵字）显示。
4. 支持 Copilot 设备登录流程，通过 LSP signInInitiate 协议完成 Copilot 账号登录。
5. 打字时自动触发 ghost text 补全，按 Tab 接受、Esc 拒绝。
6. 支持 :/model 命令切换当前 Copilot 模型（默认 GPT-5.4 mini）。
7. 保持原版 Helix 的流畅度，避免卡顿。
8. 支持 Copilot 官方支持的所有语言（C/C++/Go/Python/Rust/JS/TS 等）。
9. Go 编写的 helix-copilot CLI 工具作为辅助（登录、模型管理、配置生成）。
