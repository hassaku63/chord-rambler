import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Progression, ProgressionReading, ReadingAnswer } from '@/shared/api/client';
import { useProgressionReading } from './useProgressionReading';

const cMajor: Progression = { key: 'C', slots: [{ chord: 'C', beats: 4 }] };
const gMajor: Progression = { key: 'G', slots: [{ chord: 'G', beats: 4 }] };

/** The reading of the tonic chord alone in its own key: C in C, G in G. */
function readingOf(name: string): ProgressionReading {
  return {
    slots: [
      {
        content: 'chord',
        chord: { basis: 'definition', name },
        roman: { basis: 'computation', numeral: 'I' },
        reading: {
          basis: 'school',
          kind: 'diatonic',
          function: 'tonic',
          modes: [],
          fromParallelMinor: false,
        },
        resolution: { basis: 'school', resolution: 'none' },
        cadence: { basis: 'school', cadence: 'none' },
        tensionCautions: { basis: 'school', cautions: [] },
      },
    ],
  };
}

function answerOf(name: string): ReadingAnswer {
  return { ok: true, reading: readingOf(name) };
}

function deferred() {
  let resolve: (answer: ReadingAnswer) => void = () => undefined;
  const promise = new Promise<ReadingAnswer>((r) => (resolve = r));
  return { promise, resolve };
}

describe('useProgressionReading', () => {
  it('waits, then gives the reading of the progression', async () => {
    const read = vi.fn(() => Promise.resolve(answerOf('C')));

    const { result } = renderHook(() => useProgressionReading(cMajor, read));

    expect(result.current).toEqual({ state: 'waiting' });
    await waitFor(() => expect(result.current).toEqual({ state: 'read', reading: readingOf('C') }));
    expect(read).toHaveBeenCalledWith(cMajor, expect.any(AbortSignal));
  });

  it('keeps only the answer to the latest progression', async () => {
    const first = deferred();
    const second = deferred();
    const read = vi.fn().mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const { result, rerender } = renderHook(({ p }) => useProgressionReading(p, read), {
      initialProps: { p: cMajor },
    });

    rerender({ p: gMajor });
    second.resolve(answerOf('G'));
    await waitFor(() => expect(result.current.state).toBe('read'));
    await act(async () => {
      first.resolve(answerOf('C'));
      await first.promise;
    });

    expect(result.current).toEqual({ state: 'read', reading: readingOf('G') });
  });

  it.each([
    {
      name: 'a detail naming a slot',
      problem: { title: 'Bad Request', status: 400, detail: 'slot 2: cannot read "Hm"' },
      want: { state: 'refused', detail: 'slot 2: cannot read "Hm"', slot: 2 },
    },
    {
      name: 'a slot field',
      problem: { title: 'Bad Request', status: 400, detail: 'cannot read "Hm"', slot: 4 },
      want: { state: 'refused', detail: 'cannot read "Hm"', slot: 4 },
    },
    {
      name: 'a slot field and a detail naming another slot',
      problem: { title: 'Bad Request', status: 400, detail: 'slot 2: x', slot: 5 },
      want: { state: 'refused', detail: 'slot 2: x', slot: 5 },
    },
    {
      name: 'a detail naming no slot',
      problem: { title: 'Bad Request', status: 400, detail: 'key: cannot read "H"' },
      want: { state: 'refused', detail: 'key: cannot read "H"', slot: null },
    },
    {
      name: 'no detail',
      problem: { title: 'Payload Too Large', status: 413 },
      want: { state: 'refused', detail: 'Payload Too Large', slot: null },
    },
  ])('passes on the problem with $name', async ({ problem, want }) => {
    const read = vi.fn(() => Promise.resolve<ReadingAnswer>({ ok: false, problem }));

    const { result } = renderHook(() => useProgressionReading(cMajor, read));

    await waitFor(() => expect(result.current).toEqual(want));
  });

  it('says the reading is unreachable when the request fails', async () => {
    const read = vi.fn(() => Promise.reject(new Error('network down')));

    const { result } = renderHook(() => useProgressionReading(cMajor, read));

    await waitFor(() => expect(result.current).toEqual({ state: 'unreachable' }));
  });
});
