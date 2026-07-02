# 🤖 DevOps Agent

Welcome to the **DevOps Agent** project! 

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

### 1. Set up your environment
Ask your manager for the API Key. Once you have it, copy the example file:
```bash
cp .env.example .env
```
Open `.env` in any text editor and paste the key where it says `OPENAI_API_KEY`.

### 2. Start the Agent
You can start the visual dashboard by running:
```bash
just run
```
*(If your computer says `command not found: just`, you can type `go run cmd/agent/main.go` instead).*

### 3. How to use the Dashboard
Once the agent starts, you will see a colorful terminal application split into panels:
- **Top Bar**: Shows you how many tasks succeeded, failed, or are waiting for your approval.
- **Left Panel (Targets)**: Shows the list of servers it is trying to connect to.
  - You can use the `j` (down) and `k` (up) keys on your keyboard to select different servers.
- **Right Panel (AI Analysis)**: Shows you the output of the selected server, along with the AI's explanation of what happened.
  - You can use the `↑` and `↓` arrows on your keyboard to scroll if the text is long.
- **Bottom Panel (Prompts)**: Watch this area! If a yellow box pops up asking `Approve Execution?`, you must press `y` (yes) or `n` (no) on your keyboard to let the agent continue.

## 🛠 For Developers
If you're a new developer joining the team, please read `AGENTS.md` for strict architectural guidelines on how the code is organized (Clean Architecture) before you start programming.
