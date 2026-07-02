#!/usr/bin/env -S npx tsx
// Refresh the active wallet bearer session persisted by the hcloud SDK.
// Emits {schemaVersion, refreshed, sessionExpiresAt?}.

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
    `Usage: refresh.ts [--auth-base-url <url>] [--pretty]\n\n` +
      `Rotates the active wallet bearer pair stored under ~/.hcloud through the hcloud SDK.\n\n` +
      `Example:\n  ./refresh.ts --pretty\n`,
  );
}

async function main() {
  const { values } = parseArgs({
    options: {
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
  await client.auth.refresh().catch((error: unknown) =>
    fail('REFRESH_FAILED', (error as Error).message, typeof (error as { toJSON?: () => unknown }).toJSON === 'function' ? (error as { toJSON: () => unknown }).toJSON() : error),
  );
  const identity = await client.auth.whoami();

  const out = { schemaVersion: SCHEMA_VERSION, refreshed: true, sessionExpiresAt: identity?.sessionExpiresAt };
  process.stdout.write(JSON.stringify(out, null, values.pretty ? 2 : 0) + '\n');
}

main().catch((e) => fail('UNEXPECTED', (e as Error).message, (e as Error).stack, 3));
