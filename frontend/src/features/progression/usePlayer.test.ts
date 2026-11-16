import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fakeAudio } from '@/test/fakeAudioContext';
import type { Performance } from './schedule';
import { usePlayer } from './usePlayer';

// Two slots of one second each; a voice in each. The fake's clock reads 1 s,
// and the player starts 0.05 s ahead of it.
const performance: Performance = {
  voices: [
    { frequency: 261.63, start: 0, duration: 1, gain: 0.5 },
    { frequency: 392, start: 1, duration: 1, gain: 0.5 },
  ],
  marks: [
    { at: 0, slot: 0 },
    { at: 1, slot: 1 },
  ],
  end: 2,
};

describe('usePlayer', () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it('makes no context until the first play', () => {
    const audio = fakeAudio();

    renderHook(() => usePlayer(audio.create));

    expect(audio.made()).toBe(0);
  });

  it('starts each voice at its time with a short rise and a fade', () => {
    const audio = fakeAudio();
    const { result } = renderHook(() => usePlayer(audio.create));

    act(() => result.current.play(performance));

    expect(audio.made()).toBe(1);
    expect(audio.context.state).toBe('running');
    const [first, second] = audio.context.oscillators;
    expect(first?.type).toBe('triangle');
    expect(first?.frequency.value).toBe(261.63);
    expect(first?.startedAt).toBeCloseTo(1.05);
    expect(first?.stoppedAt).toBeCloseTo(2.05);
    expect(first?.changes).toEqual([
      ['set', 0, expect.closeTo(1.05)],
      ['ramp', 0.5, expect.closeTo(1.07)],
      ['ramp', 0, expect.closeTo(2.05)],
    ]);
    expect(second?.startedAt).toBeCloseTo(2.05);
  });

  it('says which slot sounds as the playing goes, and stops at the end', () => {
    const audio = fakeAudio();
    const { result } = renderHook(() => usePlayer(audio.create));

    act(() => result.current.play(performance));
    expect(result.current.playing).toBe(true);

    act(() => {
      vi.advanceTimersByTime(50);
    });
    expect(result.current.sounding).toBe(0);
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current.sounding).toBe(1);
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current).toMatchObject({ playing: false, sounding: null });
  });

  it('cuts every voice and forgets the slot when stopped', () => {
    const audio = fakeAudio();
    const { result } = renderHook(() => usePlayer(audio.create));
    act(() => result.current.play(performance));
    act(() => {
      vi.advanceTimersByTime(50);
    });

    act(() => result.current.stop());
    act(() => {
      vi.advanceTimersByTime(2000);
    });

    expect(audio.context.oscillators.map((o) => o.cut)).toEqual([true, true]);
    expect(result.current).toMatchObject({ playing: false, sounding: null });
  });

  it('keeps the one context and cuts the playing before for a new one', () => {
    const audio = fakeAudio();
    const { result } = renderHook(() => usePlayer(audio.create));

    act(() => result.current.play(performance));
    act(() => result.current.play(performance));

    expect(audio.made()).toBe(1);
    expect(audio.context.oscillators.map((o) => o.cut)).toEqual([true, true, false, false]);
  });

  it('closes the context when the screen goes', () => {
    const audio = fakeAudio();
    const { result, unmount } = renderHook(() => usePlayer(audio.create));
    act(() => result.current.play(performance));

    unmount();

    expect(audio.context.state).toBe('closed');
  });
});
