import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fakeAudio } from '@/test/fakeAudioContext';
import { stubServer } from '@/test/standInServer';
import { ProgressionScreen } from './ProgressionScreen';

function addChord(name: string) {
  fireEvent.change(screen.getByRole('textbox', { name: '和音名' }), { target: { value: name } });
  fireEvent.click(screen.getByRole('button', { name: '枠を足す' }));
}

const branches = () => screen.getByRole('toolbar', { name: '枝' });
const footprint = () => screen.getByRole('toolbar', { name: '探索の足跡' });
const strip = () => screen.getByRole('list', { name: '進行' });
const panel = () => screen.getByRole('complementary');
const text = (el: HTMLElement) => el.textContent;
/** The chords in the strip, slot by slot. */
const chords = () =>
  within(strip())
    .getAllByRole('listitem')
    .map((card) => within(card).getAllByText(/^(F|Fm|Dm7)$/)[0]?.textContent);

/** The branch of the chord in the group, once the answer to the latest progression is in. */
async function branch(chord: string, group = branches) {
  await waitFor(() => expect(group()).toHaveAttribute('aria-busy', 'false'));
  return within(group()).getByRole('button', { name: new RegExp(`^${chord} `) });
}

describe('ProgressionScreen with branches in the explore tab', () => {
  beforeEach(() => stubServer());
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it('grows the first candidates as branches from a start point, with chord and numeral', async () => {
    render(<ProgressionScreen />);

    expect(await screen.findByText('始まり')).toBeInTheDocument();
    await branch('F');
    expect(within(branches()).getAllByRole('button').map(text)).toEqual([
      'F IV',
      'D7 V7/V',
      'Fm iv',
    ]);
    // One Tab stop: the first branch until one is picked.
    expect(
      within(branches())
        .getAllByRole('button')
        .map((b) => b.tabIndex),
    ).toEqual([0, -1, -1]);
  });

  it('plays the last slot and a branch when the branch is pressed, and shows its labels', async () => {
    const audio = fakeAudio();
    render(<ProgressionScreen createAudioContext={audio.create} />);
    addChord('Dm7');
    const f = await branch('F');

    fireEvent.click(f);

    expect(f).toHaveAttribute('aria-pressed', 'true');
    await waitFor(() => expect(f).toHaveAttribute('aria-current', 'true'));
    // Dm7 from the start, F 4 beats later at 90 BPM.
    expect(audio.context.oscillators.map((o) => o.startedAt)).toEqual([
      1.05,
      expect.closeTo(3.717),
    ]);
    expect(within(branches()).getByText('試聴中')).toBeInTheDocument();
    expect(within(panel()).getByRole('heading', { level: 2 })).toHaveTextContent(
      'F（2 番目の枠の枝）',
    );
    const labels = within(panel()).getByRole('region', { name: 'F（IV） の札' });
    expect(labels).toHaveTextContent('S 根音の動き m3 共通音 2 クラシックでは、より少ない');
    expect(within(labels).getAllByRole('term').map(text)).toEqual([
      '種別',
      '機能',
      '根音の動き',
      '共通音',
      '好み',
    ]);
    // The other candidates stay to be added from.
    expect(within(panel()).getByRole('region', { name: /^他の候補/ })).toBeInTheDocument();
  });

  it('moves between the branches with the arrows, playing each it moves to', async () => {
    const audio = fakeAudio();
    render(<ProgressionScreen createAudioContext={audio.create} />);
    const f = await branch('F');
    act(() => f.focus());
    await waitFor(() => expect(f).toHaveAttribute('aria-current', 'true'));

    fireEvent.keyDown(f, { key: 'ArrowDown' });

    const d7 = within(branches()).getByRole('button', { name: 'D7 V7/V' });
    expect(d7).toHaveFocus();
    expect(d7).toHaveAttribute('aria-pressed', 'true');
    expect(d7).toHaveAttribute('tabindex', '0');
    expect(f).toHaveAttribute('aria-pressed', 'false');
    await waitFor(() => expect(d7).toHaveAttribute('aria-current', 'true'));
    expect(audio.context.oscillators).toHaveLength(2);

    fireEvent.keyDown(d7, { key: 'ArrowUp' });
    fireEvent.keyDown(f, { key: 'ArrowUp' });

    expect(within(branches()).getByRole('button', { name: 'Fm iv' })).toHaveFocus();
  });

  it('puts a branch in with Enter, grows new branches, and focuses the first of them', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');
    const fm = await branch('Fm');
    act(() => fm.focus());

    fireEvent.keyDown(fm, { key: 'Enter' });

    expect(chords()).toEqual(['Dm7', 'Fm']);
    const first = await branch('F');
    await waitFor(() => expect(first).toHaveFocus());
    expect(first).toHaveAttribute('aria-pressed', 'true');
    expect(
      within(screen.getByRole('list', { name: '本線' })).getAllByRole('listitem'),
    ).toHaveLength(2);
  });

  it('puts a branch in by a double click, or from the context panel', async () => {
    render(<ProgressionScreen />);

    fireEvent.doubleClick(await branch('F'));
    expect(chords()).toEqual(['F']);
    await waitFor(() => expect(within(branches()).getAllByRole('button')[0]).toHaveFocus());

    fireEvent.click(await branch('Fm'));
    fireEvent.click(within(panel()).getByRole('button', { name: 'Fm を確定する' }));

    expect(chords()).toEqual(['F', 'Fm']);
    await waitFor(() => expect(within(branches()).getAllByRole('button')[0]).toHaveFocus());
  });

  it('keeps the other branches of the step as the footprint, to put in place of the slot', async () => {
    const audio = fakeAudio();
    render(<ProgressionScreen createAudioContext={audio.create} />);
    addChord('Dm7');
    fireEvent.keyDown(await branch('F'), { key: 'Enter' });
    // The focus goes to the first new branch, which plays it, once they come.
    await waitFor(() => expect(within(branches()).getAllByRole('button')[0]).toHaveFocus());

    expect(within(footprint()).getAllByRole('button').map(text)).toEqual(['D7 V7/V', 'Fm iv']);
    const fm = within(footprint()).getByRole('button', { name: 'Fm iv' });
    fireEvent.click(fm);
    // Played in the slot it is beside: Dm7, then Fm (after the new first branch, Dm7 and F).
    await waitFor(() => expect(fm).toHaveAttribute('aria-current', 'true'));
    expect(audio.context.oscillators.map((o) => o.startedAt).slice(2)).toEqual([
      1.05,
      expect.closeTo(3.717),
    ]);
    expect(within(panel()).getByRole('heading', { level: 2 })).toHaveTextContent(
      'Fm（2 番目の枠の枝）',
    );
    expect(within(panel()).getByRole('region', { name: 'Fm（iv） の札' })).toBeInTheDocument();

    fireEvent.keyDown(fm, { key: 'Enter' });

    expect(chords()).toEqual(['Dm7', 'Fm']);
    const [, card] = within(strip()).getAllByRole('listitem');
    expect(within(card ?? document.body).getByRole('combobox', { name: '長さ' })).toHaveValue('4');
    // The footprint stays one step: the chord put in before is among it now.
    expect(within(footprint()).getAllByRole('button').map(text)).toEqual(['F IV', 'D7 V7/V']);
    await waitFor(() => expect(within(branches()).getAllByRole('button')[0]).toHaveFocus());
  });

  it('clears the footprint when the progression changes another way, and hides it in the plan tab', async () => {
    render(<ProgressionScreen />);
    fireEvent.keyDown(await branch('F'), { key: 'Enter' });
    expect(await screen.findByRole('toolbar', { name: '探索の足跡' })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('tab', { name: '計画' }));
    expect(screen.queryByRole('toolbar', { name: '探索の足跡' })).not.toBeInTheDocument();
    expect(screen.queryByRole('toolbar', { name: '枝' })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('tab', { name: '探索' }));
    expect(footprint()).toBeInTheDocument();

    addChord('Dm7');

    expect(screen.queryByRole('toolbar', { name: '探索の足跡' })).not.toBeInTheDocument();
  });

  it('lets a branch go when a node is chosen', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');
    const f = await branch('F');
    fireEvent.click(f);

    fireEvent.click(within(screen.getByRole('list', { name: '本線' })).getByRole('button'));

    expect(f).toHaveAttribute('aria-pressed', 'false');
    expect(within(panel()).getByRole('heading', { level: 2 })).toHaveTextContent(
      'Dm7（1 番目の枠）',
    );
  });
});
