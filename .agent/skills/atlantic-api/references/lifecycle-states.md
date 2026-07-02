# Lifecycle States

Atlantic query status values in the current server schema:

- `RECEIVED` - accepted, not yet actively running.
- `IN_PROGRESS` - jobs are running.
- `DONE` - terminal success.
- `FAILED` - terminal failure; inspect `errorReason`.

Some public docs and older examples use `SUCCESS` or `CANCELED`; clients should treat `DONE`, `SUCCESS`, `FAILED`, `CANCELED`, and `CANCELLED` as terminal for compatibility.

Job status values:

- `IN_PROGRESS`
- `COMPLETED`
- `FAILED`

Polling guidance:

- Start with a 5 second interval.
- Exponentially back off to a 30 second maximum.
- Preserve the query id as the durable handle.
- On `FAILED`, do not blindly resubmit; inspect `errorReason`, retry blockers, and whether a `dedupId` was used.
