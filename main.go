package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"zoom-backend/database"
	"zoom-backend/handlers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	router := gin.Default()

	// Enable CORS di router yang sama
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	database.InitDB()

	// Daftarkan route di router yang sama
	router.GET("/", handlers.Health)

	router.GET("/meetings", handlers.GetAllMeetings)
	router.GET("/meetings/:id", handlers.GetMeetingByID)
	router.POST("/meetings", handlers.CreateMeeting)
	router.PUT("/meetings/:id", handlers.UpdateMeeting)
	router.DELETE("/meetings/:id", handlers.DeleteMeeting)

	// Jalankan server pakai router yang sudah dipasang middleware & route
	router.Run(":8080")
}
