package dto

import "time"

type ErrorResponse struct {
	RequestId string    `json:"requestId"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
}
