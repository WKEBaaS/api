package router

import "github.com/danielgtaylor/huma/v2"

type ErrorResponseModel struct {
	Ok      bool     `json:"ok" doc:"指示請求是否成功 (此處為 false)"`
	Status  int      `json:"status" doc:"HTTP 狀態碼"`
	Message string   `json:"message" doc:"給使用者看的錯誤訊息"`
	Details []string `json:"details,omitempty" doc:"詳細的錯誤資訊 (可選)"`
}

func (e *ErrorResponseModel) Error() string {
	return e.Message
}

func (e *ErrorResponseModel) GetStatus() int {
	return e.Status
}

func NewCustomError(status int, message string, errs ...error) huma.StatusError {
	details := make([]string, len(errs))
	for i, err := range errs {
		details[i] = err.Error()
	}

	return &ErrorResponseModel{
		Ok:      false,
		Status:  status,
		Message: message,
		Details: details,
	}
}
