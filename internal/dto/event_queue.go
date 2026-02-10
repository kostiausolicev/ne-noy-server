package dto

type EventQueueDto struct {
	PostID      int64           `json:"post_id"`
	Text        string          `json:"text"`
	Lat         *float64        `json:"lat,omitempty"`
	Lon         *float64        `json:"lon,omitempty"`
	Address     *string         `json:"address,omitempty"`
	Poll        PollDto         `json:"poll,omitempty"`
	Photos      []PhotoDto      `json:"photos,omitempty"`
	Attachments []AttachmentDto `json:"attachments,omitempty"`
}

type PollDto struct {
	ID      int64       `json:"id"`
	Answers []AnswerDto `json:"answers"`
}

type AnswerDto struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type AttachmentDto struct {
	ID    int64  `json:"id"`
	Url   string `json:"url"`
	Title string `json:"title"`
}

type PhotoDto struct {
	ID    int64          `json:"id"`
	Sizes []PhotoSizeDto `json:"sizes"`
}

type PhotoSizeDto struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}
