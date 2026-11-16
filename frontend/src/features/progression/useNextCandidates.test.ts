import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { CandidatesAnswer, Progression, Slot } from '@/shared/api/client';
import { nextBeats, useNextCandidates } from './useNextCandidates';

const cMajor: Progression = { key: 'C', slots: [{ chord: 'C', beats: 4 }] };
const gMajor: Progression = { key: 'G', slots: [{ chord: 'G', beats: 4 }] };

function answerOf(name: string): CandidatesAnswer {
  const candidate = {
    chord: { basis: 'definition', name },
    roman: { basis: 'computation', numeral: 'I' },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: 'tonic',
      modes: [],
      fromParallelMinor: false,
    },
  };
  return { ok: true, groups: [{ kind: 'diatonic', candidates: [candidate] }] };
}

function deferred() {
  let resolve: (answer: CandidatesAnswer) => void = () => undefined;
  const promise = new Promise<CandidatesAnswer>((r) => (resolve = r));
  return { promise, resolve };
}

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

describe('useNextCandidates', () => {
  it('waits, then gives the groups, asking for the rounded beats', async () => {
    const short: Progression = { key: 'C', slots: [{ chord: 'C', beats: 14 }] };
    const get = vi.fn(() => Promise.resolve(answerOf('F')));

    const { result } = renderHook(() => useNextCandidates(short, get));

    expect(result.current).toEqual({ beats: 2, candidates: { state: 'waiting' } });
    await waitFor(() => expect(result.current.candidates.state).toBe('given'));
    expect(get).toHaveBeenCalledWith(short, 2, expect.any(AbortSignal));
  });

  it('keeps only the answer to the latest progression', async () => {
    const first = deferred();
    const second = deferred();
    const get = vi.fn().mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const { result, rerender } = renderHook(({ p }) => useNextCandidates(p, get), {
      initialProps: { p: cMajor },
    });

    rerender({ p: gMajor });
    second.resolve(answerOf('G'));
    await waitFor(() => expect(result.current.candidates.state).toBe('given'));
    await act(async () => {
      first.resolve(answerOf('C'));
      await first.promise;
    });

    const answer = answerOf('G');
    expect(result.current.candidates).toEqual({
      state: 'given',
      groups: answer.ok && answer.groups,
    });
  });

  it.each([
    {
      name: 'refuses',
      get: () =>
        Promise.resolve<CandidatesAnswer>({
          ok: false,
          problem: { title: 'Bad Request', status: 400 },
        }),
    },
    { name: 'cannot be reached', get: () => Promise.reject(new Error('network down')) },
  ])('gives no candidates when the server $name', async ({ get }) => {
    const { result } = renderHook(() => useNextCandidates(cMajor, vi.fn(get)));

    await waitFor(() => expect(result.current.candidates).toEqual({ state: 'none' }));
  });
});
