#!/usr/bin/env -S npx tsx
import { parseArgs } from 'node:util';
import { HcloudClient } from '@herodotus_dev/hcloud';
import {
  DEFAULT_ATLANTIC_BASE,
  SCHEMA_VERSION,
  apiKeyFromEnv,
  fail,
  sdkError,
  submitInputFromFlags,
  writeJson,
} from './common.ts';

function help(): void {
  process.stderr.write(
    `Usage: submit-query.ts [--api-key <key>] (--body-file <path> | --body-json <path>) [--base-url <url>] [--pretty]\n\n` +
      `Submits POST /atlantic-query and emits {schemaVersion, queryId}. JSON body files may contain {fields, files}.\n` +
      `API key resolution follows hcloud SDK order: --api-key, HCLOUD_API_KEY, HERODOTUS_API_KEY, ATLANTIC_API_KEY, then stored wallet credentials.\n` +
      `Example descriptor: {"fields":{"layout":"auto","cairoVm":"python","cairoVersion":"cairo0","result":"PROOF_GENERATION","mockFactHash":"false","declaredJobSize":"S"},"files":{"programFile":"compiled.json","inputFile":"input.json"}}\n`,
  );
}

async function main(): Promise<void> {
  const { values } = parseArgs({
    options: {
      'api-key': { type: 'string' },
      'body-file': { type: 'string' },
      'body-json': { type: 'string' },
      'file-field': { type: 'string', default: 'inputFile' },
      'base-url': { type: 'string' },
      pretty: { type: 'boolean', default: false },
      help: { type: 'boolean', default: false },
    },
  });
  if (values.help) return help();

  const apiKey = (values['api-key'] as string | undefined) ?? apiKeyFromEnv();
  const base = (values['base-url'] as string | undefined) ?? DEFAULT_ATLANTIC_BASE;
  const input = await submitInputFromFlags(
    values['body-file'] as string | undefined,
    values['body-json'] as string | undefined,
    values['file-field'] as string,
  );
  const client = new HcloudClient({ baseUrls: { atlantic: base }, ...(apiKey ? { apiKey } : {}) });

  try {
    const data = await client.atlantic.submitQuery(input);
    writeJson({ schemaVersion: SCHEMA_VERSION, queryId: data.atlanticQueryId, atlanticQueryId: data.atlanticQueryId }, values.pretty as boolean);
  } catch (error) {
    const err = sdkError(error);
    if (err.code === 'payment_challenge') {
      fail('PAYMENT_REQUIRED', 'Atlantic requires x402 payment; retry with scripts/x402-pay.ts using the same body file', err.details, err.exit);
    }
    fail(err.code, err.message, err.details, err.exit);
  }
}

main().catch((error) => fail('UNEXPECTED', (error as Error).message, (error as Error).stack, 3));
