// Package amounttool 提供人民币金额与中文大写之间的本地转换。
package amounttool

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"developer-toolbox/backend/models"
)

type AmountTool struct{}

func NewAmountTool() *AmountTool { return &AmountTool{} }
func (t *AmountTool) Info() models.Tool {
	return models.Tool{ID: "amount-cn", Name: "Chinese Amount", Description: "Convert numeric amounts to Chinese financial uppercase.", Category: "convert", Icon: "amount", Version: "0.5.0", Keywords: []string{"amount", "rmb", "人民币", "金额", "中文大写", "财务"}}
}

// Execute 默认执行阿拉伯数字到人民币大写；lower 动作用于普通中文数字展示。
func (t *AmountTool) Execute(input models.ToolInput) models.ToolOutput {
	raw := strings.TrimSpace(stringValue(input.Payload["input"], ""))
	value, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", ""), 64)
	if err != nil || math.IsInf(value, 0) || math.IsNaN(value) || math.Abs(value) >= 1e16 {
		return fail("请输入绝对值小于一亿亿的有效金额")
	}
	action := stringValue(input.Payload["action"], "upper")
	if action == "lower" {
		return ok(lowerNumber(value))
	}
	if action != "upper" {
		return fail("不支持的转换方式")
	}
	return ok(upperAmount(value))
}

var leadingZero = regexp.MustCompile(`零+`)

func upperAmount(value float64) string {
	negative := value < 0
	value = math.Abs(value)
	cents := int64(math.Round(value * 100))
	integer := cents / 100
	result := integerUpper(integer) + "元"
	jiao, fen := cents/10%10, cents%10
	if jiao == 0 && fen == 0 {
		result += "整"
	} else {
		if jiao > 0 {
			result += upperDigits[jiao] + "角"
		} else if integer > 0 && fen > 0 {
			result += "零"
		}
		if fen > 0 {
			result += upperDigits[fen] + "分"
		}
	}
	if negative && cents != 0 {
		result = "负" + result
	}
	return result
}

var upperDigits = []string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}
var lowerDigits = []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
var smallUnits = []string{"", "拾", "佰", "仟"}
var groupUnits = []string{"", "万", "亿", "万亿"}

func integerUpper(value int64) string {
	if value == 0 {
		return "零"
	}
	groups := []int{}
	for value > 0 {
		groups = append(groups, int(value%10000))
		value /= 10000
	}
	var b strings.Builder
	zeroPending := false
	for i := len(groups) - 1; i >= 0; i-- {
		group := groups[i]
		if group == 0 {
			zeroPending = b.Len() > 0
			continue
		}
		if b.Len() > 0 && (zeroPending || group < 1000) {
			b.WriteString("零")
		}
		b.WriteString(groupUpper(group))
		b.WriteString(groupUnits[i])
		zeroPending = false
	}
	return leadingZero.ReplaceAllString(b.String(), "零")
}
func groupUpper(group int) string {
	var b strings.Builder
	zero := false
	for pos := 3; pos >= 0; pos-- {
		base := int(math.Pow10(pos))
		digit := group / base % 10
		if digit == 0 {
			if b.Len() > 0 && group%base != 0 {
				zero = true
			}
			continue
		}
		if zero {
			b.WriteString("零")
			zero = false
		}
		b.WriteString(upperDigits[digit])
		b.WriteString(smallUnits[pos])
	}
	return b.String()
}
func lowerNumber(value float64) string {
	raw := strconv.FormatFloat(value, 'f', -1, 64)
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= '0' && r <= '9':
			b.WriteString(lowerDigits[r-'0'])
		case r == '.':
			b.WriteString("点")
		case r == '-':
			b.WriteString("负")
		}
	}
	return b.String()
}
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return f
}
func ok(v any) models.ToolOutput      { return models.ToolOutput{Success: true, Data: v} }
func fail(s string) models.ToolOutput { return models.Failure("AMOUNT_ERROR", s) }
