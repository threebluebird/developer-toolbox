// Package pinyintool 提供离线汉字全拼和首字母转换。
package pinyintool

import (
	"strings"
	"unicode"

	"developer-toolbox/backend/models"
	"github.com/mozillazg/go-pinyin"
)

type PinyinTool struct{}

func NewPinyinTool() *PinyinTool { return &PinyinTool{} }
func (t *PinyinTool) Info() models.Tool {
	return models.Tool{ID: "pinyin", Name: "Chinese Pinyin", Description: "Convert Chinese text to full pinyin or initials offline.", Category: "text", Icon: "pinyin", Version: "0.5.0", Keywords: []string{"pinyin", "拼音", "全拼", "首字母", "chinese"}}
}

// Execute 保留非汉字字符；separator 控制全拼词间分隔符。
func (t *PinyinTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, ok := input.Payload["input"].(string)
	if !ok {
		return models.Failure("PINYIN_ERROR", "输入必须是文本")
	}
	mode := stringValue(input.Payload["mode"], "full")
	separator := stringValue(input.Payload["separator"], " ")
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	var parts []string
	for _, r := range raw {
		values := pinyin.SinglePinyin(r, args)
		if len(values) > 0 {
			if mode == "initial" {
				parts = append(parts, string([]rune(values[0])[0]))
			} else {
				parts = append(parts, values[0])
			}
		} else if !unicode.IsSpace(r) {
			parts = append(parts, string(r))
		}
	}
	if mode != "full" && mode != "initial" {
		return models.Failure("PINYIN_ERROR", "不支持的拼音模式")
	}
	join := separator
	if mode == "initial" {
		join = ""
	}
	return models.ToolOutput{Success: true, Data: strings.Join(parts, join)}
}
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return f
}
