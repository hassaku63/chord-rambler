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

// C for 4 beats and then a slot with no chord for 2.
const emptyLast: Progression = { key: 'C', slots: progression.slots.slice(0, 2) };

const fetchNotes = vi.fn(() => Promise.resolve<NotesAnswer>(answer));

async function renderPlayback(p: Progression = progression) {
  const audio = fakeAudio();
  fetchNotes.mockClear();
  const { result, rerender } = renderHook(({ p }) => usePlayback(p, audio.create, fetchNotes), {
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

  it('plays the last slot and then a candidate, asking for the notes with it added', async () => {
    const { audio, result } = await renderPlayback();
    const withF: NotesAnswer = {
      ok: true,
      notes: [
        { pitch: 43, start: 6, beats: 4, velocity: 64 },
        { pitch: 67, start: 6, beats: 4, velocity: 64 },
        { pitch: 65, start: 10, beats: 2, velocity: 64 },
      ],
    };
    fetchNotes.mockResolvedValueOnce(withF);

    act(() => result.current.audition('F', 2));

    await waitFor(() => expect(result.current.playing).toBe('audition'));
    expect(fetchNotes).toHaveBeenLastCalledWith(
      { key: 'C', slots: [...progression.slots, { chord: 'F', beats: 2 }] },
      expect.any(AbortSignal),
    );
    expect(result.current.auditioning).toEqual({ chord: 'F', at: 3 });
    // G from 0, then F 4 beats later: 4 × 60 / 90 = 2.67 s.
    const starts = audio.context.oscillators.map((o) => o.startedAt);
    expect(starts).toEqual([1.05, 1.05, expect.closeTo(3.717)]);
  });

  it('plays the candidate alone after a slot with no chord', async () => {
    const { audio, result } = await renderPlayback(emptyLast);
    fetchNotes.mockResolvedValueOnce({
      ok: true,
      notes: [{ pitch: 65, start: 6, beats: 4, velocity: 64 }],
    });

    act(() => result.current.audition('F', 4));

    await waitFor(() => expect(result.current.playing).toBe('audition'));
    expect(audio.context.oscillators.map((o) => o.startedAt)).toEqual([1.05]);
  });

  it('plays the slots around a chord tried in a slot, each only when it holds a chord', async () => {
    const { audio, result } = await renderPlayback();
    fetchNotes.mockResolvedValueOnce({
      ok: true,
      notes: [
        { pitch: 48, start: 0, beats: 4, velocity: 64 },
        { pitch: 62, start: 4, beats: 2, velocity: 64 },
        { pitch: 43, start: 6, beats: 4, velocity: 64 },
      ],
    });

    act(() => result.current.auditionIn(1, 'Dm'));

    await waitFor(() => expect(result.current.playing).toBe('audition'));
    expect(fetchNotes).toHaveBeenLastCalledWith(
      {
        key: 'C',
        slots: [
          { chord: 'C', beats: 4 },
          { chord: 'Dm', beats: 2 },
          { chord: 'G', beats: 4 },
        ],
      },
      expect.any(AbortSignal),
    );
    expect(result.current.auditioning).toEqual({ chord: 'Dm', at: 1 });
    // C from 0, Dm 4 beats later (2.67 s), G 6 beats later (4 s).
    const starts = audio.context.oscillators.map((o) => o.startedAt);
    expect(starts).toEqual([1.05, expect.closeTo(3.717), expect.closeTo(5.05)]);

    fetchNotes.mockResolvedValueOnce(answer);
    act(() => result.current.auditionIn(0, 'C'));

    // Only C and the slot after it, which holds no chord, so C alone.
    await waitFor(() => expect(result.current.auditioning).toEqual({ chord: 'C', at: 0 }));
    expect(audio.context.oscillators.slice(3).map((o) => o.startedAt)).toEqual([
      expect.closeTo(1.05),
      expect.closeTo(1.05),
    ]);
  });

  it('drops a trial whose notes come after the progression changed', async () => {
    const { audio, result, rerender } = await renderPlayback();
    let resolve: (a: NotesAnswer) => void = () => undefined;
    fetchNotes.mockReturnValueOnce(new Promise<NotesAnswer>((r) => (resolve = r)));
    act(() => result.current.audition('F', 4));

    rerender({ p: { ...progression, key: 'G' } });
    await act(async () => {
      resolve(answer);
      await Promise.resolve();
    });

    expect(result.current.playing).toBeNull();
    expect(audio.context.oscillators).toEqual([]);
  });

  it.each([39, 201, 90.5])('keeps the tempo when given %s', async (tempo) => {
    const { result } = await renderPlayback();

    act(() => result.current.setTempo(tempo));

    expect(result.current.tempo).toBe(90);
  });
});
