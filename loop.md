# 开发日志

## 当前阶段
直接修改 Helix 源码，实现原生 ghost text（行内补全）。

## 已完成
1. ✅ fork Helix 源码仓库到本仓库
2. ✅ 在 helix-lsp-types 启用 proposed 特性（支持 InlineCompletion 相关类型）
3. ✅ 在 Document 结构体添加 inline_completion 字段 + InlineCompletion 类型
4. ✅ 在 LSP client 添加 inline_completion() 方法（发送 textDocument/inlineCompletion）
5. ✅ 在 completion handler 添加 request_inline_completion_from_servers() 触发逻辑
6. ✅ 在 editor.rs 渲染层实现 ghost text（灰色 DIM 幽灵字）
7. ✅ 添加 ghost_text_accept 命令 + C-y 快捷键接受
8. ✅ Esc 退出插入模式时清除 ghost text
9. ✅ 默认模型改为 gpt-5.4-mini
10. ✅ 编译通过，二进制安装完毕
11. ✅ Go 代理修复：
    - 注入 inlineCompletionProvider 到 capabilities
    - 注入 editorInfo + editorPluginInfo 到 init options
    - 修复竞态条件：forwardChildToHelix 统一处理所有响应
    - 测试验证通过：Copilot 返回 `println("Hello, World!")`

## 待办
- [ ] 端到端测试：启动 hx，打开文件，确认 Copilot request 被发送
- [ ] ghost text 多行支持（当前只显示第一行）
- [ ] Tab 键接受 ghost text（当前用 C-y）
- [ ] ghost text 自动重新请求（当用户继续打字时）
- [ ] 光标移动时自动清除 ghost text
- [ ] 错误处理：如果 copilot language server 断开连接
