import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Note, Progression, SlotReading } from '@/shared/api/client';
import { fakeAudio } from '@/test/fakeAudioContext';
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

/** Answers like POST /progressions/notes: one note of middle C for each slot with a chord. */
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
 * Answers like POST /progressions/readings, a 400 naming the first chord it
 * cannot read, and like POST /progressions/notes.
 */
function standInServer(url: string, init?: RequestInit): Promise<Response> {
  const progression = JSON.parse(init?.body as string) as Progression;
  if (url.endsWith('/progressions/notes')) {
    return Promise.resolve(json({ notes: notesOf(progression) }));
  }
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

    fireEvent.click(play);

    expect(play).toHaveAttribute('aria-pressed', 'false');
    expect(within(card).queryByText('鳴っています')).not.toBeInTheDocument();
    expect(audio.context.oscillators[0]?.cut).toBe(true);
  });

  it('plays the chosen slot alone, and stops', async () => {
    const audio = fakeAudio();
    render(<ProgressionScreen createAudioContext={audio.create} />);
    addChord('Dm7');
    addChord('Fm');
    const playSlot = screen.getByRole('button', { name: 'この枠だけ鳴らす' });
    const stop = screen.getByRole('button', { name: '止める' });
    await screen.findByText('iv');
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

  it('starts at 90 BPM and takes another tempo', () => {
    render(<ProgressionScreen />);
    const tempo = screen.getByRole('slider', { name: /テンポ/ });
    expect(tempo).toHaveValue('90');

    fireEvent.change(tempo, { target: { value: '120' } });

    expect(tempo).toHaveValue('120');
    expect(tempo).toHaveAttribute('aria-valuetext', '120 BPM');
  });
});
