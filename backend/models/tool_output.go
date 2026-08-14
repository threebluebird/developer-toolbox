package models

// ToolOutput is the common response payload returned from tool execution.
type ToolOutput struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ToolError `json:"error,omitempty"`
}

func Failure(code, message string) ToolOutput {
	return ToolOutput{Success: false, Error: &ToolError{Code: code, Message: message}}
}

func FailureWithDetail(code, message, detail string) ToolOutput {
	return ToolOutput{Success: false, Error: &ToolError{Code: code, Message: message, Detail: detail}}
}
