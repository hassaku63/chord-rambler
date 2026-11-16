import { describe, expect, it } from 'vitest';
import {
  parseHealth,
  parseProblem,
  parseProgressionNotes,
  parseProgressionReading,
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
  const full = {
    slots: [
      {
        chord: { basis: 'definition', name: 'D7' },
        roman: { basis: 'school', numeral: 'V7/V' },
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
        roleFit: { basis: 'school', fit: 'fits' },
      },
      {},
    ],
  };

  it('accepts the body the contract describes, leaving out what is absent', () => {
    expect(parseProgressionReading(full)).toEqual(full);
    expect(parseProgressionReading({ slots: [{}] }).slots[0]).not.toHaveProperty('chord');
  });

  it('keeps an identifier the contract does not list yet', () => {
    const body = { slots: [{ cadence: { basis: 'association', cadence: 'picardy' } }] };

    expect(parseProgressionReading(body)).toEqual(body);
  });

  it.each([
    { name: 'null', body: null },
    { name: 'no slots', body: {} },
    { name: 'a slot that is not an object', body: { slots: ['C'] } },
    { name: 'a label without basis', body: { slots: [{ chord: { name: 'C' } }] } },
    {
      name: 'a numeral that is not a string',
      body: { slots: [{ roman: { basis: 'x', numeral: 1 } }] },
    },
    {
      name: 'modes that are not strings',
      body: {
        slots: [
          {
            reading: {
              basis: 'x',
              kind: 'borrowed',
              function: 'none',
              modes: [1],
              fromParallelMinor: false,
            },
          },
        ],
      },
    },
    {
      name: 'a caution without chord tone',
      body: { slots: [{ tensionCautions: { basis: 'x', cautions: [{ tension: 'P11' }] } }] },
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

  it.each([
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
