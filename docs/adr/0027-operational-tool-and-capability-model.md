# ADR 0027: Operational tool and capability model

## Status

Accepted

## Context

The initial Autopilot must support agentic DevOps and sysadmin work on targeted hosts while remaining supervised and auditable. A shell-first execution model is flexible, but it weakens reviewability, planning discipline, and safety. The product also needs to adapt its plans to the actual operating environment of each target host.

## Decision

The initial execution model is:

- **Operational Tools** are the preferred execution surface for planned work.
- **Guarded Shell Actions** exist as a fallback when no suitable Operational Tool fits.
- The planner should use **Capability Discovery** to build a **Host Capability Profile** before planning environment-sensitive actions.
- **Capability Discovery** should run at the start of each Autopilot Run for the targeted hosts.

The Host Capability Profile includes host facts needed for operational planning, such as operating system, distribution details, service manager, package manager, available tools, and relevant versions.

## Consequences

- Planned work can target explicit tool semantics instead of treating arbitrary shell as the default abstraction.
- Shell remains available for edge cases without becoming the normal planner target.
- The system must persist or refresh host capability information strongly enough to support planning on heterogeneous hosts.
- Planning can rely on fresh host facts by default instead of assuming cached environment data is still accurate.
- Future work can add richer capability dimensions without changing the core distinction between typed tools and guarded shell fallback.
