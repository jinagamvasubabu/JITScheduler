package model

type ErrorResponse struct {
	ErrCode    int    `json:"errCode"`
	ErrMessage string `json:"errMessage"`
	Message    any    `json:"message"`
}

type CorrectResponse struct {
	Message any `json:"message"`
}
