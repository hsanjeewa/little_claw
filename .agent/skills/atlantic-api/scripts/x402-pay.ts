#!/usr/bin/env -S npx tsx
import { parseArgs } from 'node:util';
import { HcloudClient, createPrivateKeyPaymentAdapter } from '@herodotus_dev/hcloud';
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
    `Usage: x402-pay.ts (--private-key <0x...> | WALLET_PRIVATE_KEY=0x...) (--body-file <path> | --body-json <path>) [--api-key <key>] [--base-url <url>] [--pretty]\n\n` +
      `Submits POST /atlantic-query, signs the x402 v2 payment challenge on 402, retries with PAYMENT-SIGNATURE, and emits {schemaVersion, queryId, paymentTx, alreadyProcessed, network, payer}.\n`,
  );
}

async function main(): Promise<void> {
  const { values } = parseArgs({
    options: {
      'private-key': { type: 'string' },
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

  const privateKey = (values['private-key'] as string | undefined) ?? process.env.WALLET_PRIVATE_KEY;
  if (!privateKey) fail('MISSING_ENV', '--private-key or WALLET_PRIVATE_KEY is required');
  if (!/^0x[0-9a-fA-F]{64}$/.test(privateKey)) fail('INVALID_PRIVATE_KEY', 'private key must be 0x-prefixed 32-byte hex');
  const apiKey = (values['api-key'] as string | undefined) ?? apiKeyFromEnv();
  const base = (values['base-url'] as string | undefined) ?? DEFAULT_ATLANTIC_BASE;

  const input = await submitInputFromFlags(values['body-file'] as string | undefined, values['body-json'] as string | undefined, values['file-field'] as string);
  const paymentAdapter = createPrivateKeyPaymentAdapter({ privateKey: privateKey as `0x${string}` });
  const client = new HcloudClient({ baseUrls: { atlantic: base }, ...(apiKey ? { apiKey } : {}), paymentAdapter });

  try {
    const data = await client.atlantic.submitQuery(input, { anonymousPayment: !apiKey });
    writeJson(
      {
        schemaVersion: SCHEMA_VERSION,
        queryId: data.atlanticQueryId,
        atlanticQueryId: data.atlanticQueryId,
        paymentTx: data.payment?.transaction,
        alreadyProcessed: Boolean(data.payment?.alreadyProcessed),
        network: data.payment?.network,
        payer: data.payment?.payer,
      },
      values.pretty as boolean,
    );
  } catch (error) {
    const err = sdkError(error);
    fail(err.code, err.message, err.details, err.exit);
  }
}

main().catch((error) => fail('UNEXPECTED', (error as Error).message, (error as Error).stack, 3));
