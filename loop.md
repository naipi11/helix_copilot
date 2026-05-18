# 开发日志 — 2026.5.19 夜班迭代

## 当前状态
- ghost text 基础管道已通：hx → proxy → Copilot LS → proxy → hx
- 触发逻辑：随 `trigger_auto_completion` 自动触发
- 接受：Tab（无 ghost text 时 fallback 到 smart_tab）
- 拒绝：Esc

## 已知问题
1. **前缀匹配含缩进** → ✅ 已修（用 trimmed_prefix）
2. **Python 文件不触发** → 可能原因：Copilot LS 初始化慢；auto-completion 触发条件不满足
3. **光标移动不清 ghost text** → 用户在 insert 模式下按方向键 ghost text 残留
4. **long-running ghost text 反馈慢** → 打字快到一定速度时请求滞后

## 下次启动任务
1. 确认 trimmed_prefix 修复后 Go 文件 ghost text 显示正确
2. 修复 Python 不触发：改用直接从 PostInsertChar 事件触发 inline completion 的独立路径，不依赖 completion handler 的 debounce
3. 光标移动时清除 ghost text
4. 多行 ghost text 支持
