package dto

type MessageEvent struct {
	Message string `json:"message"`
}

type ErrorEvent struct {
	Message string `json:"message"`
}

type ProjectStatusEvent struct {
	Message   string `json:"message"`
	Step      int    `json:"step"`
	TotalStep int    `json:"totalStep"`
}
