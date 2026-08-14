package services

import (
	"strings"

	"developer-toolbox/backend/models"
)

type SearchService struct {
	tools *ToolService
}

func NewSearchService(tools *ToolService) *SearchService {
	return &SearchService{tools: tools}
}

func (s *SearchService) Search(query string) []models.Tool {
	// V0.1 数据量很小，大小写无关的线性匹配比引入索引更简单可靠。
	query = strings.ToLower(strings.TrimSpace(query))
	tools := s.tools.ListTools()
	if query == "" {
		return tools
	}

	matched := make([]models.Tool, 0)
	for _, tool := range tools {
		// 搜索契约覆盖名称、ID、描述和分类，前端无需重复业务规则。
		haystack := append([]string{tool.ID, tool.Name, tool.Description, tool.Category}, tool.Keywords...)
		for _, value := range haystack {
			if strings.Contains(strings.ToLower(value), query) {
				matched = append(matched, tool)
				break
			}
		}
	}
	return matched
}
