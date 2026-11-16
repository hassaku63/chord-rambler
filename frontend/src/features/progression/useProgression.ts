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
  setKey: (key: string) => void;
  addChord: (name: string) => void;
  addNoChord: () => void;
  addUnchosen: () => void;
  remove: (index: number) => void;
  setBeats: (index: number, beats: number) => void;
  select: (index: number | null) => void;
};

/** Holds the progression, kept in localStorage so that a reload brings it back. */
export function useProgression(): ProgressionState {
  const [progression, setProgression] = useState(load);
  const [selected, setSelected] = useState<number | null>(null);

  useEffect(() => save(progression), [progression]);

  const append = useCallback(
    (slot: Slot) => setProgression((p) => ({ ...p, slots: [...p.slots, slot] })),
    [],
  );

  return {
    progression,
    selected,
    setKey: useCallback((key: string) => {
      if (KEYS.some((k) => k === key)) {
        setProgression((p) => ({ ...p, key }));
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
    addNoChord: useCallback(() => append({ content: 'none', beats: DEFAULT_BEATS }), [append]),
    addUnchosen: useCallback(() => append({ content: 'unchosen', beats: DEFAULT_BEATS }), [append]),
    remove: useCallback((index: number) => {
      setProgression((p) => ({ ...p, slots: p.slots.filter((_, i) => i !== index) }));
      setSelected((s) => (s === null || s < index ? s : s === index ? null : s - 1));
    }, []),
    setBeats: useCallback((index: number, beats: number) => {
      if (isBeats(beats)) {
        setProgression((p) => ({
          ...p,
          slots: p.slots.map((slot, i) => (i === index ? { ...slot, beats } : slot)),
        }));
      }
    }, []),
    select: setSelected,
  };
}
