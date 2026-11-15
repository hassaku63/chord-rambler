import type { components } from './schema.gen';

/** The server's status and version, as the contract defines it. */
export type Health = components['schemas']['Health'];

/** The path every operation lives under: the server URL of docs/api/openapi.yaml. */
const API_PREFIX = '/api/v1';

/**
 * Reads the server's status and version. Rejects when the server does not
 * answer 200 or answers with a body the contract does not describe.
 */
export async function getHealth(signal?: AbortSignal): Promise<Health> {
  const response = await fetch(`${API_PREFIX}/health`, { signal: signal ?? null });
  if (!response.ok) {
    throw new Error(`GET /health answered ${response.status}`);
  }
  return parseHealth(await response.json());
}

/** Turns an unknown body into a Health, or throws: the one check at the boundary. */
export function parseHealth(body: unknown): Health {
  if (
    typeof body === 'object' &&
    body !== null &&
    'status' in body &&
    body.status === 'ok' &&
    'version' in body &&
    typeof body.version === 'string'
  ) {
    return { status: body.status, version: body.version };
  }
  throw new Error('GET /health answered with an unexpected body');
}
