package dto

type SendMessageError struct {
	Code    int    `json:"code"`
	Message string `json:"description"`
}

type UserSendMessageStatus struct {
	UserId string            `json:"user_id"`
	Status string            `json:"status"`
	Error  *SendMessageError `json:"error"`
}

type SendMessageResponse struct {
	Response []UserSendMessageStatus `json:"response"`
}

type UsersGet struct {
	VkId      int64  `json:"vk_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhotoId   string `json:"photo_max"`
}

type UsersGetResponse struct {
	Users []UsersGet `json:"response"`
}
