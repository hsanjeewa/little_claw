# Justfile for devops-agent

# Build the agent binary (pure Go, no CGO required)
build:
	CGO_ENABLED=0 go build -o small_claw ./cmd/agent/main.go

# Run the agent
run: build
	./small_claw

# Start simulated test servers via Docker
up:
	docker rm -f devops-web-01 devops-db-01 2>/dev/null; \
	docker run -d --name devops-web-01 --hostname web-prod-01 -p 2222:2222 -e USER_NAME=deployer -e SUDO_ACCESS=true -e PASSWORD_ACCESS=false -e PUBLIC_KEY="$(cat test_keys/id_ed25519.pub)" lscr.io/linuxserver/openssh-server:latest && \
	docker run -d --name devops-db-01 --hostname db-master -p 2223:2222 -e USER_NAME=postgres -e SUDO_ACCESS=true -e PASSWORD_ACCESS=false -e PUBLIC_KEY="$(cat test_keys/id_ed25519.pub)" lscr.io/linuxserver/openssh-server:latest && \
	echo "Waiting for containers to start..." && sleep 5 && echo "Done"

# Stop simulated test servers
down:
	docker rm -f devops-web-01 devops-db-01 2>/dev/null; true

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
	rm -f small_claw agent.db
