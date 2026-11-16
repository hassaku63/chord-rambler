import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Note, Progression, SlotCandidate } from '@/shared/api/client';
import { fakeAudio } from '@/test/fakeAudioContext';
import { ProgressionScreen } from './ProgressionScreen';

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status });
}

function candidate(name: string, numeral: string, kind: string, fn: string): SlotCandidate {
  return {
    chord: { basis: 'definition', name },
    roman: { basis: 'computation', numeral },
    reading: { basis: 'school', kind, function: fn, modes: [], fromParallelMinor: false },
  };
}

const fits = { basis: 'school', fit: 'fits', unmet: [] };
const misfitsFunction = { basis: 'school', fit: 'misfits', unmet: ['function'] };

/**
 * The candidates the stand-in server gives in C for C | (準備) | G7 (高まり) | C: for
 * the slot of 準備, those that fit first in each group; for the slot of G7,
 * G7 going on to the C that closes the phrase.
 */
const candidatesFor: Record<number, { kind: string; candidates: SlotCandidate[] }[]> = {
  1: [
    {
      kind: 'diatonic',
      candidates: [
        { ...candidate('Dm7', 'ii7', 'diatonic', 'subdominant'), roleFit: fits },
        { ...candidate('F', 'IV', 'diatonic', 'subdominant'), roleFit: fits },
        { ...candidate('Em', 'iii', 'diatonic', 'tonic'), roleFit: misfitsFunction },
      ],
    },
    {
      kind: 'secondary-dominant',
      candidates: [
        {
          ...candidate('D7', 'V7/V', 'secondary-dominant', 'none'),
          nextResolution: { basis: 'school', resolution: 'expected' },
          roleFit: misfitsFunction,
        },
      ],
    },
  ],
  2: [
    {
      kind: 'diatonic',
      candidates: [
        {
          ...candidate('G7', 'V7', 'diatonic', 'dominant'),
          nextResolution: { basis: 'school', resolution: 'expected' },
          nextCadence: { basis: 'school', cadence: 'authentic' },
          roleFit: fits,
        },
      ],
    },
  ],
};

/** One note of middle C for each slot with a chord, as POST /progressions/notes answers. */
function notesOf(progression: Progression): Note[] {
  let start = 0;
  const notes: Note[] = [];
  for (const slot of progression.slots) {
    if (slot.chord !== undefined) {
      notes.push({ pitch: 60, start, beats: slot.beats, velocity: 64 });
    }
    start += slot.beats;
  }
  return notes;
}

/**
 * Answers the readings with no labels, the notes, and the candidates for a
 * slot; refuses the next candidates and the variations, which then show none.
 */
function standInServer(url: string, init?: RequestInit): Promise<Response> {
  const body = JSON.parse(init?.body as string) as Record<string, unknown>;
  if (url.endsWith('/progressions/slot-candidates')) {
    return Promise.resolve(json({ groups: candidatesFor[body.slot as number] ?? [] }));
  }
  const progression = body as Progression;
  if (url.endsWith('/progressions/readings')) {
    return Promise.resolve(json({ slots: progression.slots.map(() => ({})) }));
  }
  if (url.endsWith('/progressions/notes')) {
    return Promise.resolve(json({ notes: notesOf(progression) }));
  }
  return Promise.resolve(json({ title: 'Bad Request', status: 400 }, 400));
}

function cards() {
  return within(screen.getByRole('list', { name: '進行' })).getAllByRole('listitem');
}

function card(index: number): HTMLElement {
  return cards()[index] ?? document.body;
}

/** Builds C | (準備) | G7 (高まり) | C, the second slot not chosen yet. */
function plan() {
  const add = (name: string) => {
    fireEvent.change(screen.getByRole('textbox', { name: '和音名' }), { target: { value: name } });
    fireEvent.click(screen.getByRole('button', { name: '枠を足す' }));
  };
  add('C');
  fireEvent.click(screen.getByRole('button', { name: 'まだ選んでいない枠を足す' }));
  add('G7');
  add('C');
  fireEvent.change(within(card(1)).getByRole('combobox', { name: '役割' }), {
    target: { value: 'prepare' },
  });
  fireEvent.change(within(card(2)).getByRole('combobox', { name: '役割' }), {
    target: { value: 'rise' },
  });
}

const addName = (item: HTMLElement) =>
  within(item)
    .getByRole('button', { name: /をこの枠に入れる$/ })
    .getAttribute('aria-label');

