import { useEffect, useState } from 'react';
import type { Variation } from '@/shared/api/client';
import { getVariations } from '@/shared/api/client';

/**
 * The variations of a chord. A chord the server refuses (one with no third,
 * as a sus chord typed in) and no answer both show no list.
 */
export type VariationsState =
  { state: 'waiting' } | { state: 'given'; variations: Variation[] } | { state: 'none' };

type Answer = { key: string; chord: string; state: VariationsState };

/**
 * Asks for the variations of the chord in the key each time either changes;
 * asks nothing for no chord. An answer to a chord or key since changed is
 * thrown away.
 */
export function useVariations(
  key: string,
  chord: string | null,
  get: typeof getVariations = getVariations,
): VariationsState {
  const [answer, setAnswer] = useState<Answer | null>(null);

  useEffect(() => {
    if (chord === null) {
      return;
    }
    const controller = new AbortController();
    const keep = (state: VariationsState) => {
      if (!controller.signal.aborted) {
        setAnswer({ key, chord, state });
      }
    };
    get(key, chord, controller.signal)
      .then((a) => keep(a.ok ? { state: 'given', variations: a.variations } : { state: 'none' }))
      .catch(() => keep({ state: 'none' }));
    return () => controller.abort();
  }, [key, chord, get]);

  if (chord === null) {
    return { state: 'none' };
  }
  return answer?.key === key && answer.chord === chord ? answer.state : { state: 'waiting' };
}
