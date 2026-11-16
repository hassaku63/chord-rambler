import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Progression } from '@/shared/api/client';
import { ProgressionScreen } from './ProgressionScreen';

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status });
}

const diatonic = {
  basis: 'school',
  kind: 'diatonic',
  function: 'subdominant',
  modes: [],
  fromParallelMinor: false,
};

/** A stand-in server: a slot reads as its chord's name, and F is the one candidate. */
function standInServer(url: string, init?: RequestInit): Promise<Response> {
  const body = JSON.parse(init?.body as string) as { progression?: Progression } & Progression;
  if (url.endsWith('/progressions/next-candidates')) {
    const f = {
      chord: { basis: 'definition', name: 'F' },
      roman: { basis: 'computation', numeral: 'IV' },
      reading: diatonic,
    };
    return Promise.resolve(json({ groups: [{ kind: 'diatonic', candidates: [f] }] }));
  }
  if (url.endsWith('/progressions/notes')) {
    return Promise.resolve(json({ notes: [] }));
  }
  const slots = body.slots.map((slot) =>
    slot.chord === undefined ? {} : { chord: { basis: 'definition', name: slot.chord } },
  );
  return Promise.resolve(json({ slots }));
}

function addChord(name: string) {
  fireEvent.change(screen.getByRole('textbox', { name: '和音名' }), { target: { value: name } });
  fireEvent.click(screen.getByRole('button', { name: '枠を足す' }));
}

function lengths(select: HTMLElement | undefined): number[] {
  return within(select ?? document.body)
    .getAllByRole('option')
    .map((o) => Number(o.getAttribute('value')));
}

describe('ProgressionScreen and the phrase', () => {
  beforeEach(() => vi.stubGlobal('fetch', vi.fn(standInServer)));
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it('gives an unchosen slot the beats left in the phrase', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');
    fireEvent.change(screen.getByRole('combobox', { name: '長さ' }), { target: { value: '15' } });

    fireEvent.click(screen.getByRole('button', { name: 'まだ選んでいない枠を足す' }));

    const cards = within(screen.getByRole('list', { name: '進行' })).getAllByRole('listitem');
    expect(within(cards[1] ?? document.body).getByRole('combobox', { name: '長さ' })).toHaveValue(
      '1',
    );
    await screen.findByRole('heading', { name: '末尾に足す候補（4 拍）' });
  });

  it('offers a slot only the lengths that stay in its phrase', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');
    fireEvent.change(screen.getByRole('combobox', { name: '長さ' }), { target: { value: '14' } });
    addChord('F');

    const [first, second] = screen.getAllByRole('combobox', { name: '長さ' });
    expect(lengths(first)).toEqual(Array.from({ length: 16 }, (_, i) => i + 1));
    expect(lengths(second)).toEqual([1, 2]);
    await screen.findByRole('heading', { name: '末尾に足す候補（4 拍）' });
  });

  it('does not let a candidate be added while the next candidates are awaited', async () => {
    let release: () => void = () => undefined;
    const held = new Promise<void>((resolve) => (release = resolve));
    vi.stubGlobal(
      'fetch',
      vi.fn(async (url: string, init?: RequestInit) => {
        const body = init?.body as string;
        if (url.endsWith('/progressions/next-candidates') && body.includes('"F"')) {
          await held;
        }
        return standInServer(url, init);
      }),
    );
    render(<ProgressionScreen />);
    addChord('Dm7');
    const add = await screen.findByRole('button', { name: 'F を足す' });

    fireEvent.click(add);

    expect(screen.getByRole('button', { name: 'F を足す' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'F を試聴する' })).toBeEnabled();
    fireEvent.click(screen.getByRole('button', { name: 'F を足す' }));
    expect(
      within(screen.getByRole('list', { name: '進行' })).getAllByRole('listitem'),
    ).toHaveLength(2);

    await act(async () => {
      release();
      await held;
    });
    await waitFor(() => expect(screen.getByRole('button', { name: 'F を足す' })).toBeEnabled());
  });
});
