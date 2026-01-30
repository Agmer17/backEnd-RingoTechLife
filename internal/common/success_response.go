package common

type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}
