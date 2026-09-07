# Developer Toolbox V0.5

Developer Toolbox 是一个面向程序员的本地桌面工具箱，基于 Go、Wails v2 和原生 JavaScript 构建。所有工具逻辑均在本地执行，不需要将输入内容发送到远程服务。

## 项目状态

当前版本：`0.5.0`

- 已完成统一 Tool 接口、Registry、Service 和 Storage 分层
- 已实现 34 个本地开发工具
- 支持 HIIT 间歇计时、文档单双页拆分、网络调试、串口通信、数据转换、离线数据库语法文档
- 支持命令面板、快捷键、History 2.0、Favorite 2.0 和本地持久化
- 支持深色、浅色、跟随系统主题
- 支持中文、英文界面即时切换
- 已覆盖所有工具包及核心服务、存储模块的单元测试

## 功能列表

| 工具 | 功能 |
| --- | --- |
| JSON Formatter | 格式化、压缩、校验、Tree 数据展示、2/4 空格缩进 |
| Base64 | UTF-8 文本编码与解码、非法输入提示 |
| URL Encoder | URL 编解码、Query String 规范化与解析 |
| Hash | MD5、SHA1、SHA256、SHA512，支持文本和文件流式计算 |
| UUID | UUID v4 单个或批量生成，数量限制为 1–100 |
| Timestamp | Unix 秒/毫秒、日期、UTC、本地时间和 ISO 8601 转换 |
| JWT Decoder | Header、Payload、Signature、`exp`、`iat` 和过期状态解析 |
| Regex Tester | 全局匹配、捕获组、匹配计数、`g/i/m/s` flags 和替换 |
| JSON ↔ YAML | JSON/YAML 双向转换及 YAML 校验 |
| XML / CSV / TOML | 格式化、校验及与 JSON 双向转换 |
| Text / SQL / Markdown / HTML | 文本处理、SQL 美化、本地预览与安全沙箱 |
| HTTP / DNS / Ping / Port / CIDR / URL Parser | 基础网络请求、诊断和地址计算 |
| Cron / Color / Code Generator / Git Tools | Cron、颜色、类型代码和 Git 辅助文本生成 |
| 费用金额中文转换 | 金额转人民币中文大写，或转普通中文数字 |
| TCP / UDP | 文本或 HEX 发送、响应编码、超时和响应大小限制 |
| 串口调试 | 枚举串口，配置波特率并收发文本或 HEX 数据 |
| 文字转拼音 | 离线转换全拼和拼音首字母，支持自定义分隔符 |
| 数据库语法文档 | PostgreSQL、MySQL、Oracle、SQL Server 常用语法及字符串、数组/集合、JSON、日期时间函数离线速查 |
| 文档单双页拆分 | 将 PDF、Word、PowerPoint 按奇数页和偶数页分别输出为 PDF，方便正反面打印；源文件保持不变 |
| HIIT 间歇计时器 | 自定义运动、休息时长与组数，支持快捷方案、暂停/继续、提示音和阶段呼吸灯 |

应用公共能力：

- 根据工具名称、ID、描述和分类搜索
- 收藏工具和展示最近使用记录
- 统一复制、下载、错误和空状态展示
- 深色、浅色、跟随系统主题
- 中文、英文界面
- JSON 缩进设置
- 自定义收藏和历史数据目录

## 界面预览

应用采用固定侧栏与响应式内容区：侧栏提供搜索、收藏、最近使用和设置入口，主区域统一承载工具卡片与双栏编辑器。窗口宽度低于 960px 时，工具卡片、设置页和编辑器自动切换为单列布局。

## 技术栈

- Go 1.25+
- Wails v2.14
- JavaScript、HTML、CSS
- Vite 7
- Playwright（端到端测试脚本）
- `gopkg.in/yaml.v3`
- `github.com/google/uuid`
- `github.com/mozillazg/go-pinyin`（离线拼音字典）
- `go.bug.st/serial`（跨平台串口访问）
- `github.com/pdfcpu/pdfcpu`（PDF 页数读取与单双页抽取）

## 目录结构

