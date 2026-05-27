package repository

import (
	"context"
	"ne_noy/internal/model/events"

	"github.com/google/uuid"
)

type EventBaseRepository interface {
	// GetAll возвращает все записи мероприятий с фильтрами
	// 	roleCode - код роли, по которой нужно отфильтровать мероприятия
	//  archived - показать архивные мероприятия
	GetAll(ctx context.Context, roleCode *string, archived *bool) ([]*events.EventView, error)

	// GetAllByOrg возвращает все мероприятия, в которых есть переданный организатор
	//  orgId - ID организатора, по которому нужно отфильтровать мероприятия
	GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*events.EventView, error)

	Delete(ctx context.Context, id uuid.UUID, eventType string) error

	// Publish устанавливает статус ACTIVE в профильной таблице мероприятия.
	Publish(ctx context.Context, id uuid.UUID) error
}
