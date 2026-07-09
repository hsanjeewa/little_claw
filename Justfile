# Justfile for devops-agent

# Dev vault master key. This is exported into the environment for the recipes
# below so the encrypted vault.enc can be opened. It is NOT committed as a real
# secret — override LITTLE_CLAW_MASTER_KEY in production (CI secret / secret
# manager) so vault.enc is protected by a real key.
export LITTLE_CLAW_MASTER_KEY := "a-very-secret-key-32-bytes-long!"

# Build the agent binary (pure Go, no CGO required)
build:
	CGO_ENABLED=0 go build -o small_claw ./cmd/agent/main.go

# Initialize the encrypted vault (vault.enc) from .env secrets (OPENAI_API_KEY,
# SUDO_PASS_*) + dev SSH keys. Headless — does not launch the TUI. Run once,
# then `just run`. After this, OPENAI_API_KEY can be removed from .env.
vault-init: build
	./small_claw init-vault

# Run the agent (master key injected via env; vault.enc holds encrypted secrets)
run: build
	./small_claw

# Start simulated test servers via Docker across multiple Linux distros.
# Each flavor exercises a different package manager + sudo path:
#   alpine  -> apk        (linuxserver/openssh-server, NOPASSWD sudo)
#   ubuntu  -> apt-get    (base image + scripts/test-server-bootstrap.sh)
#   centos  -> dnf        (base image + scripts/test-server-bootstrap.sh)
up:
	just up-alpine
	just up-ubuntu
	just up-centos
	@echo "Waiting for containers to start..." && sleep 6 && echo "Done"

# Alpine (apk) — uses the linuxserver image that honors USER_NAME/SUDO_ACCESS/PUBLIC_KEY.
up-alpine:
	docker rm -f devops-web-01 devops-db-01 2>/dev/null; \
	docker run -d --name devops-web-01 --hostname web-prod-01 -p 2222:2222 -e USER_NAME=deployer -e SUDO_ACCESS=true -e PASSWORD_ACCESS=false -e PUBLIC_KEY="$(cat test_keys/id_ed25519.pub)" lscr.io/linuxserver/openssh-server:latest && \
	docker run -d --name devops-db-01 --hostname db-master -p 2223:2222 -e USER_NAME=postgres -e SUDO_ACCESS=true -e PASSWORD_ACCESS=false -e PUBLIC_KEY="$(cat test_keys/id_ed25519.pub)" lscr.io/linuxserver/openssh-server:latest

# Ubuntu/Debian (apt-get) — bootstrapped with openssh + NOPASSWD sudo + pubkey.
up-ubuntu:
	docker rm -f devops-ubuntu-01 2>/dev/null; \
	docker run -d --name devops-ubuntu-01 --hostname ubuntu-prod-01 -p 2225:22 \
	  -e USER_NAME=ubuntu -e SUDO_ACCESS=true -e PASSWORD_ACCESS=false \
	  -v "$(pwd)/test_keys/id_ed25519.pub:/tmp/id_ed25519.pub:ro" \
	  -v "$(pwd)/scripts/test-server-bootstrap.sh:/bootstrap.sh:ro" \
	  ubuntu:22.04 bash /bootstrap.sh ubuntu /tmp/id_ed25519.pub

# CentOS/RHEL (dnf) — bootstrapped with openssh + NOPASSWD sudo + pubkey.
up-centos:
	docker rm -f devops-centos-01 2>/dev/null; \
	docker run -d --name devops-centos-01 --hostname centos-prod-01 -p 2226:22 \
	  -e USER_NAME=centos -e SUDO_ACCESS=true -e PASSWORD_ACCESS=false \
	  -v "$(pwd)/test_keys/id_ed25519.pub:/tmp/id_ed25519.pub:ro" \
	  -v "$(pwd)/scripts/test-server-bootstrap.sh:/bootstrap.sh:ro" \
	  rockylinux:9 bash /bootstrap.sh centos /tmp/id_ed25519.pub

# Password-enforced sudo host (Ubuntu) for testing the vault sudo-password feed.
# The sudo password for this host is "Sup3rSecret!"; set SUDO_PASS_SUDOPW_PROD_01
# in .env to match. The matching integration test skips if this container is down.
up-sudopw:
	docker rm -f devops-sudopw-01 2>/dev/null; \
	docker run -d --name devops-sudopw-01 --hostname sudopw-prod-01 -p 2227:22 \
	  -v "$(pwd)/test_keys/id_ed25519.pub:/tmp/id_ed25519.pub:ro" \
	  -v "$(pwd)/scripts/test-server-sudopw-bootstrap.sh:/bootstrap.sh:ro" \
	  ubuntu:22.04 bash /bootstrap.sh ubuntu /tmp/id_ed25519.pub Sup3rSecret!

# Stop all simulated test servers (all flavors)
down:
	docker rm -f devops-web-01 devops-db-01 devops-ubuntu-01 devops-centos-01 devops-sudopw-01 2>/dev/null; true

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
