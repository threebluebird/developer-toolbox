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

文档单双页拆分工具调用示例：

```json
{
  "toolId": "document-split",
  "payload": {
    "filePath": "D:\\docs\\manual.docx",
    "outputDir": "D:\\docs\\print"
  }
}
```

PDF 会直接抽取页面；Word 和 PowerPoint 会先通过 Microsoft Office（Windows）或 LibreOffice 转换为 PDF，再生成 `_odd.pdf` 与 `_even.pdf`。源文件不会被修改，同名输出会自动追加序号。单页文件只生成奇数页 PDF。

新增工具只需实现 `registry.ToolExecutor`，在 `CreateDefaultRegistry` 中注册，并为前端补充选项控件。
