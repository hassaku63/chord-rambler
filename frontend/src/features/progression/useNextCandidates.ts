import { useEffect, useState } from 'react';
import type { CandidateGroups, Progression } from '@/shared/api/client';
import { getNextCandidates } from '@/shared/api/client';
import { nextBeats } from './useProgression';

/**
 * The candidates for the slot after the last, waiting before the first
 * answer. The server's refusal, or no answer, shows no candidates: the
 * reading of the progression says why.
 */
export type CandidatesState =
  { state: 'waiting' } | { state: 'given'; groups: CandidateGroups['groups'] } | { state: 'none' };

type Answer = { for: Progression; state: CandidatesState };

/**
 * Asks for the candidates each time the progression changes, for a slot of
 * nextBeats. An answer to a progression that has since changed is thrown away.
 * Until the answer to the latest progression comes, the last answer stays and
 * `busy` is true, so that the list does not leave the screen meanwhile.
 */
export function useNextCandidates(
  progression: Progression,
  get: typeof getNextCandidates = getNextCandidates,
): { beats: number; candidates: CandidatesState; busy: boolean } {
  const [answer, setAnswer] = useState<Answer | null>(null);
  const beats = nextBeats(progression.slots);

  useEffect(() => {
    const controller = new AbortController();
    const keep = (state: CandidatesState) => {
      if (!controller.signal.aborted) {
        setAnswer({ for: progression, state });
      }
    };
    get(progression, beats, controller.signal)
      .then((a) => keep(a.ok ? { state: 'given', groups: a.groups } : { state: 'none' }))
      .catch(() => keep({ state: 'none' }));
    return () => controller.abort();
  }, [progression, beats, get]);

  return {
    beats,
    candidates: answer?.state ?? { state: 'waiting' },
    busy: answer?.for !== progression,
  };
}
