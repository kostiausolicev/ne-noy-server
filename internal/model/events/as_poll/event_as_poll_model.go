package as_poll

import "ne_noy/internal/model/events"

type AsPoll struct {
	events.EventProfile
	events.EventRelations
	ExtLinkID *string
	VkPostID  *int64
}

func (e AsPoll) TableName() string {
	return "event_as_polls"
}
