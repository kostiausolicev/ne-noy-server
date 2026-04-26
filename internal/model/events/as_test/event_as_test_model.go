package as_test

import (
	"ne_noy/internal/model/events"

	"github.com/google/uuid"
)

type EventAsTest struct {
	events.EventProfile
	events.EventRelations
	ExtLinkID *string
	Attempts  int
	EventID   *uuid.UUID
	VkPostID  *int64
}
