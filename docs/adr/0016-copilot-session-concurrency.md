# ADR 0016: Copilot session concurrency

## Status

Accepted

## Context

Copilot is a Session-oriented, human-led assistance mode. The product needs a clear concurrency model so UI structure, persistence, and advisory context remain coherent.

## Decision

For v1, Copilot supports **one active Session at a time** while preserving prior Sessions as resumable history.

## Consequences

- The Copilot mode can optimize around a single live advisory context.
- Session history and resume flows remain important, but the first version avoids multi-session coordination complexity.
- Shell status for Copilot can summarize one active Session cleanly.
