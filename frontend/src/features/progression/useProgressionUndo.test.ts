import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import type { ProgressionState } from './useProgression';
import { STORAGE_KEY, useProgression } from './useProgression';

afterEach(() => localStorage.clear());

describe('useProgression undo', () => {
  /** C | not chosen yet | G7 from the candidates, in C. */
  function built(state: ProgressionState) {
    state.addChord('C');
    state.addUnchosen();
    state.addCandidate('G7', 4);
  }

  it.each<{ name: string; change: (state: ProgressionState) => void }>([
    { name: 'adding a typed chord', change: (s) => s.addChord('F') },
    { name: 'adding a candidate', change: (s) => s.addCandidate('F', 2) },
    { name: 'adding a slot with no chord', change: (s) => s.addNoChord() },
    { name: 'adding a slot not chosen yet', change: (s) => s.addUnchosen() },
    { name: 'removing a slot', change: (s) => s.remove(0) },
    { name: 'changing a length', change: (s) => s.setBeats(0, 2) },
    { name: 'putting a role', change: (s) => s.setRole(1, 'stable') },
    { name: 'a variation', change: (s) => s.replaceChord(2, 'Gsus4') },
    { name: 'filling a slot', change: (s) => s.fillSlot(1, 'Am') },
    {
      name: 'putting in a branch of the footprint',
      change: (s) => {
        s.fillSlot(2, 'Bdim');
        s.setFootprint({ at: 2, branches: [] });
      },
    },
    {
      name: 'putting in a branch after the last',
      change: (s) => {
        s.addCandidate('C', 4);
        s.setFootprint({ at: 3, branches: [] });
      },
    },
    { name: 'changing the key', change: (s) => s.setKey('D') },
  ])('takes back $name, and saves what it brings back', ({ change }) => {
    const { result } = renderHook(() => useProgression());
    act(() => built(result.current));
    const { progression, bases } = result.current;
    act(() => change(result.current));
    expect(result.current.progression).not.toEqual(progression);

    act(() => result.current.undo());

    expect(result.current.progression).toEqual(progression);
    expect(result.current.bases).toEqual(bases);
    expect(result.current.footprint).toBeNull();
    expect(JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '')).toEqual(progression);
  });

  it('takes back one change at a time, and has nothing to undo at the start', () => {
    const { result } = renderHook(() => useProgression());
    expect(result.current.canUndo).toBe(false);
    act(() => result.current.addChord('C'));
    act(() => result.current.addChord('G'));

    act(() => result.current.undo());

    expect(result.current.progression.slots).toEqual([{ chord: 'C', beats: 4 }]);
    expect(result.current.canUndo).toBe(true);

    act(() => result.current.undo());
    act(() => result.current.undo());

    expect(result.current.progression.slots).toEqual([]);
    expect(result.current.canUndo).toBe(false);
  });

  it('keeps nothing to undo for a change that changes nothing, nor after a reload', () => {
    const { result: before, unmount } = renderHook(() => useProgression());
    act(() => before.current.addChord('C'));
    unmount();

    const { result } = renderHook(() => useProgression());

    expect(result.current.canUndo).toBe(false);
    act(() => result.current.replaceChord(0, 'C'));
    act(() => result.current.setBeats(0, 4));
    expect(result.current.canUndo).toBe(false);
  });

  it('fits what it brings back to the phrases, and tells what it shortened', () => {
    localStorage.setItem(
      STORAGE_KEY,
      '{"key":"C","slots":[{"chord":"C","beats":12},{"chord":"G","beats":8}]}',
    );
    const { result } = renderHook(() => useProgression());
    act(() => result.current.setKey('G'));

    act(() => result.current.undo());

    expect(result.current.progression).toEqual({
      key: 'C',
      slots: [
        { chord: 'C', beats: 12 },
        { chord: 'G', beats: 4 },
      ],
    });
    expect(result.current.shortened).toEqual([{ chord: 'G', beats: 4 }]);
  });

  it('lets go of a chosen slot that an undo takes away', () => {
    const { result } = renderHook(() => useProgression());
    act(() => result.current.addChord('C'));
    act(() => result.current.addChord('G'));
    act(() => result.current.select(1));
    act(() => result.current.setRole(0, 'stable'));

    act(() => result.current.undo());

    expect(result.current.selected).toBe(1);

    act(() => result.current.undo());

    expect(result.current.selected).toBeNull();
  });
});