describe('ProgressionScreen slot fill', () => {
  beforeEach(() => vi.stubGlobal('fetch', vi.fn(standInServer)));
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it('shows the candidates for a slot of 準備, those that fit it first', async () => {
    render(<ProgressionScreen />);
    plan();

    fireEvent.click(within(card(1)).getByRole('button', { name: 'まだ選んでいない' }));

    const region = await screen.findByRole('region', { name: '選んだ枠を埋める候補（4 拍）' });
    const diatonic = within(region).getByRole('list', { name: 'この枠のダイアトニックの候補' });
    const items = within(diatonic).getAllByRole('listitem');
    expect(items.map(addName)).toEqual([
      'Dm7 をこの枠に入れる',
      'F をこの枠に入れる',
      'Em をこの枠に入れる',
    ]);
    expect(items[0]).toHaveTextContent('S 役割に合うか 合う');
    expect(items[2]).toHaveTextContent('T 役割に合うか 合わない');
    const text = (el: HTMLElement) => el.textContent;
    const em = items[2] ?? document.body;
    expect(within(em).getAllByRole('term').map(text).slice(-2)).toEqual([
      '役割に合うか',
      '合わない理由',
    ]);
    expect(within(em).getAllByRole('definition').map(text).at(-1)).toBe('機能が違う 〔流儀〕');
    const d7 = within(
      within(region).getByRole('list', {
        name: 'この枠のセカンダリー・ドミナントの候補',
        hidden: true,
      }),
    ).getByRole('listitem', { hidden: true });
    expect(d7).toHaveTextContent('次の枠へ 解決 役割に合うか 合わない');
    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/progressions/slot-candidates',
      expect.objectContaining({
        body: JSON.stringify({
          progression: {
            key: 'C',
            slots: [
              { chord: 'C', beats: 4 },
              { content: 'unchosen', beats: 4, role: 'prepare' },
              { chord: 'G7', beats: 4, role: 'rise' },
              { chord: 'C', beats: 4 },
            ],
          },
          slot: 1,
        }),
      }),
    );
  });

  it('shows the candidates for a slot with a chord, with the next slot resolved and closed', async () => {
    render(<ProgressionScreen />);
    plan();

    fireEvent.click(within(card(2)).getByRole('button', { name: 'G7' }));

    const list = await screen.findByRole('list', { name: 'この枠のダイアトニックの候補' });
    expect(within(list).getByRole('listitem')).toHaveTextContent(
      'D 次の枠へ 解決 次の枠 正格終止 役割に合うか 合う',
    );
  });

  it('shows no candidates for a slot with no chord', async () => {
    render(<ProgressionScreen />);
    fireEvent.click(screen.getByRole('button', { name: '和音なしの枠を足す' }));

    fireEvent.click(within(card(0)).getByRole('button', { name: '和音なし' }));

    await screen.findByRole('region', { name: '和音なし の札' });
    expect(screen.queryByRole('region', { name: /埋める候補/ })).not.toBeInTheDocument();
    expect(fetch).not.toHaveBeenCalledWith(
      '/api/v1/progressions/slot-candidates',
      expect.anything(),
    );
  });

  it('puts a candidate in the slot, keeping its length and its role', async () => {
    render(<ProgressionScreen />);
    plan();
    fireEvent.change(within(card(1)).getByRole('combobox', { name: '長さ' }), {
      target: { value: '2' },
    });
    fireEvent.click(within(card(1)).getByRole('button', { name: 'まだ選んでいない' }));

    fireEvent.click(await screen.findByRole('button', { name: 'Dm7 をこの枠に入れる' }));

    expect(within(card(1)).getByRole('button', { name: 'Dm7' })).toBeInTheDocument();
    expect(within(card(1)).getByRole('combobox', { name: '長さ' })).toHaveValue('2');
    expect(within(card(1)).getByRole('combobox', { name: '役割' })).toHaveValue('prepare');
    expect(cards()).toHaveLength(4);
  });

  it('does not show the candidates of the slot before for a slot chosen again', async () => {
    let release: () => void = () => undefined;
    const held = new Promise<void>((resolve) => (release = resolve));
    let asked = 0;
    vi.stubGlobal(
      'fetch',
      vi.fn(async (url: string, init?: RequestInit) => {
        if (url.endsWith('/progressions/slot-candidates') && ++asked > 1) {
          await held;
        }
        return standInServer(url, init);
      }),
    );
    render(<ProgressionScreen />);
    plan();
    const unchosen = within(card(1)).getByRole('button', { name: 'まだ選んでいない' });
    fireEvent.click(unchosen);
    await screen.findByRole('button', { name: 'Dm7 をこの枠に入れる' });

    fireEvent.click(unchosen);
    fireEvent.click(within(card(0)).getByRole('button', { name: 'C の枠を消す' }));
    fireEvent.click(within(card(1)).getByRole('button', { name: 'G7' }));

    await screen.findByRole('region', { name: 'G7 の札' });
    expect(screen.queryByRole('region', { name: /埋める候補/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /をこの枠に入れる$/ })).not.toBeInTheDocument();
    await act(async () => {
      release();
      await held;
    });
    await screen.findByRole('region', { name: '選んだ枠を埋める候補（4 拍）' });
  });

  it('plays the slot before, a candidate and the slot after on trial', async () => {
    const audio = fakeAudio();
    render(<ProgressionScreen createAudioContext={audio.create} />);
    plan();
    fireEvent.click(within(card(1)).getByRole('button', { name: 'まだ選んでいない' }));
    const trial = await screen.findByRole('button', { name: 'F をこの枠で試聴する' });

    fireEvent.click(trial);

    await waitFor(() => expect(trial).toHaveAttribute('aria-pressed', 'true'));
    // C from the start, F 4 beats later and G7 8 beats later, at 90 BPM.
    expect(audio.context.oscillators.map((o) => o.startedAt)).toEqual([
      1.05,
      expect.closeTo(3.717),
      expect.closeTo(6.383),
    ]);
    const f = within(screen.getByRole('list', { name: 'この枠のダイアトニックの候補' }))
      .getAllByRole('listitem')
      .at(1);
    expect(within(f ?? document.body).getByText('試聴中')).toBeInTheDocument();

    fireEvent.click(trial);

    expect(trial).toHaveAttribute('aria-pressed', 'false');
  });
});
