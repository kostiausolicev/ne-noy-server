package group_event

type PlaceObj struct {
	ID      int64    `json:"id"`
	Lat     *float64 `json:"latitude" mapstructure:"latitude"`
	Lon     *float64 `json:"longitude" mapstructure:"longitude"`
	Address *string  `json:"title" mapstructure:"title"`
}

type GeoObject struct {
	Type  string    `json:"type"`
	Place *PlaceObj `json:"place"`
}

type AnswerObject struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type DocObject struct {
	ID    int64  `json:"id" mapstructure:"id"`
	Title string `json:"title" mapstructure:"title"`
	Url   string `json:"url" mapstructure:"url"`
}

type PhotoSizeObject struct {
	Type string `json:"type" mapstructure:"type"`
	Url  string `json:"url" mapstructure:"url"`
}

type PhotoObject struct {
	ID    int64             `json:"id" mapstructure:"id"`
	Sizes []PhotoSizeObject `json:"sizes" mapstructure:"sizes"`
}

type PollObject struct {
	ID       int64          `json:"id"`
	Question string         `json:"question"`
	Answers  []AnswerObject `json:"answers"`
}

type AttachmentObject struct {
	Type  string       `json:"type"`
	Photo *PhotoObject `json:"photo"`
	Doc   *DocObject   `json:"doc"`
	Poll  *PollObject  `json:"poll"`
}

type NewPostEvent struct {
	ID          int64              `json:"id"`
	Text        string             `json:"text"`
	Geo         GeoObject          `json:"geo"`
	Attachments []AttachmentObject `json:"attachments"`
}
