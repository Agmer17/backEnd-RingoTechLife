package common

type ErrorResponse struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
}

func NewErrorResponse(code int, msg string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: msg,
	}
}