```text
developer-toolbox/
├─ backend/
│  ├─ config/             # 常量、默认设置及设置归一化
│  ├─ models/             # 前后端共享数据契约
│  ├─ registry/           # Tool 注册、查询和稳定排序
│  ├─ services/           # 工具、搜索、历史、收藏和设置编排
│  ├─ storage/            # 文件存储及 Repository
│  └─ tools/              # 34 个独立工具模块
├─ frontend/
│  ├─ src/
│  │  ├─ services/        # Wails API 适配层
│  │  ├─ stores/          # 前端应用状态
│  │  ├─ utils/           # 剪贴板、输出、搜索辅助函数
│  │  ├─ i18n.js          # 中英文文案及工具元信息本地化
│  │  ├─ main.js          # 页面渲染、交互及工具参数组装
│  │  └─ app.css          # 响应式布局和语义主题变量
│  ├─ wailsjs/            # Wails 自动生成/同步的前端绑定
│  └─ e2e/                # GUI 端到端测试脚本
├─ docs/                  # 架构和 Tool API 文档
├─ scripts/check.ps1      # 本地质量检查脚本
├─ app.go                 # Wails 对外 API 与依赖组合根
├─ main.go                # 桌面应用入口
├─ go.mod
└─ wails.json
```

## 架构

标准工具调用链：

```text
用户操作
  → Frontend View
  → frontend/src/services/appService.js
  → Wails Binding
  → App.ExecuteTool
  → ToolService
  → Registry
  → 具体 Tool
  → ToolOutput
```

持久化调用链：

```text
App
  → History/Favorite/Config Service
  → Repository
  → Storage
  → JSON 文件
```

核心约束：

- Tool 不依赖前端、App 或 Storage
- 前端不直接调用具体 Tool，只调用统一 Wails API
- Registry 是工具发现与注册的唯一入口
- Service 负责业务编排，Repository 负责数据序列化
- 所有工具错误均返回结构化 `ToolError`

详细说明见 [架构文档](docs/architecture.md) 和 [Tool API 文档](docs/tools.md)。

## 数据契约

统一输入示例：

```json
{
  "toolId": "json",
  "payload": {
    "input": "{\"enabled\":true}",
    "action": "format",
    "indent": 2
  }
}
```

成功输出：

```json
{
  "success": true,
  "data": "{\n  \"enabled\": true\n}"
}
```

失败输出：

```json
{
  "success": false,
  "error": {
    "code": "INVALID_JSON",
    "message": "unexpected end of JSON input",
    "detail": "可选的详细信息"
  }
}
```

## 开发环境

### 前置条件

1. 安装 Go 1.25 或更高版本。
2. 安装 Node.js 20 或更高版本。
3. 安装 Wails v2 CLI：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

可使用下面的命令检查 Wails 环境：

```powershell
wails doctor
```

### 安装前端依赖

```powershell
cd frontend
npm ci
cd ..
```

### 启动开发模式

```powershell
wails dev
```

开发模式会启动 Vite 热更新服务器，并将前端调用连接到 Go 后端。

如果只需要查看前端布局，可以运行：

```powershell
cd frontend
npm run dev
```

纯浏览器预览没有 Wails Runtime，工具执行、设置持久化和原生文件选择会使用回退数据或不可用；完整功能应通过 `wails dev` 验证。

## 测试与质量检查

运行 Go 单元测试：

```powershell
go test ./...
```

运行静态检查：

```powershell
go vet ./...
```

运行前端生产构建：

```powershell
cd frontend
npm run build
```

Windows 下可运行项目提供的全量检查脚本：

```powershell
./scripts/check.ps1
```

该脚本依次检查：

1. Go 文件是否已通过 `gofmt`
2. `go test ./...`
3. `go vet ./...`
4. Service、Storage、Worker 竞态检查
5. JSON/Base64 100 KB 性能基准
6. 前端编辑器测试与 JavaScript 语法检查
7. Vite 生产构建

## 构建桌面应用

构建当前平台可执行文件：

```powershell
wails build
```

Windows 构建结果默认位于：

```text
build/bin/developer-toolbox.exe
```

也可以执行带输出校验的打包脚本：

```powershell
./scripts/package-windows.ps1
```

需要安装包时，可根据 Wails 的 NSIS 配置使用 `build/windows/installer/` 下的脚本生成。

## 设置与数据存储

