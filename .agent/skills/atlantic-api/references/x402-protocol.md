# x402 Protocol

Atlantic accepts x402 v2 HTTP payments on `POST /atlantic-query` using the public headers `PAYMENT-REQUIRED`, `PAYMENT-SIGNATURE`, and `PAYMENT-RESPONSE`.

## Flows

1. API key with sufficient credits: submit normally. Do not send `PAYMENT-SIGNATURE` preemptively.
2. API key with depleted credits: server returns `402`; payment tops up the project balance, the current query proceeds, and leftover credits remain on the project.
3. Anonymous wallet: documented as pay-once, use-once with no project balance. Current server source in `/Users/sebastiankusnierz/repos/atlantic/src/routes/atlantic/submit-query/index.ts` still requires an API key before x402 challenge generation, so treat anonymous availability as deployment/documentation-dependent until OpenAPI/docs confirm it.

## Wire Recipe

1. Submit `POST /atlantic-query` with the normal multipart body. Include the API key for the API-key flow.
2. On `402`, parse base64 JSON from `PAYMENT-REQUIRED`; the same payload may also be present as `paymentRequired` in the response body.
3. Pick one `accepts[]` requirement. Read `payTo`, `asset`, `network`, `amount`, `maxTimeoutSeconds`, and `extra` from that requirement; do not hardcode them.
4. Sign EIP-3009 `TransferWithAuthorization` typed data with `from = wallet`, `to = payTo`, `value = amount`, `validAfter = 0`, `validBefore = now + maxTimeoutSeconds`, and a fresh 32-byte nonce. Domain `name` and `version` come from `requirement.extra`; `chainId` comes from the x402 network string.
5. Retry the same `POST /atlantic-query` body with `PAYMENT-SIGNATURE: <base64 JSON>`.
6. Read `PAYMENT-RESPONSE` on success. Treat `alreadyProcessed: true` as success and do not pay again.

## Payment Payload

```json
{
  "x402Version": 2,
  "accepted": "<the chosen PaymentRequirement, verbatim>",
  "payload": {
    "signature": "0x...",
    "authorization": {
      "from": "0x...",
      "to": "0x...",
      "value": "1000000",
      "validAfter": "0",
      "validBefore": "1735689600",
      "nonce": "0x...32 bytes..."
    }
  }
}
```

## Guardrails

- Use only `POST /atlantic-query` and the public x402 headers. Do not invent payment endpoints.
- Do not hardcode `payTo`, `asset`, `network`, or `amount`.
- Do not preemptively send `PAYMENT-SIGNATURE` on the API-key flow.
- Do not reuse a signature after successful settlement.
- Do not send `dedupId` or `bucketId` on anonymous flow.
- Do not assume anonymous payments leave a credit balance.
- Match only on agent-visible error codes.

Use `scripts/x402-pay.ts` to execute this flow without copying code.
