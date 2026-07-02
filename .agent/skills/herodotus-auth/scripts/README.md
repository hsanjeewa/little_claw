# herodotus-auth scripts

Runnable TypeScript helpers for wallet-based authentication against Herodotus Cloud.

## Install

```bash
pnpm install   # or: npm install
```

Requires Node 20+. The scripts use the published `@herodotus_dev/hcloud` package from npm.

## Conventions

- Each script writes a single JSON object to **stdout** on success: `{schemaVersion: 1, ...}`.
- Errors go to **stderr** as `{schemaVersion: 1, error: {code, message, details?}}`.
- Exit codes: `0` success · `1` user error · `2` transient/network · `3` server error.
- Auth scripts accept `--auth-base-url <url>` to override the default `https://auth-billing.api.herodotus.cloud`.
- All scripts accept `--pretty` for indented JSON, `--help` for usage.

## Scripts

### `auth.ts`

Wallet auth through the hcloud SDK. Credentials are persisted under `~/.hcloud` for later Atlantic calls.

```bash
WALLET_PRIVATE_KEY=0x... ./auth.ts | jq
```

Emits `{wallet, projectId, selectedProject, apiKey}`.

### `refresh.ts`

Refresh the active wallet bearer session stored by the hcloud SDK.

```bash
./refresh.ts --pretty
```

Emits `{refreshed, sessionExpiresAt}`.

### `get-api-key.ts`

Read or list API keys for the active stored wallet, or a supplied project.

```bash
./get-api-key.ts
./get-api-key.ts --project-id "$PROJECT" --list
```

## Stored credential pattern

```bash
WALLET_PRIVATE_KEY=0x... ./auth.ts
./get-api-key.ts | jq -r .apiKey
```

## Type-check

```bash
pnpm typecheck
```
