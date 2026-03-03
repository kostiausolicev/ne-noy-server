package dto

type UsersGetDto struct {
	AccessToken string `json:"access_token"`
	UserIds     string `json:"user_ids"`
	Fields      string `json:"fields"`
}

type SendMessageDto struct {
	AccessToken string  `json:"access_token"`
	UserIds     string  `json:"user_ids"`
	Message     string  `json:"message"`
	Fragment    *string `json:"fragment"`
}
