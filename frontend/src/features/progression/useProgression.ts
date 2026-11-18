import { useCallback, useEffect, useState } from 'react';
import type { Candidate, Progression, Slot } from '@/shared/api/client';
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

/**
 * The most beats the slot at the index can have without crossing into the
 * next phrase: from its start to the end of its phrase (#171).
 */
export function maxBeatsAt(slots: Slot[], index: number): number {
  const start = slots.slice(0, index).reduce((sum, slot) => sum + slot.beats, 0);
  return PHRASE_BEATS - (start % PHRASE_BEATS);
}

/** The beats of a bar of 4/4. */
export const BAR_BEATS = 4;

/** The start of a bar within a slot: where (0 at its start, 1 at its end), which bar, and the phrase it starts, if any. */
export type BarTick = { at: number; bar: number; phrase: number | null };

/** The bars that start within each slot, counted from 1, for the bar ticks of the strip. */
export function barTicks(slots: Slot[]): BarTick[][] {
  let start = 0;
  return slots.map((slot) => {
    const ticks: BarTick[] = [];
    const first = Math.ceil(start / BAR_BEATS) * BAR_BEATS;
    for (let beat = first; beat < start + slot.beats; beat += BAR_BEATS) {
      const phrase = beat % PHRASE_BEATS === 0 ? beat / PHRASE_BEATS + 1 : null;
      ticks.push({ at: (beat - start) / slot.beats, bar: beat / BAR_BEATS + 1, phrase });
    }
    start += slot.beats;
    return ticks;
  });
}

/**
 * The slots with each that would cross the end of the phrase it starts in
 * shortened to end there, and the indexes of those shortened (#179). Going
 * from the left, a shortening moves the starts after it before they are
 * looked at, so one pass takes in a slot it makes cross.
 */
