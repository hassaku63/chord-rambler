import { describe, expect, it } from 'vitest';
import {
  parseCandidateGroups,
  parseSlotCandidateGroups,
  parseHealth,
  parseProblem,
  parseProgressionNotes,
  parseProgressionReading,
  parseVariations,
} from './client';

describe('parseHealth', () => {
  it('accepts the body the contract describes', () => {
    expect(parseHealth({ status: 'ok', version: 'abc1234' })).toEqual({
      status: 'ok',
      version: 'abc1234',
    });
  });

  it.each([
    { name: 'null', body: null },
    { name: 'a string', body: 'ok' },
    { name: 'no version', body: { status: 'ok' } },
    { name: 'no status', body: { version: 'abc1234' } },
    { name: 'another status', body: { status: 'down', version: 'abc1234' } },
    { name: 'a version that is not a string', body: { status: 'ok', version: 1 } },
  ])('rejects $name', ({ body }) => {
    expect(() => parseHealth(body)).toThrow('unexpected body');
  });
});

describe('parseProgressionReading', () => {
  const bare = {
    content: 'chord',
    chord: { basis: 'definition', name: 'D7' },
    reading: {
      basis: 'school',
      kind: 'secondary-dominant',
      function: 'none',
      modes: [],
      fromParallelMinor: false,
      target: 5,
    },
    resolution: { basis: 'school', resolution: 'expected' },
    cadence: { basis: 'school', cadence: 'none' },
    tensionCautions: { basis: 'school', cautions: [{ tension: 'P11', chordTone: 'M3' }] },
  };
  const d7 = {
    ...bare,
    roman: { basis: 'school', numeral: 'V7/V' },
    roleFit: { basis: 'school', fit: 'fits' },
  };
  const full = {
    slots: [
      d7,
      bare,
      { content: 'none' },
      { content: 'unchosen', roleFit: { basis: 'school', fit: 'undecided' } },
    ],
  };

  it('accepts the body the contract describes, leaving out what is absent', () => {
    const { slots } = parseProgressionReading(full);

    expect(slots).toEqual(full.slots);
    expect(slots[1]).not.toHaveProperty('roman');
    expect(slots[1]).not.toHaveProperty('roleFit');
    expect(slots[2]).not.toHaveProperty('roleFit');
  });

  it('keeps an identifier the contract does not list yet', () => {
    const body = { slots: [{ ...d7, cadence: { basis: 'association', cadence: 'picardy' } }] };

    expect(parseProgressionReading(body)).toEqual(body);
  });

  it.each(['chord', 'reading', 'resolution', 'cadence', 'tensionCautions'])(
    'rejects a slot with a chord without %s',
    (name) => {
      const slot = Object.fromEntries(Object.entries(d7).filter(([field]) => field !== name));

      expect(() => parseProgressionReading({ slots: [slot] })).toThrow(`slots[0].${name}`);
    },
  );

  it.each([
    { name: 'null', body: null },
    { name: 'no slots', body: {} },
    { name: 'a slot that is not an object', body: { slots: ['C'] } },
    { name: 'a slot without content', body: { slots: [{}] } },
    { name: 'a content the contract does not list', body: { slots: [{ content: 'rest' }] } },
    { name: 'a label without basis', body: { slots: [{ ...d7, chord: { name: 'C' } }] } },
    {
      name: 'a numeral that is not a string',
      body: { slots: [{ ...d7, roman: { basis: 'x', numeral: 1 } }] },
    },
    {
      name: 'modes that are not strings',
      body: { slots: [{ ...d7, reading: { ...d7.reading, kind: 'borrowed', modes: [1] } }] },
    },
    {
      name: 'a caution without chord tone',
      body: {
        slots: [{ ...d7, tensionCautions: { basis: 'x', cautions: [{ tension: 'P11' }] } }],
      },
    },
    {
      name: 'a role fit that is not a label',
      body: { slots: [{ content: 'none', roleFit: 'fits' }] },
    },
  ])('rejects $name', ({ body }) => {
    expect(() => parseProgressionReading(body)).toThrow('unexpected body');
  });
});

describe('parseProblem', () => {
  it('accepts a problem with and without detail', () => {
    expect(parseProblem({ title: 'Bad Request', status: 400, detail: 'slot 0: x' })).toEqual({
      title: 'Bad Request',
      status: 400,
      detail: 'slot 0: x',
    });
    expect(parseProblem({ title: 'Bad Request', status: 400 })).toEqual({
      title: 'Bad Request',
      status: 400,
    });
  });

  it('reads the slot field', () => {
    expect(parseProblem({ title: 'Bad Request', status: 400, slot: 3 })).toEqual({
      title: 'Bad Request',
      status: 400,
      slot: 3,
    });
  });

  it.each([
    { name: 'a slot that is not a number', body: { title: 'x', status: 400, slot: '1' } },
    { name: 'a slot that is not an integer', body: { title: 'x', status: 400, slot: 1.5 } },
    { name: 'a slot below 0', body: { title: 'x', status: 400, slot: -1 } },
    { name: 'no title', body: { status: 400 } },
    { name: 'a status that is not a number', body: { title: 'x', status: '400' } },
    { name: 'a detail that is not a string', body: { title: 'x', status: 400, detail: 1 } },
  ])('rejects $name', ({ body }) => {
    expect(() => parseProblem(body)).toThrow('unexpected body');
  });
});

