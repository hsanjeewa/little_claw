# 🤖 little_claw

Welcome to **little_claw**, the DevOps Agent project! 

The name **little_claw** was inspired by the *Game of Thrones* character **Littlefinger** ("Little Finger") — a nod to quiet influence, strategic coordination, and always knowing which lever to pull at the right time.

If you're reading this, you might be wondering what this project does. Simply put: **this is a smart assistant that helps us manage and monitor our servers safely.**

## 🌟 What does it do?
Managing servers can be scary. Sometimes engineers have to log into dozens of computers (servers) and type complex commands to update software, check disk space, or restart services. If they make a typo, things can break!

This project solves that by doing the heavy lifting automatically:
1. **It talks to our servers**: It uses secure connections (SSH) to talk to the servers we define in a file called `hosts.yaml`.
2. **It double-checks before making changes**: If the assistant is asked to install software that's already installed, it skips it to save time and prevent errors! (We call this *idempotency*).
3. **It asks for permission (Human-in-the-Loop)**: For dangerous commands (like deleting files or updating critical systems), the assistant pauses and asks you for a "Yes or No" approval directly on your screen.
4. **It has AI built-in**: When a command finishes (or fails), an artificial intelligence (Qwen 2.5) reads the technical computer jargon and translates it into a plain-English summary, so you know exactly what went wrong and why.

## 🚀 How to Run It (For Interns!)

We made sure you don't need a Ph.D. in Computer Science to run this.

