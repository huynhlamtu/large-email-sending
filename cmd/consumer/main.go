package main

import (
	"encoding/json"
	"fmt"
	"go-email-system/db"
	"go-email-system/models"
	"go-email-system/pkg/queue"
	"log"
	"math/rand"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go" // ✅ đổi sang amqp091-go
)

const workerCount = 50

func main() {
	db.Connect()
	queue.InitRabbitMQ()

	msgs, err := queue.Consume()
	if err != nil {
		log.Fatal(err)
	}

	jobs := make(chan amqp.Delivery, 10000) // ✅ dùng đúng kiểu amqp091.Delivery
	var wg sync.WaitGroup

	// Worker pool
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for msg := range jobs {
				handleMessage(msg, workerID)
			}
		}(i)
	}

	for msg := range msgs {
		jobs <- msg
	}

	close(jobs)
	wg.Wait()
}

func handleMessage(msg amqp.Delivery, workerID int) {
	var job queue.EmailJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Printf("❌ Worker %d: lỗi parse job: %v\n", workerID, err)
		msg.Ack(false)
		return
	}

	success, errMsg := sendFakeEmail(job.Email, job.Subject, job.Body)

	var logEntry models.EmailLog
	db.DB.First(&logEntry, "id = ?", job.LogID)

	logEntry.SentAt = time.Now()

	if success {
		logEntry.Status = "success"
		logEntry.Error = ""
		db.DB.Save(&logEntry)

		fmt.Printf("[Worker %d] ✅ Sent to %s\n", workerID, job.Email)
		msg.Ack(false)
		return
	}

	// Gửi thất bại
	logEntry.Status = "fail"
	logEntry.Error = errMsg
	db.DB.Save(&logEntry)

	if job.RetryCount >= 3 {
		log.Printf("❌ Worker %d: Gửi thất bại quá 3 lần, đẩy vào DLQ → %s\n", workerID, job.Email)
		msg.Nack(false, false) // Gửi vào DLQ
		return
	}

	// Retry lại
	job.RetryCount++
	job.DelayMs = int64(5 * time.Minute.Milliseconds())
	queue.Publish(job)

	log.Printf("🔁 Worker %d: Retry %d → %s\n", workerID, job.RetryCount, job.Email)
	msg.Ack(false)
}

func sendFakeEmail(to, subject, body string) (bool, string) {
	if rand.Float64() < 0.9 {
		return true, ""
	}
	return false, "SMTP giả lập: gửi thất bại"
}
