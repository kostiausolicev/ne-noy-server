package callback_dto

type GroupEvent struct {
	EventType string `json:"type"`
	GroupId   int64  `json:"group_id"`
	Secret    string `json:"secret"`
	Object    any    `json:"object"`
}
