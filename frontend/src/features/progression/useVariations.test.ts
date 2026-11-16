import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Variation, VariationsAnswer } from '@/shared/api/client';
import { useVariations } from './useVariations';

function variationOf(name: string): Variation {
  return {
    chord: { basis: 'definition', name },
    isInput: true,
    members: { basis: 'definition', notes: [] },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: 'dominant',
      modes: [],
      fromParallelMinor: false,
    },
    tensionCautions: { basis: 'school', cautions: [] },
    outOfKey: { basis: 'computation', notes: [] },
  };
}

function answerOf(name: string): VariationsAnswer {
  return { ok: true, variations: [variationOf(name)] };
}

describe('useVariations', () => {
  it('waits, then gives the variations of the chord in the key', async () => {
    const get = vi.fn(() => Promise.resolve(answerOf('G')));

    const { result } = renderHook(() => useVariations('C', 'G', get));

    expect(result.current).toEqual({ state: 'waiting' });
    await waitFor(() =>
      expect(result.current).toEqual({ state: 'given', variations: [variationOf('G')] }),
    );
    expect(get).toHaveBeenCalledWith('C', 'G', expect.any(AbortSignal));
  });

  it('asks nothing for no chord', () => {
    const get = vi.fn(() => Promise.resolve(answerOf('G')));

    const { result } = renderHook(() => useVariations('C', null, get));

    expect(result.current).toEqual({ state: 'none' });
    expect(get).not.toHaveBeenCalled();
  });

  it('keeps only the answer to the latest chord', async () => {
    let first: (a: VariationsAnswer) => void = () => undefined;
    const late = new Promise<VariationsAnswer>((r) => (first = r));
    const get = vi
      .fn()
      .mockReturnValueOnce(late)
      .mockReturnValueOnce(Promise.resolve(answerOf('F')));
    const { result, rerender } = renderHook(({ chord }) => useVariations('C', chord, get), {
      initialProps: { chord: 'G' },
    });

    rerender({ chord: 'F' });
    await waitFor(() => expect(result.current.state).toBe('given'));
    await act(async () => {
      first(answerOf('G'));
      await late;
    });

    expect(result.current).toEqual({ state: 'given', variations: [variationOf('F')] });
  });

  it('gives no list for a chord the server refuses', async () => {
    const refusal: VariationsAnswer = {
      ok: false,
      problem: { title: 'Bad Request', status: 400, detail: 'the chord has no third' },
    };
    const get = vi.fn(() => Promise.resolve(refusal));

    const { result } = renderHook(() => useVariations('C', 'Gsus4', get));

    await waitFor(() => expect(result.current).toEqual({ state: 'none' }));
  });
});
