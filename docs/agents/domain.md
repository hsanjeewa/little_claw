# Domain docs

## Layout

Single-context. The project has one global domain context:

- `CONTEXT.md` at the repo root
- `docs/adr/` at the repo root for architectural decision records

## Consumer rules

When reading domain context:

1. Start with `CONTEXT.md` for the project's domain language and boundaries.
2. Check `docs/adr/` for past architectural decisions that affect the current task.
3. If a decision isn't recorded in an ADR, assume it hasn't been decided and flag the gap.
