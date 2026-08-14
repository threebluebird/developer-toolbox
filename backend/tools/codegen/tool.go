package codegen

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"developer-toolbox/backend/models"
)

type CodeGeneratorTool struct{}

func NewCodeGeneratorTool() *CodeGeneratorTool { return &CodeGeneratorTool{} }
func (t *CodeGeneratorTool) Info() models.Tool {
	return models.Tool{ID: "code-generator", Name: "Code Generator", Description: "Generate Go, TypeScript, or Java types from JSON.", Category: "developer", Icon: "code", Version: "0.5.0", Keywords: []string{"json", "go struct", "typescript", "interface", "java", "class", "code generator"}}
}
func (t *CodeGeneratorTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, ok := input.Payload["input"].(string)
	if !ok {
		return bad("input must be a string")
	}
	var value any
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return bad(err.Error())
	}
	object, ok := value.(map[string]any)
	if !ok {
		return bad("root JSON value must be an object")
	}
	name := exportName(stringValue(input.Payload["name"], "Root"))
	language := stringValue(input.Payload["language"], "go")
	switch language {
	case "go":
		return good(generateGo(name, object))
	case "typescript":
		return good(generateTS(name, object))
	case "java":
		return good(generateJava(name, object))
	default:
		return bad("unsupported target language")
	}
}
func generateGo(name string, object map[string]any) string {
	var nested []string
	body := goStruct(name, object, &nested)
	return strings.Join(append([]string{body}, nested...), "\n\n")
}
func goStruct(name string, object map[string]any, nested *[]string) string {
	keys := sorted(object)
	var b strings.Builder
	fmt.Fprintf(&b, "type %s struct {\n", name)
	for _, key := range keys {
		field := exportName(key)
		kind := goType(field, object[key], nested)
		fmt.Fprintf(&b, "    %s %s `json:\"%s\"`\n", field, kind, key)
	}
	b.WriteString("}")
	return b.String()
}
func goType(name string, v any, nested *[]string) string {
	switch x := v.(type) {
	case string:
		return "string"
	case bool:
		return "bool"
	case json.Number:
		if strings.Contains(x.String(), ".") {
			return "float64"
		}
		return "int64"
	case nil:
		return "any"
	case map[string]any:
		*nested = append(*nested, goStruct(name, x, nested))
		return name
	case []any:
		if len(x) == 0 {
			return "[]any"
		}
		return "[]" + goType(name+"Item", x[0], nested)
	default:
		return "any"
	}
}
func generateTS(name string, object map[string]any) string {
	var nested []string
	body := tsInterface(name, object, &nested)
	return strings.Join(append([]string{body}, nested...), "\n\n")
}
func tsInterface(name string, object map[string]any, nested *[]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "export interface %s {\n", name)
	for _, key := range sorted(object) {
		fmt.Fprintf(&b, "  %s: %s;\n", safeIdentifier(key), tsType(exportName(key), object[key], nested))
	}
	b.WriteString("}")
	return b.String()
}
func tsType(name string, v any, nested *[]string) string {
	switch x := v.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case json.Number:
		return "number"
	case nil:
		return "unknown"
	case map[string]any:
		*nested = append(*nested, tsInterface(name, x, nested))
		return name
	case []any:
		if len(x) == 0 {
			return "unknown[]"
		}
		return tsType(name+"Item", x[0], nested) + "[]"
	default:
		return "unknown"
	}
}
func generateJava(name string, object map[string]any) string {
	var nested []string
	body := javaClass(name, object, &nested, true)
	return strings.Join(append([]string{body}, nested...), "\n\n")
}
func javaClass(name string, object map[string]any, nested *[]string, public bool) string {
	prefix := "static "
	if public {
		prefix = "public "
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%sclass %s {\n", prefix, name)
	for _, key := range sorted(object) {
		fmt.Fprintf(&b, "    private %s %s;\n", javaType(exportName(key), object[key], nested), safeIdentifier(key))
	}
	b.WriteString("}")
	return b.String()
}
func javaType(name string, v any, nested *[]string) string {
	switch x := v.(type) {
	case string:
		return "String"
	case bool:
		return "boolean"
	case json.Number:
		if strings.Contains(x.String(), ".") {
			return "double"
		}
		return "long"
	case nil:
		return "Object"
	case map[string]any:
		*nested = append(*nested, javaClass(name, x, nested, false))
		return name
	case []any:
		if len(x) == 0 {
			return "List<Object>"
		}
		return "List<" + boxed(javaType(name+"Item", x[0], nested)) + ">"
	default:
		return "Object"
	}
}
func boxed(v string) string {
	return strings.NewReplacer("boolean", "Boolean", "long", "Long", "double", "Double").Replace(v)
}
func exportName(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	var b strings.Builder
	for _, part := range parts {
		r := []rune(part)
		if len(r) == 0 {
			continue
		}
		b.WriteRune(unicode.ToUpper(r[0]))
		b.WriteString(string(r[1:]))
	}
	if b.Len() == 0 {
		return "Field"
	}
	return b.String()
}
func safeIdentifier(v string) string {
	if v == "" {
		return "field"
	}
	for i, r := range v {
		if !(unicode.IsLetter(r) || r == '_' || (i > 0 && unicode.IsDigit(r))) {
			return fmt.Sprintf("%q", v)
		}
	}
	return v
}
func sorted(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return f
}
func good(v any) models.ToolOutput   { return models.ToolOutput{Success: true, Data: v} }
func bad(s string) models.ToolOutput { return models.Failure("CODEGEN_ERROR", s) }
