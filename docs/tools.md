# Tool API

工具调用统一使用：

```json
{
  "toolId": "json",
  "payload": {
    "input": "{\"ok\":true}",
    "action": "format"
  }
}
```

成功输出：

```json
{"success": true, "data": "..."}
```

失败输出：

```json
{
  "success": false,
  "error": {
    "code": "INVALID_JSON",
    "message": "unexpected end of JSON input"
  }
}
```

错误代码包括 `INVALID_JSON`、`INVALID_BASE64`、`INVALID_URL`、`INVALID_HASH`、`INVALID_UUID`、`INVALID_TIMESTAMP`、`INVALID_JWT`、`INVALID_REGEX`、`INVALID_YAML` 和 `TOOL_NOT_FOUND`。

新增工具只需实现 `registry.ToolExecutor`，在 `CreateDefaultRegistry` 中注册，并为前端补充选项控件。
