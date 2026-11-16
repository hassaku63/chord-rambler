import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { stubServer, world } from '@/test/standInServer';
import { ProgressionScreen } from './ProgressionScreen';

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
  beforeEach(() => stubServer());
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

  it('shortens a slot that a removal makes cross, says so, and reads the progression', async () => {
    render(<ProgressionScreen />);
    addChord('Dm7');
    addChord('F');
    fireEvent.change(screen.getAllByRole('combobox', { name: '長さ' })[1] ?? document.body, {
      target: { value: '12' },
    });
    addChord('Fm');
    fireEvent.change(screen.getAllByRole('combobox', { name: '長さ' })[2] ?? document.body, {
      target: { value: '16' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Dm7 の枠を消す' }));

    const strip = screen.getByRole('list', { name: '進行' });
    // The server reads Fm only when no slot crosses into the next phrase.
    expect(await within(strip).findByRole('button', { name: /^Fm iv/ })).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(lengths(screen.getAllByRole('combobox', { name: '長さ' })[1])).toEqual([1, 2, 3, 4]);
    expect(screen.getAllByRole('combobox', { name: '長さ' })[1]).toHaveValue('4');
    expect(screen.getByText('Fm の枠を 4 拍に縮めました（フレーズをまたぐため）。')).toHaveRole(
      'status',
    );
    await screen.findByRole('heading', { name: '末尾に足す候補（4 拍）' });
  });

  it('does not let a candidate be added while the next candidates are awaited', async () => {
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
