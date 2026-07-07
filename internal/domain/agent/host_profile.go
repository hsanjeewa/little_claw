package agent

import "context"

type HostProfile struct {
	HostAlias   string
	OS          string
	Distro      string
	ServiceMgr  string
	PkgMgr      string
	Installed   []string
}

func NewHostProfile(alias string) HostProfile {
	return HostProfile{
		HostAlias: alias,
		Installed: []string{},
	}
}

type CapabilityDiscoverer interface {
	Discover(ctx context.Context, hostAlias string) (HostProfile, error)
}
