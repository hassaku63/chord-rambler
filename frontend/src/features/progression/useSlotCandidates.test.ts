import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Progression, SlotCandidatesAnswer } from '@/shared/api/client';
import { useSlotCandidates } from './useSlotCandidates';

const planned: Progression = {
  key: 'C',
  slots: [
    { chord: 'C', beats: 4, role: 'stable' },
    { content: 'unchosen', beats: 4, role: 'prepare' },
    { chord: 'G', beats: 4, role: 'rise' },
  ],
};
const filled: Progression = {
  key: 'C',
  slots: [
    { chord: 'C', beats: 4, role: 'stable' },
    { chord: 'F', beats: 4, role: 'prepare' },
    { chord: 'G', beats: 4, role: 'rise' },
  ],
};

function answerOf(name: string): SlotCandidatesAnswer {
  const candidate = {
    chord: { basis: 'definition', name },
    roman: { basis: 'computation', numeral: 'IV' },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: 'subdominant',
      modes: [],
      fromParallelMinor: false,
    },
    roleFit: { basis: 'school', fit: 'fits', unmet: [] },
  };
  return { ok: true, groups: [{ kind: 'diatonic', candidates: [candidate] }] };
}

function groupsOf(answer: SlotCandidatesAnswer) {
  return answer.ok ? answer.groups : [];
}

function deferred() {
  let resolve: (answer: SlotCandidatesAnswer) => void = () => undefined;
  const promise = new Promise<SlotCandidatesAnswer>((r) => (resolve = r));
  return { promise, resolve };
}

describe('useSlotCandidates', () => {
  it('asks nothing for no slot', () => {
    const get = vi.fn(() => Promise.resolve(answerOf('F')));

    const { result } = renderHook(() => useSlotCandidates(planned, null, get));

    expect(result.current).toEqual({ candidates: { state: 'none' }, busy: false });
    expect(get).not.toHaveBeenCalled();
  });

  it('waits, then gives the groups for the slot', async () => {
    const get = vi.fn(() => Promise.resolve(answerOf('F')));

    const { result } = renderHook(() => useSlotCandidates(planned, 1, get));

    expect(result.current).toEqual({ candidates: { state: 'waiting' }, busy: true });
    await waitFor(() => expect(result.current.busy).toBe(false));
    expect(result.current.candidates).toEqual({ state: 'given', groups: groupsOf(answerOf('F')) });
    expect(get).toHaveBeenCalledWith(planned, 1, expect.any(AbortSignal));
  });

  it('keeps the answer for the slot while it waits for the filled progression', async () => {
    const next = deferred();
    const get = vi
      .fn()
      .mockReturnValueOnce(Promise.resolve(answerOf('F')))
      .mockReturnValueOnce(next.promise);
    const { result, rerender } = renderHook(({ p }) => useSlotCandidates(p, 1, get), {
      initialProps: { p: planned },
    });
    await waitFor(() => expect(result.current.busy).toBe(false));

    rerender({ p: filled });

    expect(result.current).toEqual({
      candidates: { state: 'given', groups: groupsOf(answerOf('F')) },
      busy: true,
    });
    await act(async () => {
      next.resolve(answerOf('Dm'));
      await next.promise;
    });
    expect(result.current).toEqual({
      candidates: { state: 'given', groups: groupsOf(answerOf('Dm')) },
      busy: false,
    });
  });

  it('shows no answer for another slot, and drops the answer for a slot left', async () => {
    const first = deferred();
    const get = vi
      .fn()
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(new Promise(() => undefined));
    const { result, rerender } = renderHook(({ slot }) => useSlotCandidates(planned, slot, get), {
      initialProps: { slot: 1 },
    });

    rerender({ slot: 0 });
    await act(async () => {
      first.resolve(answerOf('F'));
      await first.promise;
    });

    expect(result.current).toEqual({ candidates: { state: 'waiting' }, busy: true });
    expect(get).toHaveBeenLastCalledWith(planned, 0, expect.any(AbortSignal));
  });

  it('drops the answer when the slot is let go, so that it does not show for the slot chosen again', async () => {
    const removed: Progression = { key: 'C', slots: filled.slots.slice(1) };
    const get = vi
      .fn()
      .mockReturnValueOnce(Promise.resolve(answerOf('F')))
      .mockReturnValue(new Promise(() => undefined));
    const initialProps: { p: Progression; slot: number | null } = { p: planned, slot: 1 };
    const { result, rerender } = renderHook(({ p, slot }) => useSlotCandidates(p, slot, get), {
      initialProps,
    });
    await waitFor(() => expect(result.current.busy).toBe(false));

    rerender({ p: planned, slot: null });
    rerender({ p: removed, slot: null });
    rerender({ p: removed, slot: 1 });

    expect(result.current).toEqual({ candidates: { state: 'waiting' }, busy: true });
    expect(get).toHaveBeenLastCalledWith(removed, 1, expect.any(AbortSignal));
  });

  it('gives none when the server refuses', async () => {
    const get = vi.fn(() =>
      Promise.resolve<SlotCandidatesAnswer>({
        ok: false,
        problem: { title: 'Bad Request', status: 400 },
      }),
    );

    const { result } = renderHook(() => useSlotCandidates(planned, 1, get));

    await waitFor(() => expect(result.current.candidates).toEqual({ state: 'none' }));
  });
});