### Step 0: Install Required Software
Before you start, you will need a few standard developer tools installed on your computer. If you are on a Mac, the easiest way is to use [Homebrew](https://brew.sh/) (`brew`):

1. **Go (Golang)**: The programming language this agent is built in.
   - Mac: `brew install go`
   - Windows/Linux: Download from [go.dev/dl](https://go.dev/dl/)
2. **Just**: A handy command runner (like `make`, but simpler).
   - Mac: `brew install just`
   - Windows/Linux: Follow [Just Installation](https://github.com/casey/just?tab=readme-ov-file#installation)
3. **Docker Desktop**: Used to spin up fake "test servers" safely on your laptop.
   - Download from [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop/)

### Step 1: Set up your environment
Ask your manager for the API Key. Once you have it, copy the example file:
```bash
cp .env.example .env
```
Open `.env` in any text editor and paste the key where it says `OPENAI_API_KEY`.

(Advanced configuration like timeouts and models are handled in `config/config.toml`).

### Step 2: Generate SSH Keys for Test Servers

The agent connects to servers via SSH. For local testing with Docker containers, generate a key pair:

```bash
mkdir -p test_keys
ssh-keygen -t ed25519 -f test_keys/id_ed25519 -N ""
```

This creates `test_keys/id_ed25519` (private key, used by the agent) and `test_keys/id_ed25519.pub` (public key, mounted into containers). The Docker containers in `docker-compose.yml` are pre-configured to accept this key.

> **Note**: The master secret in `cmd/agent/main.go` must be exactly **32 bytes** for AES-256 encryption of SSH keys in the vault. If it's any other length, `EncryptAndStore` silently fails and the SSH client cannot authenticate.

### Step 3: Start the Test Servers (Optional)

To test the agent safely on your computer without touching real servers, we have provided a simulated server environment:
```bash
just up
```
This boots up a Web Server (`devops-web-01` on port 2222) and a Database Server (`devops-db-01` on port 2223) for testing. (When you're done, stop them with `just down`).

### Step 4: Start the Agent

You can start the visual dashboard by running:
```bash
just run
```
*(If your computer says `command not found: just`, you can type `go run cmd/agent/main.go` instead).*

By default, Watchtower uses **SSH-backed collection** against the hosts defined in `hosts.yaml`. For UI development without real SSH targets, switch to the built-in simulator in `config/config.toml`:
```toml
[agent]
watchtower_backend = "simulator"
```

### Step 4: How to use the TUI
When the agent starts, you will land in the shell-based TUI. It has three top-level modes:

- **Watchtower** — the monitoring view
- **Autopilot** — the agent-led execution workspace
- **Copilot** — the human-led assistance workspace

#### Global navigation
little_claw uses a **leader key** inspired by terminal multiplexers:

- Press `Ctrl+a`, then `w` to go to **Watchtower**
- Press `Ctrl+a`, then `a` to go to **Autopilot**
- Press `Ctrl+a`, then `c` to go to **Copilot**
- Press `Ctrl+a`, then `z` to attach the built-in **Watchtower simulator** for UI review

If you press `Ctrl+a`, the shell will briefly show that it is waiting for the next mode key.

#### Watchtower redesign model
Watchtower is the monitoring mode and is being shaped around three internal **Watchtower Views**:

- **Fleet Aggregate** — a fleet-wide summary and triage surface
- **Fleet Matrix** — one metric family across scoped hosts as paginated cards
- **Host Detail** — a dense single-host dashboard with a strong `btop` homage

The redesign target is:

- direct Watchtower View switching with `g`, `m`, and `d`
- `b` for local Watchtower back-navigation through drill flow
- `[` / `]` for previous/next host in **Host Detail**
- `[` / `]` for previous/next page in **Fleet Matrix**
- a persistent fleet rail across all Watchtower Views
- compact Watchtower chrome that degrades by width before truncating core data
- CPU-led wide layouts for **Fleet Aggregate** and **Host Detail**, with denser supporting memory, disk, and network regions
- all four major metric families present across the redesigned surface:
  - CPU
  - Memory
  - Disk
  - Network

For the detailed interaction and release rules, see:

- `docs/plans/watchtower-redesign-spec.md`
- `docs/adr/0024-watchtower-view-model.md`
- `docs/adr/0025-watchtower-redesign-release-gate.md`

#### Current shell and Watchtower controls
Watchtower opens first by default and remains the main monitoring workspace while the redesign evolves.

- `j` / `k`
  - **Fleet Aggregate**: move focus across aggregate modules
  - **Fleet Matrix**: move the selected host within the current scope
  - **Host Detail**: move the focused metric module
- `Enter`
  - **Fleet Aggregate**: drill into **Fleet Matrix** for the focused family
  - **Fleet Matrix**: drill into **Host Detail** for the selected host
  - **Host Detail**: drill back into **Fleet Matrix** for the focused family
- `b` / `Esc` — step back through local Watchtower drill history
- `1` / `2` / `3` / `4` — switch to **Memory / CPU / Storage / Network**
- `[` / `]`
  - **Fleet Matrix**: move to the previous/next host page
  - **Host Detail**: move to the previous/next scoped host
- `r` — refresh the active visible metric family
- `a` — escalate the current Watchtower context into **Autopilot**
- `c` — escalate the current Watchtower context into **Copilot**

The simulator fleet currently exposes at least 4 hosts and bootstraps all four Watchtower metric families immediately when attached. Repeated refreshes will change the simulated values over time so you can review the UI with moving data.

#### Basic Autopilot and Copilot controls
Autopilot and Copilot are currently interactive skeletons that help you learn the layout and flow:

- `Tab` cycles focus between the command bar and panes
- `Enter` submits the current command-bar text
- both modes preserve their local state when you switch away and come back

#### Quitting
- `Ctrl+c` quits the application from any mode
- `q` also exits from non-Watchtower modes

## 📋 Configuration Reference

All configuration lives in `config/config.toml`:

```toml
[llm]
base_url = "https://openrouter.ai/api/v1"      # LLM API endpoint
model = "qwen/qwen-2.5-coder-32b-instruct"      # LLM model
timeout_seconds = 15                             # LLM request timeout

[agent]
ssh_timeout_seconds = 30                         # SSH connection timeout
inventory_path = "hosts.yaml"                    # Path to host inventory
watchtower_backend = "ssh"                       # "ssh" or "simulator"
watchtower_refresh_interval_seconds = 10         # Auto-refresh interval (0 = disabled)
database_path = "./agent.db"                     # SQLite database path
```

Set `watchtower_refresh_interval_seconds` to `0` to disable auto-refresh and rely on manual `r` keypress only.

## 🛠 For Developers
If you're a new developer joining the team, please read `AGENTS.md` for strict architectural guidelines on how the code is organized (Clean Architecture) before you start programming.
