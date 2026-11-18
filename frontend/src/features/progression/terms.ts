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
  /** The roles a slot may hold: the six structural ones, then the four colour ones. */
  role: {
    stable: '安定',
    link: 'つなぎ',
    prepare: '準備',
    rise: '高まり',
    resolve: '解決',
    close: '締め',
    spread: '広がり',
    float: '浮遊',
    wistful: '切なさ',
    linger: '余韻',
  } satisfies Record<Schemas['Role'], string>,
  /**
   * The conditions of a structural role a slot does not meet, as the table of
   * docs/architecture/theory-conventions.md (7) names them.
   */
  unmet: {
    shape: '3 度のある三和音か七の和音でない',
    function: '機能が違う',
    'previous-v': '前の枠が V でない',
    'next-dominant': '次の枠がドミナントでない',
    cadence: '終止が付かない',
    'another-root': '根音が前の枠と同じ',
    'expected-resolution': '解決が期待どおりでない',
  } satisfies Record<Schemas['RoleCondition'], string>,
} as const;

/** A role, as the contract names it. */
export type Role = Schemas['Role'];

/** The roles by kind, in the order the choice lists them (docs/glossary.md). */
export const ROLE_GROUPS: readonly { kind: string; roles: readonly Role[] }[] = [
  { kind: '構造的な役割', roles: ['stable', 'link', 'prepare', 'rise', 'resolve', 'close'] },
  { kind: '色合いの役割', roles: ['spread', 'float', 'wistful', 'linger'] },
];

/** Whether a value is one of the roles of the contract. */
export function isRole(value: unknown): value is Role {
  return typeof value === 'string' && Object.hasOwn(TERMS.role, value);
}

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

/** What a sentence calls a slot: by its chord, or by what kind of empty slot it is. */
function slotWord(slot: Schemas['Slot']): string {
  if (slot.chord !== undefined) {
    return `${slot.chord} の枠`;
  }
  return slot.content === 'none' ? '和音なしの枠' : 'まだ選んでいない枠';
}

/**
 * The one sentence that tells which slots a change shortened to the end of
 * their phrase, as they are now (#179), or an empty string for none.
 */
export function shortenedNotice(slots: readonly Schemas['Slot'][]): string {
  if (slots.length === 0) {
    return '';
  }
  const each = slots.map((slot) => `${slotWord(slot)}を ${slot.beats} 拍に`);
  return `${each.join('、')}縮めました（フレーズをまたぐため）。`;
}
