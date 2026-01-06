package dto

import "github.com/google/uuid"

// swagger:model RoleDto
type RoleDto struct {
	// ID Роли
	// format: uuid
	// required: true
	ID uuid.UUID `json:"id"`
	// Код Роли
	// required: true
	Name string `json:"name"` // TODO переделать на code
	// Название Роли
	// required: true
	DisplayName string `json:"displayName"`
}
