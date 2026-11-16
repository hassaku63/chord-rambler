import type { components } from '@/shared/api/schema.gen';

type Schemas = components['schemas'];
type ChordReading = Schemas['ChordReadingLabel'];
type Preference = Schemas['PreferenceLabel'];

/**
 * The one table from the identifiers of the readings to the words the screen
 * shows (D1 of #148). Every word is in docs/glossary.md. null marks the value
 * the contract uses for "nothing to name" (none, facts-only): no label shows.
 */
export const TERMS = {
  basis: {
    definition: '定義',
    computation: '計算',
    school: '流儀',
    convention: '慣習',
  } satisfies Record<Schemas['Basis'], string>,
  kind: {
    'facts-only': null,
    diatonic: 'ダイアトニック',
    borrowed: '借用和音',
    'secondary-dominant': 'セカンダリー・ドミナント',
  } satisfies Record<ChordReading['kind'], string | null>,
  function: {
    none: null,
    tonic: 'トニック',
    subdominant: 'サブドミナント',
    dominant: 'ドミナント',
  } satisfies Record<ChordReading['function'], string | null>,
  /** The short marks of the functions, for the slot's card. */
  functionMark: {
    none: null,
    tonic: 'T',
    subdominant: 'S',
    dominant: 'D',
  } satisfies Record<ChordReading['function'], string | null>,
  mode: {
    ionian: '長調',
    dorian: 'ドリアン',
    phrygian: 'フリジアン',
    lydian: 'リディアン',
    mixolydian: 'ミクソリディアン',
    aeolian: '自然短音階',
    locrian: 'ロクリアン',
  } satisfies Record<Schemas['Mode'], string>,
  resolution: {
    none: null,
    expected: '解決',
    irregular: '変則',
  } satisfies Record<Schemas['ResolutionLabel']['resolution'], string | null>,
  cadence: {
    none: null,
    authentic: '正格終止',
    half: '半終止',
    plagal: '変格終止',
    deceptive: '偽終止',
  } satisfies Record<Schemas['CadenceLabel']['cadence'], string | null>,
  fit: {
    fits: '合う',
    misfits: '合わない',
    undecided: '決められない',
  } satisfies Record<Schemas['RoleFitLabel']['fit'], string>,
  /** The grades of a move between functions; unstated (the source says nothing) shows no label. */
  preference: {
    common: 'よく使われる',
    'less-common': 'より少ない',
    unstated: null,
  } satisfies Record<Preference['preference'], string | null>,
  /** The style a grade is limited to, shown with it (#114). */
  style: {
    classical: 'クラシック',
  } satisfies Record<Preference['style'], string>,
} as const;

/** What a borrowed chord is borrowed from when its modes hold aeolian. */
export const PARALLEL_MINOR = '同主短調';

/**
 * The degree a secondary dominant points to, as the diatonic chord on it is
 * written (V/ii, V7/IV: docs/architecture/domain-model.md).
 */
export const TARGET_NUMERALS: Readonly<Record<number, string>> = {
  2: 'ii',
  3: 'iii',
  4: 'IV',
  5: 'V',
  6: 'vi',
};

/**
 * The word for an identifier: the table's word, null for "nothing to name",
 * or the identifier itself when the table does not know it (a newer server).
 */
export function wordFor(table: Readonly<Record<string, string | null>>, id: string): string | null {
  return Object.hasOwn(table, id) ? (table[id] ?? null) : id;
}
