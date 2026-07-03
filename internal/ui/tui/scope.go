package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/devops/agent/internal/infrastructure/inventory"
)

// TargetScopeKind identifies the supported shell-level scope shapes.
type TargetScopeKind int

const (
	// ScopeEntireInventory means every host known to the inventory is in play.
	ScopeEntireInventory TargetScopeKind = iota
	// ScopeSelectedHosts means only the explicitly chosen hosts are in play.
	ScopeSelectedHosts
)

// String returns the human-readable scope kind label.
func (k TargetScopeKind) String() string {
	switch k {
	case ScopeEntireInventory:
		return "Entire Inventory"
	case ScopeSelectedHosts:
		return "Selected Hosts"
	default:
		return "Unknown Scope"
	}
}

// TargetScope is the shell-owned primitive that describes which hosts the
// active mode and future modes are acting on. Modes consume this model
// instead of defining their own selection concepts.
type TargetScope struct {
	Kind  TargetScopeKind
	Hosts []inventory.TargetHost
}

// String returns a compact, human-readable description of the scope suitable
// for the shell chrome.
func (s TargetScope) String() string {
	switch s.Kind {
	case ScopeEntireInventory:
		return fmt.Sprintf("%s (%d hosts)", s.Kind.String(), len(s.Hosts))
	case ScopeSelectedHosts:
		aliases := make([]string, len(s.Hosts))
		for i, h := range s.Hosts {
			aliases[i] = h.Alias
		}
		sort.Strings(aliases)
		return fmt.Sprintf("%s: %s", s.Kind.String(), strings.Join(aliases, ", "))
	default:
		return s.Kind.String()
	}
}

// Clone returns a deep copy of the scope so callers cannot mutate the shell's
// internal state through the returned slice.
func (s TargetScope) Clone() TargetScope {
	return TargetScope{
		Kind:  s.Kind,
		Hosts: cloneHosts(s.Hosts),
	}
}

// Includes reports whether the given host alias is part of the scope.
func (s TargetScope) Includes(alias string) bool {
	for _, h := range s.Hosts {
		if h.Alias == alias {
			return true
		}
	}
	return false
}

func cloneHosts(src []inventory.TargetHost) []inventory.TargetHost {
	if src == nil {
		return nil
	}
	dst := make([]inventory.TargetHost, len(src))
	copy(dst, src)
	return dst
}
