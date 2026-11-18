import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Progression } from '@/shared/api/client';
import {
  getNextCandidates,
  getProgressionNotes,
  getSlotCandidates,
  getVariations,
  readProgression,
} from '@/shared/api/client';
import { stubServer } from './standInServer';

const progression: Progression = {
  key: 'C',
  slots: [
    { chord: 'Dm7', beats: 4 },
    { content: 'unchosen', beats: 4, role: 'prepare' },
    { chord: 'F', beats: 4 },
  ],
};

// The screen's hooks take a body the client cannot check as no answer, so the
// default answers are checked here, through the client that checks them.
describe('the stand-in server', () => {
  beforeEach(() => stubServer());
  afterEach(() => vi.unstubAllGlobals());

  it('answers every route with a body the client takes', async () => {
    const answers = await Promise.all([
      readProgression(progression),
      getProgressionNotes(progression),
      getNextCandidates(progression, 4),
      getSlotCandidates(progression, 0),
      getSlotCandidates(progression, 1),
      getVariations('C', 'F'),
    ]);

    expect(answers.map((a) => a.ok)).toEqual([true, true, true, true, true, true]);
  });

  it('gives the candidates for a slot between chords the labels of both', async () => {
    const answer = await getSlotCandidates(progression, 1);

    const f = answer.ok ? answer.groups[0]?.candidates[0] : undefined;
    expect(f?.resolution).toEqual({ basis: 'school', resolution: 'none' });
    expect(f?.nextResolution).toEqual({ basis: 'school', resolution: 'none' });
    expect(f?.roleFit).toEqual({ basis: 'school', fit: 'undecided', unmet: [] });
  });

  it('refuses a chord it cannot read, naming its slot', async () => {
    const answer = await readProgression({ key: 'C', slots: [{ chord: 'Hm', beats: 4 }] });

    expect(answer).toEqual({
      ok: false,
      problem: {
        title: 'Bad Request',
        status: 400,
        detail: 'slot 0: cannot read the chord name "Hm"',
      },
    });
  });

  it('refuses a slot that crosses into the next phrase, naming its slot', async () => {
    const answer = await readProgression({
      key: 'C',
      slots: [
        { chord: 'F', beats: 12 },
        { chord: 'Dm7', beats: 8 },
      ],
    });

    expect(answer).toEqual({
      ok: false,
      problem: {
        title: 'Bad Request',
        status: 400,
        detail: 'slot 1: a slot must not cross the boundary of a phrase',
        slot: 1,
      },
    });
  });
});
