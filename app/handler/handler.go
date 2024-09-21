package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetUser(c *gin.Context) {
	// dummy api controller
	c.Status(http.StatusOK)
}
