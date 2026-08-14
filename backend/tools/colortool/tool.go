package colortool

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"developer-toolbox/backend/models"
)

type ColorTool struct{}
type rgba struct {
	R, G, B int
	A       float64
}

func NewColorTool() *ColorTool { return &ColorTool{} }
func (t *ColorTool) Info() models.Tool {
	return models.Tool{ID: "color", Name: "Color Converter", Description: "Convert HEX, RGB, RGBA, HSL, and HSV colors.", Category: "developer", Icon: "color", Version: "0.5.0", Keywords: []string{"color", "hex", "rgb", "rgba", "hsl", "hsv", "css"}}
}
func (t *ColorTool) Execute(input models.ToolInput) models.ToolOutput {
	raw := strings.TrimSpace(stringValue(input.Payload["input"], ""))
	color, err := parseColor(raw)
	if err != nil {
		return bad(err.Error())
	}
	h, s, l := rgbHSL(color)
	hh, ss, v := rgbHSV(color)
	hex := fmt.Sprintf("#%02X%02X%02X", color.R, color.G, color.B)
	if color.A < 1 {
		hex += fmt.Sprintf("%02X", int(math.Round(color.A*255)))
	}
	return good(map[string]any{"hex": hex, "rgb": fmt.Sprintf("rgb(%d, %d, %d)", color.R, color.G, color.B), "rgba": fmt.Sprintf("rgba(%d, %d, %d, %.2f)", color.R, color.G, color.B, color.A), "hsl": fmt.Sprintf("hsl(%d, %d%%, %d%%)", round(h), round(s*100), round(l*100)), "hsv": fmt.Sprintf("hsv(%d, %d%%, %d%%)", round(hh), round(ss*100), round(v*100)), "channels": map[string]any{"r": color.R, "g": color.G, "b": color.B, "a": color.A, "h": round(h), "s": round(s * 100), "l": round(l * 100), "v": round(v * 100)}})
}

var functionPattern = regexp.MustCompile(`(?i)^(rgba?|hsla?|hsv)\s*\(([^)]+)\)$`)

func parseColor(raw string) (rgba, error) {
	if strings.HasPrefix(raw, "#") {
		return parseHex(raw)
	}
	match := functionPattern.FindStringSubmatch(raw)
	if len(match) != 3 {
		return rgba{}, fmt.Errorf("unsupported color format")
	}
	parts := strings.Split(match[2], ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(strings.TrimSuffix(parts[i], "%"))
	}
	switch strings.ToLower(match[1]) {
	case "rgb", "rgba":
		if len(parts) < 3 || len(parts) > 4 {
			return rgba{}, fmt.Errorf("invalid RGB color")
		}
		r, e1 := number(parts[0])
		g, e2 := number(parts[1])
		b, e3 := number(parts[2])
		if e1 != nil || e2 != nil || e3 != nil || r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
			return rgba{}, fmt.Errorf("RGB channels must be 0-255")
		}
		a := 1.0
		if len(parts) == 4 {
			a, _ = strconv.ParseFloat(parts[3], 64)
			if a < 0 || a > 1 {
				return rgba{}, fmt.Errorf("alpha must be 0-1")
			}
		}
		return rgba{int(r), int(g), int(b), a}, nil
	case "hsl", "hsla", "hsv":
		if len(parts) < 3 || len(parts) > 4 {
			return rgba{}, fmt.Errorf("invalid color")
		}
		h, e1 := number(parts[0])
		s, e2 := number(parts[1])
		lv, e3 := number(parts[2])
		if e1 != nil || e2 != nil || e3 != nil || s < 0 || s > 100 || lv < 0 || lv > 100 {
			return rgba{}, fmt.Errorf("invalid HSL/HSV channels")
		}
		a := 1.0
		if len(parts) == 4 {
			a, _ = strconv.ParseFloat(parts[3], 64)
		}
		if strings.HasPrefix(strings.ToLower(match[1]), "hsl") {
			return hslRGB(h, s/100, lv/100, a), nil
		}
		return hsvRGB(h, s/100, lv/100, a), nil
	}
	return rgba{}, fmt.Errorf("unsupported color")
}
func parseHex(raw string) (rgba, error) {
	value := strings.TrimPrefix(raw, "#")
	if len(value) == 3 || len(value) == 4 {
		expanded := ""
		for _, r := range value {
			expanded += string(r) + string(r)
		}
		value = expanded
	}
	if len(value) != 6 && len(value) != 8 {
		return rgba{}, fmt.Errorf("HEX must have 3, 4, 6, or 8 digits")
	}
	n, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return rgba{}, fmt.Errorf("invalid HEX")
	}
	if len(value) == 6 {
		return rgba{int(n >> 16), int(n>>8) & 255, int(n) & 255, 1}, nil
	}
	return rgba{int(n >> 24), int(n>>16) & 255, int(n>>8) & 255, float64(n&255) / 255}, nil
}
func rgbHSL(c rgba) (float64, float64, float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	maxV, minV := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	l := (maxV + minV) / 2
	if maxV == minV {
		return 0, 0, l
	}
	d := maxV - minV
	s := d / (1 - math.Abs(2*l-1))
	h := 0.0
	if maxV == r {
		h = 60 * math.Mod((g-b)/d, 6)
	} else if maxV == g {
		h = 60 * ((b-r)/d + 2)
	} else {
		h = 60 * ((r-g)/d + 4)
	}
	if h < 0 {
		h += 360
	}
	return h, s, l
}
func rgbHSV(c rgba) (float64, float64, float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	maxV, minV := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	d := maxV - minV
	h := 0.0
	if d != 0 {
		if maxV == r {
			h = 60 * math.Mod((g-b)/d, 6)
		} else if maxV == g {
			h = 60 * ((b-r)/d + 2)
		} else {
			h = 60 * ((r-g)/d + 4)
		}
	}
	if h < 0 {
		h += 360
	}
	s := 0.0
	if maxV != 0 {
		s = d / maxV
	}
	return h, s, maxV
}
func hslRGB(h, s, l, a float64) rgba {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	r, g, b := sector(h, c, x)
	return rgba{round((r + m) * 255), round((g + m) * 255), round((b + m) * 255), a}
}
func hsvRGB(h, s, v, a float64) rgba {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	r, g, b := sector(h, c, x)
	return rgba{round((r + m) * 255), round((g + m) * 255), round((b + m) * 255), a}
}
func sector(h, c, x float64) (float64, float64, float64) {
	switch int(math.Mod(h, 360) / 60) {
	case 0:
		return c, x, 0
	case 1:
		return x, c, 0
	case 2:
		return 0, c, x
	case 3:
		return 0, x, c
	case 4:
		return x, 0, c
	default:
		return c, 0, x
	}
}
func number(s string) (float64, error) { return strconv.ParseFloat(s, 64) }
func round(v float64) int              { return int(math.Round(v)) }
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return f
}
func good(v any) models.ToolOutput   { return models.ToolOutput{Success: true, Data: v} }
func bad(s string) models.ToolOutput { return models.Failure("COLOR_ERROR", s) }
