package dto

// ErrorResponse エラーレスポンスの標準形式
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse 成功レスポンスの標準形式
type SuccessResponse struct {
	Message string `json:"message"`
}
