#!/usr/bin/env -S npx tsx
// List or read API keys for a project.
// Inputs: optional --project-id <id> [--list]
// Default: emits first key as {schemaVersion, apiKey, type, projectId}.
// With --list: emits {schemaVersion, projectId, apiKeys: [...]}.

import { parseArgs } from 'node:util';
import { HcloudClient } from '@herodotus_dev/hcloud';

const DEFAULT_BASE = 'https://auth-billing.api.herodotus.cloud';
const SCHEMA_VERSION = 1;

function fail(code: string, message: string, details?: unknown, exit = 1): never {
  process.stderr.write(JSON.stringify({ schemaVersion: SCHEMA_VERSION, error: { code, message, details } }) + '\n');
  process.exit(exit);
}

function printHelp() {
  process.stderr.write(
    `Usage: get-api-key.ts [--project-id <id>] [--list] [--auth-base-url <url>] [--pretty]\n\n` +
      `Reads or lists API keys through the hcloud SDK using active stored wallet credentials.\n\n` +
      `Example:\n  ./get-api-key.ts | jq -r .apiKey\n`,
  );
}

async function main() {
  const { values } = parseArgs({
    options: {
      'project-id': { type: 'string' },
      list: { type: 'boolean', default: false },
      'auth-base-url': { type: 'string' },
      pretty: { type: 'boolean', default: false },
      help: { type: 'boolean', default: false },
    },
  });
  if (values.help) {
    printHelp();
    return;
  }

  const base = (values['auth-base-url'] as string | undefined) ?? DEFAULT_BASE;
  const client = new HcloudClient({ baseUrls: { 'auth-billing': base } });
  const identity = await client.auth.whoami();
  const projectId = (values['project-id'] as string | undefined) ?? identity?.projectId;
  if (!projectId) fail('MISSING_ARG', '--project-id is required when no active wallet is stored');

  const data = await client.auth.apiKeys.list({ projectId, limit: 50, offset: 0 }).catch((error: unknown) =>
    fail('API_KEYS_FAILED', (error as Error).message, typeof (error as { toJSON?: () => unknown }).toJSON === 'function' ? (error as { toJSON: () => unknown }).toJSON() : error),
  );

  let out: object;
  if (values.list) {
    out = { schemaVersion: SCHEMA_VERSION, projectId, apiKeys: data.data };
  } else {
    const first = data.data[0];
    if (!first) fail('NO_API_KEY', 'no API keys found for project', data);
    out = { schemaVersion: SCHEMA_VERSION, projectId, apiKey: first.apiKey, type: first.type };
  }
  process.stdout.write(JSON.stringify(out, null, values.pretty ? 2 : 0) + '\n');
}

main().catch((e) => fail('UNEXPECTED', (e as Error).message, (e as Error).stack, 3));
