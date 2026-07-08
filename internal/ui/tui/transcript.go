package tui

type TranscriptEntryKind int

const (
	TranscriptOperator TranscriptEntryKind = iota
	TranscriptAgent
	TranscriptSystem
)

type TranscriptEntry struct {
	Kind TranscriptEntryKind
	Text string
}

func (e TranscriptEntry) String() string {
	switch e.Kind {
	case TranscriptOperator:
		return "> " + e.Text
	case TranscriptAgent:
		return "🤖 " + e.Text
	case TranscriptSystem:
		return "ℹ️  " + e.Text
	default:
		return e.Text
	}
}

func (e TranscriptEntry) OperatorText() bool {
	return e.Kind == TranscriptOperator
}

func (e TranscriptEntry) AgentText() bool {
	return e.Kind == TranscriptAgent
}

func (e TranscriptEntry) SystemText() bool {
	return e.Kind == TranscriptSystem
}