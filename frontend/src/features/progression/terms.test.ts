import { describe, expect, expectTypeOf, it } from 'vitest';
import type { components } from '@/shared/api/schema.gen';
import { TERMS, wordFor } from './terms';

type Schemas = components['schemas'];
type ChordReading = Schemas['ChordReadingLabel'];
type Preference = Schemas['PreferenceLabel'];

// Every enumeration of the readings in docs/api/openapi.yaml, written out. The
// type checks below fail when the contract (through schema.gen.ts) adds a value
// that is missing here.
const basis = ['definition', 'computation', 'school', 'convention'] as const;
const kind = ['facts-only', 'diatonic', 'borrowed', 'secondary-dominant'] as const;
const fn = ['none', 'tonic', 'subdominant', 'dominant'] as const;
const mode = [
  'ionian',
  'dorian',
  'phrygian',
  'lydian',
  'mixolydian',
  'aeolian',
  'locrian',
] as const;
const resolution = ['none', 'expected', 'irregular'] as const;
const cadence = ['none', 'authentic', 'half', 'plagal', 'deceptive'] as const;
const fit = ['fits', 'misfits', 'undecided'] as const;
const candidateKind = ['diatonic', 'secondary-dominant', 'borrowed'] as const;
const preference = ['common', 'less-common', 'unstated'] as const;
const style = ['classical'] as const;
const moveFunction = ['tonic', 'subdominant', 'dominant'] as const;

// Type checks only (tsc runs them): each list is exactly its enumeration.
expectTypeOf<(typeof basis)[number]>().toEqualTypeOf<Schemas['Basis']>();
expectTypeOf<(typeof kind)[number]>().toEqualTypeOf<ChordReading['kind']>();
expectTypeOf<(typeof fn)[number]>().toEqualTypeOf<ChordReading['function']>();
expectTypeOf<(typeof mode)[number]>().toEqualTypeOf<Schemas['Mode']>();
expectTypeOf<(typeof resolution)[number]>().toEqualTypeOf<
  Schemas['ResolutionLabel']['resolution']
>();
expectTypeOf<(typeof cadence)[number]>().toEqualTypeOf<Schemas['CadenceLabel']['cadence']>();
expectTypeOf<(typeof fit)[number]>().toEqualTypeOf<Schemas['RoleFitLabel']['fit']>();
expectTypeOf<(typeof candidateKind)[number]>().toEqualTypeOf<Schemas['CandidateKind']>();
expectTypeOf<(typeof preference)[number]>().toEqualTypeOf<Preference['preference']>();
expectTypeOf<(typeof style)[number]>().toEqualTypeOf<Preference['style']>();
expectTypeOf<(typeof moveFunction)[number]>().toEqualTypeOf<Preference['from']>();
expectTypeOf<(typeof moveFunction)[number]>().toEqualTypeOf<Preference['to']>();

describe('TERMS', () => {
  it.each([
    { name: 'basis', table: TERMS.basis, ids: basis },
    { name: 'kind', table: TERMS.kind, ids: kind },
    { name: 'function', table: TERMS.function, ids: fn },
    { name: 'function mark', table: TERMS.functionMark, ids: fn },
    { name: 'mode', table: TERMS.mode, ids: mode },
    { name: 'resolution', table: TERMS.resolution, ids: resolution },
    { name: 'cadence', table: TERMS.cadence, ids: cadence },
    { name: 'fit', table: TERMS.fit, ids: fit },
    { name: 'candidate kind', table: TERMS.kind, ids: candidateKind },
    { name: 'preference', table: TERMS.preference, ids: preference },
    { name: 'style', table: TERMS.style, ids: style },
    { name: 'function of a move', table: TERMS.function, ids: moveFunction },
  ])('has an entry for every $name of the contract', ({ table, ids }) => {
    const missing = ids.filter((id) => !Object.hasOwn(table, id));

    expect(missing).toEqual([]);
  });

  it('gives the word, nothing for "nothing to name", and an unknown identifier as it is', () => {
    expect(wordFor(TERMS.cadence, 'authentic')).toBe('正格終止');
    expect(wordFor(TERMS.cadence, 'none')).toBeNull();
    expect(wordFor(TERMS.cadence, 'picardy')).toBe('picardy');
  });
});
