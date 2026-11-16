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

/**
 * Open for each member of a tagged union, with the tag kept as the contract
 * lists it, so the union still narrows on the tag.
 */
type OpenTagged<T, Tag extends keyof T> = T extends unknown
  ? { [K in keyof T]: K extends Tag ? T[K] : Open<T[K]> }
  : never;

/** The labels of one slot: those of a slot with a chord, or of a slot with no chord. */
export type SlotReading = OpenTagged<components['schemas']['SlotReading'], 'content'>;
/** The labels of a slot with a chord: the chord's labels are always there. */
type ChordSlotReading = Extract<SlotReading, { content: 'chord' }>;
/** The readings of a progression's slots, in the order of the slots. */
export type ProgressionReading = Omit<components['schemas']['ProgressionReading'], 'slots'> & {
  slots: SlotReading[];
};
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
  const { ok, body } = await post('/progressions/readings', progression, signal);
  if (ok) {
    return { ok: true, reading: parseProgressionReading(body) };
  }
  return { ok: false, problem: parseProblem(body) };
}

/** The answer to a request for notes: the notes, or the problem the server named. */
export type NotesAnswer = { ok: true; notes: Note[] } | { ok: false; problem: Problem };

/**
 * Reads the notes that play a progression. Resolves with the problem when the
 * server refuses the progression; rejects when the server cannot be reached or
 * answers with a body the contract does not describe.
 */
export async function getProgressionNotes(
  progression: Progression,
  signal?: AbortSignal,
): Promise<NotesAnswer> {
  const { ok, body } = await post('/progressions/notes', progression, signal);
  if (ok) {
    return { ok: true, notes: parseProgressionNotes(body) };
  }
  return { ok: false, problem: parseProblem(body) };
}

/** The groups of the candidates for the slot after the last, in the contract's order. */
export type CandidateGroups = Open<components['schemas']['CandidateGroups']>;
/** One candidate and its labels. */
export type Candidate = Open<components['schemas']['Candidate']>;

/** The answer to a request for candidates: the groups, or the problem the server named. */
export type CandidatesAnswer =
  { ok: true; groups: CandidateGroups['groups'] } | { ok: false; problem: Problem };

/**
 * Asks for the candidates for a slot of the beats given, added after the last.
 * Resolves with the problem when the server refuses; rejects when the server
 * cannot be reached or answers with a body the contract does not describe.
 */
export async function getNextCandidates(
  progression: Progression,
  beats: number,
  signal?: AbortSignal,
): Promise<CandidatesAnswer> {
  const { ok, body } = await post('/progressions/next-candidates', { progression, beats }, signal);
  if (ok) {
    return { ok: true, groups: parseCandidateGroups(body).groups };
  }
  return { ok: false, problem: parseProblem(body) };
}

/** The groups of the candidates for a slot to fill, in the contract's order. */
export type SlotCandidateGroups = Open<components['schemas']['SlotCandidateGroups']>;
/** One candidate for a slot to fill, with the labels that look at the next slot and the role. */
export type SlotCandidate = Open<components['schemas']['SlotCandidate']>;

/** The answer to a request for the candidates for a slot: the groups, or the problem. */
export type SlotCandidatesAnswer =
  { ok: true; groups: SlotCandidateGroups['groups'] } | { ok: false; problem: Problem };

/**
 * Asks for the candidates for the slot of the progression (counted from 0).
 * Resolves with the problem when the server refuses; rejects when the server
 * cannot be reached or answers with a body the contract does not describe.
 */
export async function getSlotCandidates(
  progression: Progression,
  slot: number,
  signal?: AbortSignal,
): Promise<SlotCandidatesAnswer> {
  const { ok, body } = await post('/progressions/slot-candidates', { progression, slot }, signal);
  if (ok) {
    return { ok: true, groups: parseSlotCandidateGroups(body).groups };
  }
  return { ok: false, problem: parseProblem(body) };
}

/** One variation of a chord and its labels. */
export type Variation = Open<components['schemas']['Variation']>;

/** The answer to a request for variations: the list, or the problem the server named. */
export type VariationsAnswer =
  { ok: true; variations: Variation[] } | { ok: false; problem: Problem };

