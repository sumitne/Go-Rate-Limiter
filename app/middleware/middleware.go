package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

var ctx = context.Background()

// RateLimiterMiddleware to limit requests
func RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "user:" + c.ClientIP() // Rate-limit per client IP
		limit := 5                    // Limit to 5 requests
		window := time.Minute

		allowed, err := RateLimit(key, limit, window)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}
// Fixed Window Counter

// RateLimit function as defined earlier
func RateLimit(key string, limit int, window time.Duration) (bool, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis address
		Password: "",               // no password set
		DB:       0,                // use default DB
	})

	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		err = rdb.Expire(ctx, key, window).Err()
		if err != nil {
			return false, err
		}
	}

	if count > int64(limit) {
		return false, nil
	}

	return true, nil
}
