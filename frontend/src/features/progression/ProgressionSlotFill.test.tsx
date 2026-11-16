import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fakeAudio } from '@/test/fakeAudioContext';
import type { AfterChord, Answers, BeforeChord, Schemas } from '@/test/standInServer';
import { stubServer } from '@/test/standInServer';
import { ProgressionScreen } from './ProgressionScreen';

type Reading = Schemas['ChordReadingLabel'];

function candidate(
  name: string,
  numeral: string,
  kind: Reading['kind'],
  fn: Reading['function'],
): Schemas['SlotCandidate'] {
  return {
    chord: { basis: 'definition', name },
    roman: { basis: 'computation', numeral },
    reading: { basis: 'school', kind, function: fn, modes: [], fromParallelMinor: false },
  };
}

type Fit = Schemas['CandidateRoleFitLabel'];
type Resolution = Schemas['ResolutionLabel'];
const fits: Fit = { basis: 'school', fit: 'fits', unmet: [] };
const misfitsFunction: Fit = { basis: 'school', fit: 'misfits', unmet: ['function'] };
const none: Resolution = { basis: 'school', resolution: 'none' };
const expected: Resolution = { basis: 'school', resolution: 'expected' };

/** For the slot of 準備, between C and G7: those that fit first in each group. */
const betweenCAndG7: { kind: Schemas['CandidateKind']; candidates: AfterChord<BeforeChord>[] }[] = [
  {
    kind: 'diatonic',
    candidates: [
      {
        ...candidate('Dm7', 'ii7', 'diatonic', 'subdominant'),
        resolution: none,
        nextResolution: none,
        roleFit: fits,
      },
      {
        ...candidate('F', 'IV', 'diatonic', 'subdominant'),
        resolution: none,
        nextResolution: none,
        roleFit: fits,
      },
      {
        ...candidate('Em', 'iii', 'diatonic', 'tonic'),
        resolution: none,
        nextResolution: none,
        roleFit: misfitsFunction,
      },
    ],
  },
  {
    kind: 'secondary-dominant',
    candidates: [
      {
        ...candidate('D7', 'V7/V', 'secondary-dominant', 'none'),
        resolution: none,
        nextResolution: expected,
        roleFit: misfitsFunction,
      },
    ],
  },
];

/** For the slot of G7, after the slot not chosen: G7 going on to the C that closes the phrase. */
const beforeTheLastC: { kind: Schemas['CandidateKind']; candidates: BeforeChord[] }[] = [
  {
    kind: 'diatonic',
    candidates: [
      {
        ...candidate('G7', 'V7', 'diatonic', 'dominant'),
        nextResolution: expected,
        nextCadence: { basis: 'school', cadence: 'authentic' },
        roleFit: fits,
      },
    ],
  },
];

/** The candidates above for the slots 1 and 2 of C | (準備) | G7 (高まり) | C. */
const slotCandidates: Answers['/progressions/slot-candidates'] = ({ slot }) => ({
  groups: slot === 1 ? betweenCAndG7 : slot === 2 ? beforeTheLastC : [],
});

/** The stand-in server's answers: readings with no labels, and the candidates above. */
const answers: Partial<Answers> = {
  '/progressions/readings': ({ slots }) => ({ slots: slots.map(() => ({})) }),
  '/progressions/slot-candidates': slotCandidates,
};

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
  beforeEach(() => stubServer(answers));
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
    stubServer({
      ...answers,
      '/progressions/slot-candidates': async (request) => {
        if (++asked > 1) {
          await held;
        }
        return slotCandidates(request);
      },
    });
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
