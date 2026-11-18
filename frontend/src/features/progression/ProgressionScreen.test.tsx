import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fakeAudio } from '@/test/fakeAudioContext';
import { stubServer, world } from '@/test/standInServer';
import { ProgressionScreen } from './ProgressionScreen';

function addChord(name: string) {
  fireEvent.change(screen.getByRole('textbox', { name: '和音名' }), { target: { value: name } });
  fireEvent.click(screen.getByRole('button', { name: '枠を足す' }));
}

describe('ProgressionScreen', () => {
  beforeEach(() => stubServer());
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it('shows the chord name and the Roman numeral on the card of an added chord', async () => {
    render(<ProgressionScreen />);

    addChord('Dm7');

    const strip = screen.getByRole('list', { name: '進行' });
    const card = within(strip).getByRole('listitem');
    expect(await within(card).findByText('ii7')).toBeInTheDocument();
    expect(within(card).getByText('Dm7')).toBeInTheDocument();
    expect(within(card).getByText('ダイアトニック S')).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('shows the detail of the problem and marks the slot it names', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');

    addChord('Hm');

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'slot 1: cannot read the chord name "Hm"',
    );
    const cards = within(screen.getByRole('list', { name: '進行' })).getAllByRole('listitem');
    expect(cards.map((card) => within(card).queryByText('この枠を読めません') !== null)).toEqual([
      false,
      true,
    ]);
  });

  it('shows the labels of the chosen slot with the kind of knowledge each rests on', async () => {
    render(<ProgressionScreen />);
    addChord('Fm');
    await within(screen.getByRole('list', { name: '進行' })).findByText('iv');

    fireEvent.click(screen.getByRole('button', { name: 'Fm iv 借用和音' }));

    const detail = screen.getByRole('region', { name: 'Fm（iv） の札' });
    const text = (el: HTMLElement) => el.textContent;
    expect(within(detail).getAllByRole('term').map(text)).toEqual(['種別', '借り先', '調の外の音']);
    expect(within(detail).getAllByRole('definition').map(text)).toEqual([
      '借用和音 〔流儀〕',
      '同主短調 〔流儀〕',
      'A♭ 〔計算〕',
    ]);
  });

  it('adds empty slots, changes a length and removes a slot', () => {
    render(<ProgressionScreen />);
    fireEvent.click(screen.getByRole('button', { name: '和音なしの枠を足す' }));
    fireEvent.click(screen.getByRole('button', { name: 'まだ選んでいない枠を足す' }));

    const [firstLength] = screen.getAllByRole('combobox', { name: '長さ' });
    fireEvent.change(firstLength ?? document.body, { target: { value: '2' } });
    fireEvent.click(screen.getByRole('button', { name: 'まだ選んでいない の枠を消す' }));

    const [card, ...rest] = within(screen.getByRole('list', { name: '進行' })).getAllByRole(
      'listitem',
    );
    expect(rest).toEqual([]);
    expect(card).toHaveTextContent('和音なし');
    expect(screen.getByRole('combobox', { name: '長さ' })).toHaveValue('2');
  });

  it('sends the key chosen with the progression', () => {
    render(<ProgressionScreen />);

    fireEvent.change(screen.getByRole('combobox', { name: '調' }), { target: { value: 'B♭' } });

    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/progressions/readings',
      expect.objectContaining({ body: JSON.stringify({ key: 'B♭', slots: [] }) }),
    );
  });

  it('plays the progression, marks the slot sounding, and stops when pressed again', async () => {
    const audio = fakeAudio();
    render(<ProgressionScreen createAudioContext={audio.create} />);
    const play = screen.getByRole('button', { name: '進行を鳴らす' });
    expect(play).toBeDisabled();
    addChord('Dm7');
    await waitFor(() => expect(play).toBeEnabled());

    fireEvent.click(play);

    expect(play).toHaveAttribute('aria-pressed', 'true');
    expect(audio.context.oscillators).toHaveLength(1);
    const card = within(screen.getByRole('list', { name: '進行' })).getByRole('listitem');
    expect(await within(card).findByText('鳴っています')).toBeInTheDocument();
    expect(card).toHaveAttribute('aria-current', 'true');
    const node = within(screen.getByRole('list', { name: '本線' })).getByRole('listitem');
    expect(node).toHaveAttribute('aria-current', 'true');
    expect(node).toHaveTextContent('鳴っています');

    fireEvent.click(play);

    expect(play).toHaveAttribute('aria-pressed', 'false');
    expect(within(card).queryByText('鳴っています')).not.toBeInTheDocument();
    expect(node).not.toHaveAttribute('aria-current');
    expect(audio.context.oscillators[0]?.cut).toBe(true);
  });

  it('plays the chosen slot alone, and stops', async () => {
    const audio = fakeAudio();
    render(<ProgressionScreen createAudioContext={audio.create} />);
    addChord('Dm7');
    addChord('Fm');
    const playSlot = screen.getByRole('button', { name: 'この枠だけ鳴らす' });
    const stop = screen.getByRole('button', { name: '止める' });
    await within(screen.getByRole('list', { name: '進行' })).findByText('iv');
    expect(playSlot).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: 'Fm iv 借用和音' }));
    await waitFor(() => expect(playSlot).toBeEnabled());
    fireEvent.click(playSlot);

    expect(playSlot).toHaveAttribute('aria-pressed', 'true');
    expect(audio.context.oscillators.map((o) => o.startedAt)).toEqual([1.05]);
    fireEvent.click(stop);
    expect(playSlot).toHaveAttribute('aria-pressed', 'false');
    expect(stop).toBeDisabled();
  });

  it('shows the candidates by group, only the diatonic group open, with their labels', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');

    const diatonic = await screen.findByRole('list', { name: 'ダイアトニックの候補' });
    const f = within(diatonic).getByRole('listitem');
    await waitFor(() =>
      expect(f).toHaveTextContent('S 根音の動き m3 共通音 2 クラシックでは、より少ない'),
    );
    expect(within(f).getByText('F')).toBeInTheDocument();
    expect(within(f).getByText('IV')).toBeInTheDocument();
    // A closed details element hides all but its summary.
    expect(diatonic).toBeVisible();
    for (const name of ['セカンダリー・ドミナントの候補', '借用和音の候補']) {
      expect(screen.getByRole('list', { name, hidden: true })).not.toBeVisible();
    }
    const text = (el: HTMLElement) => el.textContent;
    expect(within(f).getAllByRole('term').map(text)).toEqual([
      '種別',
      '機能',
      '根音の動き',
      '共通音',
      '好み',
    ]);
    expect(within(f).getAllByRole('definition').map(text).at(-1)).toBe(
      'サブドミナント → サブドミナント、クラシックでは、より少ない 〔慣習〕',
    );
    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/progressions/next-candidates',
      expect.objectContaining({
        body: JSON.stringify({
          progression: { key: 'C', slots: [{ chord: 'Dm7', beats: 4 }] },
          beats: 4,
        }),
      }),
    );
  });

  it('adds a candidate after the last slot', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');

    fireEvent.click(await screen.findByRole('button', { name: 'F を足す' }));

    const cards = within(screen.getByRole('list', { name: '進行' })).getAllByRole('listitem');
    expect(cards).toHaveLength(2);
    expect(await within(cards[1] ?? document.body).findByText('IV')).toBeInTheDocument();
  });

  it('keeps the list, the open groups and the focus while the next candidates come', async () => {
    let release: () => void = () => undefined;
    const held = new Promise<void>((resolve) => (release = resolve));
    stubServer({
      '/progressions/next-candidates': async (request) => {
        if (request.progression.slots.some((slot) => slot.chord === 'F')) {
          await held;
        }
        return world['/progressions/next-candidates'](request);
      },
    });
    render(<ProgressionScreen />);
    addChord('Dm7');
    const add = await screen.findByRole('button', { name: 'F を足す' });
    fireEvent.click(screen.getByText('借用和音（1）'));
    await waitFor(() => expect(screen.getByRole('list', { name: '借用和音の候補' })).toBeVisible());

    add.focus();
    fireEvent.click(add);

    const region = screen.getByRole('region', { name: /^他の候補/ });
    expect(screen.getByRole('heading', { name: /^他の候補/ })).toHaveFocus();
    expect(region).toHaveAttribute('aria-busy', 'true');
    expect(within(region).getByText('新しい候補を待っています。')).toBeInTheDocument();
    expect(within(region).getByRole('list', { name: '借用和音の候補' })).toBeVisible();

    await act(async () => {
      release();
      await held;
    });

    await waitFor(() => expect(region).toHaveAttribute('aria-busy', 'false'));
    expect(within(region).queryByText('新しい候補を待っています。')).not.toBeInTheDocument();
    expect(within(region).getByRole('list', { name: '借用和音の候補' })).toBeVisible();
    expect(within(region).getByRole('list', { name: 'ダイアトニックの候補' })).toBeVisible();
    expect(screen.getByRole('heading', { name: /^他の候補/ })).toHaveFocus();
  });

  it('cuts a typed chord to the beats left in the phrase', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');
    fireEvent.change(screen.getByRole('combobox', { name: '長さ' }), { target: { value: '15' } });

    addChord('F');

    const cards = within(screen.getByRole('list', { name: '進行' })).getAllByRole('listitem');
    expect(cards).toHaveLength(2);
    expect(within(cards[1] ?? document.body).getByRole('combobox', { name: '長さ' })).toHaveValue(
      '1',
    );
    await screen.findByRole('heading', { name: '他の候補（末尾に足す、4 拍）' });
  });

  it('puts a variation in place of the chosen chord, keeping its length and the list', async () => {
    render(<ProgressionScreen />);
    fireEvent.click(await screen.findByRole('button', { name: 'F を足す' }));
    fireEvent.change(screen.getByRole('combobox', { name: '長さ' }), { target: { value: '2' } });
    // In the strip: the branches have an F of their own.
    const strip = screen.getByRole('list', { name: '進行' });
    fireEvent.click(await within(strip).findByRole('button', { name: /^F IV/ }));
    const list = await screen.findByRole('list', { name: 'F の変形' });
    expect(within(list).getByRole('button', { name: 'F' })).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(within(list).getByRole('button', { name: 'Fsus4' }));

    const card = within(screen.getByRole('list', { name: '進行' })).getByRole('listitem');
    expect(await within(card).findByText('Fsus4')).toBeInTheDocument();
    expect(within(card).getByRole('combobox', { name: '長さ' })).toHaveValue('2');
    // The list still comes from F, the chord the slot was added with.
    const again = await screen.findByRole('list', { name: 'F の変形' });
    expect(within(again).getByRole('button', { name: 'Fsus4' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
  });

  it('names the notes outside the key beside a variation that has any, not in its button', async () => {
    render(<ProgressionScreen />);
    addChord('F');
    // In the strip: the branches have an F of their own.
    const strip = screen.getByRole('list', { name: '進行' });
    fireEvent.click(await within(strip).findByRole('button', { name: /^F IV/ }));

    const list = await screen.findByRole('list', { name: 'F の変形' });
    const text = (el: HTMLElement) => el.textContent;
    expect(within(list).getAllByRole('listitem').map(text)).toEqual([
      'F ダイアトニック S',
      'Fsus4 借用和音 調の外の音 B♭',
    ]);
    expect(within(list).getAllByRole('button').map(text)).toEqual(['F', 'Fsus4']);
  });

  it('shows no variations for a chord with no third typed in', async () => {
    render(<ProgressionScreen />);
    addChord('Fsus4');

    // The card's name grows with its reading's labels once the reading comes.
    fireEvent.click(await screen.findByRole('button', { name: /^Fsus4(?! の枠を消す)/ }));

    await waitFor(() =>
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/variations',
        expect.objectContaining({ body: JSON.stringify({ key: 'C', chord: 'Fsus4' }) }),
      ),
    );
    // Let the refusal come in.
    await act(() => new Promise((resolve) => setTimeout(resolve, 0)));
    expect(screen.queryByRole('list', { name: /の変形/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('plays the last chord and then a candidate on trial', async () => {
    const audio = fakeAudio();
    render(<ProgressionScreen createAudioContext={audio.create} />);
    addChord('Dm7');
    const trial = await screen.findByRole('button', { name: 'F を試聴する' });

    fireEvent.click(trial);

    await waitFor(() => expect(trial).toHaveAttribute('aria-pressed', 'true'));
    // Dm7 from the start, F 4 beats later at 90 BPM.
    const starts = audio.context.oscillators.map((o) => o.startedAt);
    expect(starts).toEqual([1.05, expect.closeTo(3.717)]);
    const f = within(screen.getByRole('list', { name: 'ダイアトニックの候補' })).getByRole(
      'listitem',
    );
    expect(within(f).getByText('試聴中')).toBeInTheDocument();
    expect(f).toHaveAttribute('aria-current', 'true');

    fireEvent.click(trial);

    expect(trial).toHaveAttribute('aria-pressed', 'false');
    expect(within(f).queryByText('試聴中')).not.toBeInTheDocument();
    expect(f).not.toHaveAttribute('aria-current');
  });

  it('starts at 90 BPM and takes another tempo', () => {
    render(<ProgressionScreen />);
    const tempo = screen.getByRole('slider', { name: /テンポ/ });
    expect(tempo).toHaveValue('90');

    fireEvent.change(tempo, { target: { value: '120' } });

    expect(tempo).toHaveValue('120');
    expect(tempo).toHaveAttribute('aria-valuetext', '120 BPM');
  });
});
