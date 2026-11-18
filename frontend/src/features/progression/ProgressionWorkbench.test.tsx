import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Schemas } from '@/test/standInServer';
import { stubServer } from '@/test/standInServer';
import { ProgressionScreen } from './ProgressionScreen';

/** A chord's reading, as in C: G7 resolving as expected, and C closing with an authentic cadence. */
function readingOf(chord: string): Schemas['SlotReading'] {
  return {
    content: 'chord',
    chord: { basis: 'definition', name: chord },
    roman: { basis: 'computation', numeral: { Dm7: 'ii7', G7: 'V7' }[chord] ?? 'I' },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: chord === 'G7' ? 'dominant' : 'tonic',
      modes: [],
      fromParallelMinor: false,
    },
    resolution: { basis: 'school', resolution: chord === 'G7' ? 'expected' : 'none' },
    cadence: { basis: 'school', cadence: chord === 'C' ? 'authentic' : 'none' },
    tensionCautions: { basis: 'school', cautions: [] },
  };
}

function readingsOf({ slots }: Schemas['Progression']): Schemas['ProgressionReading'] {
  return {
    slots: slots.map((slot) =>
      slot.chord === undefined ? { content: slot.content ?? 'unchosen' } : readingOf(slot.chord),
    ),
  };
}

function addChord(name: string) {
  fireEvent.change(screen.getByRole('textbox', { name: '和音名' }), { target: { value: name } });
  fireEvent.click(screen.getByRole('button', { name: '枠を足す' }));
}

const mainLine = () => screen.getByRole('list', { name: '本線' });
const panel = () => screen.getByRole('complementary');
const text = (el: HTMLElement) => el.textContent;

describe('ProgressionScreen as the workbench', () => {
  beforeEach(() => stubServer({ '/progressions/readings': readingsOf }));
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it('shows a node for each slot, and the judgments of the reading on the lines between', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');
    addChord('G7');
    addChord('C');
    fireEvent.click(screen.getByRole('button', { name: 'まだ選んでいない枠を足す' }));

    const nodes = within(mainLine()).getAllByRole('listitem');
    await within(nodes[1] ?? document.body).findByText('V7');
    // G7 resolves on to C, and C closes the phrase; Dm7 → G7 names nothing.
    expect(nodes.map(text)).toEqual([
      '1 Dm7 ii7',
      '→2 G7 V7',
      '解決 正格終止→3 C I',
      '→4 まだ選んでいない',
    ]);
  });

  it('chooses the slot of a node, as the strip does, and names it in the context panel', async () => {
    render(<ProgressionScreen />);
    addChord('C');
    addChord('G7');
    const node = await within(mainLine()).findByRole('button', { name: '2 G7 V7' });
    expect(within(panel()).getByRole('heading', { level: 2 })).toHaveTextContent(
      '枠を選んでいません',
    );

    fireEvent.click(node);

    expect(node).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: /^G7 V7/ })).toHaveAttribute('aria-pressed', 'true');
    expect(within(panel()).getByRole('heading', { level: 2 })).toHaveTextContent(
      'G7（2 番目の枠）',
    );
    expect(within(panel()).getByRole('region', { name: 'G7（V7） の札' })).toBeInTheDocument();
    // In the explore tab the candidates for the slot after the last make way for the chosen slot.
    expect(screen.queryByRole('region', { name: /^他の候補/ })).not.toBeInTheDocument();

    fireEvent.click(node);

    expect(node).toHaveAttribute('aria-pressed', 'false');
    expect(await screen.findByRole('region', { name: /^他の候補/ })).toBeInTheDocument();
  });

  it('marks in the strip where each bar and each phrase starts', () => {
    render(<ProgressionScreen />);
    addChord('C');
    fireEvent.change(screen.getByRole('combobox', { name: '長さ' }), { target: { value: '8' } });
    addChord('G7');
    addChord('C');
    addChord('G7');

    const cards = within(screen.getByRole('list', { name: '進行' })).getAllByRole('listitem');
    const ticks = (card: HTMLElement) =>
      within(card)
        .getAllByText(/^(フレーズ|小節) \d+$/)
        .map(text);
    expect(cards.map(ticks)).toEqual([
      ['フレーズ 1', '小節 2'],
      ['小節 3'],
      ['小節 4'],
      ['フレーズ 2'],
    ]);
  });

  it('tells the tab in words and by its state, and keeps the progression across tabs', async () => {
    render(<ProgressionScreen />);
    addChord('C');
    await within(mainLine()).findByText('I');
    const [explore, plan] = screen.getAllByRole('tab');
    expect(explore).toHaveAttribute('aria-selected', 'true');
    expect(within(panel()).getByText('探索のタブ')).toBeInTheDocument();

    fireEvent.keyDown(explore ?? document.body, { key: 'ArrowRight' });

    expect(plan).toHaveAttribute('aria-selected', 'true');
    expect(plan).toHaveFocus();
    expect(screen.getByRole('tabpanel', { name: '計画' })).toBeInTheDocument();
    expect(within(panel()).getByText('計画のタブ')).toBeInTheDocument();
    expect(
      within(panel()).getByText('枠を選ぶと、その枠を埋める候補が出ます。'),
    ).toBeInTheDocument();
    expect(screen.queryByRole('region', { name: /^他の候補/ })).not.toBeInTheDocument();
    expect(within(mainLine()).getAllByRole('listitem').map(text)).toEqual(['1 C I']);
  });
});
