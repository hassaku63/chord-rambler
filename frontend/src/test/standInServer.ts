import { vi } from 'vitest';
import type { components, paths } from '@/shared/api/schema.gen';

/** The schemas of the contract, as generated from docs/api/openapi.yaml. */
export type Schemas = components['schemas'];

/** Every path of the contract the screen posts to. A path added to the contract joins it. */
export type Route = {
  [P in keyof paths]-?: paths[P]['post'] extends undefined ? never : P;
}[keyof paths];

type PostOf<R extends Route> = paths[R]['post'];

/** The body the contract takes on a route. */
export type RequestOf<R extends Route> =
  PostOf<R> extends { requestBody: { content: { 'application/json': infer B } } } ? B : never;

/** The body the contract gives with 200 on a route. */
export type AnswerOf<R extends Route> =
  PostOf<R> extends { responses: { 200: { content: { 'application/json': infer A } } } }
    ? A
    : never;

/** A refusal: the problem document the server answers with 400. */
export type Refusal = { refused: Schemas['Problem'] };

/** How the stand-in server answers a route: with the contract's body or a refusal. */
export type Answer<R extends Route> = (
  request: RequestOf<R>,
) => AnswerOf<R> | Refusal | Promise<AnswerOf<R> | Refusal>;

/** An answer for every route of the contract: missing one fails the type check. */
export type Answers = { [R in Route]: Answer<R> };

/** A 400 with the detail given, as the server refuses a request. */
export function refuse(detail?: string): Refusal {
  const problem: Schemas['Problem'] = { title: 'Bad Request', status: 400 };
  return { refused: detail === undefined ? problem : { ...problem, detail } };
}

/**
 * A candidate after a slot with a chord: the contract gives it the resolution
 * of that chord ("Present when the slot before holds a chord").
 */
export type AfterChord<C extends Schemas['Candidate']> = C &
  Required<Pick<Schemas['Candidate'], 'resolution'>>;

/**
 * A candidate for a slot before a slot with a chord: the contract gives it how
 * it goes on to that chord ("Present when the next slot holds a chord").
 */
export type BeforeChord = Schemas['SlotCandidate'] &
  Required<Pick<Schemas['SlotCandidate'], 'nextResolution'>>;

const API_PREFIX = '/api/v1';
const STRUCTURAL_ROLES: readonly Schemas['Role'][] = [
  'stable',
  'link',
  'prepare',
  'rise',
  'resolve',
  'close',
];

const diatonic: Schemas['ChordReadingLabel'] = {
  basis: 'school',
  kind: 'diatonic',
  function: 'subdominant',
  modes: [],
  fromParallelMinor: false,
};

/** The readings the stand-in server knows, as the contract writes them. */
const known: Record<string, Schemas['SlotReading']> = {
  Dm7: {
    chord: { basis: 'definition', name: 'Dm7' },
    roman: { basis: 'computation', numeral: 'ii7' },
    reading: diatonic,
    resolution: { basis: 'school', resolution: 'none' },
    cadence: { basis: 'school', cadence: 'none' },
    tensionCautions: { basis: 'school', cautions: [] },
  },
  F: {
    chord: { basis: 'definition', name: 'F' },
    roman: { basis: 'computation', numeral: 'IV' },
  },
  Fsus4: { chord: { basis: 'definition', name: 'Fsus4' } },
  Fm: {
    chord: { basis: 'definition', name: 'Fm' },
    roman: { basis: 'computation', numeral: 'iv' },
    reading: {
      basis: 'school',
      kind: 'borrowed',
      function: 'none',
      modes: ['dorian', 'aeolian'],
      fromParallelMinor: true,
    },
  },
};

/** The candidates after the last slot: one in each group, F with its labels after a chord. */
function candidatesAfter(progression: Schemas['Progression']): Schemas['CandidateGroups'] {
  const after = progression.slots.at(-1)?.chord !== undefined;
  const candidate = (
    name: string,
    numeral: string,
    kind: Schemas['ChordReadingLabel']['kind'],
  ): Schemas['Candidate'] => ({
    chord: { basis: 'definition', name },
    roman: { basis: 'computation', numeral },
    reading: { ...diatonic, kind, function: 'none' },
    ...(after && { resolution: { basis: 'school', resolution: 'none' } }),
  });
  const f: Schemas['Candidate'] = { ...candidate('F', 'IV', 'diatonic'), reading: diatonic };
  if (after) {
    f.rootMotion = { basis: 'computation', interval: 'm3' };
    f.commonTones = { basis: 'computation', count: 2 };
    f.preference = {
      basis: 'convention',
      from: 'subdominant',
      to: 'subdominant',
      preference: 'less-common',
      style: 'classical',
    };
  }
  return {
    groups: [
      { kind: 'diatonic', candidates: [f] },
      { kind: 'secondary-dominant', candidates: [candidate('D7', 'V7/V', 'secondary-dominant')] },
      { kind: 'borrowed', candidates: [candidate('Fm', 'iv', 'borrowed')] },
    ],
  };
}

