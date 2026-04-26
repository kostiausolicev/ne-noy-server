package model

type Attachment struct {
	BaseModel
	ID       int64
	Url      string
	Filename string
}
