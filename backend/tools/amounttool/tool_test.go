package amounttool

import (
	"developer-toolbox/backend/models"
	"testing"
)

func TestUpperAmount(t *testing.T) {
	cases := map[string]string{"0": "零元整", "1234.56": "壹仟贰佰叁拾肆元伍角陆分", "1001.01": "壹仟零壹元零壹分", "-10.20": "负壹拾元贰角"}
	for input, want := range cases {
		out := NewAmountTool().Execute(models.ToolInput{Payload: map[string]any{"input": input, "action": "upper"}})
		if !out.Success || out.Data != want {
			t.Errorf("%s: got %#v want %s", input, out.Data, want)
		}
	}
}
