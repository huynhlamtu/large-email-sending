package models

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string
	Name      string
	Age       int
	Gender    string
	CreatedAt int64
}
