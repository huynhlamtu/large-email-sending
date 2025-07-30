package db

import (
	"fmt"
	"go-email-system/models"
	"math/rand"
)

func SeedUsers(n int) {
	Connect()
	for i := 0; i < n; i++ {
		user := models.User{
			Email:  fmt.Sprintf("user%06d@example.com", i),
			Name:   fmt.Sprintf("User %d", i),
			Age:    rand.Intn(60) + 18,
			Gender: []string{"male", "female"}[rand.Intn(2)],
		}
		DB.Create(&user)
	}
}
