package dto

import "github.com/google/uuid"

type RoleDto struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"` // TODO переделать на code
	DisplayName string    `json:"displayName"`
}
