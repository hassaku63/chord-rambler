import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Progression, SlotReading } from '@/shared/api/client';
import { ProgressionScreen } from './ProgressionScreen';

// The readings the stand-in server knows, as the contract writes them.
const known: Record<string, SlotReading> = {
  Dm7: {
    chord: { basis: 'definition', name: 'Dm7' },
    roman: { basis: 'computation', numeral: 'ii7' },
    reading: {
      basis: 'school',
      kind: 'diatonic',
      function: 'subdominant',
      modes: [],
      fromParallelMinor: false,
    },
    resolution: { basis: 'school', resolution: 'none' },
    cadence: { basis: 'school', cadence: 'none' },
    tensionCautions: { basis: 'school', cautions: [] },
  },
  Fm: {
    chord: { basis: 'definition', name: 'Fm' },
    roman: { basis: 'computation', numeral: 'iv' },
    reading: {
      basis: 'school',
      kind: 'borrowed',
      function: 'none',
      modes: ['dorian', 'aeolian'],
      fromParallelMinor: true,
    },
  },
};

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status });
}

/** Answers like POST /progressions/readings: a 400 naming the first chord it cannot read. */
function standInServer(_url: string, init?: RequestInit): Promise<Response> {
  const progression = JSON.parse(init?.body as string) as Progression;
  const slots: SlotReading[] = [];
  for (const [i, slot] of progression.slots.entries()) {
    const reading = slot.chord === undefined ? {} : known[slot.chord];
    if (reading === undefined) {
      const detail = `slot ${i}: cannot read the chord name "${slot.chord ?? ''}"`;
      return Promise.resolve(json({ title: 'Bad Request', status: 400, detail }, 400));
    }
    slots.push(reading);
  }
  return Promise.resolve(json({ slots }));
}

function addChord(name: string) {
  fireEvent.change(screen.getByRole('textbox', { name: '和音名' }), { target: { value: name } });
  fireEvent.click(screen.getByRole('button', { name: '枠を足す' }));
}

describe('ProgressionScreen', () => {
  beforeEach(() => vi.stubGlobal('fetch', vi.fn(standInServer)));
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
    await screen.findByText('iv');

    fireEvent.click(screen.getByRole('button', { name: 'Fm iv 借用和音' }));

    const detail = screen.getByRole('region', { name: 'Fm（iv） の札' });
    const text = (el: HTMLElement) => el.textContent;
    expect(within(detail).getAllByRole('term').map(text)).toEqual(['種別', '借り先']);
    expect(within(detail).getAllByRole('definition').map(text)).toEqual([
      '借用和音 〔流儀〕',
      '同主短調 〔流儀〕',
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

    expect(fetch).toHaveBeenLastCalledWith(
      '/api/v1/progressions/readings',
      expect.objectContaining({ body: JSON.stringify({ key: 'B♭', slots: [] }) }),
    );
  });
});
