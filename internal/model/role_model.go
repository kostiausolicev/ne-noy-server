package model

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
}
