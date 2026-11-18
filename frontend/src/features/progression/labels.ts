import type { Candidate, SlotCandidate, SlotReading } from '@/shared/api/client';
import { PARALLEL_MINOR, TARGET_NUMERALS, TERMS, wordFor } from './terms';

/** One label of the detailed view: what it is about, its word, and its basis. */
export type LabelRow = { term: string; word: string; basis: string };

function row(term: string, word: string | null, basis: string): LabelRow[] {
  return word === null ? [] : [{ term, word, basis: wordFor(TERMS.basis, basis) ?? basis }];
}

/** The main reading of a chord, as a slot with a chord, a candidate and a variation carry it. */
type ChordReading = Candidate['reading'];

function borrowedFrom(reading: ChordReading): string | null {
  if (reading.fromParallelMinor) {
    return PARALLEL_MINOR;
  }
  const modes = reading.modes.map((mode) => wordFor(TERMS.mode, mode) ?? mode);
  return modes.length === 0 ? null : modes.join('、');
}

function chordRows(reading: ChordReading): LabelRow[] {
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
  const { roleFit } = slot;
  const fit = roleFit ? row('役割に合うか', wordFor(TERMS.fit, roleFit.fit), roleFit.basis) : [];
  if (slot.content !== 'chord') {
    return fit;
  }
  const { reading, resolution, cadence, tensionCautions, outOfKey } = slot;
  const cautions = tensionCautions.cautions
    .map((c) => `${c.tension}（${c.chordTone} の半音上）`)
    .join('、');
  const outside = outOfKey?.notes.join('、');
  return [
    ...chordRows(reading),
    ...row('解決', wordFor(TERMS.resolution, resolution.resolution), resolution.basis),
    ...row('終止', wordFor(TERMS.cadence, cadence.cadence), cadence.basis),
    ...(cautions ? row('テンションの注意', cautions, tensionCautions.basis) : []),
    ...(outOfKey && outside ? row('調の外の音', outside, outOfKey.basis) : []),
    ...fit,
  ];
}

/** The Roman numeral of a slot with a chord, when the chord has one. */
export function numeralOf(slot: SlotReading | undefined): string | undefined {
  return slot?.content === 'chord' ? slot.roman?.numeral : undefined;
}

/** The mark of a role fit, with what it is about (役割に合うか 合わない). */
function fitMark(roleFit: SlotReading['roleFit']): string | undefined {
  return roleFit && `役割に合うか ${wordFor(TERMS.fit, roleFit.fit) ?? roleFit.fit}`;
}

/** The short marks of a chord's reading: kind and function. */
export function readingMarks(reading: ChordReading): string[] {
  const marks = [wordFor(TERMS.kind, reading.kind), wordFor(TERMS.functionMark, reading.function)];
  return marks.filter((mark): mark is string => typeof mark === 'string');
}

/** The short marks on a slot's card: kind, function, resolution or cadence, and role fit. */
export function cardMarks(slot: SlotReading): string[] {
  const chord =
    slot.content === 'chord'
      ? [
          ...readingMarks(slot.reading),
          wordFor(TERMS.resolution, slot.resolution.resolution),
          wordFor(TERMS.cadence, slot.cadence.cadence),
        ]
      : [];
  const marks = [...chord, fitMark(slot.roleFit)];
  return marks.filter((mark): mark is string => typeof mark === 'string');
}

/**
 * The grade of a move between functions with the style it is limited to
 * (「クラシックでは、より少ない」, #114), or null when it names none.
 */
export function preferenceWord(preference: NonNullable<Candidate['preference']>): string | null {
  const grade = wordFor(TERMS.preference, preference.preference);
  return grade === null ? null : `${wordFor(TERMS.style, preference.style) ?? ''}では、${grade}`;
}

/** The move between functions a preference is about (トニック → サブドミナント). */
function moveOf(preference: NonNullable<Candidate['preference']>): string {
  const from = wordFor(TERMS.function, preference.from) ?? preference.from;
  const to = wordFor(TERMS.function, preference.to) ?? preference.to;
  return `${from} → ${to}`;
}

