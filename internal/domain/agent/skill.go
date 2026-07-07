package agent

type Skill struct {
	Name        string
	Description string
}

type SkillRegistry struct {
	skills []Skill
}

func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: []Skill{},
	}
}

func (r *SkillRegistry) Register(s Skill) {
	r.skills = append(r.skills, s)
}

func (r *SkillRegistry) Lookup(prefix string) []Skill {
	var matches []Skill
	for _, s := range r.skills {
		if len(prefix) > 0 && len(s.Name) >= len(prefix) && s.Name[:len(prefix)] == prefix {
			matches = append(matches, s)
		}
	}
	return matches
}
