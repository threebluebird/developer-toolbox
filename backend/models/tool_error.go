package models

// ToolError describes a structured error from tool execution.
type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *ToolError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}
