package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"taskflow/internal/db"
	"taskflow/internal/utils"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type CreateProjectInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func CreateProject(c *gin.Context) {
	var input CreateProjectInput

	// bind + validate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, utils.FormatValidationError(err))
		return
	}

	userID, _ := c.Get("user_id")

	var id string
	var updatedAt time.Time

	err := db.DB.QueryRow(c.Request.Context(),
		`INSERT INTO projects (name, description, owner_id, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 RETURNING id, updated_at`,
		input.Name, input.Description, userID,
	).Scan(&id, &updatedAt)

	if err != nil {
		c.JSON(500, gin.H{"error": "could not create project"})
		return
	}

	c.JSON(201, gin.H{
		"id":          id,
		"name":        input.Name,
		"description": input.Description,
		"updated_at":  updatedAt,
	})
}
func GetProjects(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	rows, err := db.DB.Query(c.Request.Context(),
		`SELECT DISTINCT p.id, p.name, p.description, p.updated_at
		 FROM projects p
		 LEFT JOIN tasks t ON p.id = t.project_id
		 WHERE p.owner_id = $1 OR t.assignee_id = $1
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
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
		var updatedAt *time.Time

		rows.Scan(&id, &name, &description, &updatedAt)

		projects = append(projects, gin.H{
			"id":          id,
			"name":        name,
			"description": description,
			"updated_at":  updatedAt,
		})
	}

	c.JSON(200, gin.H{"projects": projects})
}

func GetProject(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var project map[string]interface{} = make(map[string]interface{})
	var pid, name, ownerID string
	var description *string
	var updatedAt time.Time

	err := db.DB.QueryRow(c.Request.Context(),
		`SELECT id, name, description, owner_id, updated_at FROM projects WHERE id = $1`, id).
		Scan(&pid, &name, &description, &ownerID, &updatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}

	if ownerID != userID {
		var assigned bool
		err = db.DB.QueryRow(
			c.Request.Context(),
			`SELECT EXISTS (SELECT 1 FROM tasks WHERE project_id=$1 AND assignee_id=$2)`,
			id,
			userID,
		).Scan(&assigned)
		if err != nil {
			c.JSON(500, gin.H{"error": "database error"})
			return
		}
		if !assigned {
			c.JSON(403, gin.H{"error": "forbidden"})
			return
		}
	}

	project["id"] = pid
	project["name"] = name
	project["description"] = description
	project["owner_id"] = ownerID
	project["updated_at"] = updatedAt

	rows, err := db.DB.Query(c.Request.Context(),
		`SELECT id, title, status, priority, assignee_id, due_date, created_at, updated_at FROM tasks WHERE project_id = $1`, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var tasks []gin.H
	for rows.Next() {
		var t_id, t_title, t_status, t_priority string
		var t_updated_at *time.Time
		var t_assignee_id, t_due_date, t_created_at *string
		rows.Scan(&t_id, &t_title, &t_status, &t_priority, &t_assignee_id, &t_due_date, &t_created_at, &t_updated_at)
		tasks = append(tasks, gin.H{
			"id":          t_id,
			"title":       t_title,
			"status":      t_status,
			"priority":    t_priority,
			"assignee_id": t_assignee_id,
			"due_date":    t_due_date,
			"created_at":  t_created_at,
			"updated_at":  t_updated_at,
		})
	}
	if tasks == nil {
		tasks = []gin.H{}
	}
	project["tasks"] = tasks

	c.JSON(200, project)
}

func GetProjectStats(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var projectExists bool
	err := db.DB.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1)`, id).Scan(&projectExists)
	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}
	if !projectExists {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	var allowed bool
	err = db.DB.QueryRow(c.Request.Context(),
		`SELECT EXISTS (
			SELECT 1 FROM projects p
			LEFT JOIN tasks t ON p.id = t.project_id
			WHERE p.id=$1 AND (p.owner_id=$2 OR t.assignee_id=$2)
		)`, id, userID,
	).Scan(&allowed)
	if err != nil || !allowed {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	byStatus := make(map[string]int)
	byAssignee := make(map[string]int)
	total := 0

	rows, err := db.DB.Query(c.Request.Context(), `SELECT status, COUNT(*) FROM tasks WHERE project_id=$1 GROUP BY status`, id)
	if err == nil {
		for rows.Next() {
			var status string
			var count int
			rows.Scan(&status, &count)
			byStatus[status] = count
			total += count
		}
		rows.Close()
	}

	rowsAss, err := db.DB.Query(c.Request.Context(), `SELECT assignee_id, COUNT(*) FROM tasks WHERE project_id=$1 GROUP BY assignee_id`, id)
	if err == nil {
		for rowsAss.Next() {
			var assignee *string
			var count int
			rowsAss.Scan(&assignee, &count)
			if assignee == nil {
				byAssignee["unassigned"] = count
			} else {
				byAssignee[*assignee] = count
			}
		}
		rowsAss.Close()
	}

	c.JSON(200, gin.H{
		"total":       total,
		"by_status":   byStatus,
		"by_assignee": byAssignee,
	})
}

type UpdateProjectInput struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func UpdateProject(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var projectExists bool
	err := db.DB.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1)`, id).Scan(&projectExists)
	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}
	if !projectExists {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	var exists bool
	err = db.DB.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1 AND owner_id=$2)`, id, userID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	var input UpdateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, utils.FormatValidationError(err))
		return
	}

	query := "UPDATE projects SET updated_at = NOW()"
	args := []interface{}{}
	i := 1

	if input.Name != nil {
		query += fmt.Sprintf(", name=$%d", i)
		args = append(args, *input.Name)
		i++
	}
	if input.Description != nil {
		query += fmt.Sprintf(", description=$%d", i)
		args = append(args, *input.Description)
		i++
	}

	if input.UpdatedAt != nil {
		query += fmt.Sprintf(" WHERE id=$%d AND updated_at=$%d", i, i+1)
		args = append(args, id, *input.UpdatedAt)
	} else {
		query += fmt.Sprintf(" WHERE id=$%d", i)
		args = append(args, id)
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

	var updatedName string
	var updatedDesc, newUpdatedAt *string
	db.DB.QueryRow(c.Request.Context(), `SELECT name, description, updated_at FROM projects WHERE id=$1`, id).Scan(&updatedName, &updatedDesc, &newUpdatedAt)

	c.JSON(200, gin.H{
		"id":          id,
		"name":        updatedName,
		"description": updatedDesc,
		"owner_id":    userID,
		"updated_at":  newUpdatedAt,
	})
}

func DeleteProject(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	updatedAt := c.Query("updated_at")

	var projectExists bool
	err := db.DB.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1)`, id).Scan(&projectExists)
	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}
	if !projectExists {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	var exists bool
	err = db.DB.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1 AND owner_id=$2)`, id, userID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	query := `DELETE FROM projects WHERE id=$1`
	args := []interface{}{id}

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
