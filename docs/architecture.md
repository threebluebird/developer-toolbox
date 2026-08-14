# Architecture

Developer Toolbox 使用单向依赖链：

```text
Frontend → Wails App → Services → Registry → Tools
                           ↓
                      Repositories → Storage
```

## Backend

- `models` 定义前后端共用的数据契约及结构化 `ToolError`。
- `registry` 是工具注册与发现的唯一入口。
- `services` 编排工具执行、搜索、历史、收藏与配置。
- `storage` 通过受限 key 的文件存储实现 JSON 持久化。
- `tools` 中每个目录是独立工具模块，不依赖 UI 或 Storage。
- 根目录 `App` 负责组装服务并暴露 Wails API，不承载工具算法。

所有工具实现统一的 `ToolExecutor` 接口：

```go
type ToolExecutor interface {
    Info() models.Tool
    Execute(input models.ToolInput) models.ToolOutput
}
```

失败输出使用 `{code, message, detail}` 结构，前端无需分析错误字符串。

## Frontend

- `src/services/appService.js` 是 Wails API 的统一适配层。
- `src/utils/output.js` 统一处理结果和错误显示。
- `src/main.js` 管理 V0.1 页面状态和通用 Tool Page。
- 所有工具共享输入、选项、输出、复制和下载布局。

后续若 UI 继续增长，可按 Home、Tool、Settings 将 `main.js` 拆为独立 view 模块，不改变后端契约。
