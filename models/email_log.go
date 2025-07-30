package models

import (
	"time"

	"github.com/google/uuid"
)

type EmailLog struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          uuid.UUID
	ScheduleID      uuid.UUID
	SubjectRendered string
	BodyRendered    string
	Status          string // pending | success | fail
	Error           string
	SentAt          time.Time
}
