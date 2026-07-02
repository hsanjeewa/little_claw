---
name: atlantic-api
description: Build, operate, and troubleshoot Atlantic proving workflows with strong source-grounding and minimal hallucination risk.
---

# Herodotus AI Skill: Atlantic API (v1)

## Purpose

Use this skill to build, operate, and troubleshoot Atlantic proving workflows with strong source-grounding and minimal hallucination risk.

## When to use

> **Precondition:** Atlantic requires an API key. If you don't already have one, run the `herodotus-auth` skill first to obtain one programmatically via wallet auth (https://docs.herodotus.cloud/skills/herodotus-auth). In Claude Code: `/herodotus-skills:herodotus-auth`.

- Submit Cairo proving jobs.
- Track query/job lifecycle and terminal states.
- Download artifacts (PIE/PROOF/metadata).
- Integrate L1/L2/offchain verification steps.

## Source-of-truth

- https://docs.herodotus.cloud/atlantic-api/introduction
- https://docs.herodotus.cloud/atlantic-api/getting-started
- https://docs.herodotus.cloud/atlantic-api/sending-query
- https://docs.herodotus.cloud/atlantic-api/status
- https://docs.herodotus.cloud/atlantic-api/downloading-files
- https://docs.herodotus.cloud/atlantic-api/webhooks
- https://docs.herodotus.cloud/atlantic-api/steps/trace-generation
- https://docs.herodotus.cloud/atlantic-api/steps/proof-generation
- https://docs.herodotus.cloud/atlantic-api/steps/l1-proof-verification
- https://docs.herodotus.cloud/atlantic-api/steps/l2-proof-verification
- https://docs.herodotus.cloud/atlantic-api/x402-payments
- OpenAPI spec: `openapi-atlantic.json` (available in the docs repo) and https://atlantic.api.herodotus.cloud/docs/json

## Architecture pattern

Treat Atlantic as proving infrastructure, not your business workflow engine:

`planner -> submitter -> lifecycle tracker -> artifact store -> verifier adapters -> app settlement`

## Implementation workflow

