import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { NotesAnswer, Progression } from '@/shared/api/client';
import { fakeAudio } from '@/test/fakeAudioContext';
import { usePlayback } from './usePlayback';

// C for 4 beats, a slot with no chord for 2, G for 4, and the notes the server gives.
const progression: Progression = {
  key: 'C',
  slots: [
    { chord: 'C', beats: 4 },
    { content: 'none', beats: 2 },
    { chord: 'G', beats: 4 },
  ],
};
const answer: NotesAnswer = {
  ok: true,
  notes: [
    { pitch: 48, start: 0, beats: 4, velocity: 64 },
    { pitch: 60, start: 0, beats: 4, velocity: 64 },
    { pitch: 43, start: 6, beats: 4, velocity: 64 },
    { pitch: 67, start: 6, beats: 4, velocity: 64 },
  ],
};

async function renderPlayback(p: Progression = progression) {
  const audio = fakeAudio();
  const get = vi.fn(() => Promise.resolve(answer));
  const { result, rerender } = renderHook(({ p }) => usePlayback(p, audio.create, get), {
    initialProps: { p },
  });
  await waitFor(() => expect(result.current.playable).toBe(true));
  return { audio, result, rerender };
}

describe('usePlayback', () => {
  it('cannot play until the notes come', () => {
    const get = vi.fn(() => new Promise<NotesAnswer>(() => undefined));

    const { result } = renderHook(() => usePlayback(progression, fakeAudio().create, get));

    expect(result.current.playable).toBe(false);
    expect(get).toHaveBeenCalledWith(progression, expect.any(AbortSignal));
  });

  it('plays the whole progression at 90 BPM', async () => {
    const { audio, result } = await renderPlayback();

    act(() => result.current.playProgression());

    expect(result.current.playing).toBe('progression');
    // G starts on beat 6: 6 × 60 / 90 = 4 s after the start, 1.05 s on the fake's clock.
    const starts = audio.context.oscillators.map((o) => o.startedAt);
    expect(starts).toEqual([1.05, 1.05, expect.closeTo(5.05), expect.closeTo(5.05)]);
  });

  it('plays only the chosen slot, from the start, for its length', async () => {
    const { audio, result } = await renderPlayback();
    act(() => result.current.setTempo(120));

    act(() => result.current.playSlot(2));

    expect(result.current.playing).toBe('slot');
    const played = audio.context.oscillators.map((o) => [o.frequency.value, o.startedAt]);
    expect(played).toEqual([
      [expect.closeTo(98, 2), 1.05],
      [expect.closeTo(392, 2), 1.05],
    ]);
    // 4 beats at 120 BPM: 2 s.
    expect(audio.context.oscillators[0]?.stoppedAt).toBeCloseTo(3.05);
  });

  it('stops', async () => {
    const { audio, result } = await renderPlayback();
    act(() => result.current.playProgression());

    act(() => result.current.stop());

    expect(result.current.playing).toBeNull();
    expect(audio.context.oscillators.every((o) => o.cut)).toBe(true);
  });

  it('stops when the progression changes', async () => {
    const { result, rerender } = await renderPlayback();
    act(() => result.current.playProgression());

    rerender({ p: { ...progression, key: 'G' } });

    expect(result.current.playing).toBeNull();
  });

  it('says a slot with no chord has nothing to play', async () => {
    const { result } = await renderPlayback();

    expect([0, 1, 2].map(result.current.slotPlayable)).toEqual([true, false, true]);
  });

  it.each([39, 201, 90.5])('keeps the tempo when given %s', async (tempo) => {
    const { result } = await renderPlayback();

    act(() => result.current.setTempo(tempo));

    expect(result.current.tempo).toBe(90);
  });
});
