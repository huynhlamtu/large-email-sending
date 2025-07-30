package main

import (
	"fmt"
	"go-email-system/db"
	"go-email-system/models"
	"go-email-system/pkg/queue"
	"strings"
	"time"
)

func main() {
	db.Connect()
	queue.InitRabbitMQ()

	var schedules []models.EmailSchedule
	now := time.Now().Unix()
	start := now + 600 // 10 phút sau
	end := now + 720   // 12 phút sau
	// Lấy các schedule cần gửi
	db.DB.Where("scheduled_time BETWEEN ? AND ?", start, end).Find(&schedules)

	for _, schedule := range schedules {
		fmt.Printf("Processing schedule: %s\n", schedule.ID)

		// B1. Build raw query
		query := fmt.Sprintf("SELECT * FROM users WHERE %s", schedule.FilterQuery)

		// B2. Thực thi filter_query
		var users []models.User
		if err := db.DB.Raw(query).Scan(&users).Error; err != nil {
			fmt.Println("Error filtering users:", err)
			continue
		}

		fmt.Printf(" → Found %d users\n", len(users))

		// B3. Render từng email & insert vào email_logs
		for _, user := range users {
			renderedSubject := renderTemplate(schedule.Subject, user)
			renderedBody := renderTemplate(schedule.Body, user)

			log := models.EmailLog{
				UserID:          user.ID,
				ScheduleID:      schedule.ID,
				SubjectRendered: renderedSubject,
				BodyRendered:    renderedBody,
				Status:          "pending",
			}

			db.DB.Create(&log)

			// Gửi vào RabbitMQ
			job := queue.EmailJob{
				LogID:   log.ID,
				Email:   user.Email,
				Subject: renderedSubject,
				Body:    renderedBody,
			}
			queue.Publish(job)
		}
	}
}

// Template engine đơn giản
func renderTemplate(template string, user models.User) string {
	output := strings.ReplaceAll(template, "{{name}}", user.Name)
	output = strings.ReplaceAll(output, "{{age}}", fmt.Sprintf("%d", user.Age))
	output = strings.ReplaceAll(output, "{{email}}", user.Email)
	output = strings.ReplaceAll(output, "{{gender}}", user.Gender)
	return output
}
