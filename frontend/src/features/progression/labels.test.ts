import { describe, expect, it } from 'vitest';
import type { SlotReading, Variation } from '@/shared/api/client';
import { labelRows, variationMarks } from './labels';

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

describe('variationMarks', () => {
  const variation: Variation = {
    chord: { basis: 'definition', name: 'Fm7' },
    isInput: false,
    members: { basis: 'definition', notes: ['F', 'A♭', 'C', 'E♭'] },
    reading: fm7.reading,
    tensionCautions: { basis: 'school', cautions: [] },
    outOfKey: { basis: 'computation', notes: ['A♭', 'E♭'] },
  };

  it('names the notes outside the key after the reading, joined as the detail joins them', () => {
    expect(variationMarks(variation)).toEqual(['借用和音', '調の外の音 A♭、E♭']);
  });

  it('shows no mark for the notes outside the key when there are none', () => {
    const inKey = { ...variation, outOfKey: { basis: 'computation', notes: [] } };

    expect(variationMarks(inKey)).toEqual(['借用和音']);
  });
});