describe('parseProgressionNotes', () => {
  it('accepts the body the contract describes', () => {
    const notes = [
      { pitch: 48, start: 0, beats: 4, velocity: 64 },
      { pitch: 60, start: 0, beats: 4, velocity: 64 },
    ];

    expect(parseProgressionNotes({ notes })).toEqual(notes);
    expect(parseProgressionNotes({ notes: [] })).toEqual([]);
  });

  it.each([
    { name: 'no notes', body: {} },
    { name: 'a note that is not an object', body: { notes: [60] } },
    { name: 'no velocity', body: { notes: [{ pitch: 60, start: 0, beats: 4 }] } },
    {
      name: 'a pitch that is not an integer',
      body: { notes: [{ pitch: 60.5, start: 0, beats: 4, velocity: 64 }] },
    },
    {
      name: 'a start that is not a number',
      body: { notes: [{ pitch: 60, start: '0', beats: 4, velocity: 64 }] },
    },
  ])('rejects $name', ({ body }) => {
    expect(() => parseProgressionNotes(body)).toThrow('unexpected body');
  });
});

describe('parseCandidateGroups', () => {
  const bare = {
    chord: { basis: 'definition', name: 'F' },
    roman: { basis: 'computation', numeral: 'IV' },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: 'subdominant',
      modes: [],
      fromParallelMinor: false,
    },
  };
  const full = {
    ...bare,
    rootMotion: { basis: 'computation', interval: 'P4' },
    commonTones: { basis: 'computation', count: 1 },
    resolution: { basis: 'school', resolution: 'none' },
    cadence: { basis: 'school', cadence: 'none' },
    preference: {
      basis: 'convention',
      from: 'tonic',
      to: 'subdominant',
      preference: 'common',
      style: 'classical',
    },
  };
  const bodyOf = (candidate: object) => ({
    groups: [{ kind: 'diatonic', candidates: [candidate] }],
  });

  it.each([
    { name: 'every label', body: bodyOf(full) },
    { name: 'the labels that are absent left out', body: bodyOf(bare) },
    { name: 'an empty group', body: { groups: [{ kind: 'borrowed', candidates: [] }] } },
  ])('accepts a body with $name', ({ body }) => {
    expect(parseCandidateGroups(body)).toEqual(body);
  });

  it.each([
    { name: 'no groups', body: {} },
    { name: 'a group without kind', body: { groups: [{ candidates: [] }] } },
    { name: 'a candidate without roman', body: bodyOf({ ...bare, roman: undefined }) },
    {
      name: 'a count that is not an integer',
      body: bodyOf({ ...bare, commonTones: { basis: 'x', count: '1' } }),
    },
    {
      name: 'a preference without style',
      body: bodyOf({ ...bare, preference: { ...full.preference, style: undefined } }),
    },
  ])('rejects $name', ({ body }) => {
    expect(() => parseCandidateGroups(body)).toThrow('unexpected body');
  });
});

describe('parseVariations', () => {
  const g = {
    chord: { basis: 'definition', name: 'G' },
    isInput: true,
    members: { basis: 'definition', notes: ['G', 'B', 'D'] },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: 'dominant',
      modes: [],
      fromParallelMinor: false,
    },
    tensionCautions: { basis: 'school', cautions: [] },
    outOfKey: { basis: 'computation', notes: [] },
  };

  it('accepts the body the contract describes', () => {
    expect(parseVariations({ variations: [g] })).toEqual([g]);
  });

  it.each([
    { name: 'no variations', body: {} },
    { name: 'an isInput that is not a boolean', body: { variations: [{ ...g, isInput: 'yes' }] } },
    {
      name: 'members that are not strings',
      body: { variations: [{ ...g, members: { basis: 'definition', notes: [7] } }] },
    },
    { name: 'no outOfKey', body: { variations: [{ ...g, outOfKey: undefined }] } },
  ])('rejects $name', ({ body }) => {
    expect(() => parseVariations(body)).toThrow('unexpected body');
  });
});

describe('parseSlotCandidateGroups', () => {
  const bare = {
    chord: { basis: 'definition', name: 'Dm7' },
    roman: { basis: 'computation', numeral: 'ii7' },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: 'subdominant',
      modes: [],
      fromParallelMinor: false,
    },
  };
  const full = {
    ...bare,
    resolution: { basis: 'school', resolution: 'none' },
    nextResolution: { basis: 'school', resolution: 'none' },
    nextCadence: { basis: 'school', cadence: 'half' },
    roleFit: { basis: 'school', fit: 'misfits', unmet: ['shape', 'function'] },
  };
  const bodyOf = (candidate: object) => ({
    groups: [{ kind: 'diatonic', candidates: [candidate] }],
  });

  it.each([
    { name: 'every label', body: bodyOf(full) },
    { name: 'the labels that are absent left out', body: bodyOf(bare) },
  ])('accepts a body with $name', ({ body }) => {
    expect(parseSlotCandidateGroups(body)).toEqual(body);
  });

  it.each([
    { name: 'a candidate without reading', body: bodyOf({ ...bare, reading: undefined }) },
    {
      name: 'a next cadence without cadence',
      body: bodyOf({ ...bare, nextCadence: { basis: 'school' } }),
    },
    {
      name: 'a role fit without unmet',
      body: bodyOf({ ...bare, roleFit: { basis: 'school', fit: 'fits' } }),
    },
    {
      name: 'a condition that is not a string',
      body: bodyOf({ ...bare, roleFit: { basis: 'school', fit: 'misfits', unmet: [1] } }),
    },
  ])('rejects $name', ({ body }) => {
    expect(() => parseSlotCandidateGroups(body)).toThrow('unexpected body');
  });
});
