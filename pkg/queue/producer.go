package queue

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	mqConn       *amqp.Connection
	mqChan       *amqp.Channel
	queueName    = "email_send_queue"
	exchangeName = "email_delayed_exchange"
	dlqName      = "email_send_dlq"
)

type EmailJob struct {
	LogID      uuid.UUID
	Email      string
	Subject    string
	Body       string
	DelayMs    int64
	RetryCount int
}

func InitRabbitMQ() {
	var err error
	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "rabbitmq"
	}
	dsn := "amqp://guest:guest@" + host + ":5672/"

	// Retry connect
	for i := 1; i <= 10; i++ {
		log.Printf("➡️ Kết nối đến %s (lần %d)...", dsn, i)
		mqConn, err = amqp.Dial(dsn)
		if err == nil {
			break
		}
		log.Printf("⚠️ Kết nối RabbitMQ thất bại (thử %d): %v\n", i, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("❌ Không thể kết nối RabbitMQ sau nhiều lần thử:", err)
	}

	mqChan, err = mqConn.Channel()
	if err != nil {
		log.Fatal("Không tạo được channel:", err)
	}

	// Hạn chế prefetch quá nhiều
	mqChan.Qos(100, 0, false)

	// Declare DLQ
	_, err = mqChan.QueueDeclare(dlqName, true, false, false, false, nil)
	if err != nil {
		log.Fatal("Không tạo được DLQ:", err)
	}

	// Declare exchange x-delayed
	err = mqChan.ExchangeDeclare(
		exchangeName,
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-delayed-type": "direct",
		},
	)
	if err != nil {
		log.Fatal("Không tạo được delayed exchange:", err)
	}

	// Declare main queue và liên kết với DLQ
	_, err = mqChan.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "", // default exchange
			"x-dead-letter-routing-key": dlqName,
		},
	)
	if err != nil {
		log.Fatal("Không tạo được queue chính:", err)
	}

	// Bind queue vào exchange
	err = mqChan.QueueBind(queueName, "send_email", exchangeName, false, nil)
	if err != nil {
		log.Fatal("Không bind queue với exchange:", err)
	}
}

func Publish(job EmailJob) {
	if mqChan == nil {
		log.Println("❌ mqChan is nil: Did you call InitRabbitMQ()?")
		return
	}

	data, _ := json.Marshal(job)
	headers := amqp.Table{"x-delay": job.DelayMs}

	err := mqChan.Publish(
		exchangeName,
		"send_email",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Headers:      headers,
			Body:         data,
		},
	)
	if err != nil {
		log.Println("Lỗi gửi vào RabbitMQ:", err)
	}
}
