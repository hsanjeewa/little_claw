package llm

func BuildPlanContext(goal, scope, capabilities, constraints, watchtowerContext string) []string {
	schema := `You are a DevOps planner. Respond with a JSON object containing a "reasoning" field and a "steps" array.
The response must be valid JSON. Every key and string value must be enclosed in double quotes.

Expected format:
{
  "reasoning": "Brief explanation of your plan",
  "steps": [
    {
      "description": "Human-readable step description",
      "command": "Actual shell command to execute",
      "is_mutative": false,
      "verify_command": "Optional verification command"
    }
  ]
}

CRITICAL: Each key like "command" must be followed by a colon (":") and then a quoted value. For example:
  "command": "sudo apt-get update"
NOT "command: sudo apt-get update" (that is invalid JSON).

Return ONLY valid JSON, no markdown, no extra text.`

	targetScope := "Target Scope: " + scope
	hostCapabilities := "Host Capabilities: " + capabilities
	opsConstraints := "Constraints: " + constraints
	userGoal := "User Goal: " + goal
	soulContent := "SOUL.md content placeholder"
	identityContent := "IDENTITY.md content placeholder"

	return []string{
		schema,
		targetScope,
		hostCapabilities,
		opsConstraints,
		watchtowerContext,
		userGoal,
		soulContent,
		identityContent,
	}
}