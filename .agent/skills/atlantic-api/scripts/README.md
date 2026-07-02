# atlantic-api scripts

Runnable TypeScript helpers for Atlantic proving workflows.

## Install

```bash
pnpm install
```

Requires Node 20+. The scripts use the published `@herodotus_dev/hcloud` package from npm.

## Conventions

- Success writes one JSON object to stdout: `{schemaVersion: 1, ...}`.
- Errors write one JSON object to stderr: `{schemaVersion: 1, error: {code, message, details?}}`.
- Exit codes: `0` success, `1` user/API 4xx, `2` transient/network, `3` server/unexpected.
- All scripts accept `--base-url`; default is `https://atlantic.api.herodotus.cloud`.
- API key resolution follows the hcloud SDK: `--api-key`, `HCLOUD_API_KEY`, `HERODOTUS_API_KEY`, `ATLANTIC_API_KEY`, then stored wallet credentials from `~/.hcloud`.
- `--pretty` indents JSON; `--help` prints usage.

## Body descriptors

Use `--body-json` or `--body-file` with a JSON descriptor:

```json
{
  "fields": {
    "layout": "auto",
    "cairoVm": "python",
    "cairoVersion": "cairo0",
    "result": "PROOF_GENERATION",
    "mockFactHash": "false",
    "declaredJobSize": "S"
  },
  "files": {
    "programFile": "compiled.json",
    "inputFile": "input.json"
  }
}
```

File paths are resolved relative to the descriptor file.

## Scripts

```bash
HCLOUD_API_KEY=... ./submit-query.ts --body-json query.json
HCLOUD_API_KEY=... ./poll-status.ts --query-id "$QUERY_ID"
HCLOUD_API_KEY=... ./download-artifacts.ts --query-id "$QUERY_ID"
WALLET_PRIVATE_KEY=0x... HCLOUD_API_KEY=... ./x402-pay.ts --body-json query.json
```

`submit-query.ts` exits with `PAYMENT_REQUIRED` on 402. Use `x402-pay.ts` with the same body descriptor to let the hcloud SDK sign and retry with the configured payment adapter.

## Type-check

```bash
pnpm typecheck
```
