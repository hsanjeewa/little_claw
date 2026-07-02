---
name: data-processor
description: Implement verifiable computation pipelines with HDP and enforce soundness constraints over historical on-chain data.
---

# Herodotus AI Skill: Data Processor (HDP) (v1)

## Purpose

Use this skill to implement verifiable computation pipelines with HDP and enforce soundness constraints over historical on-chain data.

## When to use

- Build Cairo modules that consume proof-backed chain data.
- Run `dry-run -> fetch-proofs -> sound-run`.
- Enforce correctness over candidate data sourced off-chain.

## Source-of-truth

- https://docs.herodotus.cloud/data-processor/introduction
- https://docs.herodotus.cloud/data-processor/architecture
- https://docs.herodotus.cloud/data-processor/pipeline
- https://docs.herodotus.cloud/data-processor/capabilities
- https://docs.herodotus.cloud/data-processor/verification
- https://docs.herodotus.cloud/data-processor/state-management
- https://docs.herodotus.cloud/data-processor/state-server
- https://docs.herodotus.cloud/data-processor/output
- https://docs.herodotus.cloud/data-processor/eth-call
- https://docs.herodotus.cloud/data-processor/design-patterns
- https://docs.herodotus.cloud/data-processor/examples
- https://docs.herodotus.cloud/data-processor/reference-cli
- https://docs.herodotus.cloud/data-processor/reference-configuration
- https://docs.herodotus.cloud/data-processor/reference-types
- https://docs.herodotus.cloud/data-processor/debugging

## Architecture pattern

Treat HDP as the soundness layer, not the query/discovery layer:

`indexer/discovery -> candidate set -> HDP constrained validation -> verifiable output -> app decision`

## Constraint patterns for query-like products

- Completeness constraint: ensure expected bounded set is fully represented.
- No-duplication constraint: deterministic ordering + uniqueness checks.
- Membership constraint: recompute keys/fields from proof-backed values.
- Anchoring constraint: compare derived commitment to trusted root/counter/checkpoint.
- Fail-closed policy: reject on any mismatch or missing proof.

## Implementation workflow

1. Build candidate dataset off-chain with provenance metadata.
2. Feed candidates + policy parameters into module inputs.
3. Run `hdp dry-run` to discover touched keys.
4. Run `hdp fetch-proofs` to assemble proof inputs.
5. Run `hdp sound-run` for deterministic verified execution.
6. Use `task_hash`, `output_root`, and `mmr_metas` as final outputs.

## Anti-hallucination guardrails

- Do not claim HDP is a general indexer or SQL engine.
- Do not claim local stage success guarantees settlement.
- Do not invent HDP APIs/flags outside documented surfaces.
- Do not mix HDP framework semantics with Data Processor API semantics.

## Self-contained reference example

```bash
# 1) Generate candidate set (off-chain)
./indexer_job --from <start> --to <end> --out candidates.json

# 2) Discover keys
hdp dry-run -m module.compiled_contract_class.json --print_output

# 3) Fetch proofs
hdp fetch-proofs --inputs dry_run_output.json --output proofs.json

# 4) Deterministic verified run
hdp sound-run -m module.compiled_contract_class.json --proofs proofs.json --print_output
```

## Output checklist

- Constraints documented and testable
- Candidate provenance retained
- Proof input and dry-run output matched
- Deterministic result validated across reruns
