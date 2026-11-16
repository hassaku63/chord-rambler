import { useCallback, useEffect, useState } from 'react';
import type { Progression, Slot } from '@/shared/api/client';
import type { Role } from './terms';
import { isRole } from './terms';

/** The 12 major keys the screen offers (ADR-0006), round the circle of fifths. */
export const KEYS = ['C', 'G', 'D', 'A', 'E', 'B', 'F♯', 'D♭', 'A♭', 'E♭', 'B♭', 'F'] as const;
/** The length a new slot gets: one bar of 4/4. */
export const DEFAULT_BEATS = 4;
export const MIN_BEATS = 1;
export const MAX_BEATS = 16;
/** The beats of a phrase: four bars of 4/4, cut from the start (docs/api/openapi.yaml). */
export const PHRASE_BEATS = 16;

/**
 * The length of a slot added after the last: one bar, or the beats left in
 * the phrase when fewer are left, so that the slot does not cross into the
 * next phrase (#116).
 */
export function nextBeats(slots: Slot[]): number {
  const used = slots.reduce((sum, slot) => sum + slot.beats, 0) % PHRASE_BEATS;
  return Math.min(DEFAULT_BEATS, PHRASE_BEATS - used);
}

/** Where the progression is kept in the browser (the C1 of #110). */
export const STORAGE_KEY = 'chord-rambler.progression';

const EMPTY: Progression = { key: 'C', slots: [] };

function isBeats(value: unknown): value is number {
  return (
    Number.isInteger(value) && (value as number) >= MIN_BEATS && (value as number) <= MAX_BEATS
  );
}

/** The slot with the role, or with none when the role is null: no role leaves no field. */
function withRole({ chord, content, beats }: Slot, role: Role | null): Slot {
  const slot: Slot =
    chord !== undefined ? { chord, beats } : content !== undefined ? { content, beats } : { beats };
  return role === null ? slot : { ...slot, role };
}

function parseSlot(value: unknown): Slot | null {
  if (typeof value !== 'object' || value === null) {
    return null;
  }
  const fields = value as Record<string, unknown>;
  const names = Object.keys(fields);
  if (
    !isBeats(fields.beats) ||
    names.some((n) => !['chord', 'content', 'beats', 'role'].includes(n)) ||
    ('role' in fields && !isRole(fields.role))
  ) {
    return null;
  }
  const { chord, content, beats } = fields;
  const role = isRole(fields.role) ? fields.role : null;
  if (typeof chord === 'string' && content === undefined && chord.trim() !== '') {
    return withRole({ chord, beats }, role);
  }
  if (chord === undefined && (content === 'none' || content === 'unchosen')) {
    return withRole({ content, beats }, role);
  }
  return null;
}

/**
 * Turns a stored value into a Progression, or null when it is not one the
 * screen could have saved: the one check at the storage boundary.
 */
export function parseProgression(value: unknown): Progression | null {
  if (typeof value !== 'object' || value === null) {
    return null;
  }
  const { key, slots } = value as Record<string, unknown>;
  if (!KEYS.some((k) => k === key) || !Array.isArray(slots)) {
    return null;
  }
  const parsed = slots.map(parseSlot);
  if (parsed.some((slot) => slot === null)) {
    return null;
  }
  return { key: key as string, slots: parsed as Slot[] };
}

function load(): Progression {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored === null ? EMPTY : (parseProgression(JSON.parse(stored)) ?? EMPTY);
  } catch {
    return EMPTY;
  }
}

function save(progression: Progression) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(progression));
  } catch {
    // Storage full or switched off: the progression lives until the page closes.
  }
}

/** The progression being built, the slot chosen in it, and the operations on them. */
export type ProgressionState = {
  progression: Progression;
  selected: number | null;
  /**
   * For each slot, the chord it was added with from the candidates, or null:
   * the base its variations come from (#115). Kept in memory only.
   */
  bases: (string | null)[];
  setKey: (key: string) => void;
  /** Adds a slot holding a typed chord, as long as nextBeats gives. */
  addChord: (name: string) => void;
  /** Adds a slot holding a candidate, which becomes the base of the slot. */
  addCandidate: (name: string, beats: number) => void;
  addNoChord: () => void;
  addUnchosen: () => void;
  /** Puts the chord in place of the slot's, keeping its length, its role and its base. */
  replaceChord: (index: number, name: string) => void;
  /**
   * Puts a candidate in the slot, keeping its length and its role; the
   * candidate becomes the base of the slot, as one added after the last.
   */
  fillSlot: (index: number, name: string) => void;
  remove: (index: number) => void;
  setBeats: (index: number, beats: number) => void;
  /** Puts a role of the contract on the slot; an empty string takes its role away. */
  setRole: (index: number, role: string) => void;
  select: (index: number | null) => void;
};

