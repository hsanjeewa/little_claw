# Artifact Types

`GET /atlantic-query/:id` returns `metadataUrls`, a list of downloadable object URLs under `queries/<queryId>/`.

Common file names:

- `program.<cairoVersion>.json` - uploaded or normalized program artifact.
- `input.<cairoVersion>.json` or `input.<cairoVersion>.txt` - input artifact.
- `pie.<cairoVersion>.zip` - PIE archive.
- `proof.json` - generated proof.
- `metadata.json` - trace/proof metadata.
- `calldata` - verifier calldata.

Use `scripts/download-artifacts.ts` to download all URLs. It writes to `artifacts/<queryId>` by default and uses a temporary directory plus rename so interrupted downloads do not leave a completed-looking file.
