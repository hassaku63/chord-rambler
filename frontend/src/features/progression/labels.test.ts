import { describe, expect, it } from 'vitest';
import type { Candidate, SlotReading, Variation } from '@/shared/api/client';
import { candidateRows, edgeWords, footprintRows, labelRows, variationMarks } from './labels';

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

describe('edgeWords', () => {
  const g7 = { ...fm7, resolution: { basis: 'school', resolution: 'expected' } } as const;
  const c = { ...fm7, cadence: { basis: 'school', cadence: 'authentic' } } as const;
  const unchosen: SlotReading = { content: 'unchosen' };
  it.each([
    { name: 'nothing to name', before: fm7, after: fm7, want: [] },
    { name: 'the resolution of the chord before', before: g7, after: fm7, want: ['解決'] },
    { name: 'the cadence of the slot after', before: fm7, after: c, want: ['正格終止'] },
    { name: 'both', before: g7, after: c, want: ['解決', '正格終止'] },
    { name: 'not the resolution of the slot after', before: fm7, after: g7, want: [] },
    { name: 'a slot with no chord', before: unchosen, after: unchosen, want: [] },
    { name: 'no reading yet', before: undefined, after: undefined, want: [] },
  ])('names the judgments of the line: $name', ({ before, after, want }) => {
    expect(edgeWords(before, after)).toEqual(want);
  });
});

describe('footprintRows', () => {
  it('names the resolution as that of the slot before, where the candidate rows name the last', () => {
    const g7: Candidate = {
      chord: { basis: 'definition', name: 'G7' },
      roman: { basis: 'computation', numeral: 'V7' },
      reading: { ...fm7.reading, kind: 'diatonic', function: 'dominant', modes: [] },
      resolution: { basis: 'school', resolution: 'expected' },
    };
    const term = (rows: { term: string }[]) => rows.map((row) => row.term).at(-1);

    expect(term(candidateRows(g7))).toBe('末尾の和音の解決');
    expect(term(footprintRows(g7))).toBe('前の枠の和音の解決');
  });
});
