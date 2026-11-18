import { describe, expect, it } from 'vitest';
import type { SlotReading } from '@/shared/api/client';
import { labelRows } from './labels';

const fm7: Extract<SlotReading, { content: 'chord' }> = {
  content: 'chord',
  chord: { basis: 'definition', name: 'Fm7' },
  reading: {
    basis: 'school',
    kind: 'borrowed',
    function: 'none',
    modes: ['dorian', 'phrygian', 'aeolian'],
    fromParallelMinor: true,
  },
  resolution: { basis: 'school', resolution: 'none' },
  cadence: { basis: 'school', cadence: 'none' },
  tensionCautions: { basis: 'school', cautions: [] },
};

describe('labelRows', () => {
  it('names the notes outside the key in one row, in the order the server gives them', () => {
    const slot = { ...fm7, outOfKey: { basis: 'computation', notes: ['A♭', 'E♭'] } };

    expect(labelRows(slot).find((row) => row.term === '調の外の音')).toEqual({
      term: '調の外の音',
      word: 'A♭、E♭',
      basis: '計算',
    });
  });

  it('shows no row for the notes outside the key when the reading has none', () => {
    expect(labelRows(fm7).map((row) => row.term)).not.toContain('調の外の音');
  });
});
