package dto

type MessageEvent struct {
	Message string `json:"content"`
}

type ErrorEvent struct {
	Message string `json:"message"`
}

type ProjectCreatedEvent struct {
	Reference string `json:"reference"`
	Message   string `json:"message"`
}
