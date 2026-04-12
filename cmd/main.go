package main

import (
	"log"
	"os"

	"taskflow/internal/db"
	"taskflow/internal/handlers"
	authMiddleware "taskflow/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db.Connect()
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// r.POST("/auth/register", handlers.Register)

	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}
	protected := r.Group("/")
	protected.Use(authMiddleware.AuthMiddleware())
	protected.POST("/projects", handlers.CreateProject)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(r.Run(":" + port))
}
