import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Schemas } from '@/test/standInServer';
import { stubServer } from '@/test/standInServer';
import { ProgressionScreen } from './ProgressionScreen';

/** The readings of Dm7, and of a slot with no chord, before the role fit. */
const dm7: Schemas['SlotReading'] = {
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
};
const COLOUR_ROLES: readonly Schemas['Role'][] = ['spread', 'float', 'wistful', 'linger'];

/**
 * Reads Dm7 and slots with no chord: a slot with a structural role gets
 * undecided when it holds no chord and misfits when it holds Dm7; a colour
 * role gets no role fit.
 */
function readingsOf(progression: Schemas['Progression']): Schemas['ProgressionReading'] {
  const slots = progression.slots.map((slot): Schemas['SlotReading'] => {
    const reading = slot.chord === undefined ? {} : dm7;
    if (slot.role === undefined || COLOUR_ROLES.includes(slot.role)) {
      return reading;
    }
    const fit = slot.chord === undefined ? 'undecided' : 'misfits';
    return { ...reading, roleFit: { basis: 'school', fit } };
  });
  return { slots };
}

function cards() {
  return within(screen.getByRole('list', { name: '進行' })).getAllByRole('listitem');
}

/** The progression the last request for the readings sent. */
function lastRead(): unknown {
  const reads = vi
    .mocked(fetch)
    .mock.calls.filter(([url]) => url === '/api/v1/progressions/readings');
  return JSON.parse(reads.at(-1)?.[1]?.body as string);
}

function roleOf(card: HTMLElement | undefined) {
  return within(card ?? document.body).getByRole('combobox', { name: '役割' });
}

describe('ProgressionScreen roles', () => {
  beforeEach(() => stubServer({ '/progressions/readings': readingsOf }));
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it('offers no role, the structural roles and the colour roles, by kind', () => {
    render(<ProgressionScreen />);
    fireEvent.click(screen.getByRole('button', { name: 'まだ選んでいない枠を足す' }));

    const role = roleOf(cards()[0]);

    expect(role).toHaveValue('');
    expect(
      within(role)
        .getAllByRole('option')
        .map((o) => o.textContent),
    ).toEqual([
      'なし',
      '安定',
      'つなぎ',
      '準備',
      '高まり',
      '解決',
      '締め',
      '広がり',
      '浮遊',
      '切なさ',
      '余韻',
    ]);
    expect(
      within(role)
        .getAllByRole('group')
        .map((g) => g.getAttribute('label')),
    ).toEqual(['構造的な役割', '色合いの役割']);
  });

  it('puts a role on a slot, sends it and shows whether the slot fits it', async () => {
    render(<ProgressionScreen />);
    fireEvent.change(screen.getByRole('textbox', { name: '和音名' }), {
      target: { value: 'Dm7' },
    });
    fireEvent.click(screen.getByRole('button', { name: '枠を足す' }));
    fireEvent.click(screen.getByRole('button', { name: 'まだ選んでいない枠を足す' }));
    fireEvent.click(screen.getByRole('button', { name: '和音なしの枠を足す' }));
    const [first, second, third] = cards();

    fireEvent.change(roleOf(first), { target: { value: 'rise' } });
    fireEvent.change(roleOf(second), { target: { value: 'prepare' } });
    fireEvent.change(roleOf(third), { target: { value: 'wistful' } });

    expect(roleOf(first)).toHaveValue('rise');
    expect(lastRead()).toEqual({
      key: 'C',
      slots: [
        { chord: 'Dm7', beats: 4, role: 'rise' },
        { content: 'unchosen', beats: 4, role: 'prepare' },
        { content: 'none', beats: 4, role: 'wistful' },
      ],
    });
    const card = first ?? document.body;
    expect(await within(card).findByText('ダイアトニック S 役割に合うか 合わない')).toBeVisible();
    expect(within(second ?? document.body).getByText('役割に合うか 決められない')).toBeVisible();
    expect(within(third ?? document.body).queryByText(/役割に合うか/)).not.toBeInTheDocument();

    fireEvent.click(within(card).getByRole('button', { name: /^Dm7 ii7/ }));

    const detail = screen.getByRole('region', { name: 'Dm7（ii7） の札' });
    const text = (el: HTMLElement) => el.textContent;
    expect(within(detail).getAllByRole('term').map(text)).toEqual(['種別', '機能', '役割に合うか']);
    expect(within(detail).getAllByRole('definition').map(text).at(-1)).toBe('合わない 〔流儀〕');
  });

  it('takes a role away with なし', () => {
    render(<ProgressionScreen />);
    fireEvent.click(screen.getByRole('button', { name: 'まだ選んでいない枠を足す' }));
    fireEvent.change(roleOf(cards()[0]), { target: { value: 'close' } });

    fireEvent.change(roleOf(cards()[0]), { target: { value: '' } });

    expect(roleOf(cards()[0])).toHaveValue('');
    expect(lastRead()).toEqual({ key: 'C', slots: [{ content: 'unchosen', beats: 4 }] });
  });
});
