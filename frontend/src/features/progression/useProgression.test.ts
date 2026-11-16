import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import type { Slot } from '@/shared/api/client';
import { STORAGE_KEY, fitToPhrases, maxBeatsAt, nextBeats, useProgression } from './useProgression';

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

describe('maxBeatsAt', () => {
  it.each([
    { name: 'the first slot', slots: slotsOf(4), index: 0, want: 16 },
    { name: 'after 4 beats', slots: slotsOf(4, 4), index: 1, want: 12 },
    { name: 'after 15 beats', slots: slotsOf(15, 1), index: 1, want: 1 },
    { name: 'the start of the second phrase', slots: slotsOf(16, 4), index: 1, want: 16 },
  ])('gives the beats to the end of the phrase: $name', ({ slots, index, want }) => {
    expect(maxBeatsAt(slots, index)).toBe(want);
  });
});

describe('fitToPhrases', () => {
  it.each([
    {
      name: 'after removing the first of [4, 12, 16]',
      beats: [12, 16],
      want: [12, 4],
      shortened: [1],
    },
    {
      name: 'after shortening the first of [4, 4, 8, 4] to 2',
      beats: [2, 4, 8, 4],
      want: [2, 4, 8, 2],
      shortened: [3],
    },
    {
      name: 'after lengthening the first of [2, 14] to 4',
      beats: [4, 14],
      want: [4, 12],
      shortened: [1],
    },
    {
      name: 'when no slot crosses',
      beats: [4, 4, 8, 16, 2],
      want: [4, 4, 8, 16, 2],
      shortened: [],
    },
    {
      // Shortening the second moves the fourth back by 2 beats, across the end of the second phrase.
      name: 'when shortening one makes a later one cross',
      beats: [12, 6, 14, 4],
      want: [12, 4, 14, 2],
      shortened: [1, 3],
    },
  ])('shortens each slot to the end of its phrase: $name', ({ beats, want, shortened }) => {
    const fitted = fitToPhrases(slotsOf(...beats));

    expect(fitted.slots).toEqual(slotsOf(...want));
    expect(fitted.shortened).toEqual(shortened);
  });

  it('keeps what a shortened slot holds', () => {
    const fitted = fitToPhrases([
      { chord: 'C', beats: 12 },
      { chord: 'G', beats: 8, role: 'rise' },
    ]);

    expect(fitted.slots).toEqual([
      { chord: 'C', beats: 12 },
      { chord: 'G', beats: 4, role: 'rise' },
    ]);
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

  it('shortens a slot that a removal makes cross, and tells which', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addChord('C'));
    act(() => result.current.addChord('F'));
    act(() => result.current.setBeats(1, 12));
    act(() => result.current.addChord('G'));
    act(() => result.current.setBeats(2, 16));

    act(() => result.current.remove(0));

    expect(result.current.progression.slots).toEqual([
      { chord: 'F', beats: 12 },
      { chord: 'G', beats: 4 },
    ]);
    expect(result.current.shortened).toEqual([{ chord: 'G', beats: 4 }]);
  });

  it('shortens a slot that a change of length before it makes cross', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addChord('C'));
    act(() => result.current.setBeats(0, 2));
    act(() => result.current.addUnchosen());
    act(() => result.current.setBeats(1, 14));

    act(() => result.current.setBeats(0, 4));

    expect(result.current.progression.slots).toEqual([
      { chord: 'C', beats: 4 },
      { content: 'unchosen', beats: 12 },
    ]);
    expect(result.current.shortened).toEqual([{ content: 'unchosen', beats: 12 }]);
  });

  it('forgets what it shortened at the next change of the progression', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addChord('C'));
    act(() => result.current.addChord('F'));
    act(() => result.current.setBeats(1, 12));
    act(() => result.current.addChord('G'));
    act(() => result.current.setBeats(2, 16));
    act(() => result.current.remove(0));

    act(() => result.current.select(1));

    expect(result.current.shortened).toHaveLength(1);

    act(() => result.current.setRole(0, 'stable'));

    expect(result.current.shortened).toEqual([]);
  });

  it('leaves a stored slot that crosses as it is until the first change', () => {
    localStorage.setItem(
      STORAGE_KEY,
      '{"key":"C","slots":[{"chord":"C","beats":12},{"chord":"G","beats":8}]}',
    );
    const { result } = renderHook(() => useProgression());

    expect(result.current.progression.slots).toEqual([
      { chord: 'C', beats: 12 },
      { chord: 'G', beats: 8 },
    ]);
    expect(result.current.shortened).toEqual([]);

    act(() => result.current.setKey('G'));

    expect(result.current.progression.slots).toEqual([
      { chord: 'C', beats: 12 },
      { chord: 'G', beats: 4 },
    ]);
    expect(result.current.shortened).toEqual([{ chord: 'G', beats: 4 }]);
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

  it('puts a role on a slot, keeps it through a length and a chord, and takes it away', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addUnchosen());
    act(() => result.current.addChord('C'));

    act(() => result.current.setRole(0, 'prepare'));
    act(() => result.current.setRole(1, 'wistful'));
    act(() => result.current.setBeats(0, 2));
    act(() => result.current.replaceChord(0, 'F'));

    expect(result.current.progression.slots).toEqual([
      { chord: 'F', beats: 2, role: 'prepare' },
      { chord: 'C', beats: 4, role: 'wistful' },
    ]);

    act(() => result.current.setRole(1, ''));
    act(() => result.current.setRole(0, 'tension'));

    expect(result.current.progression.slots).toEqual([
      { chord: 'F', beats: 2, role: 'prepare' },
      { chord: 'C', beats: 4 },
    ]);
  });

  it('fills a slot with a candidate, keeping its length and role, and makes it the base', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addCandidate('F', 2));
    act(() => result.current.addUnchosen());
    act(() => result.current.setRole(0, 'prepare'));

    act(() => result.current.fillSlot(0, 'Dm7'));
    act(() => result.current.fillSlot(1, 'G7'));

    expect(result.current.progression.slots).toEqual([
      { chord: 'Dm7', beats: 2, role: 'prepare' },
      { chord: 'G7', beats: 4 },
    ]);
    expect(result.current.bases).toEqual(['Dm7', 'G7']);
  });

  it('brings the roles back after a reload', () => {
    const { result: before, unmount } = renderHook(() => useProgression());
    act(() => before.current.addUnchosen());
    act(() => before.current.addNoChord());
    act(() => before.current.addChord('G'));
    act(() => before.current.setRole(0, 'stable'));
    act(() => before.current.setRole(1, 'linger'));
    act(() => before.current.setRole(2, 'rise'));
    unmount();

    const { result } = renderHook(() => useProgression());

    expect(result.current.progression.slots).toEqual([
      { content: 'unchosen', beats: 4, role: 'stable' },
      { content: 'none', beats: 4, role: 'linger' },
      { chord: 'G', beats: 4, role: 'rise' },
    ]);
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
    { name: 'an unknown role', stored: '{"key":"C","slots":[{"chord":"C","beats":4,"role":"x"}]}' },
    { name: 'a role of null', stored: '{"key":"C","slots":[{"chord":"C","beats":4,"role":null}]}' },
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
