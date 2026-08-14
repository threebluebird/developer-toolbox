package xmltool

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"

	"developer-toolbox/backend/models"
)

type XMLTool struct{}
type node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Nodes   []node     `xml:",any"`
}

func NewXMLTool() *XMLTool { return &XMLTool{} }
func (t *XMLTool) Info() models.Tool {
	return models.Tool{ID: "xml", Name: "XML Toolkit", Description: "Format, validate, and convert XML and JSON.", Category: "data", Icon: "xml", Version: "0.5.0", Keywords: []string{"xml", "json", "format", "validate", "convert"}}
}
func (t *XMLTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, ok := input.Payload["input"].(string)
	if !ok {
		return fail("input must be a string")
	}
	action, _ := input.Payload["action"].(string)
	switch action {
	case "format":
		result, err := formatXML(raw)
		if err != nil {
			return fail(err.Error())
		}
		return success(result)
	case "validate":
		var v any
		if err := xml.Unmarshal([]byte(raw), &v); err != nil {
			return fail(err.Error())
		}
		return success("valid")
	case "xmlToJson":
		var root node
		if err := xml.Unmarshal([]byte(raw), &root); err != nil {
			return fail(err.Error())
		}
		b, _ := json.MarshalIndent(map[string]any{root.XMLName.Local: nodeValue(root)}, "", "  ")
		return success(string(b))
	case "jsonToXml":
		var value any
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			return fail(err.Error())
		}
		root := "root"
		if object, ok := value.(map[string]any); ok && len(object) == 1 {
			for k, v := range object {
				root = k
				value = v
			}
		}
		var b bytes.Buffer
		b.WriteString(xml.Header)
		writeXML(&b, root, value, 0)
		return success(b.String())
	default:
		return fail("unknown action")
	}
}

func formatXML(raw string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(raw))
	var b bytes.Buffer
	encoder := xml.NewEncoder(&b)
	encoder.Indent("", "  ")
	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		if err := encoder.EncodeToken(token); err != nil {
			return "", err
		}
	}
	if err := encoder.Flush(); err != nil {
		return "", err
	}
	return b.String(), nil
}
func nodeValue(n node) any {
	result := map[string]any{}
	for _, a := range n.Attrs {
		result["@"+a.Name.Local] = a.Value
	}
	for _, child := range n.Nodes {
		v := nodeValue(child)
		if old, exists := result[child.XMLName.Local]; exists {
			if list, ok := old.([]any); ok {
				result[child.XMLName.Local] = append(list, v)
			} else {
				result[child.XMLName.Local] = []any{old, v}
			}
		} else {
			result[child.XMLName.Local] = v
		}
	}
	text := strings.TrimSpace(n.Content)
	if len(result) == 0 {
		return text
	}
	if text != "" {
		result["#text"] = text
	}
	return result
}
func writeXML(b *bytes.Buffer, name string, value any, depth int) {
	indent := strings.Repeat("  ", depth)
	switch v := value.(type) {
	case []any:
		for _, item := range v {
			writeXML(b, name, item, depth)
		}
	case map[string]any:
		b.WriteString(indent + "<" + name)
		for k, val := range v {
			if strings.HasPrefix(k, "@") {
				b.WriteString(" " + k[1:] + "=\"")
				_ = xml.EscapeText(b, []byte(toString(val)))
				b.WriteString("\"")
			}
		}
		b.WriteString(">\n")
		if text, ok := v["#text"]; ok {
			b.WriteString(strings.Repeat("  ", depth+1))
			_ = xml.EscapeText(b, []byte(toString(text)))
			b.WriteByte('\n')
		}
		for k, val := range v {
			if !strings.HasPrefix(k, "@") {
				writeXML(b, k, val, depth+1)
			}
		}
		b.WriteString(indent + "</" + name + ">\n")
	default:
		b.WriteString(indent + "<" + name + ">")
		_ = xml.EscapeText(b, []byte(toString(v)))
		b.WriteString("</" + name + ">\n")
	}
}
func toString(v any) string {
	b, _ := json.Marshal(v)
	if s, ok := v.(string); ok {
		return s
	}
	return string(b)
}
func success(v any) models.ToolOutput { return models.ToolOutput{Success: true, Data: v} }
func fail(s string) models.ToolOutput { return models.Failure("INVALID_XML", s) }
