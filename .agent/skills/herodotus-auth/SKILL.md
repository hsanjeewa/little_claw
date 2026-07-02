---
name: herodotus-auth
description: Authenticate to Herodotus Cloud programmatically with an EVM wallet (EIP-712) to obtain a Bearer access token and an API key. Required precondition for every Herodotus API skill (Atlantic, Storage Proof, Data Processor, etc.).
---

# Herodotus AI Skill: Wallet Authentication (v1)

## Purpose

Programmatic, browser-free authentication for AI agents and CLIs. Exchange an EIP-712-signed challenge for a Bearer-channel access token, then read or create your API key. Every other Herodotus API skill assumes this skill has already produced a usable API key.

## When to use

- An agent, CLI, server, or notebook needs an API key for Atlantic API, Storage Proof API, Data Processor API, Data Structure Indexer API, or Satellite contract orchestration.
- Cookie-based session is not viable (no cookie jar, cross-origin, serverless, headless).
- You have an EVM wallet — any signer that can produce an EIP-712 signature works (private key in env, KMS, hardware wallet, MetaMask, ethers, viem).

**Out of scope.** This skill is wallet auth only. The GitHub OAuth path (`/auth/github/...`) is cookie-only by design and is not exposed through this protocol — do not send `channel: "bearer"` in the body against it.

## Source-of-truth

- Docs: https://docs.herodotus.cloud/documentation/authentication#programmatic-wallet-authentication
- Skill page: https://docs.herodotus.cloud/skills/herodotus-auth
- Prod base URL: `https://auth-billing.api.herodotus.cloud`
- Endpoints used by this skill:
  - `GET  /auth/web3/challenge?wallet=0x…`
  - `POST /auth/web3/session` (body field `channel: "bearer"`)
  - `POST /auth/refresh-token` (`Authorization: Bearer <refreshToken>`)
  - `GET  /api-keys?projectId=<selectedProject>&limit=10&offset=0`
  - `POST /api-keys` (mint additional keys)

## Protocol contract (do not deviate)

1. **Fetch challenge.** `GET /auth/web3/challenge?wallet=<addr>` returns `{ challengeToken, nonce, issuedAt, expiresAt, statement, eip712 }`. The `eip712` object contains `domain`, `types`, `primaryType`, and `message` — **sign exactly those fields, do not reconstruct them client-side**.
2. **Sign typed data.** Use any EIP-712 signer to produce a signature over `eip712.domain` / `eip712.types` / `eip712.primaryType` / `eip712.message`.
3. **Exchange for a Bearer session.** `POST /auth/web3/session` with body `{ wallet, challengeToken, signature, channel: "bearer" }`. Response body returns `{ accessToken, refreshToken, expiresAt, selectedProject }`. **No `Set-Cookie` is sent for the bearer path.** Persist all four fields.
4. **Use the access token.** Set `Authorization: Bearer <accessToken>` on every subsequent call. Do **not** put the token in a cookie — channel binding will reject it.
5. **Refresh.** Before expiry, `POST /auth/refresh-token` with `Authorization: Bearer <refreshToken>`. Response body returns a fresh `{ accessToken, refreshToken, expiresAt }`. Old refresh token is invalidated.
6. **Get API key.** First-time wallets are auto-provisioned with a Personal project and one active API key. Retrieve it with `GET /api-keys?projectId=<selectedProject>&limit=10&offset=0`; read `data[0].apiKey`. Mint additional keys with `POST /api-keys` body `{ projectId, type: { name, color } }`.

## Signers and Channel Binding

Signing is caller-owned. `scripts/auth.ts` uses the hcloud SDK `privateKeySigner`; ethers, browser wallets, KMS, and hardware wallets are covered in `references/eip712-signing.md`.

Tokens are channel-bound. A token issued with `channel: "bearer"` must be sent via `Authorization: Bearer`; cookie-issued tokens must stay cookies. See `references/channel-binding.md` for the security property and curl recipe.

## Run It

```bash
cd scripts
pnpm install
WALLET_PRIVATE_KEY=0x... ./auth.ts | jq -r .apiKey
```

## Anti-hallucination guardrails

- Do not invent endpoints. The five listed under "Source-of-truth" are the entire surface this skill needs.
- Do not hardcode the EIP-712 `domain`, `types`, `primaryType`, or `statement` — read them from the challenge response on every login. The server may rotate them.
- Do not extract a cookie-issued JWT and forward it as `Authorization: Bearer`. The server enforces channel binding and will reject it.
- Do not assume a default `projectId`. Always read `selectedProject` from the session response.
- Do not assume `POST /api-keys` is required. New wallets get one auto-provisioned; only call POST if you need additional keys.
- If a behavior is undocumented in the source-of-truth list above, mark it unknown and ask for clarification rather than inventing it.

## Output checklist

- Challenge fetched with the exact wallet address that will sign.
- Signature produced over the verbatim `eip712` payload from the challenge response.
- `channel: "bearer"` set on the session request.
- `accessToken`, `refreshToken`, `expiresAt`, `selectedProject`, and API key persisted by the hcloud SDK credential store.
- API key retrieved (or minted) and stored for downstream skills.
- Refresh path verified before access-token expiry: `POST /auth/refresh-token` with `Authorization: Bearer <refreshToken>` returns a new pair.
- All subsequent Herodotus API calls use `Authorization: Bearer <accessToken>` — never cookies.

## Index

### scripts/ (runnable TypeScript helpers)
- `scripts/auth.ts` — wallet auth through hcloud SDK → persisted credentials + API key JSON
- `scripts/refresh.ts` — refresh active stored wallet session before expiry
- `scripts/get-api-key.ts` — read/list API keys for the active stored wallet or explicit project
- `scripts/README.md`, `scripts/package.json`, `scripts/pnpm-lock.yaml`, `scripts/tsconfig.json`, `scripts/.gitignore` — install and type-check support

### examples/ (full scenario recipes)
- `examples/ethers-v6-signer.ts` — alternative ethers v6 signer
- `examples/kms-signer-pattern.md` — production KMS/hardware signer boundary
- `examples/e2e-auth-then-atlantic.ts` — auth → Atlantic submit composition
- `examples/README.md` — recipe index
- `examples/package.json`, `examples/pnpm-lock.yaml`, `examples/tsconfig.json`, `examples/.gitignore` — install support for the Atlantic composition example

### references/ (deep docs, load on demand)
- `references/eip712-signing.md` — signer-specific notes
- `references/channel-binding.md` — bearer/cookie channel enforcement
- `references/api-keys-management.md` — project model and key creation

## Next skill

Once you have an API key, load the product-specific skill: `atlantic-api`, `storage-proof-api`, `data-processor-api`, `data-structure-indexer-api`, `satellite-contracts`, or `data-processor`.