export function fitToPhrases(slots: Slot[]): { slots: Slot[]; shortened: number[] } {
  const shortened: number[] = [];
  let start = 0;
  const fitted = slots.map((slot, i) => {
    const left = PHRASE_BEATS - (start % PHRASE_BEATS);
    const fit = slot.beats > left ? { ...slot, beats: left } : slot;
    if (fit !== slot) {
      shortened.push(i);
    }
    start += fit.beats;
    return fit;
  });
  return { slots: fitted, shortened };
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

/**
 * The branches of the step before, kept beside the slot a branch was put in
 * (the footprint, docs/product/screen-promise.md, 4): where, and all the
 * branches of that step, the one in the slot among them.
 */
export type Footprint = { at: number; branches: Candidate[] };

/** The progression being built, the slot chosen in it, and the operations on them. */
export type ProgressionState = {
  progression: Progression;
  selected: number | null;
  /**
   * For each slot, the chord it was added with from the candidates, or null:
   * the base its variations come from (#115). Kept in memory only.
   */
  bases: (string | null)[];
  /**
   * The slots the last change of the progression shortened to the end of
   * their phrase, as they are now; empty when it shortened none.
   */
  shortened: Slot[];
  /** The footprint of the last branch put in; any other change of the progression clears it. */
  footprint: Footprint | null;
  setKey: (key: string) => void;
  /** Adds a slot holding a typed chord, as long as nextBeats gives. */
  addChord: (name: string) => void;
  /**
   * Adds a slot holding a candidate, which becomes the base of the slot. For
   * a branch, the branches of its step are given, and kept as the footprint.
   */
  addCandidate: (name: string, beats: number, branches?: Candidate[]) => void;
  /** Puts a branch of the footprint in the slot it is beside, as fillSlot does, keeping the footprint. */
  replaceFromFootprint: (name: string) => void;
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

/** The slots, their bases, what the last change shortened and the footprint, which change together. */
type Held = {
  progression: Progression;
  bases: (string | null)[];
  shortened: Slot[];
  footprint: Footprint | null;
};

/** What a change gives: the footprint only when the change keeps or makes one. */
type Changed = Omit<Held, 'shortened' | 'footprint'> & { footprint?: Footprint | null };

function initial(): Held {
  const progression = load();
  return { progression, bases: progression.slots.map(() => null), shortened: [], footprint: null };
}

/**
 * The state after a change: no slot crosses into the next phrase, and the
 * slots shortened for that are the ones the change shortened.
 */
function fitted({ progression, bases, footprint = null }: Changed): Held {
  const { slots, shortened } = fitToPhrases(progression.slots);
  return {
    progression: { ...progression, slots },
    bases,
    shortened: slots.filter((_, i) => shortened.includes(i)),
    footprint,
  };
}

/** The slots and their bases with the chord in the slot at the index, which keeps its length and role. */
function filled({ progression, bases }: Held, index: number, chord: string): Changed {
  return {
    progression: {
      ...progression,
      slots: progression.slots.map((slot, i) =>
        i === index ? withRole({ chord, beats: slot.beats }, slot.role ?? null) : slot,
      ),
    },
    bases: bases.map((base, i) => (i === index ? chord : base)),
  };
}

/** The branch of the footprint in the slot it is beside, keeping the footprint. */
function refilled(h: Held, chord: string): Changed {
  return h.footprint === null || chord === ''
    ? h
    : { ...filled(h, h.footprint.at, chord), footprint: h.footprint };
}

/** The slots and their bases without the slot at the index. */
function without({ progression, bases }: Held, index: number): Changed {
  return {
    progression: { ...progression, slots: progression.slots.filter((_, i) => i !== index) },
    bases: bases.filter((_, i) => i !== index),
  };
}

/** Holds the progression, kept in localStorage so that a reload brings it back. */
export function useProgression(): ProgressionState {
  const [held, setHeld] = useState(initial);
  const [selected, setSelected] = useState<number | null>(null);
  const { progression, bases, shortened, footprint } = held;
  /**
   * Changes the progression; every change goes through here, so none leaves a
   * slot crossing, and none keeps the footprint unless it says so.
   */
  const change = useCallback((next: (h: Held) => Changed) => setHeld((h) => fitted(next(h))), []);

  useEffect(() => save(progression), [progression]);

  /** Adds a slot after the last; a function makes it from the slots before it. */
  const append = useCallback(
    (slot: Slot | ((slots: Slot[]) => Slot), base: string | null = null, branches?: Candidate[]) =>
      change((h) => {
        const added = typeof slot === 'function' ? slot(h.progression.slots) : slot;
        return {
          progression: { ...h.progression, slots: [...h.progression.slots, added] },
          bases: [...h.bases, base],
          footprint: branches ? { at: h.progression.slots.length, branches } : null,
        };
      }),
    [change],
  );
  /** Changes the slots in place: the bases stay with them. */
  const changeSlots = useCallback(
    (each: (slot: Slot, i: number) => Slot) =>
      change((h) => ({
        bases: h.bases,
        progression: { ...h.progression, slots: h.progression.slots.map(each) },
      })),
    [change],
  );

  return {
    progression,
    selected,
    bases,
    shortened,
    footprint,
    setKey: useCallback(
      (key: string) => {
        if (KEYS.some((k) => k === key)) {
          change((h) => ({ bases: h.bases, progression: { ...h.progression, key } }));
        }
      },
      [change],
    ),
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
      (chord: string, beats: number, branches?: Candidate[]) => {
        if (chord !== '' && isBeats(beats)) {
          append({ chord, beats }, chord, branches);
        }
      },
      [append],
    ),
    replaceFromFootprint: useCallback(
      (chord: string) => change((h) => refilled(h, chord)),
      [change],
    ),
    addNoChord: useCallback(
      () => append((slots) => ({ content: 'none', beats: nextBeats(slots) })),
      [append],
    ),
    addUnchosen: useCallback(
      () => append((slots) => ({ content: 'unchosen', beats: nextBeats(slots) })),
      [append],
    ),
    replaceChord: useCallback(
      (index: number, chord: string) =>
        changeSlots((slot, i) =>
          i === index && chord !== ''
            ? withRole({ chord, beats: slot.beats }, slot.role ?? null)
            : slot,
        ),
      [changeSlots],
    ),
    fillSlot: useCallback(
      (index: number, chord: string) => {
        if (chord !== '') {
          change((h) => filled(h, index, chord));
        }
      },
      [change],
    ),
    remove: useCallback(
      (index: number) => {
        change((h) => without(h, index));
        setSelected((s) => (s === null || s < index ? s : s === index ? null : s - 1));
      },
      [change],
    ),
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
