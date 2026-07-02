#!/usr/bin/env -S npx tsx
// Wallet auth → persisted hcloud credentials.
// Reads HCLOUD_WALLET_PRIVATE_KEY or WALLET_PRIVATE_KEY from env, optionally --auth-base-url.
// Emits {schemaVersion, wallet, projectId, selectedProject, apiKey} on stdout.

import { parseArgs } from 'node:util';
import { HcloudClient, privateKeySigner } from '@herodotus_dev/hcloud';

const DEFAULT_BASE = 'https://auth-billing.api.herodotus.cloud';
const SCHEMA_VERSION = 1;

function fail(code: string, message: string, details?: unknown, exit = 1): never {
  process.stderr.write(JSON.stringify({ schemaVersion: SCHEMA_VERSION, error: { code, message, details } }) + '\n');
  process.exit(exit);
}

function log(msg: string) {
  process.stderr.write(`[auth] ${msg}\n`);
}

function printHelp() {
  process.stderr.write(
    `Usage: auth.ts [--auth-base-url <url>] [--pretty] [--help]\n\n` +
      `Performs EIP-712 wallet auth through the hcloud SDK, persists credentials under ~/.hcloud, and returns wallet + API key JSON.\n\n` +
      `Env:\n  WALLET_PRIVATE_KEY  required, 0x-prefixed EVM private key\n\n` +
      `Example:\n  WALLET_PRIVATE_KEY=0x... ./auth.ts | jq -r .apiKey\n`,
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
  const pk = process.env.HCLOUD_WALLET_PRIVATE_KEY ?? process.env.WALLET_PRIVATE_KEY;
  if (!pk) fail('MISSING_ENV', 'WALLET_PRIVATE_KEY env var is required');
  if (!/^0x[0-9a-fA-F]{64}$/.test(pk!)) fail('INVALID_PRIVATE_KEY', 'WALLET_PRIVATE_KEY must be 0x-prefixed 32-byte hex');

  const signer = privateKeySigner(pk as `0x${string}`);
  const wallet = await signer.getAddress();
  log(`wallet=${wallet} authBase=${base}`);

  const client = new HcloudClient({
    baseUrls: { 'auth-billing': base },
    signer,
  });

  const result = await client.auth.login().catch((error: unknown) =>
    fail('AUTH_FAILED', (error as Error).message, typeof (error as { toJSON?: () => unknown }).toJSON === 'function' ? (error as { toJSON: () => unknown }).toJSON() : error),
  );

  const out = {
    schemaVersion: SCHEMA_VERSION,
    wallet: result.wallet,
    projectId: result.projectId,
    selectedProject: result.projectId,
    apiKey: result.apiKey,
  };
  process.stdout.write(JSON.stringify(out, null, values.pretty ? 2 : 0) + '\n');
}

main().catch((e) => fail('UNEXPECTED', (e as Error).message, (e as Error).stack, 3));
