
Copy code
# Rate Limiter

A flexible and efficient rate limiting implementation in Go, supporting multiple algorithms including Fixed Window Counter, Sliding Window Log, Sliding Window Counter, Leaky Bucket, and Token Bucket. This project is designed for use in APIs and microservices to control the rate of incoming requests.

## Table of Contents

- [Features](#features)
- [Algorithms](#algorithms)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- Implements various rate-limiting algorithms:
  - **Fixed Window Counter**
  - **Sliding Window Log**
  - **Sliding Window Counter**
  - **Leaky Bucket**
  - **Token Bucket**
- Uses Redis for efficient request counting and state management.
- Easily configurable limits and time windows.
- Suitable for microservices and API rate limiting.

## Algorithms

### 1. Fixed Window Counter
A simple approach that counts the number of requests in a fixed time window. When the limit is exceeded, further requests are blocked until the window resets.

### 2. Sliding Window Log
Maintains a log of request timestamps to calculate the number of requests in the last N seconds. This approach provides a more granular control over request limits.

### 3. Sliding Window Counter
Combines the features of the Fixed Window and Sliding Window Log algorithms. It maintains a count of requests that fall within a defined window.

### 4. Leaky Bucket
Models the concept of a bucket that leaks tokens at a fixed rate. Requests can only proceed if there are tokens available in the bucket.

### 5. Token Bucket
Allows a burst of requests by replenishing tokens at a defined rate. If tokens are available, requests are allowed; otherwise, they are rate-limited.

## Getting Started

### Prerequisites

- Go 1.15+
- Redis installed and running
- Docker (optional, for Redis)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/rate-limiter.git
   cd rate-limiter
Install dependencies (if any):

bash
Copy code
go mod tidy
Start your Redis server. You can run Redis using Docker:

bash
Copy code
docker run -d -p 6379:6379 redis
Usage
Example
Here's a quick example of how to use the rate limiter:

go
Copy code
package main

import (
    "fmt"
    "time"
)

func main() {
    key := "user123"
    limit := 5
    window := 10 * time.Second

    allowed, err := RateLimitFixedWindowCounter(key, limit, window)
    if err != nil {
        fmt.Println("Error:", err)
    }
    if allowed {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Request rate-limited")
    }
}
Configuration
You can configure the rate limits and time windows as needed based on your application requirements.

Testing
To run the tests for the rate-limiting algorithms, execute:

bash
Copy code
go test -v
This will run all the test cases and provide detailed output.

Contributing
Contributions are welcome! Please open an issue or submit a pull request for any features or improvements.

Fork the repository.
Create your feature branch (git checkout -b feature/AmazingFeature).
Commit your changes (git commit -m 'Add some AmazingFeature').
Push to the branch (git push origin feature/AmazingFeature).
Open a pull request.
License
This project is licensed under the MIT License. See the LICENSE file for more details.

Feel free to customize any sections based on your specific implementation details or project goals. Good luck with your project!
