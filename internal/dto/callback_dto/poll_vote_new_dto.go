package callback_dto

type PollVoteNewDto struct {
	PollID   int64 `json:"poll_id" mapstructure:"poll_id"`
	OptionID int64 `json:"option_id" mapstructure:"option_id"`
	UserID   int64 `json:"user_id" mapstructure:"user_id"`
}
