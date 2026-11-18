import { describe, expect, it } from 'vitest';
import type { Slot } from '@/shared/api/client';
import { nextUnchosen } from './usePlanFill';

describe('nextUnchosen', () => {
  const u: Slot = { content: 'unchosen', beats: 4 };
  const c: Slot = { chord: 'C', beats: 4 };
  it.each<{ name: string; slots: Slot[]; index: number; want: number | null }>([
    { name: 'the next one after', slots: [c, u, c, u], index: 1, want: 3 },
    { name: 'the first after, not the one at the start', slots: [u, c, u, u], index: 1, want: 2 },
    { name: 'from the start when none is after', slots: [c, u, c, u], index: 3, want: 1 },
    { name: 'not the slot itself', slots: [c, u, c], index: 1, want: null },
    { name: 'none in a progression all chosen', slots: [c, c], index: 0, want: null },
    {
      name: 'none for a slot with no chord',
      slots: [{ content: 'none', beats: 4 }, c],
      index: 1,
      want: null,
    },
  ])('gives the next slot not chosen yet: $name', ({ slots, index, want }) => {
    expect(nextUnchosen(slots, index)).toBe(want);
  });
});
