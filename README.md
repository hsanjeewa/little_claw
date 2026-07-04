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

### Step 2: Start the Test Servers (Optional)
To test the agent safely on your computer without touching real servers, we have provided a simulated server environment. Run:
```bash
just up
```
This boots up a fake Web Server and a fake Database Server specifically for testing! (When you're done, you can stop them with `just down`).

### Step 3: Start the Agent
You can start the visual dashboard by running:
```bash
just run
```
*(If your computer says `command not found: just`, you can type `go run cmd/agent/main.go` instead).*

If you want Watchtower to start directly on the built-in simulator backend instead of SSH-backed hosts, set this in `config/config.toml`:
```toml
[agent]
watchtower_backend = "simulator"
```

### Step 4: How to use the TUI
When the agent starts, you will land in the new shell-based TUI. It currently has three top-level modes:

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

#### Basic Watchtower controls
Watchtower opens first by default and focuses on server health across your inventory.

- `j` / `k` — move between hosts in the fleet matrix
- `Enter` — open **Host Detail** for the selected host
- `b` — go back from Host Detail to the fleet matrix
- `1` — switch to the **Memory** metric family
- `2` — switch to the **CPU** metric family
- `3` — switch to the **Storage** metric family
- `4` — switch to the **Network** metric family
- `r` — refresh the current metric family
- `a` — escalate the current Watchtower context into **Autopilot**
- `c` — escalate the current Watchtower context into **Copilot**

The simulator fleet currently exposes at least 4 hosts and supports all four Watchtower metric families. Repeated refreshes will change the simulated values over time so you can review the UI with moving data.

#### Basic Autopilot and Copilot controls
Autopilot and Copilot are currently interactive skeletons that help you learn the layout and flow:

- `Tab` cycles focus between the command bar and panes
- `Enter` submits the current command-bar text
- both modes preserve their local state when you switch away and come back

#### Quitting
- `Ctrl+c` quits the application from any mode
- `q` also exits from non-Watchtower modes

## 🛠 For Developers
If you're a new developer joining the team, please read `AGENTS.md` for strict architectural guidelines on how the code is organized (Clean Architecture) before you start programming.
