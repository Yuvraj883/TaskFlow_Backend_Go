package handlers

import (
	"context"
	"strings"
	// "net/http"

	"taskflow/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CreateProjectInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func CreateProject(c *gin.Context) {
	var input CreateProjectInput

	// bind + validate
	if err := c.ShouldBindJSON(&input); err != nil {
		fields := gin.H{}

		// extract validation errors
		if errs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range errs {
				field := strings.ToLower(e.Field())
				fields[field] = "is required"
			}
		}

		c.JSON(400, gin.H{
			"error":  "validation failed",
			"fields": fields,
		})
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
func GetProjects(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rows, err := db.DB.Query(context.Background(),
		`SELECT DISTINCT p.id, p.name, p.description
		 FROM projects p
		 LEFT JOIN tasks t ON p.id = t.project_id
		 WHERE p.owner_id = $1 OR t.assignee_id = $1`,
		userID,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch projects"})
		return
	}
	defer rows.Close()

	var projects []gin.H

	for rows.Next() {
		var id, name string
		var description *string

		rows.Scan(&id, &name, &description)

		projects = append(projects, gin.H{
			"id":          id,
			"name":        name,
			"description": description,
		})
	}

	c.JSON(200, gin.H{"projects": projects})
}
