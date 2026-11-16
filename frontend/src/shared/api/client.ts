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

/** A progression, as the contract defines it: what the screen sends and keeps. */
export type Progression = components['schemas']['Progression'];
/** One slot of a progression. */
export type Slot = components['schemas']['Slot'];

/**
 * The contract's type with every string widened to string. A server newer than
 * the screen may answer an identifier the contract does not list yet; the
 * screen shows it as it comes instead of refusing the whole answer.
 */
type Open<T> = T extends string
  ? string
  : T extends readonly (infer U)[]
    ? Open<U>[]
    : T extends object
      ? { [K in keyof T]: Open<T[K]> }
      : T;

/** The readings of a progression's slots, in the order of the slots. */
export type ProgressionReading = Open<components['schemas']['ProgressionReading']>;
/** The labels of one slot. */
export type SlotReading = Open<components['schemas']['SlotReading']>;
/** A problem document (RFC 9457). */
export type Problem = components['schemas']['Problem'];

/** One note to play, as the contract defines it. */
export type Note = components['schemas']['Note'];

/** The answer to a reading: the readings, or the problem the server named. */
export type ReadingAnswer =
  { ok: true; reading: ProgressionReading } | { ok: false; problem: Problem };

/**
 * Reads every slot of a progression. Resolves with the problem when the server
 * refuses the progression; rejects when the server cannot be reached or answers
 * with a body the contract does not describe.
 */
export async function readProgression(
  progression: Progression,
  signal?: AbortSignal,
): Promise<ReadingAnswer> {
  const response = await fetch(`${API_PREFIX}/progressions/readings`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(progression),
    signal: signal ?? null,
  });
  const body: unknown = await response.json();
  if (response.ok) {
    return { ok: true, reading: parseProgressionReading(body) };
  }
  return { ok: false, problem: parseProblem(body) };
}

type Fields = Record<string, unknown>;

function isFields(value: unknown): value is Fields {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

class UnexpectedBody extends Error {
  constructor(where: string) {
    super(`POST /progressions/readings answered with an unexpected body (${where})`);
  }
}

function fieldsAt(value: unknown, where: string): Fields {
  if (!isFields(value)) {
    throw new UnexpectedBody(where);
  }
  return value;
}

function stringAt(fields: Fields, name: string, where: string): string {
  const value = fields[name];
  if (typeof value !== 'string') {
    throw new UnexpectedBody(`${where}.${name}`);
  }
  return value;
}

function arrayAt(fields: Fields, name: string, where: string): unknown[] {
  const value = fields[name];
  if (!Array.isArray(value)) {
    throw new UnexpectedBody(`${where}.${name}`);
  }
  return value;
}

/** A label's fields and its basis, which every label carries. */
function labelAt(value: unknown, where: string): { fields: Fields; basis: string } {
  const fields = fieldsAt(value, where);
  return { fields, basis: stringAt(fields, 'basis', where) };
}

/** A label whose one other field is a string: chord, roman, resolution, cadence, roleFit. */
function oneStringLabel<N extends string>(value: unknown, name: N, where: string) {
  const { fields, basis } = labelAt(value, where);
  return { basis, [name]: stringAt(fields, name, where) } as { basis: string } & Record<N, string>;
}

function parseChordReadingLabel(value: unknown, where: string) {
  const { fields, basis } = labelAt(value, where);
  const modes = arrayAt(fields, 'modes', where).map((mode, i) => {
    if (typeof mode !== 'string') {
      throw new UnexpectedBody(`${where}.modes[${i}]`);
    }
    return mode;
  });
  const fromParallelMinor = fields.fromParallelMinor;
  if (typeof fromParallelMinor !== 'boolean') {
    throw new UnexpectedBody(`${where}.fromParallelMinor`);
  }
  const label = {
    basis,
    kind: stringAt(fields, 'kind', where),
    function: stringAt(fields, 'function', where),
    modes,
    fromParallelMinor,
  };
  if (!('target' in fields)) {
    return label;
  }
  const target = fields.target;
  if (typeof target !== 'number' || !Number.isInteger(target)) {
    throw new UnexpectedBody(`${where}.target`);
  }
  return { ...label, target };
}

function parseTensionCautionsLabel(value: unknown, where: string) {
  const { fields, basis } = labelAt(value, where);
  const cautions = arrayAt(fields, 'cautions', where).map((caution, i) => {
    const at = `${where}.cautions[${i}]`;
    const c = fieldsAt(caution, at);
    return { tension: stringAt(c, 'tension', at), chordTone: stringAt(c, 'chordTone', at) };
  });
  return { basis, cautions };
}

function parseSlotReading(value: unknown, where: string): SlotReading {
  const fields = fieldsAt(value, where);
  const slot: SlotReading = {};
  const at = (name: string) => `${where}.${name}`;
  if ('chord' in fields) slot.chord = oneStringLabel(fields.chord, 'name', at('chord'));
  if ('roman' in fields) slot.roman = oneStringLabel(fields.roman, 'numeral', at('roman'));
  if ('reading' in fields) slot.reading = parseChordReadingLabel(fields.reading, at('reading'));
  if ('resolution' in fields) {
    slot.resolution = oneStringLabel(fields.resolution, 'resolution', at('resolution'));
  }
  if ('cadence' in fields) {
    slot.cadence = oneStringLabel(fields.cadence, 'cadence', at('cadence'));
  }
  if ('tensionCautions' in fields) {
    slot.tensionCautions = parseTensionCautionsLabel(fields.tensionCautions, at('tensionCautions'));
  }
  if ('roleFit' in fields) slot.roleFit = oneStringLabel(fields.roleFit, 'fit', at('roleFit'));
  return slot;
}

/**
 * Turns an unknown body into a ProgressionReading, or throws: the one check at
 * the boundary. Identifiers are checked to be strings, not to be listed in the
 * contract (see Open).
 */
export function parseProgressionReading(body: unknown): ProgressionReading {
  const fields = fieldsAt(body, 'body');
  return {
    slots: arrayAt(fields, 'slots', 'body').map((slot, i) => parseSlotReading(slot, `slots[${i}]`)),
  };
}

/** Turns an unknown body into a Problem, or throws. */
export function parseProblem(body: unknown): Problem {
  const fields = fieldsAt(body, 'problem');
  const title = stringAt(fields, 'title', 'problem');
  const status = fields.status;
  if (typeof status !== 'number' || !Number.isInteger(status)) {
    throw new UnexpectedBody('problem.status');
  }
  const problem: Problem = { title, status };
  if ('detail' in fields) {
    problem.detail = stringAt(fields, 'detail', 'problem');
  }
  return problem;
}
