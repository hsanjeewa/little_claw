# ADR 0033: Mutating step verification

## Status

Accepted

## Context

Agentic DevOps and sysadmin work is unsafe if state-changing steps are treated as successful based only on command completion. The product needs a clear verification expectation that fits supervised execution and heterogeneous hosts.

## Decision

Each mutating **Execution Branch** step must be followed by an explicit **Verification Step**.

Verification should prefer tool-aware checks when available. If no typed verification path exists, the system may use guarded read-only inspection. Command exit status alone is not sufficient verification for important state changes.

If verification fails for only part of a branch, the branch may split: verified hosts may continue, while the failed hosts become a **Blocked Subset** that requires operator review before further mutating work.

For a Blocked Subset, the operator may initiate explicit **Recovery Actions**: inspect evidence, retry the failed step, skip the affected hosts for the remainder of the run, or abort the remaining work for that subset. Automatic replanning for failed hosts is not supported in v1.

## Consequences

- Planned work becomes more trustworthy and reviewable.
- Successful hosts do not need to be held back by unrelated failures in the same branch.
- Tool design must consider both mutation and verification surfaces.
- Run presentation should make verification visible as part of operational progress rather than treating it as incidental logging.
