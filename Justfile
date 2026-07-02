# Justfile for devops-agent

# Build the agent binary (pure Go, no CGO required)
build:
	CGO_ENABLED=0 go build -o agent ./cmd/agent/main.go

# Run the agent
run: build
	./agent

# Start simulated test servers via Docker Compose
up:
	docker-compose up -d

# Stop simulated test servers
down:
	docker-compose down

# View logs for test servers
logs:
	docker-compose logs -f

# Format code
fmt:
	go fmt ./...

# Run tests
test:
	go test -v ./...

# Tidy go modules
tidy:
	go mod tidy

# Build Docker image
docker-build:
	docker build -t devops-agent .

# Clean build artifacts
clean:
	rm -f agent agent.db
