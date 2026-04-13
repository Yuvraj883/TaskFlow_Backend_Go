package handlers

import (
	"fmt"
	"strconv"
	"time"

	// "net/http"

	"taskflow/internal/db"
	"taskflow/internal/utils"

	"github.com/gin-gonic/gin"
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
		c.JSON(400, utils.FormatValidationError(err))
		return
	}

	userID, _ := c.Get("user_id")
	projectID := c.Param("id")

	var userExists bool

	err := db.DB.QueryRow(c.Request.Context(),
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
	err = db.DB.QueryRow(c.Request.Context(),
		`SELECT EXISTS (
			SELECT 1 FROM projects WHERE id=$1 AND owner_id=$2
		)`,
		projectID, userID,
	).Scan(&exists)

	if err != nil || !exists {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	var taskID, updatedAt string

	err = db.DB.QueryRow(c.Request.Context(),
		`INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, due_date, updated_at)
		 VALUES ($1, $2, 'todo', $3, $4, $5, $6, NOW())
		 RETURNING id, updated_at`,
		input.Title,
		input.Description,
		input.Priority,
		projectID,
		input.AssigneeID,
		input.DueDate,
	).Scan(&taskID, &updatedAt)

	if err != nil {
		c.JSON(500, gin.H{"error": "could not create task"})
		return
	}

	c.JSON(201, gin.H{
		"id":         taskID,
		"updated_at": updatedAt,
	})
}

func GetTasks(c *gin.Context) {
	projectID := c.Param("id")
	status := c.Query("status")
	assignee := c.Query("assignee")

	query := `SELECT id, title, status, priority, updated_at FROM tasks WHERE project_id=$1`
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
		i++
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	rows, err := db.DB.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch tasks"})
		return
	}
	defer rows.Close()

	var tasks []gin.H

	for rows.Next() {
		var id, title, status, priority string
		var updatedAt *time.Time
		rows.Scan(&id, &title, &status, &priority, &updatedAt)

		tasks = append(tasks, gin.H{
			"id":         id,
			"title":      title,
			"status":     status,
			"priority":   priority,
			"updated_at": updatedAt,
		})
	}

	c.JSON(200, gin.H{"tasks": tasks})
}

func DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := c.Get("user_id")

	updatedAt := c.Query("updated_at")

	var allowed bool

	err := db.DB.QueryRow(c.Request.Context(),
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

	query := `DELETE FROM tasks WHERE id=$1`
	args := []interface{}{taskID}

	if updatedAt != "" {
		query += ` AND updated_at=$2`
		args = append(args, updatedAt)
	}

	tag, err := db.DB.Exec(c.Request.Context(), query, args...)

	if err != nil {
		c.JSON(500, gin.H{"error": "delete failed"})
		return
	}
	if tag.RowsAffected() == 0 && updatedAt != "" {
		c.JSON(409, gin.H{"error": "conflict: resource was modified"})
		return
	}

	c.Status(204)
}

type UpdateTaskInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority    *string    `json:"priority" binding:"omitempty,oneof=low medium high"`
	AssigneeID  *string    `json:"assignee_id"`
	DueDate     *string    `json:"due_date"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := c.Get("user_id")
	var input UpdateTaskInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, utils.FormatValidationError(err))
		return
	}

	// 🔒 Check authorization
	var allowed bool
	err := db.DB.QueryRow(c.Request.Context(),
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

	query := "UPDATE tasks SET updated_at=NOW(), "
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

	if input.UpdatedAt != nil {
		query += fmt.Sprintf(" WHERE id=$%d AND updated_at=$%d", i, i+1)
		args = append(args, taskID, *input.UpdatedAt)
	} else {
		query += fmt.Sprintf(" WHERE id=$%d", i)
		args = append(args, taskID)
	}

	tag, err := db.DB.Exec(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": "update failed"})
		return
	}
	if tag.RowsAffected() == 0 && input.UpdatedAt != nil {
		c.JSON(409, gin.H{"error": "conflict: resource was modified"})
		return
	}

	var newUpdatedAt *string
	db.DB.QueryRow(c.Request.Context(), `SELECT updated_at FROM tasks WHERE id=$1`, taskID).Scan(&newUpdatedAt)

	c.JSON(200, gin.H{"message": "updated", "updated_at": newUpdatedAt})
}
