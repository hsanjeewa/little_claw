# Channel Binding

Herodotus auth tokens are bound to the transport channel used at issuance.

- `channel: "bearer"` tokens must be sent as `Authorization: Bearer <accessToken>`.
- Cookie-issued tokens must be sent via cookies.
- A cookie token sent as bearer, or a bearer token sent as a cookie, is rejected server-side with channel mismatch behavior.

Do not work around this. If a workflow needs both browser and headless access, create one cookie session and one bearer session.

## Curl Recipe

```bash
WALLET=0x...
BASE=https://auth-billing.api.herodotus.cloud

curl -s "$BASE/auth/web3/challenge?wallet=$WALLET" > challenge.json

# Sign challenge.json's eip712 payload with your wallet out of band.
SIG=0x...

curl -s -X POST "$BASE/auth/web3/session" \
  -H 'content-type: application/json' \
  -d "{\"wallet\":\"$WALLET\",\"challengeToken\":\"$(jq -r .challengeToken challenge.json)\",\"signature\":\"$SIG\",\"channel\":\"bearer\"}" \
  > session.json

ACCESS=$(jq -r .accessToken session.json)
PROJECT=$(jq -r .selectedProject session.json)

curl -s "$BASE/api-keys?projectId=$PROJECT&limit=10&offset=0" \
  -H "authorization: Bearer $ACCESS" \
  | jq -r '.data[0].apiKey'
```
