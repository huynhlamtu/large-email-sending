# 📧 Large Email Sending System

A scalable email sending system capable of handling **millions of scheduled emails** efficiently using Go, RabbitMQ, and PostgreSQL. Designed with modular components for scheduling, queuing, sending, and logging — all containerized with Docker.

---

## 🚀 Features

- Schedule and send emails to millions of users
- Cron-based batch preparation every 10 minutes
- Email queuing using RabbitMQ with delayed and retry support
- Consumer processes messages and sends emails
- Separate **log writer** to offload database pressure
- Built-in delay, retry, and DLQ (Dead Letter Queue)
- Optimized for batch insert and horizontal scaling

---

## 🛠️ Tech Stack

| Component        | Technology       |
|------------------|------------------|
| Language         | Golang           |
| Queue System     | RabbitMQ         |
| Database         | PostgreSQL       |
| Containerization | Docker Compose   |
| Cron Job         | Go binary        |
| Email Transport  | Simulated SMTP   |

---

## 🧱 System Architecture

```text
[ API / Schedule Email ] --> [ API Service ] --> [ RabbitMQ Queue ]
                                              ↘
                                           [ Email Consumer ]
                                              ↘
                                     [ Log Writer → PostgreSQL ]


- api: Accepts email schedule requests or manual triggers
- cron: Periodically prepares emails due for sending
- consumer: Consumes messages and sends emails via SMTP
- log-writer: Batches and logs email delivery results

---
## ⚙️ Run Locally
# Build & start the entire system
docker-compose up --build -d

# Check running containers
docker-compose ps