/**
 * Asks for the variations of the chord in the major key. Resolves with the
 * problem when the server refuses (a chord with no third, among others);
 * rejects when the server cannot be reached or answers with a body the
 * contract does not describe.
 */
export async function getVariations(
  key: string,
  chord: string,
  signal?: AbortSignal,
): Promise<VariationsAnswer> {
  const { ok, body } = await post('/variations', { key, chord }, signal);
  if (ok) {
    return { ok: true, variations: parseVariations(body) };
  }
  return { ok: false, problem: parseProblem(body) };
}

async function post(path: string, payload: unknown, signal?: AbortSignal) {
  const response = await fetch(`${API_PREFIX}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    signal: signal ?? null,
  });
  const body: unknown = await response.json();
  return { ok: response.ok, body };
}

type Fields = Record<string, unknown>;

function isFields(value: unknown): value is Fields {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

class UnexpectedBody extends Error {
  constructor(where: string) {
    super(`the server answered with an unexpected body (${where})`);
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

/**
 * The labels of one slot, in the shape its content names. A content the
 * contract does not list is refused: the screen cannot tell which labels the
 * slot has.
 */
function parseSlotReading(value: unknown, where: string): SlotReading {
  const fields = fieldsAt(value, where);
  const at = (name: string) => `${where}.${name}`;
  const content = stringAt(fields, 'content', where);
  const roleFit =
    'roleFit' in fields ? { roleFit: oneStringLabel(fields.roleFit, 'fit', at('roleFit')) } : {};
  if (content === 'none' || content === 'unchosen') {
    return { content, ...roleFit };
  }
  if (content !== 'chord') {
    throw new UnexpectedBody(at('content'));
  }
  const slot: ChordSlotReading = {
    content,
    chord: oneStringLabel(fields.chord, 'name', at('chord')),
    reading: parseChordReadingLabel(fields.reading, at('reading')),
    resolution: oneStringLabel(fields.resolution, 'resolution', at('resolution')),
    cadence: oneStringLabel(fields.cadence, 'cadence', at('cadence')),
    tensionCautions: parseTensionCautionsLabel(fields.tensionCautions, at('tensionCautions')),
    ...roleFit,
  };
  if ('roman' in fields) slot.roman = oneStringLabel(fields.roman, 'numeral', at('roman'));
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

function parseCandidate(value: unknown, where: string): Candidate {
  const fields = fieldsAt(value, where);
  const at = (name: string) => `${where}.${name}`;
  const candidate: Candidate = {
    chord: oneStringLabel(fields.chord, 'name', at('chord')),
    roman: oneStringLabel(fields.roman, 'numeral', at('roman')),
    reading: parseChordReadingLabel(fields.reading, at('reading')),
  };
  if ('rootMotion' in fields) {
    candidate.rootMotion = oneStringLabel(fields.rootMotion, 'interval', at('rootMotion'));
  }
  if ('commonTones' in fields) {
    const { fields: c, basis } = labelAt(fields.commonTones, at('commonTones'));
    candidate.commonTones = { basis, count: integerAt(c, 'count', at('commonTones')) };
  }
  if ('resolution' in fields) {
    candidate.resolution = oneStringLabel(fields.resolution, 'resolution', at('resolution'));
  }
  if ('cadence' in fields) {
    candidate.cadence = oneStringLabel(fields.cadence, 'cadence', at('cadence'));
  }
  if ('preference' in fields) {
    const where = at('preference');
    const { fields: p, basis } = labelAt(fields.preference, where);
    candidate.preference = {
      basis,
      from: stringAt(p, 'from', where),
      to: stringAt(p, 'to', where),
      preference: stringAt(p, 'preference', where),
      style: stringAt(p, 'style', where),
    };
  }
  return candidate;
}

function parseSlotCandidate(value: unknown, where: string): SlotCandidate {
  const fields = fieldsAt(value, where);
  const at = (name: string) => `${where}.${name}`;
  const candidate: SlotCandidate = parseCandidate(fields, where);
  if ('nextResolution' in fields) {
    candidate.nextResolution = oneStringLabel(
      fields.nextResolution,
      'resolution',
      at('nextResolution'),
    );
  }
  if ('nextCadence' in fields) {
    candidate.nextCadence = oneStringLabel(fields.nextCadence, 'cadence', at('nextCadence'));
  }
  if ('roleFit' in fields) {
    const { fields: r, basis } = labelAt(fields.roleFit, at('roleFit'));
    const unmet = arrayAt(r, 'unmet', at('roleFit')).map((condition, i) => {
      if (typeof condition !== 'string') {
        throw new UnexpectedBody(`${at('roleFit')}.unmet[${i}]`);
      }
      return condition;
    });
    candidate.roleFit = { basis, fit: stringAt(r, 'fit', at('roleFit')), unmet };
  }
  return candidate;
}

/** The groups of a body, each candidate turned by the function given. */
function parseGroups<C>(body: unknown, parse: (value: unknown, where: string) => C) {
  const fields = fieldsAt(body, 'body');
  return {
    groups: arrayAt(fields, 'groups', 'body').map((value, i) => {
      const where = `groups[${i}]`;
      const group = fieldsAt(value, where);
      return {
        kind: stringAt(group, 'kind', where),
        candidates: arrayAt(group, 'candidates', where).map((c, j) =>
          parse(c, `${where}.candidates[${j}]`),
        ),
      };
    }),
  };
}

/** Turns an unknown body into the groups of candidates, or throws. */
export function parseCandidateGroups(body: unknown): CandidateGroups {
  return parseGroups(body, parseCandidate);
}

/** Turns an unknown body into the groups of the candidates for a slot, or throws. */
export function parseSlotCandidateGroups(body: unknown): SlotCandidateGroups {
  return parseGroups(body, parseSlotCandidate);
}

function parseNotesLabel(value: unknown, where: string) {
  const { fields, basis } = labelAt(value, where);
  const notes = arrayAt(fields, 'notes', where).map((note, i) => {
    if (typeof note !== 'string') {
      throw new UnexpectedBody(`${where}.notes[${i}]`);
    }
    return note;
  });
  return { basis, notes };
}

/** Turns an unknown body into the variations of a chord, or throws. */
export function parseVariations(body: unknown): Variation[] {
  const fields = fieldsAt(body, 'body');
  return arrayAt(fields, 'variations', 'body').map((value, i) => {
    const where = `variations[${i}]`;
    const variation = fieldsAt(value, where);
    const isInput = variation.isInput;
    if (typeof isInput !== 'boolean') {
      throw new UnexpectedBody(`${where}.isInput`);
    }
    const at = (name: string) => `${where}.${name}`;
    return {
      chord: oneStringLabel(variation.chord, 'name', at('chord')),
      isInput,
      members: parseNotesLabel(variation.members, at('members')),
      reading: parseChordReadingLabel(variation.reading, at('reading')),
      tensionCautions: parseTensionCautionsLabel(variation.tensionCautions, at('tensionCautions')),
      outOfKey: parseNotesLabel(variation.outOfKey, at('outOfKey')),
    };
  });
}

function integerAt(fields: Fields, name: string, where: string): number {
  const value = fields[name];
  if (typeof value !== 'number' || !Number.isInteger(value)) {
    throw new UnexpectedBody(`${where}.${name}`);
  }
  return value;
}

/** Turns an unknown body into the notes of a progression, or throws. */
export function parseProgressionNotes(body: unknown): Note[] {
  const fields = fieldsAt(body, 'body');
  return arrayAt(fields, 'notes', 'body').map((value, i) => {
    const where = `notes[${i}]`;
    const note = fieldsAt(value, where);
    return {
      pitch: integerAt(note, 'pitch', where),
      start: integerAt(note, 'start', where),
      beats: integerAt(note, 'beats', where),
      velocity: integerAt(note, 'velocity', where),
    };
  });
}

/** Turns an unknown body into a Problem, or throws. */
export function parseProblem(body: unknown): Problem {
  const fields = fieldsAt(body, 'problem');
  const title = stringAt(fields, 'title', 'problem');
  const problem: Problem = { title, status: integerAt(fields, 'status', 'problem') };
  if ('detail' in fields) {
    problem.detail = stringAt(fields, 'detail', 'problem');
  }
  if ('slot' in fields) {
    const slot = integerAt(fields, 'slot', 'problem');
    if (slot < 0) {
      throw new UnexpectedBody('problem.slot');
    }
    problem.slot = slot;
  }
  return problem;
}
