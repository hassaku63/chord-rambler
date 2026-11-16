import { useEffect, useState } from 'react';
import type { Progression, SlotCandidateGroups } from '@/shared/api/client';
import { getSlotCandidates } from '@/shared/api/client';

/**
 * The candidates for a slot to fill: waiting before the first answer for the
 * slot, none for no slot, a refusal or no answer (the reading says why).
 */
export type SlotCandidatesState =
  | { state: 'waiting' }
  | { state: 'given'; groups: SlotCandidateGroups['groups'] }
  | { state: 'none' };

type Answer = { for: Progression; slot: number; state: SlotCandidatesState };

/**
 * Asks for the candidates for the slot (counted from 0) each time the
 * progression or the slot changes; asks nothing for no slot. An answer to a
 * progression or a slot since changed is thrown away. Until the answer to the
 * latest progression comes, the last answer for the same slot stays and
 * `busy` is true; the answer for another slot never shows.
 */
export function useSlotCandidates(
  progression: Progression,
  slot: number | null,
  get: typeof getSlotCandidates = getSlotCandidates,
): { candidates: SlotCandidatesState; busy: boolean } {
  const [answer, setAnswer] = useState<Answer | null>(null);

  useEffect(() => {
    if (slot === null) {
      return;
    }
    const controller = new AbortController();
    const keep = (state: SlotCandidatesState) => {
      if (!controller.signal.aborted) {
        setAnswer({ for: progression, slot, state });
      }
    };
    get(progression, slot, controller.signal)
      .then((a) => keep(a.ok ? { state: 'given', groups: a.groups } : { state: 'none' }))
      .catch(() => keep({ state: 'none' }));
    return () => controller.abort();
  }, [progression, slot, get]);

  if (slot === null) {
    return { candidates: { state: 'none' }, busy: false };
  }
  if (answer?.slot !== slot) {
    return { candidates: { state: 'waiting' }, busy: true };
  }
  return { candidates: answer.state, busy: answer.for !== progression };
}
