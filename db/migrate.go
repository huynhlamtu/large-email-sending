package db

import (
	"go-email-system/models"
)

func Migrate() {
	Connect()
	DB.AutoMigrate(&models.User{}, &models.EmailSchedule{}, &models.EmailLog{})
}
