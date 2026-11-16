import type { Note, Slot } from '@/shared/api/client';

/** The tempo the screen starts with, and the range it offers, in beats per minute. */
export const DEFAULT_TEMPO = 90;
export const MIN_TEMPO = 40;
export const MAX_TEMPO = 200;

/**
 * One oscillator to play: its frequency in Hz, its start and length in seconds
 * from the start of the playing, and its loudness from 0 to 1.
 */
export type Voice = { frequency: number; start: number; duration: number; gain: number };

/** The moment, in seconds from the start of the playing, that a slot begins to sound. */
export type Mark = { at: number; slot: number };

/** What the player plays: the voices, the slots as they come, and when the playing ends. */
export type Performance = { voices: Voice[]; marks: Mark[]; end: number };

/** The frequency of a MIDI note number in equal temperament, A4 = 69 = 440 Hz. */
export function frequencyOf(pitch: number): number {
  return 440 * 2 ** ((pitch - 69) / 12);
}

export function secondsPerBeat(tempo: number): number {
  return 60 / tempo;
}

/** The voices of the notes at the tempo. The velocity, 0 to 127, becomes the loudness. */
export function voicesOf(notes: Note[], tempo: number): Voice[] {
  const beat = secondsPerBeat(tempo);
  return notes.map((note) => ({
    frequency: frequencyOf(note.pitch),
    start: note.start * beat,
    duration: note.beats * beat,
    gain: note.velocity / 127,
  }));
}

/** The beat each slot starts on, counted from 0. */
export function slotStarts(slots: Slot[]): number[] {
  let start = 0;
  return slots.map((slot) => {
    const at = start;
    start += slot.beats;
    return at;
  });
}

/** The whole progression: every note, and every slot marked as it begins. */
export function progressionPerformance(notes: Note[], slots: Slot[], tempo: number): Performance {
  const beat = secondsPerBeat(tempo);
  const total = slots.reduce((sum, slot) => sum + slot.beats, 0);
  return {
    voices: voicesOf(notes, tempo),
    marks: slotStarts(slots).map((start, slot) => ({ at: start * beat, slot })),
    end: total * beat,
  };
}

/** The notes of one slot, moved to start at beat 0 and cut to the slot's length. */
export function notesOfSlot(notes: Note[], slots: Slot[], index: number): Note[] {
  const slot = slots[index];
  const from = slotStarts(slots)[index];
  if (slot === undefined || from === undefined) {
    return [];
  }
  const to = from + slot.beats;
  return notes
    .filter((note) => note.start >= from && note.start < to)
    .map((note) => ({
      ...note,
      start: note.start - from,
      beats: Math.min(note.beats, to - note.start),
    }));
}

/**
 * The slots from `from` up to `to` (not included) alone: their notes from the
 * start of the first, each slot marked as it begins, for their length.
 */
export function passagePerformance(
  notes: Note[],
  slots: Slot[],
  from: number,
  to: number,
  tempo: number,
): Performance {
  const beat = secondsPerBeat(tempo);
  const starts = slotStarts(slots);
  const offset = starts[from] ?? 0;
  const range = Array.from({ length: Math.max(0, to - from) }, (_, i) => from + i);
  const at = (index: number) => ((starts[index] ?? 0) - offset) * beat;
  return {
    voices: range.flatMap((index) =>
      voicesOf(notesOfSlot(notes, slots, index), tempo).map((voice) => ({
        ...voice,
        start: voice.start + at(index),
      })),
    ),
    marks: range.map((index) => ({ at: at(index), slot: index })),
    end: range.reduce((sum, index) => sum + (slots[index]?.beats ?? 0), 0) * beat,
  };
}

/** One slot alone, for its length. */
export function slotPerformance(
  notes: Note[],
  slots: Slot[],
  index: number,
  tempo: number,
): Performance {
  return passagePerformance(notes, slots, index, index + 1, tempo);
}
