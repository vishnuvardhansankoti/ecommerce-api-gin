package handlers

import (
	"net/http"
	"strings"

	"ecommerce-api-gin/internal/services"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type badRequestError struct {
	message string
}

func (e badRequestError) Error() string {
	return e.message
}

func errBadRequest(message string) error {
	return badRequestError{message: strings.TrimSpace(message)}
}

func isBadRequestError(err error) bool {
	_, ok := err.(badRequestError)
	return ok
}

func writeError(c *gin.Context, err error, defaultStatus int) {
	switch {
	case services.IsValidation(err), isBadRequestError(err):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case services.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(defaultStatus, gin.H{"error": err.Error()})
	}
}
