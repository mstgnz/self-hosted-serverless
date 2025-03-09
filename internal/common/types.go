package common

// FunctionHandler is the interface that all serverless functions must implement
type FunctionHandler interface {
	Execute(input map[string]any) (any, error)
}

// FunctionInfo represents metadata about a registered function
type FunctionInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Runtime     string `json:"runtime"`
}