默认数据目录是应用工作目录下的 `data/`：

```text
data/
├─ settings.json
├─ favorites.json
└─ history.json
```

- `settings.json`：主题、语言、自定义数据目录和 JSON 缩进
- `favorites.json`：工具、输入和 HTTP 请求收藏；敏感字段自动脱敏
- `history.json`：最近输入与输出快照，最多保留 50 条、单项最多 64 KB
- `developer-toolbox.log`：INFO 级应用生命周期和工具错误日志

设置页指定自定义目录后：

- 设置文件继续保存在默认 `data/settings.json`
- 收藏和历史记录立即切换到自定义目录
- 无需重启应用即可生效
- 下次启动时应用会先读取默认设置文件，再恢复自定义目录

Storage 仅允许 Repository 使用简单文件名作为 key，拒绝绝对路径和路径穿越。

## 主题与语言

主题支持：

- `dark`：深色
- `light`：浅色
- `system`：跟随操作系统，并监听系统主题实时变化

语言支持：

- `en`：English
- `zh`：中文

语言切换会即时更新侧栏、首页、设置页、工具名称、工具描述、分类、日期格式和提示消息。旧版本或非法设置会由后端自动归一化为合法默认值。

## 新增工具

1. 在 `backend/tools/<tool>/` 创建独立包。
2. 实现 `registry.ToolExecutor`：

```go
type ToolExecutor interface {
    Info() models.Tool
    Execute(ctx models.ToolContext, input models.ToolInput) models.ToolOutput
}
```

简单的纯本地工具也可以继续实现兼容接口 `Execute(input models.ToolInput)`。

3. 在 `backend/registry/default_registry.go` 的 `CreateDefaultRegistry` 中注册。
4. 为正常输入、空输入、非法输入和边界输入添加测试。
5. 在 `frontend/src/main.js` 的 `setupToolEditor` 和 `executeToolCommand` 中补充专属选项。
6. 在 `frontend/src/i18n.js` 添加中英文工具名称和描述。

新增工具不应修改现有 Tool 的代码，也不应将工具算法写入 App 或前端。

## 常见问题

### 切换主题后没有变化

请确认运行的是最新构建，并重新执行：

```powershell
cd frontend
npm run build
```

应用当前使用 `data-theme="light|dark"` 和 CSS 语义变量控制全部颜色。`system` 模式取决于操作系统当前主题。

### 语言或设置重启后丢失

检查应用是否有权限写入工作目录下的 `data/settings.json`。损坏或旧版本字段会被自动补全；完全无法读取时应用会回退默认设置。

### 浏览器预览提示 Wails API 不存在

这是预期行为。`npm run dev` 仅运行前端，完整后端能力需要使用：

```powershell
wails dev
```

### 文件 Hash 无法选择文件

原生文件选择依赖 Wails Runtime，不能在普通浏览器预览中使用。

## 新增通信工具说明

- TCP、UDP 和串口工具用于开发调试，请只连接你有权访问的设备和服务。
- TCP 响应最多读取 1 MiB；UDP 单个数据报最多 65,507 字节。
- 串口单次发送和读取最多 1 MiB，操作结束后立即关闭端口。
- 所有主机、端口、波特率和超时参数均由后端校验，不经过 Shell。
- 数据库语法文档完全离线，不创建数据库连接，也不执行 SQL。

## 版本范围

V0.5 已完成 Tool Framework、统一编辑器、数据与网络工具、开发辅助、工作流和稳定性建设。JWT 仍仅做解码，不执行签名验证，也不用于认证决策。

## License

本项目基于 [Apache License 2.0](LICENSE) 开源。

在遵守许可证条款的前提下，你可以免费使用、复制、修改、发布和分发本项目，也可以将其用于商业用途。再分发源代码或衍生作品时，需要：

- 随附 Apache License 2.0 许可证副本；
- 对修改过的文件作出明确说明；
- 保留适用的版权、专利、商标和署名声明；
- 如果发行包包含 `NOTICE` 文件，按许可证要求保留其中适用的声明。

Apache License 2.0 同时包含专利授权，并明确软件按“原样”提供，不附带任何明示或默示担保。完整、具有约束力的条款以仓库中的 [LICENSE](LICENSE) 文件为准。
