# API Keys Management

New wallets are auto-provisioned with a Personal project and one active API key. Prefer reading the existing key before minting another one.

## Read Keys

```http
GET /api-keys?projectId=<selectedProject>&limit=10&offset=0
Authorization: Bearer <accessToken>
```

Read `data[0].apiKey` for the default key. Use `scripts/get-api-key.ts --list` when you need to inspect all keys for the project.

## Create a Key

```http
POST /api-keys
Authorization: Bearer <accessToken>
Content-Type: application/json
```

Body:

```json
{
  "projectId": "project-id-from-session",
  "type": {
    "name": "agent",
    "color": "blue"
  }
}
```

The server owns key status and client ownership. Do not assume a project id; use `selectedProject` from the bearer session response.
