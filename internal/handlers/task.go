package handlers

import (
	"context"
	"fmt"

	// "net/http"

	"taskflow/internal/db"

	"github.com/gin-gonic/gin"
	// "github.com/go-playground/validator/v10"
)

type CreateTaskInput struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Priority    string  `json:"priority" binding:"required,oneof=low medium high"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func CreateTask(c *gin.Context) {
	var input CreateTaskInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "validation failed"})
		return
	}

	userID, _ := c.Get("user_id")
	projectID := c.Param("id")

	var userExists bool

	err := db.DB.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`,
		input.AssigneeID,
	).Scan(&userExists)

	if input.AssigneeID != nil && !userExists {
		c.JSON(400, gin.H{
			"error": "validation failed",
			"fields": gin.H{
				"assignee_id": "invalid user",
			},
		})
		return
	}

	// 🔒 Check if user owns project
	var exists bool
	err = db.DB.QueryRow(context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM projects WHERE id=$1 AND owner_id=$2
		)`,
		projectID, userID,
	).Scan(&exists)

	if err != nil || !exists {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	var taskID string

	err = db.DB.QueryRow(context.Background(),
		`INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, due_date)
		 VALUES ($1, $2, 'todo', $3, $4, $5, $6)
		 RETURNING id`,
		input.Title,
		input.Description,
		input.Priority,
		projectID,
		input.AssigneeID,
		input.DueDate,
	).Scan(&taskID)

	if err != nil {
		c.JSON(500, gin.H{"error": "could not create task"})
		return
	}

	c.JSON(201, gin.H{
		"id": taskID,
	})
}

func GetTasks(c *gin.Context) {
	projectID := c.Param("id")
	status := c.Query("status")
	assignee := c.Query("assignee")

	query := `SELECT id, title, status, priority FROM tasks WHERE project_id=$1`
	args := []interface{}{projectID}
	i := 2

	if status != "" {
		query += " AND status=$" + fmt.Sprint(i)
		args = append(args, status)
		i++
	}

	if assignee != "" {
		query += " AND assignee_id=$" + fmt.Sprint(i)
		args = append(args, assignee)
	}

	rows, err := db.DB.Query(context.Background(), query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch tasks"})
		return
	}
	defer rows.Close()

	var tasks []gin.H

	for rows.Next() {
		var id, title, status, priority string
		rows.Scan(&id, &title, &status, &priority)

		tasks = append(tasks, gin.H{
			"id":       id,
			"title":    title,
			"status":   status,
			"priority": priority,
		})
	}

	c.JSON(200, gin.H{"tasks": tasks})
}

func DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := c.Get("user_id")

	var allowed bool

	err := db.DB.QueryRow(context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM tasks t
			JOIN projects p ON t.project_id = p.id
			WHERE t.id=$1 AND (p.owner_id=$2)
		)`,
		taskID, userID,
	).Scan(&allowed)

	if err != nil || !allowed {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	_, err = db.DB.Exec(context.Background(),
		`DELETE FROM tasks WHERE id=$1`, taskID)

	if err != nil {
		c.JSON(500, gin.H{"error": "delete failed"})
		return
	}

	c.Status(204)
}

type UpdateTaskInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status" binding:"omitempty,oneof=todo in_progress completed"`
	Priority    *string `json:"priority" binding:"omitempty,oneof=low medium high"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := c.Get("user_id")
	var input UpdateTaskInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "validation failed"})
		return
	}

	// 🔒 Check authorization
	var allowed bool
	err := db.DB.QueryRow(context.Background(),
		`SELECT EXISTS (
            SELECT 1 FROM tasks t
            JOIN projects p ON t.project_id = p.id
            WHERE t.id=$1 AND p.owner_id=$2
        )`,
		taskID, userID,
	).Scan(&allowed)

	if err != nil || !allowed {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	query := "UPDATE tasks SET "
	args := []interface{}{}
	i := 1

	if input.Title != nil {
		query += fmt.Sprintf("title=$%d,", i)
		args = append(args, *input.Title)
		i++
	}
	if input.Description != nil {
		query += fmt.Sprintf("description=$%d,", i)
		args = append(args, *input.Description)
		i++
	}
	if input.Status != nil {
		query += fmt.Sprintf("status=$%d,", i)
		args = append(args, *input.Status)
		i++
	}
	if input.Priority != nil {
		query += fmt.Sprintf("priority=$%d,", i)
		args = append(args, *input.Priority)
		i++
	}
	if input.AssigneeID != nil {
		query += fmt.Sprintf("assignee_id=$%d,", i)
		args = append(args, *input.AssigneeID)
		i++
	}
	if input.DueDate != nil {
		query += fmt.Sprintf("due_date=$%d,", i)
		args = append(args, *input.DueDate)
		i++
	}

	if len(args) == 0 {
		c.JSON(400, gin.H{"error": "no fields to update"})
		return
	}

	query = query[:len(query)-1] // remove trailing comma
	query += fmt.Sprintf(" WHERE id=$%d", i)
	args = append(args, taskID)

	_, err = db.DB.Exec(context.Background(), query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": "update failed"})
		return
	}

	c.JSON(200, gin.H{"message": "updated"})
}
