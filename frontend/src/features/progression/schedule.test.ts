import { describe, expect, it } from 'vitest';
import type { Note, Slot } from '@/shared/api/client';
import {
  frequencyOf,
  notesOfSlot,
  progressionPerformance,
  secondsPerBeat,
  slotPerformance,
  voicesOf,
} from './schedule';

describe('frequencyOf', () => {
  // Equal temperament with A4 = 440 Hz; the values are those of the usual tables.
  it.each([
    { name: 'A4', pitch: 69, hz: 440 },
    { name: 'A3', pitch: 57, hz: 220 },
    { name: 'A5', pitch: 81, hz: 880 },
    { name: 'C4', pitch: 60, hz: 261.63 },
    { name: 'C5', pitch: 72, hz: 523.25 },
    { name: 'C2', pitch: 36, hz: 65.41 },
    { name: 'E4', pitch: 64, hz: 329.63 },
  ])('gives $hz Hz for $name ($pitch)', ({ pitch, hz }) => {
    expect(frequencyOf(pitch)).toBeCloseTo(hz, 2);
  });
});

describe('secondsPerBeat', () => {
  it.each([
    { tempo: 90, seconds: 0.667 },
    { tempo: 60, seconds: 1 },
    { tempo: 120, seconds: 0.5 },
    { tempo: 40, seconds: 1.5 },
    { tempo: 200, seconds: 0.3 },
  ])('gives $seconds s at $tempo BPM', ({ tempo, seconds }) => {
    expect(secondsPerBeat(tempo)).toBeCloseTo(seconds, 3);
  });
});

describe('voicesOf', () => {
  it('turns pitch, beats and velocity into frequency, seconds and loudness', () => {
    const [voice] = voicesOf([{ pitch: 60, start: 4, beats: 4, velocity: 64 }], 120);

    expect(voice?.frequency).toBeCloseTo(261.63, 2);
    expect(voice?.start).toBe(2);
    expect(voice?.duration).toBe(2);
    expect(voice?.gain).toBeCloseTo(0.504, 3);
  });

  it('gives full loudness at velocity 127 and none at 0', () => {
    const voices = voicesOf(
      [
        { pitch: 60, start: 0, beats: 1, velocity: 127 },
        { pitch: 60, start: 0, beats: 1, velocity: 0 },
      ],
      60,
    );

    expect(voices.map((v) => v.gain)).toEqual([1, 0]);
  });
});

// C for 4 beats, a slot with no chord for 2, G for 4: notes as the server gives them.
const slots: Slot[] = [
  { chord: 'C', beats: 4 },
  { content: 'none', beats: 2 },
  { chord: 'G', beats: 4 },
];
const notes: Note[] = [
  { pitch: 48, start: 0, beats: 4, velocity: 64 },
  { pitch: 60, start: 0, beats: 4, velocity: 64 },
  { pitch: 43, start: 6, beats: 4, velocity: 64 },
  { pitch: 67, start: 6, beats: 4, velocity: 64 },
];

describe('progressionPerformance', () => {
  it('marks each slot as it begins and ends after the last', () => {
    const performance = progressionPerformance(notes, slots, 60);

    expect(performance.voices.map((v) => v.start)).toEqual([0, 0, 6, 6]);
    expect(performance.marks).toEqual([
      { at: 0, slot: 0 },
      { at: 4, slot: 1 },
      { at: 6, slot: 2 },
    ]);
    expect(performance.end).toBe(10);
  });
});

describe('notesOfSlot', () => {
  it.each([
    {
      name: 'the first slot',
      index: 0,
      want: [
        { pitch: 48, start: 0, beats: 4, velocity: 64 },
        { pitch: 60, start: 0, beats: 4, velocity: 64 },
      ],
    },
    {
      name: 'a later slot, moved to beat 0',
      index: 2,
      want: [
        { pitch: 43, start: 0, beats: 4, velocity: 64 },
        { pitch: 67, start: 0, beats: 4, velocity: 64 },
      ],
    },
    { name: 'a slot with no chord', index: 1, want: [] },
    { name: 'a slot that is not there', index: 3, want: [] },
  ])('gives the notes of $name', ({ index, want }) => {
    expect(notesOfSlot(notes, slots, index)).toEqual(want);
  });

  it('cuts a note to the length of the slot', () => {
    const long: Note[] = [{ pitch: 60, start: 0, beats: 8, velocity: 64 }];

    expect(notesOfSlot(long, [{ chord: 'C', beats: 4 }], 0)).toEqual([
      { pitch: 60, start: 0, beats: 4, velocity: 64 },
    ]);
  });
});

describe('slotPerformance', () => {
  it('plays the slot alone from 0 for its length', () => {
    const performance = slotPerformance(notes, slots, 2, 120);

    expect(performance.voices.map((v) => [v.start, v.duration])).toEqual([
      [0, 2],
      [0, 2],
    ]);
    expect(performance.marks).toEqual([{ at: 0, slot: 2 }]);
    expect(performance.end).toBe(2);
  });
});
