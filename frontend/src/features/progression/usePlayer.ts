import { useCallback, useEffect, useRef, useState } from 'react';
import type { Performance, Voice } from './schedule';

/** Makes the AudioContext the player sounds through; tests pass a stand-in. */
export type AudioContextFactory = () => AudioContext;

export const newAudioContext: AudioContextFactory = () => new AudioContext();

/** How far ahead of now the first voice starts, so that it is not cut. */
export const LEAD_SECONDS = 0.05;
/** The rise of each voice, for a soft start; after it, the voice fades to its end. */
export const ATTACK_SECONDS = 0.02;
/** The loudness of all the voices together: a chord's voices add up, and must not clip. */
export const MASTER_GAIN = 0.2;

export type Player = {
  playing: boolean;
  /** The slot sounding now, or null. */
  sounding: number | null;
  play: (performance: Performance) => void;
  stop: () => void;
};

type Output = { context: AudioContext; master: GainNode };

type Live = {
  audio: Output | null;
  oscillators: OscillatorNode[];
  timers: ReturnType<typeof setTimeout>[];
};

function silence(live: Live) {
  live.timers.forEach(clearTimeout);
  live.timers = [];
  for (const oscillator of live.oscillators) {
    oscillator.stop();
    oscillator.disconnect();
  }
  live.oscillators = [];
}

function withMaster(context: AudioContext): Output {
  const master = context.createGain();
  master.gain.value = MASTER_GAIN;
  master.connect(context.destination);
  return { context, master };
}

/** Starts one voice through the master: a triangle wave with a short rise and a fade. */
function startVoice({ context, master }: Output, at: number, voice: Voice) {
  const oscillator = context.createOscillator();
  oscillator.type = 'triangle';
  oscillator.frequency.value = voice.frequency;
  const envelope = context.createGain();
  const start = at + voice.start;
  const end = start + voice.duration;
  envelope.gain.setValueAtTime(0, start);
  envelope.gain.linearRampToValueAtTime(
    voice.gain,
    start + Math.min(ATTACK_SECONDS, voice.duration / 2),
  );
  envelope.gain.linearRampToValueAtTime(0, end);
  oscillator.connect(envelope);
  envelope.connect(master);
  oscillator.start(start);
  oscillator.stop(end);
  return oscillator;
}

/**
 * Plays performances through Web Audio. The context is made, or resumed, on
 * the first play, which comes from a click (browsers hold sound back until the
 * page is used). A new play stops the one before.
 */
export function usePlayer(create: AudioContextFactory = newAudioContext): Player {
  const live = useRef<Live>({ audio: null, oscillators: [], timers: [] });
  const [playing, setPlaying] = useState(false);
  const [sounding, setSounding] = useState<number | null>(null);

  const stop = useCallback(() => {
    silence(live.current);
    setPlaying(false);
    setSounding(null);
  }, []);

  const play = useCallback(
    (performance: Performance) => {
      const current = live.current;
      silence(current);
      const audio = current.audio ?? withMaster(create());
      current.audio = audio;
      if (audio.context.state === 'suspended') {
        void audio.context.resume();
      }
      const at = audio.context.currentTime + LEAD_SECONDS;
      current.oscillators = performance.voices.map((voice) => startVoice(audio, at, voice));
      current.timers = [
        ...performance.marks.map((mark) =>
          setTimeout(() => setSounding(mark.slot), (LEAD_SECONDS + mark.at) * 1000),
        ),
        setTimeout(stop, (LEAD_SECONDS + performance.end) * 1000),
      ];
      setPlaying(true);
      setSounding(null);
    },
    [create, stop],
  );

  useEffect(() => {
    const current = live.current;
    return () => {
      silence(current);
      void current.audio?.context.close();
      current.audio = null;
    };
  }, []);

  return { playing, sounding, play, stop };
}
