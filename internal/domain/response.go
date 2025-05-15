package domain

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Errors:  nil,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(errors ...string) APIResponse {
	return APIResponse{
		Success: false,
		Data:    nil,
		Errors:  errors,
	}
}
