package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// FormatValidationError formats the validator.ValidationErrors into the specific JSON format expected
func FormatValidationError(err error) gin.H {
	fields := gin.H{}
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := strings.ToLower(e.Field())
			message := "is invalid"
			switch e.Tag() {
			case "required":
				message = "is required"
			case "email":
				message = "must be a valid email"
			case "min":
				message = "must be at least " + e.Param() + " characters"
			case "oneof":
				message = "must be one of: " + e.Param()
			}
			fields[field] = message
		}
	} else {
		// Non validation error (e.g. JSON syntax error)
		fields["payload"] = "malformed JSON or wrong type"
	}

	return gin.H{
		"error":  "validation failed",
		"fields": fields,
	}
}
