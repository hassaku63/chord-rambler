import { useEffect, useState } from 'react';
import type { CandidateGroups, Progression, Slot } from '@/shared/api/client';
import { getNextCandidates } from '@/shared/api/client';
import { DEFAULT_BEATS } from './useProgression';

/** The beats of a phrase: four bars of 4/4, cut from the start (docs/api/openapi.yaml). */
export const PHRASE_BEATS = 16;

/**
 * The length of the slot a candidate goes in: one bar, or the beats left in
 * the phrase when fewer are left, so that the slot does not cross into the
 * next phrase (#116).
 */
export function nextBeats(slots: Slot[]): number {
  const used = slots.reduce((sum, slot) => sum + slot.beats, 0) % PHRASE_BEATS;
  return Math.min(DEFAULT_BEATS, PHRASE_BEATS - used);
}

/**
 * The candidates for the slot after the last. The server's refusal, or no
 * answer, shows no candidates: the reading of the progression says why.
 */
export type CandidatesState =
  { state: 'waiting' } | { state: 'given'; groups: CandidateGroups['groups'] } | { state: 'none' };

type Answer = { for: Progression; state: CandidatesState };

/**
 * Asks for the candidates each time the progression changes, for a slot of
 * nextBeats. An answer to a progression that has since changed is thrown away.
 */
export function useNextCandidates(
  progression: Progression,
  get: typeof getNextCandidates = getNextCandidates,
): { beats: number; candidates: CandidatesState } {
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

  return { beats, candidates: answer?.for === progression ? answer.state : { state: 'waiting' } };
}
