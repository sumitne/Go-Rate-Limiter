REDIS

// making the redis persit 
// docker run --name redis-container -p 6379:6379 -v redis-data:/data -d redis


This command mounts the Redis data to a volume (redis-data) to persist it even if the container restarts

// Algos:
// Fixed Window Counter 
A fixed time window (e.g., 1 minute) is defined.
All incoming requests within that window are counted.
Once the request count exceeds the defined limit within the current window, further requests are blocked.
When the window expires, the request count resets, and the process starts again for the next window.

