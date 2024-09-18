package main

import (
	"github.com/gin-gonic/gin"
	H "github.com/sumitne/api/app/handler"
	mid "github.com/sumitne/api/app/middleware"
)

func main() {
	r := gin.Default()

	// Apply the rate limiter middleware
	r.Use(mid.RateLimiterMiddleware())

	r.Handle("GET", "/users", H.GetUser)
   r.Run(":8080")
}

