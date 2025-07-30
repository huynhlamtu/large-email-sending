package main

import (
	"fmt"
	"go-email-system/db"
	"go-email-system/models"
	"go-email-system/pkg/queue"
	"strings"
	"time"
)

const (
	scanIntervalSec = 100
	producerCount   = 20
)

func main() {
	db.Connect()
	queue.InitRabbitMQ()

	now := time.Now().Unix()
	end := now + 5*60 // now + 5 minutes

	// Debug: in toàn bộ schedule
	var allSchedules []models.EmailSchedule
	db.DB.Find(&allSchedules)
	for i, schedule := range allSchedules {
		fmt.Println("[DEBUG] schedule:", i, schedule.ID, time.Unix(schedule.ScheduledTime, 0).Format(time.RFC3339))
	}

	formattedNow := time.Unix(now, 0).Format(time.RFC3339)
	formattedTimeEnd := time.Unix(end, 0).Format(time.RFC3339)
	fmt.Println("[DEBUG] now", formattedNow)
	fmt.Println("[DEBUG] scanning pending schedules until", formattedTimeEnd)

	var schedules []models.EmailSchedule
	db.DB.Where("scheduled_time <= ? AND status = ?", end, "pending").Find(&schedules)

	jobs := make(chan queue.EmailJob, 10000)

	// Launch producer goroutine
	for i := 0; i < producerCount; i++ {
		go func(id int) {
			for job := range jobs {
				queue.Publish(job)
			}
		}(i)
	}

	for _, schedule := range schedules {
		fmt.Printf("⏳ Preparing schedule %s\n", schedule.ID)
		query := fmt.Sprintf("SELECT * FROM users WHERE %s", schedule.FilterQuery)

		var users []models.User
		if err := db.DB.Raw(query).Scan(&users).Error; err != nil {
			fmt.Println("❌ Lỗi filter user:", err)
			continue
		}

		for _, user := range users {
			subject := render(schedule.Subject, user)
			body := render(schedule.Body, user)

			log := models.EmailLog{
				UserID:          user.ID,
				ScheduleID:      schedule.ID,
				SubjectRendered: subject,
				BodyRendered:    body,
				Status:          "pending",
			}
			db.DB.Create(&log)

			delayMs := max(0, (schedule.ScheduledTime-time.Now().Unix())*1000)

			jobs <- queue.EmailJob{
				LogID:      log.ID,
				Email:      user.Email,
				Subject:    subject,
				Body:       body,
				DelayMs:    delayMs,
				RetryCount: 0,
			}
		}

		fmt.Printf("✅ Prepared %d emails for schedule %s\n", len(users), schedule.ID)

		schedule.Status = "scheduled"
		db.DB.Save(&schedule)

		// Tạo bản sao cho schedule kế tiếp
		nextSchedule := models.EmailSchedule{
			FilterQuery:    schedule.FilterQuery,
			Subject:        schedule.Subject,
			Body:           schedule.Body,
			Type:           schedule.Type,
			RootScheduleID: &schedule.ID,
			Status:         "pending",
		}

		switch schedule.Type {
		case 1:
			nextSchedule.ScheduledTime = schedule.ScheduledTime + 24*3600
		case 2:
			nextSchedule.ScheduledTime = schedule.ScheduledTime + 7*24*3600
		case 3:
			nextSchedule.ScheduledTime = schedule.ScheduledTime + 30*24*3600
		case 4:
			nextSchedule.ScheduledTime = schedule.ScheduledTime + 30
		default:
			continue // once
		}

		fmt.Printf("🔄 Creating next schedule at %s\n", time.Unix(nextSchedule.ScheduledTime, 0).Format(time.RFC3339))
		db.DB.Create(&nextSchedule)
	}

	close(jobs)
}

func render(tpl string, user models.User) string {
	s := strings.ReplaceAll(tpl, "{{name}}", user.Name)
	s = strings.ReplaceAll(s, "{{age}}", fmt.Sprintf("%d", user.Age))
	s = strings.ReplaceAll(s, "{{email}}", user.Email)
	s = strings.ReplaceAll(s, "{{gender}}", user.Gender)
	return s
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
