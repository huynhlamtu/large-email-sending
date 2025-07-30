package models

import "github.com/google/uuid"

type EmailSchedule struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ScheduledTime  int64
	FilterQuery    string
	Subject        string
	Body           string
	Type           int8       // 0: once, 1: daily, 2: weekly, 3: monthly, 4: after 30 seconds
	RootScheduleID *uuid.UUID // for daily, weekly, monthly, after 30 seconds
	Status         string     // pending | scheduled
}
