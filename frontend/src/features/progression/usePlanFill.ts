import { useEffect, useRef } from 'react';
import type { Slot } from '@/shared/api/client';
import type { ProgressionState } from './useProgression';

/**
 * The slot not chosen yet that comes next after the one at the index: the
 * first after it, else the first from the start, else null.
 */
export function nextUnchosen(slots: Slot[], index: number): number | null {
  const after = slots.findIndex((slot, i) => i > index && slot.content === 'unchosen');
  const from = slots.findIndex((slot, i) => i < index && slot.content === 'unchosen');
  return after !== -1 ? after : from !== -1 ? from : null;
}

/**
 * Filling a slot in the plan tab (docs/product/screen-promise.md): the
 * candidate goes in, and the next slot not chosen yet is chosen, from after
 * it and then from the start; with none, the slot stays chosen. When the
 * choice moves, the focus goes to the heading of the context panel, which
 * names the slot now chosen, since the candidates it was on go away until
 * the new slot's come.
 */
export function usePlanFill(state: ProgressionState) {
  const headingRef = useRef<HTMLHeadingElement>(null);
  const pending = useRef(false);
  const { selected } = state;

  useEffect(() => {
    if (pending.current) {
      pending.current = false;
      headingRef.current?.focus();
    }
  }, [selected]);

  const fill = (index: number, chord: string) => {
    const next = nextUnchosen(state.progression.slots, index);
    state.fillSlot(index, chord);
    if (next !== null) {
      pending.current = true;
      state.select(next);
    }
  };

  return { fill, headingRef };
}
