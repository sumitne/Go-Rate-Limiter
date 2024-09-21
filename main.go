package main

import (
	"flag"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	H "github.com/sumitne/api/app/handler"
	mid "github.com/sumitne/api/app/middleware"
)

func main() {
	algo := flag.String("algo", mid.FIXED_WINDOW_COUNTER, "Rate limiting algorithm input")
	flag.Parse()

	r := gin.Default()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis address
		Password: "",               // no password set
		DB:       0,                // use default DB
	})

	// Apply the rate limiter middleware
	r.Use(mid.RateLimiterMiddleware(*algo, rdb))

	r.Handle("GET", "/users", H.GetUser)
	r.Run(":8080")
}
