package main

import (
	"fmt"
	"go-email-system/db"
	"go-email-system/models"
	"math/rand"
	"time"
)

func main() {
	db.Connect()

	var logs []models.EmailLog

	// Lấy các email đang pending
	if err := db.DB.Where("status = ?", "pending").Find(&logs).Error; err != nil {
		fmt.Println("Không thể lấy email_logs:", err)
		return
	}

	fmt.Printf("🔔 Bắt đầu gửi %d email...\n", len(logs))

	for _, log := range logs {
		// Lấy thông tin user để gửi
		var user models.User
		if err := db.DB.First(&user, "id = ?", log.UserID).Error; err != nil {
			fmt.Println("Không tìm thấy user:", log.UserID)
			continue
		}

		// Giả lập gửi email
		success, errMsg := sendFakeEmail(user.Email, log.SubjectRendered, log.BodyRendered)

		if success {
			log.Status = "success"
			log.SentAt = time.Now()
			log.Error = ""
		} else {
			log.Status = "fail"
			log.SentAt = time.Now()
			log.Error = errMsg
		}

		db.DB.Save(&log)
		fmt.Printf("→ Gửi đến %-25s [%s]\n", user.Email, log.Status)
	}
}

func sendFakeEmail(to, subject, body string) (bool, string) {
	// 90% thành công
	if rand.Float64() < 0.9 {
		return true, ""
	}
	return false, "Giả lập lỗi SMTP: hộp thư không tồn tại"
}
