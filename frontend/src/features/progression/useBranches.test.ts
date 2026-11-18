import { describe, expect, it } from 'vitest';
import type { Candidate } from '@/shared/api/client';
import { BRANCH_COUNT, branchesOf } from './useBranches';

function candidate(name: string): Candidate {
  return {
    chord: { basis: 'definition', name },
    roman: { basis: 'computation', numeral: 'I' },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: 'tonic',
      modes: [],
      fromParallelMinor: false,
    },
  };
}

const group = (kind: string, names: string[]) => ({ kind, candidates: names.map(candidate) });
const names = (list: Candidate[]) => list.map((c) => c.chord.name);

describe('branchesOf', () => {
  it('takes the first five of the answer, in the order of its groups (screen-promise.md, 5)', () => {
    const groups = [
      group('diatonic', ['Dm', 'Dm7', 'F', 'Fmaj7']),
      group('secondary-dominant', ['A', 'A7']),
      group('borrowed', ['E♭']),
    ];
    expect(BRANCH_COUNT).toBe(5);
    expect(names(branchesOf(groups))).toEqual(['Dm', 'Dm7', 'F', 'Fmaj7', 'A']);
  });

  it('takes all of a shorter answer, and none of an empty one', () => {
    expect(names(branchesOf([group('diatonic', ['C']), group('borrowed', ['Fm'])]))).toEqual([
      'C',
      'Fm',
    ]);
    expect(branchesOf([])).toEqual([]);
  });
});
