package handlers

import (
	// "net/http"
	// "time"

	"taskflow/internal/db"
	"taskflow/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func Register(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, utils.FormatValidationError(err))
		return
	}

	// Check if user already exists
	var exists bool
	err := db.DB.QueryRow(c.Request.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", input.Email).Scan(&exists)
	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}
	if exists {
		c.JSON(400, gin.H{"error": "validation failed", "fields": gin.H{"email": "already exists"}})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 12)

	id := uuid.New()

	_, err = db.DB.Exec(c.Request.Context(),
		"INSERT INTO users (id, name, email, password) VALUES ($1,$2,$3,$4)",
		id, input.Name, input.Email, string(hashed),
	)

	if err != nil {
		c.JSON(400, gin.H{"error": "user creation failed"})
		return
	}

	c.JSON(201, gin.H{"message": "user created"})
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, utils.FormatValidationError(err))
		return
	}

	var id, name, email, hashedPassword string

	err := db.DB.QueryRow(c.Request.Context(),
		"SELECT id, name, email, password FROM users WHERE email=$1",
		input.Email,
	).Scan(&id, &name, &email, &hashedPassword)

	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	// compare password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password))
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	// generate token
	token, err := utils.GenerateToken(id, email)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":    id,
			"name":  name,
			"email": email,
		},
	})
}