/** The slots and their bases, which change together. */
type Held = { progression: Progression; bases: (string | null)[] };

function initial(): Held {
  const progression = load();
  return { progression, bases: progression.slots.map(() => null) };
}

/** Holds the progression, kept in localStorage so that a reload brings it back. */
export function useProgression(): ProgressionState {
  const [held, setHeld] = useState(initial);
  const [selected, setSelected] = useState<number | null>(null);
  const { progression, bases } = held;

  useEffect(() => save(progression), [progression]);

  /** Adds a slot after the last; a function makes it from the slots before it. */
  const append = useCallback(
    (slot: Slot | ((slots: Slot[]) => Slot), base: string | null = null) =>
      setHeld((h) => {
        const added = typeof slot === 'function' ? slot(h.progression.slots) : slot;
        return {
          progression: { ...h.progression, slots: [...h.progression.slots, added] },
          bases: [...h.bases, base],
        };
      }),
    [],
  );
  /** Changes the slots in place: the bases stay with them. */
  const changeSlots = useCallback(
    (change: (slot: Slot, i: number) => Slot) =>
      setHeld((h) => ({
        ...h,
        progression: { ...h.progression, slots: h.progression.slots.map(change) },
      })),
    [],
  );

  return {
    progression,
    selected,
    bases,
    setKey: useCallback((key: string) => {
      if (KEYS.some((k) => k === key)) {
        setHeld((h) => ({ ...h, progression: { ...h.progression, key } }));
      }
    }, []),
    addChord: useCallback(
      (name: string) => {
        const chord = name.trim();
        if (chord !== '') {
          append((slots) => ({ chord, beats: nextBeats(slots) }));
        }
      },
      [append],
    ),
    addCandidate: useCallback(
      (chord: string, beats: number) => {
        if (chord !== '' && isBeats(beats)) {
          append({ chord, beats }, chord);
        }
      },
      [append],
    ),
    addNoChord: useCallback(() => append({ content: 'none', beats: DEFAULT_BEATS }), [append]),
    addUnchosen: useCallback(() => append({ content: 'unchosen', beats: DEFAULT_BEATS }), [append]),
    replaceChord: useCallback(
      (index: number, chord: string) =>
        changeSlots((slot, i) =>
          i === index && chord !== ''
            ? withRole({ chord, beats: slot.beats }, slot.role ?? null)
            : slot,
        ),
      [changeSlots],
    ),
    fillSlot: useCallback((index: number, chord: string) => {
      if (chord !== '') {
        setHeld((h) => ({
          progression: {
            ...h.progression,
            slots: h.progression.slots.map((slot, i) =>
              i === index ? withRole({ chord, beats: slot.beats }, slot.role ?? null) : slot,
            ),
          },
          bases: h.bases.map((base, i) => (i === index ? chord : base)),
        }));
      }
    }, []),
    remove: useCallback((index: number) => {
      setHeld((h) => ({
        progression: { ...h.progression, slots: h.progression.slots.filter((_, i) => i !== index) },
        bases: h.bases.filter((_, i) => i !== index),
      }));
      setSelected((s) => (s === null || s < index ? s : s === index ? null : s - 1));
    }, []),
    setBeats: useCallback(
      (index: number, beats: number) => {
        if (isBeats(beats)) {
          changeSlots((slot, i) => (i === index ? { ...slot, beats } : slot));
        }
      },
      [changeSlots],
    ),
    setRole: useCallback(
      (index: number, role: string) => {
        if (role === '' || isRole(role)) {
          changeSlots((slot, i) =>
            i === index ? withRole(slot, role === '' ? null : role) : slot,
          );
        }
      },
      [changeSlots],
    ),
    select: setSelected,
  };
}
