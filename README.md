## Multi-threaded Proxy server
It is a multi-threaded forward proxy server built in Go.

## 🚀 Features
- HTTP proxy Support
- Concurrent request handling with semaphore limiting
- Caching Strategy
  - In-memory LRU Cache
- Handling requests on the TCP level

## 📂 Project Structure
```
├── app.go                    # contains server classes
├── go.mod                    # go module file
├── helper.go                 # helper functions
├── internal
│   └── cache
│       └── lru.go           # LRU cache implementation
├── main.go
├── proxy.go                  # proxy implementation
```

## 🛠️ Setup & Run
### Clone the repository
``` git clone <repository-url> ```
``` cd multithreaded-proxy-web-server ```

### Install Go dependencies
``` go mod tidy ```

### Build the binary
``` go build -o bin/proxy . ```

### Execute the binary
``` ./bin/proxy -port=9000 ```

now the proxy server is running on port 9000

### Test Locally using Postman
First, setup proxy setting to port 9000 in postman application, and you are all good to send your requests.

## 📺 Demo
[Watch on YouTube](https://youtu.be/7NFOJhJmL60?si=81Ok1bC4maj_QROg)

## To be implemented (open for contribution)
- HTTPS support
- Cache Invalidation policy
