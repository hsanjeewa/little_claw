---
name: satellite-contracts
description: Integrate with Satellite contracts as the trust-minimized on-chain data layer in Herodotus workflows.
---

# Herodotus AI Skill: Satellite Contracts (v1)

## Purpose

Use this skill to integrate with Satellite contracts as the trust-minimized on-chain data layer in Herodotus workflows.

## When to use

- Read verified historical values on-chain.
- Integrate `ISatellite` in Solidity projects.
- Resolve environment/chain deployment addresses.
- Design contract-side trust boundaries around proof-backed data.

## Source-of-truth

- https://docs.herodotus.cloud/satellite-contracts/introduction
- https://docs.herodotus.cloud/satellite-contracts/for-end-users
- https://docs.herodotus.cloud/satellite-contracts/architecture
- https://docs.herodotus.cloud/satellite-contracts/module-map
- https://docs.herodotus.cloud/satellite-contracts/herodotus-stack
- https://docs.herodotus.cloud/satellite-contracts/deployments-and-integration
- https://github.com/HerodotusDev/satellite

## Architecture pattern

`proof producer (Storage Proof/HDP) -> Satellite state -> consumer contracts -> app policy`

Satellite is trust layer; application logic remains your responsibility.

## Integration workflow

1. Select chain and environment.
2. Resolve Satellite address from deployment manifests.
3. Import `ISatellite`.
4. Implement minimum required read surface.
5. Gate business actions on verified reads.

## Safety requirements

- Prefer safe read variants where availability is uncertain.
- Keep chain/environment/address config externalized.
- Separate read adapter logic from business policy logic.
- Fail closed if verified data is absent or inconsistent.

## Anti-hallucination guardrails

- Do not invent addresses or module availability.
- Do not assume all module surfaces exist on every chain.
- Do not infer undocumented messaging paths.
- If docs and repo differ, prefer repo interfaces/manifests for integration details.

## Self-contained reference example

```solidity
function isEligible(
    address satellite,
    uint256 chainId,
    uint256 blockNumber,
    address account,
    uint256 minBalance
) external view returns (bool) {
    (bool ok, bytes32 v) = ISatellite(satellite).accountFieldSafe(
        chainId,
        blockNumber,
        account,
        IEvmFactRegistryModule.AccountField.BALANCE
    );
    if (!ok) return false;
    return uint256(v) >= minBalance;
}
```

## Output checklist

- Address resolution path documented
- ABI/interface version pinned
- Safe-read fallback behavior defined
- Business rules gated on verified data only
