package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"time"
)

var ctx = context.Background()

const (
	FIXED_WINDOW_COUNTER   = "fixed_window_counter"
	SLIDING_WINDOW_LOG     = "sliding_window_log"
	TOKEN_BUCKET           = "token_bucket"
	SLIDING_WINDOW_COUNTER = "sliding_window_counter"
	LEAKY_BUCKET           = "leaky_bucket"
)

// RateLimiterMiddleware to limit requests
func RateLimiterMiddleware(Algo string, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "user:" + c.ClientIP() // Rate-limit per client IP
		limit := 10                   // Max 10 requests
		window := 60 * time.Second    // Time window of 60 seconds
		subWindow := 10 * time.Second // Sub-window for sliding window counter
		ratePerSec := 2.0             // 2 requests per second for leaky/token bucket

		var allowed bool
		var err error

		switch Algo {
		case FIXED_WINDOW_COUNTER:
			allowed, err = RateLimitFixedWindowCounter(key, limit, window, rdb)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				c.Abort()
				return
			}
		case SLIDING_WINDOW_LOG:
			allowed, err = RateLimitFixedWindowCounter(key, limit, window, rdb)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				c.Abort()
				return
			}
		case TOKEN_BUCKET:
			allowed, err = RateLimitTokenBucket(key, limit, ratePerSec, rdb)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				c.Abort()
				return
			}
		case SLIDING_WINDOW_COUNTER:
			allowed, err = RateLimitSlidingWindowCounter(key, limit, window, subWindow, rdb)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				c.Abort()
				return
			}
		case LEAKY_BUCKET:
			allowed, err = RateLimitLeakyBucket(key, limit, ratePerSec, rdb)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				c.Abort()
				return
			}
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Undefined Rate Limit Algo"})

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
func RateLimitFixedWindowCounter(key string, limit int, window time.Duration, rdb *redis.Client) (bool, error) {

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

// Sliding window log
func RateLimitSlidingWindowLog(key string, limit int, window time.Duration, rdb *redis.Client) (bool, error) {

	now := time.Now().UnixNano()
	windowStart := now - int64(window)

	// Use sorted set to store request timestamps
	key = "sliding-window-log:" + key

	// Remove timestamps outside of the current window
	err := rdb.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10)).Err()
	if err != nil {
		return false, err
	}

	// Get current count of requests within the window
	count, err := rdb.ZCard(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Check if limit has been reached
	if count >= int64(limit) {
		return false, nil
	}

	// Add current request timestamp to the sorted set
	err = rdb.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: now,
	}).Err()
	if err != nil {
		return false, err
	}

	// Set expiry to avoid keeping stale data
	err = rdb.Expire(ctx, key, window).Err()
	if err != nil {
		return false, err
	}

	return true, nil
}

// token bucket
func RateLimitTokenBucket(key string, limit int, ratePerSec float64, rdb *redis.Client) (bool, error) {

	now := float64(time.Now().UnixNano()) / 1e9

	// Get token data from Redis
	result, err := rdb.HMGet(ctx, key, "tokens", "lastTime").Result()
	if err != nil {
		return false, err
	}

	var tokens float64
	var lastTime float64
	if result[0] != nil {
		tokens, _ = strconv.ParseFloat(result[0].(string), 64)
	}
	if result[1] != nil {
		lastTime, _ = strconv.ParseFloat(result[1].(string), 64)
	}

	// Add tokens based on time elapsed
	elapsed := now - lastTime
	tokens += elapsed * ratePerSec

	// Cap tokens at the limit
	if tokens > float64(limit) {
		tokens = float64(limit)
	}

	// If there are no tokens, reject the request
	if tokens < 1 {
		return false, nil
	}

	// Use one token
	tokens -= 1

	// Update Redis with the new token count and timestamp
	_, err = rdb.HMSet(ctx, key, "tokens", tokens, "lastTime", now).Result()
	if err != nil {
		return false, err
	}

	return true, nil
}

// sliding window counter
func RateLimitSlidingWindowCounter(key string, limit int, window, subWindow time.Duration, rdb *redis.Client) (bool, error) {

	now := time.Now().UnixNano()
	currentSubWindow := now / int64(subWindow)

	// Generate keys for current and previous sub-windows
	currKey := fmt.Sprintf("sliding-window-counter:%s:%d", key, currentSubWindow)
	prevKey := fmt.Sprintf("sliding-window-counter:%s:%d", key, currentSubWindow-1)

	// Increment count for the current sub-window
	currCount, err := rdb.Incr(ctx, currKey).Result()
	if err != nil {
		return false, err
	}

	// Set expiration for the current sub-window key
	err = rdb.Expire(ctx, currKey, subWindow*2).Err()
	if err != nil {
		return false, err
	}

	// Get count for the previous sub-window
	prevCount, err := rdb.Get(ctx, prevKey).Int64()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// Calculate total count by interpolating between current and previous windows
	timeIntoSubWindow := float64(now%int64(subWindow)) / float64(subWindow)
	totalCount := float64(prevCount)*(1-timeIntoSubWindow) + float64(currCount)

	// Check if limit exceeded
	if totalCount > float64(limit) {
		return false, nil
	}

	return true, nil
}

// leaky bucket
func RateLimitLeakyBucket(key string, limit int, ratePerSec float64, rdb *redis.Client) (bool, error) {

	now := float64(time.Now().UnixNano()) / 1e9

	// Get bucket data: tokens (available) and last checked time
	result, err := rdb.HMGet(ctx, key, "tokens", "lastTime").Result()
	if err != nil {
		return false, err
	}

	var tokens float64
	var lastTime float64
	if result[0] != nil {
		tokens, _ = strconv.ParseFloat(result[0].(string), 64)
	}
	if result[1] != nil {
		lastTime, _ = strconv.ParseFloat(result[1].(string), 64)
	}

	// Calculate new tokens based on elapsed time
	elapsed := now - lastTime
	tokens += elapsed * ratePerSec

	// Cap tokens at the bucket limit
	if tokens > float64(limit) {
		tokens = float64(limit)
	}

	// Request token
	if tokens < 1 {
		return false, nil
	}

	tokens -= 1

	// Update Redis with new token count and timestamp
	_, err = rdb.HMSet(ctx, key, "tokens", tokens, "lastTime", now).Result()
	if err != nil {
		return false, err
	}

	return true, nil
}
