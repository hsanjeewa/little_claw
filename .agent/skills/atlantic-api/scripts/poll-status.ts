#!/usr/bin/env -S npx tsx
import { parseArgs } from 'node:util';
import { HcloudClient, waitForQuery } from '@herodotus_dev/hcloud';
import { DEFAULT_ATLANTIC_BASE, SCHEMA_VERSION, apiKeyFromEnv, fail, sdkError, writeJson } from './common.ts';

function help(): void {
  process.stderr.write(
    `Usage: poll-status.ts --api-key <key> --query-id <id> [--timeout-s 600] [--interval-s 5] [--base-url <url>] [--pretty]\n\n` +
      `Polls GET /atlantic-query/:id until a terminal status is reached or timeout expires.\n`,
  );
}

async function main(): Promise<void> {
  const { values } = parseArgs({
    options: {
      'api-key': { type: 'string' },
      'query-id': { type: 'string' },
      'timeout-s': { type: 'string', default: '600' },
      'interval-s': { type: 'string', default: '5' },
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
  const timeoutMs = Number(values['timeout-s']) * 1000;
  const intervalMs = Number(values['interval-s']) * 1000;
  if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) fail('INVALID_ARG', '--timeout-s must be positive');
  if (!Number.isFinite(intervalMs) || intervalMs <= 0) fail('INVALID_ARG', '--interval-s must be positive');

  const client = new HcloudClient({ baseUrls: { atlantic: base }, ...(apiKey ? { apiKey } : {}) });
  try {
    const data = await waitForQuery(client.atlantic, queryId, { timeoutMs, intervalMs });
    writeJson(
      {
        schemaVersion: SCHEMA_VERSION,
        queryId,
        status: data.atlanticQuery.status,
        terminal: true,
        error: data.atlanticQuery.errorReason ? { message: data.atlanticQuery.errorReason } : undefined,
        atlanticQuery: data.atlanticQuery,
        metadataUrls: data.metadataUrls,
        observedStatuses: data.observedStatuses,
      },
      values.pretty as boolean,
    );
  } catch (error) {
    const err = sdkError(error);
    fail(err.code, err.message, err.details, err.exit);
  }
}

main().catch((error) => fail('UNEXPECTED', (error as Error).message, (error as Error).stack, 3));
