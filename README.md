# GoGate - Lightweight Cloud-Native API Gateway

GoGate is a high-performance, lightweight API Gateway built for cloud-native environments, designed with simplicity and developer experience in mind.

## Features

Core functionality implemented:

- **Reverse Proxy**: Forward requests to backend services

  - Path rewriting support
  - Standard proxy header handling (X-Forwarded-For, X-Real-IP, etc.)
  - Flexible route matching rules

- **Load Balancing**: Intelligently distribute requests to multiple backend services

  - Weighted Round Robin algorithm
  - Dynamic service node management
  - Smooth request distribution

- **JWT Authentication**: Secure API access

  - Standard JWT token validation
  - Configurable path exclusion list
  - User information propagation

- **Rate Limiting**: Prevent service overload
  - Token bucket algorithm implementation
  - Global and path-level rate limiting
  - Configurable rate and burst settings

## Installation and Usage

### Prerequisites

- Go 1.16+
- Git

### Installation Steps

```bash
# Clone the repository
git clone https://github.com/ilukemagic/gogate.git
cd gogate

# Download dependencies
go mod download

# Build
go build -o gogate cmd/server/main.go
```

### Configuration

GoGate uses YAML configuration files. Modify the settings in `configs/config.yaml`:

```yaml
proxy:
  listen: ":8080" # Gateway listening address
  routes:
    "/api/test": # Route path
      targets:
        - url: "http://localhost:8081"
          weight: 3 # Weight of 3
        - url: "http://localhost:8082"
          weight: 2 # Weight of 2

jwt:
  secretKey: "your-secret-key-here"
  exclude: # Paths that don't require JWT validation
    - "/health"

rateLimit:
  enable: true
  rate: 100 # Global rate limit: 100 requests per second
  burst: 50 # Allow burst of 50 requests
  routes:
    "/api/test":
      rate: 10 # Path-specific rate limit: 10 requests per second
      burst: 5 # Allow burst of 5 requests
```

### Running

```bash
# Run with default configuration
./gogate

# Specify a configuration file
./gogate -config path/to/config.yaml
```

## Testing

### Reverse Proxy and Load Balancing Test

1. Start the test servers:

```bash
go run test/test_servers.go
```

2. Start GoGate:

```bash
go run cmd/server/main.go
```

3. Get a JWT token:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'
```

4. Test API access:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/test
```

### Load Balancing Test

```bash
# Make multiple requests to observe load distribution
for i in {1..10}; do
  curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/test
done
```

### Rate Limiting Test

```bash
# Use the test script
./scripts/test_ratelimit.sh
```

## Contributing

Contributions are welcome, whether it's code contributions, bug reports, or feature suggestions.

## License

This project is licensed under the MIT License.

[Chinese Version README (中文版)](README_CN.md)
