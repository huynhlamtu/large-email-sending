package main

import (
	"go-email-system/db"
	"go-email-system/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Connect()
	r := gin.Default()

	r.POST("/schedules", func(c *gin.Context) {
		var req struct {
			ScheduledTime string `json:"scheduled_time"`
			FilterQuery   string `json:"filter_query"`
			Subject       string `json:"subject"`
			Body          string `json:"body"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		parsedTime, err := time.Parse(time.RFC3339, req.ScheduledTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled_time format (use RFC3339)"})
			return
		}

		schedule := models.EmailSchedule{
			ScheduledTime: parsedTime.Unix(),
			FilterQuery:   req.FilterQuery,
			Subject:       req.Subject,
			Body:          req.Body,
		}

		if err := db.DB.Create(&schedule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Schedule created successfully",
			"schedule_id":  schedule.ID,
			"scheduled_at": time.Unix(schedule.ScheduledTime, 0).Format(time.RFC3339),
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	r.Run(":8080")
}
