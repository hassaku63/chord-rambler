import type { SlotReading } from '@/shared/api/client';
import { PARALLEL_MINOR, TARGET_NUMERALS, TERMS, wordFor } from './terms';

/** One label of the detailed view: what it is about, its word, and its basis. */
export type LabelRow = { term: string; word: string; basis: string };

function row(term: string, word: string | null, basis: string): LabelRow[] {
  return word === null ? [] : [{ term, word, basis: wordFor(TERMS.basis, basis) ?? basis }];
}

function borrowedFrom(reading: NonNullable<SlotReading['reading']>): string | null {
  if (reading.fromParallelMinor) {
    return PARALLEL_MINOR;
  }
  const modes = reading.modes.map((mode) => wordFor(TERMS.mode, mode) ?? mode);
  return modes.length === 0 ? null : modes.join('、');
}

function chordRows(reading: SlotReading['reading']): LabelRow[] {
  if (reading === undefined) {
    return [];
  }
  const target =
    reading.target === undefined ? null : (TARGET_NUMERALS[reading.target] ?? `${reading.target}`);
  return [
    ...row('種別', wordFor(TERMS.kind, reading.kind), reading.basis),
    ...row('機能', wordFor(TERMS.function, reading.function), reading.basis),
    ...row('借り先', borrowedFrom(reading), reading.basis),
    ...row('向かう先', target, reading.basis),
  ];
}

/** The labels of a slot, in the order the detailed view lists them. */
export function labelRows(slot: SlotReading): LabelRow[] {
  const { resolution, cadence, tensionCautions, roleFit } = slot;
  const cautions = tensionCautions?.cautions
    .map((c) => `${c.tension}（${c.chordTone} の半音上）`)
    .join('、');
  return [
    ...chordRows(slot.reading),
    ...(resolution
      ? row('解決', wordFor(TERMS.resolution, resolution.resolution), resolution.basis)
      : []),
    ...(cadence ? row('終止', wordFor(TERMS.cadence, cadence.cadence), cadence.basis) : []),
    ...(tensionCautions && cautions
      ? row('テンションの注意', cautions, tensionCautions.basis)
      : []),
    ...(roleFit ? row('役割に合うか', wordFor(TERMS.fit, roleFit.fit), roleFit.basis) : []),
  ];
}

/** The short marks on a slot's card: kind, function, and resolution or cadence. */
export function cardMarks(slot: SlotReading): string[] {
  const { reading, resolution, cadence } = slot;
  const marks = [
    reading && wordFor(TERMS.kind, reading.kind),
    reading && wordFor(TERMS.functionMark, reading.function),
    resolution && wordFor(TERMS.resolution, resolution.resolution),
    cadence && wordFor(TERMS.cadence, cadence.cadence),
  ];
  return marks.filter((mark): mark is string => typeof mark === 'string');
}
