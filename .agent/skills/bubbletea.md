# Bubble Tea UI Expert Skill

You are an expert at building Terminal UIs using the Charm ecosystem in Go.
This skill applies to: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, and `github.com/charmbracelet/bubbles`.

---

## CRITICAL RULES (NEVER VIOLATE)

### 1. Never Write to stdout/stderr During Runtime
```go
// FORBIDDEN - corrupts the terminal
fmt.Println("debug")
log.Printf("info")

// CORRECT - route through the message system
return m, tea.Printf("info")            // bubbletea logger
return m, logToChan("info")             // custom channel
```

### 2. Commands Must Be Non-Blocking
```go
// FORBIDDEN - blocks the UI thread
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    output, err := ssh.Execute(...) // BLOCKS!
    return m, nil
}

// CORRECT - return a Cmd that runs in a goroutine
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m, func() tea.Msg {
        output, err := ssh.Execute(...)
        return SSHResultMsg{Output: output, Err: err}
    }
}
```

### 3. Never Compute Layout in View()
```go
// FORBIDDEN - View() must be pure
func (m model) View() string {
    width := getTerminalWidth() // UNSTABLE
    return render(width)
}

// CORRECT - store dimensions in model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}
func (m model) View() string {
    return render(m.width, m.height) // STABLE
}
```

### 4. Always Handle tea.WindowSizeMsg
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    // Also resize child components
    m.viewport.Width = msg.Width
    m.viewport.Height = msg.Height - headerHeight - footerHeight
    m.list.SetSize(msg.Width, msg.Height)
```

### 5. Initialize Viewport Only After First WindowSizeMsg
```go
type model struct {
    ready    bool
    viewport viewport.Model
}

case tea.WindowSizeMsg:
    if !m.ready {
        m.viewport = viewport.New(msg.Width, msg.Height)
        m.ready = true
    } else {
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height
    }
```

---

## Lipgloss Sizing Rules

### Padding and Borders ADD to Rendered Size
```go
// This style has Padding(1) and a Border
style := lipgloss.NewStyle().Padding(1).BorderStyle(lipgloss.NormalBorder())

// style.Width(10) means CONTENT+PADDING = 10
// Rendered total = 10 + 2 border chars = 12
// Vertical: 10 + 2 border lines = 12
```

### GetFrameSize() for Exact Calculation
```go
hPad, vPad := style.GetHorizontalFrameSize(), style.GetVerticalFrameSize()
// Use these when calculating child dimensions
childWidth := parentWidth - hPad
childHeight := parentHeight - vPad
```

### Always Subtract Frames From Width Assignments
```go
// WRONG - will overflow terminal
leftPanel := leftStyle.Width(m.width / 2).Render("hello")
rightPanel := rightStyle.Width(m.width / 2).Render("world")

// CORRECT - account for borders
frameSize := leftStyle.GetHorizontalFrameSize() + rightStyle.GetHorizontalFrameSize()
leftPanel := leftStyle.Width(m.width/2 - leftStyle.GetHorizontalFrameSize()).Render("hello")
rightPanel := rightStyle.Width(m.width/2 - rightStyle.GetHorizontalFrameSize()).Render("world")
```

### Truncate Before Styling
```go
// FORBIDDEN - may wrap and break height
line := lipgloss.NewStyle().Width(20).Render(longString)

// CORRECT - truncate first
truncated := ansi.Truncate(longString, 20, "")
line := lipgloss.NewStyle().Width(20).Render(truncated)
```

---

## Common Patterns

### Channel Listener Pattern (Background -> UI)
```go
// In your domain/infrastructure layer:
type LogMsg struct { Entry string }

func listenForLogs(logChan <-chan string) tea.Cmd {
    return func() tea.Msg {
        entry := <-logChan
        return LogMsg{Entry: entry}
    }
}

// In Update():
case LogMsg:
    m.logs = append(m.logs, msg.Entry)
    return m, listenForLogs(m.logChan) // Re-subscribe
```

### Side-by-Side Panels
```go
left := leftStyle.Width(leftW).Render(m.leftContent)
right := rightStyle.Width(rightW).Render(m.rightContent)
main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

header := headerStyle.Width(m.width).Render("Title")
footer := footerStyle.Width(m.width).Render("Status")

return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
```

### Human-in-the-Loop Approval Gate
```go
type model struct {
    approvalPending *ApprovalRequest
}

type ApprovalRequest struct {
    Prompt   string
    Response chan bool
}

case tea.KeyMsg:
    if m.approvalPending != nil {
        switch msg.String() {
        case "y", "Y":
            m.approvalPending.Response <- true
            m.approvalPending = nil
        case "n", "N":
            m.approvalPending.Response <- false
            m.approvalPending = nil
        }
        return m, nil
    }
    // Normal key handling...
```

### Multi-Source Message Listener
```go
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        listenForTask(m.taskChan),
        listenForLog(m.logChan),
        listenForAlert(m.alertChan),
    )
}

// Each listener re-subscribes after firing
func listenForTask(ch <-chan Task) tea.Cmd {
    return func() tea.Msg {
        return TaskMsg{Task: <-ch}
    }
}
```

---

## Error Handling

### Commands Return Errors as Messages
```go
typeErrMsg struct{ err error }

func fetchData() tea.Cmd {
    return func() tea.Msg {
        data, err := http.Get("...")
        if err != nil {
            return errMsg{err}
        }
        return dataMsg{data}
    }
}

case errMsg:
    m.error = msg.err.Error()
    return m, nil
```

### Graceful Context Cancellation
```go
func sshExec(ctx context.Context, cmd string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
        // ... run command
    }
}
```

---

## Component Patterns

### Viewport with Header/Footer
```go
func (m model) View() string {
    if !m.ready {
        return "  Initializing..."
    }
    header := m.headerView()
    footer := m.footerView()
    return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

func (m model) headerView() string {
    title := titleStyle.Render("Title")
    line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
    return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
    info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
    line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
    return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
```

### Adaptive Colors (Light/Dark Terminal)
```go
style := lipgloss.NewStyle().
    Foreground(lipgloss.AdaptiveColor{
        Light: "#000000",
        Dark:  "#FFFFFF",
    })
```

### Dynamic Style Variables (Package Level)
```go
var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#7D56F4")).
        MarginLeft(2)

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF0000")).
        Bold(true)

    subtleStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#888888"))
)
```

---

## Checklist for Every Bubble Tea Feature

- [ ] Does `Update()` handle `tea.WindowSizeMsg`?
- [ ] Is viewport/bubbles component sized using stored `m.width`/`m.height`?
- [ ] Are all I/O operations wrapped in `tea.Cmd` (never in `Update` directly)?
- [ ] Do `tea.Cmd` functions return error messages instead of panicking?
- [ ] Are Lipgloss frame sizes subtracted when assigning child widths?
- [ ] Is text truncated with `ansi.Truncate` before `.Width()` styling?
- [ ] Does `View()` only read from model state (pure function)?
- [ ] Are channel listeners re-subscribed after each message?
- [ ] Is `m.ready` checked before accessing initialized components?
