package test

import (
	"context"
	"github.com/go-redis/redis/v8"
	mid "github.com/sumitne/api/app/middleware"
	"testing"
	"time"
)

var ctx = context.Background()

// Create a Redis client for testing purposes
func createTestRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

// Helper function to reset a key between tests
func resetKey(t *testing.T, rdb *redis.Client, key string) {
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		t.Fatalf("Could not reset key: %v", err)
	}
}

func TestRateLimitFixedWindowCounter(t *testing.T) {
	rdb := createTestRedisClient()
	key := "test_fixed_window_counter"
	limit := 5
	window := 10 * time.Second

	// Reset the Redis key before starting the test
	resetKey(t, rdb, key)

	for i := 1; i <= 6; i++ {
		allowed, err := mid.RateLimitFixedWindowCounter(key, limit, window, rdb)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if i <= limit && !allowed {
			t.Errorf("Expected request %d to be allowed, but it was rate-limited", i)
		} else if i > limit && allowed {
			t.Errorf("Expected request %d to be rate-limited, but it was allowed", i)
		}
	}
}

func TestRateLimitSlidingWindowLog(t *testing.T) {
	rdb := createTestRedisClient()
	key := "test_sliding_window_log"
	limit := 5
	window := 10 * time.Second

	// Reset the Redis key before starting the test
	resetKey(t, rdb, key)

	for i := 1; i <= 6; i++ {
		allowed, err := mid.RateLimitSlidingWindowLog(key, limit, window, rdb)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if i <= limit && !allowed {
			t.Errorf("Expected request %d to be allowed, but it was rate-limited", i)
		} else if i > limit && allowed {
			t.Errorf("Expected request %d to be rate-limited, but it was allowed", i)
		}
		time.Sleep(1 * time.Second)
	}
}
func TestRateLimitSlidingWindowCounter(t *testing.T) {
    rdb := createTestRedisClient()
    defer rdb.Close()

    key := "test_sliding_window_counter"
    limit := 5
    window := 10 * time.Second
    subWindow := 2 * time.Second

    // Reset the Redis key before starting the test
    resetKey(t, rdb, key)

    for i := 1; i <= 6; i++ {
        allowed, err := mid.RateLimitSlidingWindowCounter(key, limit, window, subWindow, rdb)
        if err != nil {
            t.Fatalf("Unexpected error on request %d: %v", i, err)
        }

        t.Logf("Request %d: allowed = %v", i, allowed)

        if i <= limit {
            if !allowed {
                t.Errorf("Expected request %d to be allowed, but it was rate-limited", i)
            }
        } else {
            if allowed {
                t.Errorf("Expected request %d to be rate-limited, but it was allowed", i)
            }
        }

        // Add a small delay between requests
        time.Sleep(100 * time.Millisecond)
    }
}
func TestRateLimitLeakyBucket(t *testing.T) {
    rdb := createTestRedisClient()
    defer rdb.Close()

    key := "test_leaky_bucket"
    limit := 5
    ratePerSec := 1.0

    // Reset the Redis key before starting the test
    resetKey(t, rdb, key)

    for i := 1; i <= 6; i++ {
        allowed, err := mid.RateLimitLeakyBucket(key, limit, ratePerSec, rdb)
        if err != nil {
            t.Fatalf("Unexpected error on request %d: %v", i, err)
        }

        t.Logf("Request %d: allowed = %v", i, allowed)

        if i <= limit {
            if !allowed {
                t.Errorf("Expected request %d to be allowed, but it was rate-limited", i)
            }
        } else {
            if allowed {
                t.Errorf("Expected request %d to be rate-limited, but it was allowed", i)
            }
        }

        // Slow down requests to simulate real-world scenario
        time.Sleep(500 * time.Millisecond)
    }

    // Additional check: verify the current count in Redis
    count, err := rdb.Get(context.Background(), key).Float64()
    if err != nil {
        t.Fatalf("Failed to get count from Redis: %v", err)
    }
    t.Logf("Final count in Redis: %f", count)
}

// func TestRateLimitLeakyBucket(t *testing.T) {
// 	rdb := createTestRedisClient()
// 	key := "test_leaky_bucket"
// 	limit := 5
// 	ratePerSec := 1.0

// 	// Reset the Redis key before starting the test
// 	resetKey(t, rdb, key)

// 	for i := 1; i <= 6; i++ {
// 		allowed, err := mid.RateLimitLeakyBucket(key, limit, ratePerSec, rdb)
// 		if err != nil {
// 			t.Fatalf("Unexpected error: %v", err)
// 		}
// 		if i <= limit && !allowed {
// 			t.Errorf("Expected request %d to be allowed, but it was rate-limited", i)
// 		} else if i > limit && allowed {
// 			t.Errorf("Expected request %d to be rate-limited, but it was allowed", i)
// 		}
// 		time.Sleep(500 * time.Millisecond) // Slow down requests
// 	}
// }

func TestRateLimitTokenBucket(t *testing.T) {
	rdb := createTestRedisClient()
	key := "test_token_bucket"
	limit := 5
	ratePerSec := 1.0

	// Reset the Redis key before starting the test
	resetKey(t, rdb, key)

	for i := 1; i <= 6; i++ {
		allowed, err := mid.RateLimitTokenBucket(key, limit, ratePerSec, rdb)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if i <= limit && !allowed {
			t.Errorf("Expected request %d to be allowed, but it was rate-limited", i)
		} else if i > limit && allowed {
			t.Errorf("Expected request %d to be rate-limited, but it was allowed", i)
		}
		time.Sleep(500 * time.Millisecond) // Slow down requests
	}
}
