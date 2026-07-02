#!/usr/bin/env -S npx tsx
import { mkdir } from 'node:fs/promises';
import path from 'node:path';
import { parseArgs } from 'node:util';
import { HcloudClient } from '@herodotus_dev/hcloud';
import { DEFAULT_ATLANTIC_BASE, SCHEMA_VERSION, apiKeyFromEnv, atomicWrite, fail, responseError, sdkError, writeJson } from './common.ts';

function help(): void {
  process.stderr.write(
    `Usage: download-artifacts.ts --api-key <key> --query-id <id> [--out-dir ./artifacts/<id>] [--base-url <url>] [--pretty]\n\n` +
      `Reads metadataUrls from GET /atlantic-query/:id and downloads each artifact with atomic writes.\n`,
  );
}

function artifactType(url: string): string {
  const clean = new URL(url).pathname.split('/').pop() ?? url;
  return clean.split('.')[0] || 'artifact';
}

async function main(): Promise<void> {
  const { values } = parseArgs({
    options: {
      'api-key': { type: 'string' },
      'query-id': { type: 'string' },
      'out-dir': { type: 'string' },
      'base-url': { type: 'string' },
      pretty: { type: 'boolean', default: false },
      help: { type: 'boolean', default: false },
    },
  });
  if (values.help) return help();

  const queryId = values['query-id'] as string | undefined;
  if (!queryId) fail('MISSING_ARG', '--query-id is required');
  const apiKey = (values['api-key'] as string | undefined) ?? apiKeyFromEnv();
  const base = (values['base-url'] as string | undefined) ?? DEFAULT_ATLANTIC_BASE;
  const outDir = (values['out-dir'] as string | undefined) ?? path.join('artifacts', queryId);
  await mkdir(outDir, { recursive: true });

  const client = new HcloudClient({ baseUrls: { atlantic: base }, ...(apiKey ? { apiKey } : {}) });
  const details = await client.atlantic.getQuery(queryId).catch((error: unknown) => {
    const err = sdkError(error);
    fail(err.code, err.message, err.details, err.exit);
  });
  const urls = details.metadataUrls;
  const artifacts = [];
  for (const url of urls) {
    const res = await fetch(url).catch((error) => fail('NETWORK_ERROR', `artifact download failed: ${(error as Error).message}`, { url }, 2));
    if (!res.ok) {
      const err = await responseError(res);
      fail(err.code, err.message, { url, details: err.details }, err.exit);
    }
    const data = Buffer.from(await res.arrayBuffer());
    const name = decodeURIComponent(new URL(url).pathname.split('/').pop() ?? `${artifactType(url)}.bin`);
    const filePath = path.join(outDir, name);
    await atomicWrite(filePath, data);
    artifacts.push({ type: artifactType(url), path: filePath, size: data.byteLength });
  }

  writeJson({ schemaVersion: SCHEMA_VERSION, queryId, outDir, artifacts }, values.pretty as boolean);
}

main().catch((error) => fail('UNEXPECTED', (error as Error).message, (error as Error).stack, 3));
