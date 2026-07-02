---
name: data-structure-indexer-api
description: Query accumulator and remapper data for planning proof-backed workflows.
---

# Herodotus AI Skill: Data Structure Indexer API (v1)

## Purpose

Use this skill to query accumulator/remapper data for planning proof-backed workflows.

## When to use

- Discover accumulators and MMR metadata.
- Retrieve peaks/proofs/remapper paths.
- Map timestamp-oriented requests into block candidates.

## Source-of-truth

- https://docs.herodotus.cloud/data-structure-indexer-api/introduction
- OpenAPI spec: `openapi-data-structure-indexer-api.json` (available in the docs repo)

## Architecture pattern

Use as discovery substrate, then validate via proof-backed pipeline:

`indexer query -> candidate planning -> proof-backed validation (HDP/Storage Proof) -> trusted consumption`

## Implementation workflow

1. Choose environment server.
2. Query accumulator/remapper endpoints with explicit filters.
3. Persist candidate provenance (endpoint + params + response version).
4. Feed candidates to proof-backed path.
5. Consume only verified outcomes downstream.

## Anti-hallucination guardrails

- Do not invent endpoint semantics beyond OpenAPI.
- Do not claim this API alone proves on-chain truth.
- Keep unknown semantics explicit where docs are sparse.
- Keep chain/filter parameters explicit and validated.

## Self-contained reference example

```ts
async function resolveVerifiedBlockForTimestamp(ts: number) {
  const candidate = await getBlockHeaderByTimestamp(ts); // indexer step
  const proofJob = await submitProofForBlock(candidate.blockNumber); // proof-backed step
  await waitProofCompletion(proofJob.id);
  return candidate.blockNumber;
}
```

## Output checklist

- Candidate provenance persisted
- Filter constraints validated
- Proof-backed validation step required
- Downstream consumption guarded on verified status