/**
 * The candidates for a slot: those after the slot before, with how they go on
 * to a chord in the next slot and, for a structural role, a role fit.
 */
function candidatesFor({
  progression,
  slot,
}: Schemas['SlotCandidatesRequest']): Schemas['SlotCandidateGroups'] {
  const next = progression.slots[slot + 1]?.chord !== undefined;
  const role = progression.slots[slot]?.role;
  const structural = role !== undefined && STRUCTURAL_ROLES.includes(role);
  const before = { ...progression, slots: progression.slots.slice(0, slot) };
  return {
    groups: candidatesAfter(before).groups.map(({ kind, candidates }) => ({
      kind,
      candidates: candidates.map((c) => ({
        ...c,
        ...(next && { nextResolution: { basis: 'school', resolution: 'none' } }),
        ...(structural && { roleFit: { basis: 'school', fit: 'undecided', unmet: [] } }),
      })),
    })),
  };
}

/** The variations of F, F and Fsus4; a 400 for any other chord, as for a chord with no third. */
function variationsOf({ chord }: Schemas['VariationsRequest']): Schemas['Variations'] | Refusal {
  if (chord !== 'F') {
    return refuse('the chord has no third');
  }
  const variation = (name: string): Schemas['Variation'] => ({
    chord: { basis: 'definition', name },
    isInput: name === chord,
    members: { basis: 'definition', notes: [] },
    reading: diatonic,
    tensionCautions: { basis: 'school', cautions: [] },
    outOfKey: { basis: 'computation', notes: [] },
  });
  return { variations: [variation('F'), variation('Fsus4')] };
}

/** One note of middle C for each slot with a chord. */
function notesOf(progression: Schemas['Progression']): Schemas['ProgressionNotes'] {
  let start = 0;
  const notes: Schemas['Note'][] = [];
  for (const slot of progression.slots) {
    if (slot.chord !== undefined) {
      notes.push({ pitch: 60, start, beats: slot.beats, velocity: 64 });
    }
    start += slot.beats;
  }
  return { notes };
}

/** The readings of the chords it knows; a 400 naming the first chord it cannot read. */
function readingsOf(progression: Schemas['Progression']): Schemas['ProgressionReading'] | Refusal {
  const slots: Schemas['SlotReading'][] = [];
  for (const [i, slot] of progression.slots.entries()) {
    const reading = slot.chord === undefined ? {} : known[slot.chord];
    if (reading === undefined) {
      return refuse(`slot ${i}: cannot read the chord name "${slot.chord ?? ''}"`);
    }
    slots.push(reading);
  }
  return { slots };
}

/** How the stand-in server answers each route unless a test says otherwise. */
export const world: Answers = {
  '/progressions/readings': readingsOf,
  '/progressions/notes': notesOf,
  '/progressions/next-candidates': ({ progression }) => candidatesAfter(progression),
  '/progressions/slot-candidates': candidatesFor,
  '/variations': variationsOf,
};

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status });
}

/**
 * A fetch that answers every route of the contract like the server: as the
 * world does, or as the answers given. A path outside the contract throws.
 */
export function standInServer(answers: Partial<Answers> = {}) {
  const routes: Answers = { ...world, ...answers };
  return async (url: string, init?: RequestInit): Promise<Response> => {
    const route = url.slice(API_PREFIX.length);
    if (!url.startsWith(API_PREFIX) || !(route in routes)) {
      throw new Error(`the stand-in server does not know ${url}`);
    }
    const answer = routes[route as Route] as (request: unknown) => ReturnType<Answer<Route>>;
    const reply = await answer(JSON.parse(init?.body as string));
    return 'refused' in reply ? json(reply.refused, reply.refused.status) : json(reply);
  };
}

/** Puts the stand-in server in place of fetch; vi.unstubAllGlobals takes it away. */
export function stubServer(answers: Partial<Answers> = {}) {
  vi.stubGlobal('fetch', vi.fn(standInServer(answers)));
}
