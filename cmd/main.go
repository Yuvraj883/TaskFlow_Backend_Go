package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskflow/internal/db"
	"taskflow/internal/handlers"
	"taskflow/internal/middleware"
	"taskflow/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	utils.InitLogger()
	err := godotenv.Load(".env")
	if err != nil {
		godotenv.Load("../.env") // fallback
		slog.Warn("Could not load primary .env, trying fallback")
	}
	db.Connect()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.CORSMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// r.POST("/auth/register", handlers.Register)

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Use(middleware.RateLimiter(5, 10*time.Second))
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/projects", handlers.CreateProject)
	protected.GET("/projects", handlers.GetProjects)
	protected.GET("/projects/:id", handlers.GetProject)
	protected.GET("/projects/:id/stats", handlers.GetProjectStats)
	protected.PATCH("/projects/:id", handlers.UpdateProject)
	protected.DELETE("/projects/:id", handlers.DeleteProject)
	
	protected.POST("/projects/:id/tasks", handlers.CreateTask)
	protected.GET("/projects/:id/tasks", handlers.GetTasks)
	protected.DELETE("/tasks/:id", handlers.DeleteTask)
	protected.PATCH("/tasks/:id", handlers.UpdateTask)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("Starting server", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ListenAndServe failed", slog.Any("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	// gracefully turn off on interrupt (CTRL+C) or SIGTERM (Docker)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
	}

	slog.Info("Server exiting")
}
