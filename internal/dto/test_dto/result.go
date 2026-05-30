package test_dto

import "ne_noy/internal/dto"

type MyTestResultDto struct {
	Question          QuestionDto `json:"question"`
	SelectedAnswerIds []string    `json:"selected_answer_ids"`
}

type UserTestResultDto struct {
	User     dto.UserMiniDto  `json:"user"`
	Attempts []TestAttemptDto `json:"attempts"`
}

type TestAttemptDto struct {
	CorrectCount int `json:"correct_count"`
	TotalCount   int `json:"total_count"`
}

type TestUserResultDetailDto struct {
	User     dto.UserMiniDto        `json:"user"`
	Attempts []UserAttemptDetailDto `json:"attempts"`
}
