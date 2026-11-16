import { useCallback, useEffect, useState } from 'react';
import type { Progression, Slot } from '@/shared/api/client';

/** The 12 major keys the screen offers (ADR-0006), round the circle of fifths. */
export const KEYS = ['C', 'G', 'D', 'A', 'E', 'B', 'F♯', 'D♭', 'A♭', 'E♭', 'B♭', 'F'] as const;
/** The length a new slot gets: one bar of 4/4. */
export const DEFAULT_BEATS = 4;
export const MIN_BEATS = 1;
export const MAX_BEATS = 16;

/** Where the progression is kept in the browser (the C1 of #110). */
export const STORAGE_KEY = 'chord-rambler.progression';

const EMPTY: Progression = { key: 'C', slots: [] };

function isBeats(value: unknown): value is number {
  return (
    Number.isInteger(value) && (value as number) >= MIN_BEATS && (value as number) <= MAX_BEATS
  );
}

function parseSlot(value: unknown): Slot | null {
  if (typeof value !== 'object' || value === null) {
    return null;
  }
  const fields = value as Record<string, unknown>;
  const names = Object.keys(fields);
  if (!isBeats(fields.beats) || names.some((n) => !['chord', 'content', 'beats'].includes(n))) {
    return null;
  }
  const { chord, content, beats } = fields;
  if (typeof chord === 'string' && content === undefined && chord.trim() !== '') {
    return { chord, beats };
  }
  if (chord === undefined && (content === 'none' || content === 'unchosen')) {
    return { content, beats };
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
  addChord: (name: string) => void;
  /** Adds a slot holding a candidate, which becomes the base of the slot. */
  addCandidate: (name: string, beats: number) => void;
  addNoChord: () => void;
  addUnchosen: () => void;
  /** Puts the chord in place of the slot's, keeping its length and its base. */
  replaceChord: (index: number, name: string) => void;
  remove: (index: number) => void;
  setBeats: (index: number, beats: number) => void;
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

  const append = useCallback(
    (slot: Slot, base: string | null = null) =>
      setHeld((h) => ({
        progression: { ...h.progression, slots: [...h.progression.slots, slot] },
        bases: [...h.bases, base],
      })),
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
          append({ chord, beats: DEFAULT_BEATS });
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
          i === index && chord !== '' ? { chord, beats: slot.beats } : slot,
        ),
      [changeSlots],
    ),
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
    select: setSelected,
  };
}
