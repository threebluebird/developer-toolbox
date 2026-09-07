package registry

import (
	"errors"
	"sync"

	"developer-toolbox/backend/models"
	"developer-toolbox/backend/tools/amounttool"
	"developer-toolbox/backend/tools/base64tool"
	"developer-toolbox/backend/tools/cidrtool"
	"developer-toolbox/backend/tools/codegen"
	"developer-toolbox/backend/tools/colortool"
	"developer-toolbox/backend/tools/crontool"
	"developer-toolbox/backend/tools/csvtool"
	"developer-toolbox/backend/tools/dbdocs"
	"developer-toolbox/backend/tools/dnstool"
	"developer-toolbox/backend/tools/documentsplit"
	"developer-toolbox/backend/tools/gittool"
	hashtool "developer-toolbox/backend/tools/hash"
	"developer-toolbox/backend/tools/hiittimer"
	"developer-toolbox/backend/tools/htmltool"
	"developer-toolbox/backend/tools/httpclient"
	jsontool "developer-toolbox/backend/tools/json"
	"developer-toolbox/backend/tools/jwttool"
	"developer-toolbox/backend/tools/markdowntool"
	"developer-toolbox/backend/tools/pingtool"
	"developer-toolbox/backend/tools/pinyintool"
	"developer-toolbox/backend/tools/porttool"
	"developer-toolbox/backend/tools/regex"
	"developer-toolbox/backend/tools/serialtool"
	"developer-toolbox/backend/tools/sqltool"
	"developer-toolbox/backend/tools/tcptool"
	"developer-toolbox/backend/tools/texttool"
	"developer-toolbox/backend/tools/timestamptool"
	"developer-toolbox/backend/tools/tomltool"
	"developer-toolbox/backend/tools/udptool"
	"developer-toolbox/backend/tools/urlparser"
	"developer-toolbox/backend/tools/urltool"
	uuidtool "developer-toolbox/backend/tools/uuid"
	"developer-toolbox/backend/tools/xmltool"
	"developer-toolbox/backend/tools/yamltool"
)

var (
	ErrToolAlreadyRegistered = errors.New("tool already registered")
	ErrToolNotFound          = errors.New("tool not found")
)

// DefaultRegistry implements Registry with a thread-safe in-memory map.
type DefaultRegistry struct {
	mu    sync.RWMutex
	tools map[string]ToolDefinition
	order []string
}

// NewDefaultRegistry creates an empty registry instance.
func NewDefaultRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		tools: make(map[string]ToolDefinition),
		order: make([]string, 0),
	}
}

// CreateDefaultRegistry 集中注册 V0.1 的全部工具。
// order 会保留这里的注册顺序，因此新增工具只需在此追加。
func CreateDefaultRegistry() *DefaultRegistry {
	r := NewDefaultRegistry()
	_ = r.Register(jsontool.NewJSONTool())
	_ = r.Register(base64tool.NewBase64Tool())
	_ = r.Register(urltool.NewURLTool())
	_ = r.Register(hashtool.NewHashTool())
	_ = r.Register(uuidtool.NewUUIDTool())
	_ = r.Register(timestamptool.NewTimestampTool())
	_ = r.Register(jwttool.NewJWTTool())
	_ = r.Register(regex.NewRegexTool())
	_ = r.Register(yamltool.NewYAMLTool())
	_ = r.Register(xmltool.NewXMLTool())
	_ = r.Register(csvtool.NewCSVTool())
	_ = r.Register(tomltool.NewTOMLTool())
	_ = r.Register(texttool.NewTextTool())
	_ = r.Register(sqltool.NewSQLTool())
	_ = r.Register(markdowntool.NewMarkdownTool())
	_ = r.Register(htmltool.NewHTMLTool())
	_ = r.Register(httpclient.NewHTTPClientTool())
	_ = r.Register(dnstool.NewDNSTool())
	_ = r.Register(pingtool.NewPingTool())
	_ = r.Register(porttool.NewPortTool())
	_ = r.Register(cidrtool.NewCIDRTool())
	_ = r.Register(urlparser.NewURLParserTool())
	_ = r.Register(crontool.NewCronTool())
	_ = r.Register(colortool.NewColorTool())
	_ = r.Register(codegen.NewCodeGeneratorTool())
	_ = r.Register(gittool.NewGitTool())
	_ = r.Register(amounttool.NewAmountTool())
	_ = r.Register(tcptool.NewTCPTool())
	_ = r.Register(udptool.NewUDPTool())
	_ = r.Register(serialtool.NewSerialTool())
	_ = r.Register(pinyintool.NewPinyinTool())
	_ = r.Register(dbdocs.NewDatabaseDocsTool())
	_ = r.Register(documentsplit.NewTool())
	_ = r.Register(hiittimer.NewTool())
	return r
}

// Register adds a new tool to the registry.
// It returns ErrToolAlreadyRegistered when the tool ID is already present.
func (r *DefaultRegistry) Register(tool ToolDefinition) error {
	if tool == nil {
		return errors.New("tool cannot be nil")
	}

	info := tool.Info()
	if info.ID == "" {
		return errors.New("tool id cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 工具 ID 是前端、历史和收藏使用的稳定标识，禁止重复注册覆盖旧实现。
	if _, exists := r.tools[info.ID]; exists {
		return ErrToolAlreadyRegistered
	}

	r.tools[info.ID] = tool
	r.order = append(r.order, info.ID)
	return nil
}

// Get returns a tool executor by its ID.
func (r *DefaultRegistry) Get(toolID string) (ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[toolID]
	return tool, ok
}

// List returns metadata for all registered tools in registration order.
func (r *DefaultRegistry) List() []models.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 不直接遍历 map，保证每次返回给前端的工具顺序稳定。
	tools := make([]models.Tool, 0, len(r.order))
	for _, id := range r.order {
		if tool, ok := r.tools[id]; ok {
			tools = append(tools, tool.Info())
		}
	}
	return tools
}
