# x402 Error Codes

| Agent-visible code | Status | When |
|---|---:|---|
| `X402_NOT_ENABLED` | 503 | x402 disabled in this Atlantic deployment; back off, do not retry. |
| `MISSING_API_KEY` | 400 | Anonymous flow disabled or API key absent where required; fall back to `herodotus-auth`. |
| `X402_CHALLENGE_FAILED` | 502 | Upstream challenge build failed; transient, retry with backoff. |
| `X402_SETTLEMENT_FAILED` | 402 | Verify/settle rejected; fetch a new challenge. |
| `X402_SETTLEMENT_RESPONSE_INVALID` | 502 | Settlement may have succeeded but response schema failed; check query status before retrying. |
| `X402_SERVICE_AUTH_NOT_CONFIGURED` | 500 | Server misconfiguration; surface to the user. |
| `WALLET_FLOW_DEDUP_ID_NOT_SUPPORTED` | 400 | Drop `dedupId` from anonymous flow and resubmit. |
| `WALLET_FLOW_BUCKET_NOT_SUPPORTED` | 400 | Drop `bucketId` from anonymous flow and resubmit. |
| `WALLET_FLOW_NOT_RETRIABLE` | 400 | Anonymous query cannot be retried; submit a new one. |

Current `auth-billing` source also raises lower-level `BadRequest`, `BillingNotConfigured`, and `InvalidChallenge` errors inside settlement. Treat those as implementation details unless surfaced by Atlantic as stable agent-visible codes.
