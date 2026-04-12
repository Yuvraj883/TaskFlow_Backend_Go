package handlers

import (
	"context"
	// "net/http"

	"taskflow/internal/db"

	"github.com/gin-gonic/gin"
)

type CreateProjectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateProject(c *gin.Context) {
	var input CreateProjectInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "validation failed"})
		return
	}

	userID, _ := c.Get("user_id")

	var id string

	err := db.DB.QueryRow(context.Background(),
		`INSERT INTO projects (name, description, owner_id)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		input.Name, input.Description, userID,
	).Scan(&id)

	if err != nil {
		c.JSON(500, gin.H{"error": "could not create project"})
		return
	}

	c.JSON(201, gin.H{
		"id":          id,
		"name":        input.Name,
		"description": input.Description,
	})
}
