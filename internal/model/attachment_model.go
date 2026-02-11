package model

import (
	"time"
)

type Attachment struct {
	ID        int64
	Url       string
	Filename  string
	CreatedAt time.Time
}
