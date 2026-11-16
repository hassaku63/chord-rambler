import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import type { Slot } from '@/shared/api/client';
import { STORAGE_KEY, nextBeats, useProgression } from './useProgression';

afterEach(() => localStorage.clear());

function slotsOf(...beats: number[]): Slot[] {
  return beats.map((b) => ({ content: 'none', beats: b }));
}

describe('nextBeats', () => {
  it.each([
    { name: 'an empty progression', slots: [], want: 4 },
    { name: 'a phrase with 6 beats left', slots: slotsOf(4, 4, 2), want: 4 },
    { name: 'a phrase with 3 beats left', slots: slotsOf(4, 4, 4, 1), want: 3 },
    { name: 'a phrase with 1 beat left', slots: slotsOf(15), want: 1 },
    { name: 'a full phrase', slots: slotsOf(16), want: 4 },
    { name: 'the second phrase with 2 beats left', slots: slotsOf(16, 14), want: 2 },
  ])('gives one bar or the beats left: $name', ({ slots, want }) => {
    expect(nextBeats(slots)).toBe(want);
  });
});

describe('useProgression', () => {
  it('starts empty in C', () => {
    const { result } = renderHook(() => useProgression());

    expect(result.current.progression).toEqual({ key: 'C', slots: [] });
    expect(result.current.selected).toBeNull();
  });

  it('adds slots at the end, one bar long, and changes the key', () => {
    const { result } = renderHook(() => useProgression());

    act(() => result.current.addChord(' Dm7 '));
    act(() => result.current.addNoChord());
    act(() => result.current.addUnchosen());
    act(() => result.current.setKey('E♭'));

    expect(result.current.progression).toEqual({
      key: 'E♭',
      slots: [
        { chord: 'Dm7', beats: 4 },
        { content: 'none', beats: 4 },
        { content: 'unchosen', beats: 4 },
      ],
    });
  });

  it('cuts a typed chord to the beats left in the phrase', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addNoChord());
    act(() => result.current.setBeats(0, 14));

    act(() => {
      result.current.addChord('F');
      result.current.addChord('G');
    });

    expect(result.current.progression.slots).toEqual([
      { content: 'none', beats: 14 },
      { chord: 'F', beats: 2 },
      { chord: 'G', beats: 4 },
    ]);
  });

  it('ignores an empty chord name and a key that is not one of the 12', () => {
    const { result } = renderHook(() => useProgression());

    act(() => result.current.addChord('  '));
    act(() => result.current.setKey('E♯'));

    expect(result.current.progression).toEqual({ key: 'C', slots: [] });
  });

  it('adds a candidate of the beats given, and ignores beats outside 1 to 16', () => {
    const { result } = renderHook(() => useProgression());

    act(() => result.current.addCandidate('F', 2));
    act(() => result.current.addCandidate('G', 0));

    expect(result.current.progression.slots).toEqual([{ chord: 'F', beats: 2 }]);
    expect(result.current.bases).toEqual(['F']);
  });

  it('keeps the base of a candidate with its slot, through a removal and a replacement', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addChord('C'));
    act(() => result.current.addCandidate('G', 2));
    act(() => result.current.addNoChord());

    act(() => result.current.remove(0));
    act(() => result.current.replaceChord(0, 'Gsus4'));

    expect(result.current.progression.slots).toEqual([
      { chord: 'Gsus4', beats: 2 },
      { content: 'none', beats: 4 },
    ]);
    expect(result.current.bases).toEqual(['G', null]);
  });

  it('puts a chord in place of an empty slot, keeping its length', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addUnchosen());
    act(() => result.current.setBeats(0, 2));

    act(() => result.current.replaceChord(0, 'F'));

    expect(result.current.progression.slots).toEqual([{ chord: 'F', beats: 2 }]);
    expect(result.current.bases).toEqual([null]);
  });

  it('changes a length within 1 to 16 beats and ignores one outside', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addChord('C'));

    act(() => result.current.setBeats(0, 16));
    act(() => result.current.setBeats(0, 17));
    act(() => result.current.setBeats(0, 0));

    expect(result.current.progression.slots).toEqual([{ chord: 'C', beats: 16 }]);
  });

  it('removes a slot and keeps the chosen one chosen', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addChord('C'));
    act(() => result.current.addChord('F'));
    act(() => result.current.addChord('G'));
    act(() => result.current.select(2));

    act(() => result.current.remove(0));

    expect(result.current.progression.slots).toEqual([
      { chord: 'F', beats: 4 },
      { chord: 'G', beats: 4 },
    ]);
    expect(result.current.selected).toBe(1);

    act(() => result.current.remove(1));

    expect(result.current.selected).toBeNull();
  });

  it('brings the progression back after a reload', () => {
    const { result: before, unmount } = renderHook(() => useProgression());
    act(() => before.current.setKey('A'));
    act(() => before.current.addChord('F♯m'));
    act(() => before.current.addUnchosen());
    unmount();

    const { result } = renderHook(() => useProgression());

    expect(result.current.progression).toEqual({
      key: 'A',
      slots: [
        { chord: 'F♯m', beats: 4 },
        { content: 'unchosen', beats: 4 },
      ],
    });
    expect(JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '')).toEqual(result.current.progression);
  });

  it.each([
    { name: 'not JSON', stored: '{key' },
    { name: 'not an object', stored: '"C"' },
    { name: 'a key outside the 12', stored: '{"key":"E♯","slots":[]}' },
    { name: 'no slots', stored: '{"key":"C"}' },
    {
      name: 'a slot with both chord and content',
      stored: '{"key":"C","slots":[{"chord":"C","content":"none","beats":4}]}',
    },
    { name: 'a slot with neither', stored: '{"key":"C","slots":[{"beats":4}]}' },
    { name: 'an unknown content', stored: '{"key":"C","slots":[{"content":"rest","beats":4}]}' },
    { name: 'a length of 0', stored: '{"key":"C","slots":[{"chord":"C","beats":0}]}' },
    { name: 'a length of 1.5', stored: '{"key":"C","slots":[{"chord":"C","beats":1.5}]}' },
    {
      name: 'a field the screen does not write',
      stored: '{"key":"C","slots":[{"chord":"C","beats":4,"x":1}]}',
    },
  ])('starts empty when the stored value is $name', ({ stored }) => {
    localStorage.setItem(STORAGE_KEY, stored);

    const { result } = renderHook(() => useProgression());

    expect(result.current.progression).toEqual({ key: 'C', slots: [] });
  });
});
