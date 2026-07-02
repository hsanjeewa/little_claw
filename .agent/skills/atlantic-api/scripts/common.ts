import { mkdir, readFile, rename, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import type { SubmitQueryInput } from '@herodotus_dev/hcloud';

export const SCHEMA_VERSION = 1;
export const DEFAULT_ATLANTIC_BASE = 'https://atlantic.api.herodotus.cloud';

export type JsonRecord = Record<string, unknown>;

export function fail(code: string, message: string, details?: unknown, exit = 1): never {
  process.stderr.write(JSON.stringify({ schemaVersion: SCHEMA_VERSION, error: { code, message, details } }) + '\n');
  process.exit(exit);
}

export function writeJson(value: unknown, pretty?: boolean): void {
  process.stdout.write(JSON.stringify(value, null, pretty ? 2 : 0) + '\n');
}

export function apiKeyFromEnv(): string | undefined {
  return process.env.HCLOUD_API_KEY ?? process.env.HERODOTUS_API_KEY ?? process.env.ATLANTIC_API_KEY;
}

export function sdkError(error: unknown): { code: string; message: string; details: unknown; exit: number } {
  const value = error as { code?: string; status?: number; kind?: string; message?: string; details?: unknown; toJSON?: () => unknown };
  const status = typeof value.status === 'number' ? value.status : undefined;
  const code = value.code ?? value.kind ?? 'SDK_ERROR';
  return {
    code,
    message: value.message ?? 'Atlantic SDK request failed',
    details: typeof value.toJSON === 'function' ? value.toJSON() : value.details ?? error,
    exit: status && status >= 500 ? 3 : value.kind === 'timeout' ? 2 : 1,
  };
}

export async function submitInputFromFlags(bodyFile?: string, bodyJson?: string, fileField = 'inputFile'): Promise<SubmitQueryInput> {
  const source = bodyJson ?? bodyFile;
  if (!source) fail('MISSING_ARG', 'provide --body-file <path> or --body-json <path>');

  const raw = await readFile(source);
  let parsed: JsonRecord;
  try {
    parsed = JSON.parse(raw.toString('utf8')) as JsonRecord;
  } catch {
    return {
      declaredJobSize: 'S',
      [fileField]: await fileFor(source),
    } as unknown as SubmitQueryInput;
  }

  const input: Record<string, unknown> = {};

  if (parsed.fields && typeof parsed.fields === 'object') {
    Object.assign(input, parsed.fields);
  }
  if (parsed.files && typeof parsed.files === 'object') {
    for (const [fieldName, filePathValue] of Object.entries(parsed.files as Record<string, string>)) {
      input[fieldName] = await fileFor(path.resolve(path.dirname(source), filePathValue));
    }
  }
  if (!parsed.fields && !parsed.files) {
    for (const [key, value] of Object.entries(parsed)) {
      if (key.endsWith('File') && typeof value === 'string') {
        input[key] = await fileFor(path.resolve(path.dirname(source), value));
      } else if (value !== undefined && value !== null) {
        input[key] = value;
      }
    }
  }

  return input as unknown as SubmitQueryInput;
}

function contentTypeFor(filePath: string): string {
  if (filePath.endsWith('.zip')) return 'application/zip';
  if (filePath.endsWith('.txt') || filePath.endsWith('.cairo')) return 'text/plain';
  return 'application/json';
}

async function fileFor(filePath: string): Promise<File> {
  const data = await readFile(filePath);
  const arrayBuffer = new ArrayBuffer(data.byteLength);
  new Uint8Array(arrayBuffer).set(data);
  return new File([arrayBuffer], path.basename(filePath), { type: contentTypeFor(filePath) });
}

export function parseMaybeJson(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

export async function responseError(res: Response): Promise<{ code: string; message: string; details: unknown; exit: number }> {
  const bodyText = await res.text();
  const details = parseMaybeJson(bodyText);
  const code =
    typeof details === 'object' && details !== null && 'message' in details && typeof (details as { message?: unknown }).message === 'string'
      ? String((details as { message: string }).message)
      : `HTTP_${res.status}`;
  return {
    code,
    message: `request returned ${res.status}`,
    details,
    exit: res.status >= 500 ? 3 : 1,
  };
}

export async function atomicWrite(filePath: string, data: Buffer): Promise<void> {
  const tmpDir = path.join(path.dirname(filePath), '.tmp');
  await mkdir(tmpDir, { recursive: true });
  const tmpPath = path.join(tmpDir, `${path.basename(filePath)}.${process.pid}.${Date.now()}`);
  await writeFile(tmpPath, data);
  await rename(tmpPath, filePath);
  await rm(tmpDir, { recursive: true, force: true });
}