/**
 * The short marks of a candidate: function, root motion, common tones,
 * resolution, cadence and preference. Its kind is the heading of its group.
 */
export function candidateMarks(candidate: Candidate): string[] {
  const { reading, rootMotion, commonTones, resolution, cadence, preference } = candidate;
  const marks = [
    wordFor(TERMS.functionMark, reading.function),
    rootMotion && `根音の動き ${rootMotion.interval}`,
    commonTones && `共通音 ${commonTones.count}`,
    resolution && wordFor(TERMS.resolution, resolution.resolution),
    cadence && wordFor(TERMS.cadence, cadence.cadence),
    preference && preferenceWord(preference),
  ];
  return marks.filter((mark): mark is string => typeof mark === 'string');
}

/**
 * The labels of a candidate, in the order its detailed view lists them; the
 * resolution of the chord before it under the term given.
 */
function rowsOf(candidate: Candidate, resolutionTerm: string): LabelRow[] {
  const { rootMotion, commonTones, resolution, cadence, preference } = candidate;
  const grade = preference && preferenceWord(preference);
  return [
    ...chordRows(candidate.reading),
    ...(rootMotion ? row('根音の動き', rootMotion.interval, rootMotion.basis) : []),
    ...(commonTones ? row('共通音', `${commonTones.count}`, commonTones.basis) : []),
    ...(resolution
      ? row(resolutionTerm, wordFor(TERMS.resolution, resolution.resolution), resolution.basis)
      : []),
    ...(cadence ? row('終止', wordFor(TERMS.cadence, cadence.cadence), cadence.basis) : []),
    ...(preference && grade
      ? row('好み', `${moveOf(preference)}、${grade}`, preference.basis)
      : []),
  ];
}

/** The labels of a candidate for the slot after the last. */
export function candidateRows(candidate: Candidate): LabelRow[] {
  return rowsOf(candidate, '末尾の和音の解決');
}

/** A word with what it is about (次の枠へ 解決), or null when there is no word. */
function about(what: string, word: string | null): string | null {
  return word === null ? null : `${what} ${word}`;
}

/**
 * The short marks of a candidate for a slot to fill: those of a candidate,
 * then how it goes on to the next slot, the cadence of the next slot, and its
 * role fit.
 */
export function slotCandidateMarks(candidate: SlotCandidate): string[] {
  const { nextResolution, nextCadence, roleFit } = candidate;
  const marks = [
    ...candidateMarks(candidate),
    nextResolution && about('次の枠へ', wordFor(TERMS.resolution, nextResolution.resolution)),
    nextCadence && about('次の枠', wordFor(TERMS.cadence, nextCadence.cadence)),
    fitMark(roleFit),
  ];
  return marks.filter((mark): mark is string => typeof mark === 'string');
}

/** The labels of a candidate for a slot to fill, the conditions it does not meet last. */
export function slotCandidateRows(candidate: SlotCandidate): LabelRow[] {
  const { nextResolution, nextCadence, roleFit } = candidate;
  const unmet = roleFit?.unmet.map((id) => wordFor(TERMS.unmet, id) ?? id).join('、');
  return [
    ...rowsOf(candidate, '前の枠の和音の解決'),
    ...(nextResolution
      ? row(
          '次の枠への解決',
          wordFor(TERMS.resolution, nextResolution.resolution),
          nextResolution.basis,
        )
      : []),
    ...(nextCadence
      ? row('次の枠の終止', wordFor(TERMS.cadence, nextCadence.cadence), nextCadence.basis)
      : []),
    ...(roleFit ? row('役割に合うか', wordFor(TERMS.fit, roleFit.fit), roleFit.basis) : []),
    ...(roleFit && unmet ? row('合わない理由', unmet, roleFit.basis) : []),
  ];
}
