import { useEffect, useState } from 'react';
import type { Progression, ProgressionReading, ReadingAnswer } from '@/shared/api/client';
import { readProgression } from '@/shared/api/client';

/**
 * The reading of the current progression. A progression just changed is
 * waiting; a refused one carries the problem's detail and, when the detail
 * names one, the slot (counted from 0): the problem's slot field, else the slot
 * the detail's text names.
 */
export type ReadingState =
  | { state: 'waiting' }
  | { state: 'read'; reading: ProgressionReading }
  | { state: 'refused'; detail: string; slot: number | null }
  | { state: 'unreachable' };

type Answer = { for: Progression; state: ReadingState };

/** The fallback for a problem with no slot field: the detail is written "slot 2: ...". */
function slotNamedIn(detail: string): number | null {
  const match = /^slot (\d+):/.exec(detail);
  return match?.[1] === undefined ? null : Number(match[1]);
}

function stateOf(answer: ReadingAnswer): ReadingState {
  if (answer.ok) {
    return { state: 'read', reading: answer.reading };
  }
  const detail = answer.problem.detail ?? answer.problem.title;
  return { state: 'refused', detail, slot: answer.problem.slot ?? slotNamedIn(detail) };
}

/**
 * Reads the progression each time it changes. An answer to a progression that
 * has since changed is thrown away, so only the latest one shows.
 */
export function useProgressionReading(
  progression: Progression,
  read: typeof readProgression = readProgression,
): ReadingState {
  const [answer, setAnswer] = useState<Answer | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    read(progression, controller.signal)
      .then((a) => {
        if (!controller.signal.aborted) {
          setAnswer({ for: progression, state: stateOf(a) });
        }
      })
      .catch(() => {
        if (!controller.signal.aborted) {
          setAnswer({ for: progression, state: { state: 'unreachable' } });
        }
      });
    return () => controller.abort();
  }, [progression, read]);

  return answer?.for === progression ? answer.state : { state: 'waiting' };
}
