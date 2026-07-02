---
name: storage-proof-api
description: Request proof-backed data, track completion, and consume verified values on-chain via Satellite.
---

# Herodotus AI Skill: Storage Proof API (v1)

## Purpose

Use this skill to request proof-backed data, track completion, and consume verified values on-chain via Satellite.

## When to use

- Request account/storage/header/timestamp proof data.
- Build cross-chain historical verification workflows.
- Read proven values on-chain with safe access patterns.

## Source-of-truth

- https://docs.herodotus.cloud/storage-proofs-api/introduction
- https://docs.herodotus.cloud/storage-proofs-api/use-cases
- https://docs.herodotus.cloud/storage-proofs-api/quick-start-guide
- https://docs.herodotus.cloud/storage-proofs-api/key-concepts
- https://docs.herodotus.cloud/storage-proofs-api/contracts/accessing-data
- https://docs.herodotus.cloud/storage-proofs-api/contracts/example-smart-contract
- https://docs.herodotus.cloud/storage-proofs-api/contracts/contract-addresses
- OpenAPI spec: `openapi-storage-proof-api.json` (available in the docs repo)

## Architecture pattern

`request builder -> proof submission -> lifecycle tracker -> Satellite read adapter -> app rule engine`

Only make business decisions after terminal success and successful on-chain read.

## Implementation workflow

1. Build batch query payload from app intent.
2. Submit request and persist request ID + payload hash.
3. Poll status until terminal success/failure.
4. On success, read values from Satellite using safe methods.
5. Feed verified values to business logic.

## On-chain access pattern

Use documented Satellite reads:

- `headerFieldSafe`
- `accountFieldSafe`
- `storageSlotSafe`
- `timestampSafe`

Prefer safe variants to avoid revert-driven control flow.

## Anti-hallucination guardrails

- Do not invent chain IDs, statuses, or fields.
- Do not conflate with Data Processor API.
- Resolve docs/OpenAPI auth differences explicitly and configure auth per environment.
- Treat unknown behavior as unknown; do not infer.

## Self-contained reference example

```ts
async function proveThenReadBalance(req: BatchQuery, readArgs: ReadArgs) {
  const id = await submitBatchQuery(req);
  const status = await waitBatchDone(id);
  if (status !== "DONE") throw new Error("proof flow not completed");
  const result = await satelliteAccountFieldSafe(readArgs);
  if (!result.exists) throw new Error("verified value missing on-chain");
  return result.value;
}
```

## Output checklist

- Proof request tracked by immutable ID
- Terminal status required before consumption
- Safe-read fallback implemented
- Address/network validation in place