1. Authenticate and obtain an API key via the `herodotus-auth` skill (https://docs.herodotus.cloud/skills/herodotus-auth). New wallets get a Personal project + active API key auto-provisioned on the first session — `GET /api-keys?projectId=<selectedProject>` returns it.
2. Prefer the hcloud CLI/SDK for Atlantic submission, polling, artifact download, auth reuse, and x402 handling.
3. If the user does not want to use the hcloud CLI/SDK, implement a custom client from the OpenAPI contract at https://atlantic.api.herodotus.cloud/docs/json.
4. Build request payload from OpenAPI.
5. Submit query.
6. Poll status/jobs with backoff.
7. On terminal success, download artifacts.
8. Route to verifier adapter (L1/L2/offchain) as needed.
9. Persist query state, artifacts, and verification outcome.

## Agent interaction workflow

When the user wants to submit an Atlantic query, collect readiness and payload details before asking for final submission confirmation:

1. First ask whether the user already has a Herodotus/Atlantic API key.
2. If they do not have an API key, suggest using the `herodotus-auth` skill to obtain one before continuing. Do not ask whether to submit the query until the API key path is clear.
3. Before submit, ask for the full payload details required by the OpenAPI contract and workflow context, including at minimum:
   - target network/environment (`testnet` or `mainnet`)
   - Cairo version and VM choice when applicable
   - result/terminal artifact goal, for example trace/proof/verification flow
   - program/input files or JSON body descriptor
   - any verifier path requirements, such as L1, L2, or offchain verification
   - project-specific constraints such as declared job size, mocked fact hash policy, webhook URL, or idempotency key if used
4. Recommend `layout: "auto"` unless the user has a specific documented reason to override it. Atlantic can then choose the optimal layout.
5. After the user confirms the final payload and API-key/payment path, submit the query.
6. After submit, always surface the Atlantic Console status URL derived from the returned query id:

```text
https://www.herodotus.cloud/en/atlantic/<query-id>
```

Tell the user that inspecting this status/details page in Herodotus Console requires logging in with the same wallet that submitted the query.

For example, if `POST /atlantic-query` returns `01KR19Q2PV7EV61DS1MGP69ER0`, tell the user they can inspect status and query details at:

```text
https://www.herodotus.cloud/en/atlantic/01KR19Q2PV7EV61DS1MGP69ER0
```

## Reliability requirements

- Keep Atlantic query ID as a first-class DB entity.
- Add idempotency/dedup key handling on submit.
- Implement retry budget + exponential backoff.
- Support both polling and webhook-driven progression.
- Record full error payloads and transitions for debugging.

## Paying with x402

Atlantic accepts x402 v2 HTTP payments on `POST /atlantic-query`. The documented headers are `PAYMENT-REQUIRED`, `PAYMENT-SIGNATURE`, and `PAYMENT-RESPONSE`. Use x402 when an agent has an EVM wallet and either no API key, or an API key on a project whose prepaid credits ran out.

Read `references/x402-protocol.md` for the full wire recipe and deployment caveats, and `references/x402-error-codes.md` for the error taxonomy. To execute the flow without copying code, use `scripts/x402-pay.ts`.

## Run It

```bash
cd scripts
pnpm install
HCLOUD_API_KEY=... ./submit-query.ts --body-json query.json
HCLOUD_API_KEY=... ./poll-status.ts --query-id <id>
HCLOUD_API_KEY=... ./download-artifacts.ts --query-id <id>
```

The runnable TypeScript helpers import the published `@herodotus_dev/hcloud` package from npm.

## Anti-hallucination guardrails

- Do not invent endpoints/fields/statuses absent in OpenAPI/docs.
- Do not conflate Atlantic API with Data Processor API.
- If docs prose conflicts with OpenAPI, prefer OpenAPI for wire contract.
- If a behavior is undocumented, mark as unknown and ask for clarification.

## Self-contained reference example

```ts
async function runAtlanticJob(payload: unknown) {
  const queryId = await submitAtlanticQuery(payload); // POST /atlantic-query
  const status = await waitUntilTerminal(queryId); // poll with backoff
  if (status.kind !== "success") throw new Error(status.error);
  const artifacts = await downloadAtlanticArtifacts(queryId);
  return { queryId, artifacts };
}
```

## Output checklist

- Query ID captured
- Console status/details URL shown to the user as `https://www.herodotus.cloud/en/atlantic/<query-id>`
- Same-wallet login requirement for Herodotus Console status inspection communicated
- Terminal status captured
- Artifacts persisted
- Verification path selected
- Failure taxonomy and retry policy documented

## Index

### scripts/ (runnable TypeScript helpers)
- `scripts/submit-query.ts` — submit `POST /atlantic-query`
- `scripts/poll-status.ts` — poll query status to terminal state
- `scripts/download-artifacts.ts` — download `metadataUrls` artifacts
- `scripts/x402-pay.ts` — handle 402, sign EIP-3009, retry submit
- `scripts/common.ts`, `scripts/README.md`, `scripts/package.json`, `scripts/pnpm-lock.yaml`, `scripts/tsconfig.json`, `scripts/.gitignore` — shared helpers and install support

### examples/ (full scenario recipes)
- `examples/README.md`, `examples/LICENSE`, `examples/NOTICE` — index and Apache 2.0 attribution
- `examples/cairo0-python-vm/` — Cairo 0 Python VM
- `examples/cairo0-rust-vm/` — Cairo 0 Rust VM
- `examples/cairo1-rust-vm/` — Cairo 1 Rust VM
- `examples/l1-verification-contract/` — Solidity verification
- `examples/l2-verification-contract/` — Starknet verification
- `examples/offchain-verification/` — offchain verification
- `examples/anonymous-x402-flow.ts` — wallet-only x402 flow
- `examples/apikey-x402-topup.ts` — depleted API-key x402 top-up
- `examples/e2e-auth-then-submit.ts` — auth → submit composition
- `examples/package.json`, `examples/pnpm-lock.yaml`, `examples/tsconfig.json`, `examples/.gitignore` — install support for TS examples that import `@herodotus_dev/hcloud`

### references/ (deep docs, load on demand)
- `references/x402-protocol.md` — full x402 wire recipe
- `references/x402-error-codes.md` — x402 error taxonomy
- `references/lifecycle-states.md` — query/job states and retry policy
- `references/artifact-types.md` — PIE/proof/metadata artifacts
