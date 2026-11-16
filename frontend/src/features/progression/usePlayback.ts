import { useCallback, useEffect, useRef, useState } from 'react';
import type { Note, Progression, Slot } from '@/shared/api/client';
import { getProgressionNotes } from '@/shared/api/client';
import {
  DEFAULT_TEMPO,
  MAX_TEMPO,
  MIN_TEMPO,
  notesOfSlot,
  passagePerformance,
  progressionPerformance,
  slotPerformance,
} from './schedule';
import type { AudioContextFactory } from './usePlayer';
import { usePlayer } from './usePlayer';

type Fetched = { for: Progression; notes: Note[] | null };

/**
 * The notes of the progression, asked for each time it changes. Null until
 * they come, and when the server refuses or cannot be reached: the reading
 * already says why. An answer to a progression that has since changed is
 * thrown away.
 */
function useNotes(progression: Progression, get: typeof getProgressionNotes): Note[] | null {
  const [fetched, setFetched] = useState<Fetched | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    const keep = (notes: Note[] | null) => {
      if (!controller.signal.aborted) {
        setFetched({ for: progression, notes });
      }
    };
    get(progression, controller.signal)
      .then((answer) => keep(answer.ok ? answer.notes : null))
      .catch(() => keep(null));
    return () => controller.abort();
  }, [progression, get]);

  return fetched?.for === progression ? fetched.notes : null;
}

/** The chord on trial in the slot at, or null when none is tried there. */
export function triedAt(auditioning: PlaybackState['auditioning'], at: number): string | null {
  return auditioning !== null && auditioning.at === at ? auditioning.chord : null;
}

export type PlaybackState = {
  tempo: number;
  setTempo: (tempo: number) => void;
  /** What is playing: the whole progression, one slot, a candidate on trial, or nothing. */
  playing: 'progression' | 'slot' | 'audition' | null;
  /**
   * The chord on trial while one plays and the slot it is tried in (the
   * number of slots for a slot after the last), or null.
   */
  auditioning: { chord: string; at: number } | null;
  /** The slot sounding now, or null. */
  sounding: number | null;
  /** Whether the progression has notes to play. */
  playable: boolean;
  /** Whether the slot has notes to play. */
  slotPlayable: (index: number) => boolean;
  playProgression: () => void;
  playSlot: (index: number) => void;
  /**
   * Plays the last slot, when it holds a chord, and then the chord in a slot
   * of the beats added after it: the notes are those of the progression with
   * the chord added.
   */
  audition: (chord: string, beats: number) => void;
  /**
   * Plays the slot before, the chord in place of the slot's, and the slot
   * after, each slot around only when it holds a chord: the notes are those
   * of the progression with the chord in the slot.
   */
  auditionIn: (index: number, chord: string) => void;
  stop: () => void;
};

/**
 * Plays the progression, one slot of it, or a candidate after the last slot or
 * in a slot, at the tempo the screen holds. A change to the progression stops the
 * playing; a change to the tempo takes effect at the next play.
 */
export function usePlayback(
  progression: Progression,
  create?: AudioContextFactory,
  get: typeof getProgressionNotes = getProgressionNotes,
): PlaybackState {
  const notes = useNotes(progression, get);
  const player = usePlayer(create);
  const { play, stop: silence } = player;
  const [tempo, setTempo] = useState(DEFAULT_TEMPO);
  const [what, setWhat] = useState<'progression' | 'slot' | 'audition'>('progression');
  const [auditioned, setAuditioned] = useState({ chord: '', at: 0 });
  // The request for the notes of a trial not yet played.
  const pending = useRef<{ trial: AbortController | null }>({ trial: null });
  const { slots } = progression;

  const stop = useCallback(() => {
    pending.current.trial?.abort();
    silence();
  }, [silence]);

  useEffect(() => stop, [progression, stop]);

  const start = (performance: Parameters<typeof play>[0], kind: typeof what) => {
    pending.current.trial?.abort();
    play(performance);
    setWhat(kind);
  };

  /** Plays the slots around the slot at, in the progression with the chord in it. */
  const trial = (chord: string, at: number, withChord: Slot[]) => {
    pending.current.trial?.abort();
    const controller = new AbortController();
    pending.current.trial = controller;
    const from = withChord[at - 1]?.chord === undefined ? at : at - 1;
    const to = withChord[at + 1]?.chord === undefined ? at + 1 : at + 2;
    get({ ...progression, slots: withChord }, controller.signal)
      .then((answer) => {
        if (!controller.signal.aborted && answer.ok) {
          start(passagePerformance(answer.notes, withChord, from, to, tempo), 'audition');
          setAuditioned({ chord, at });
        }
      })
      .catch(() => undefined);
  };

  return {
    tempo,
    setTempo: useCallback((value: number) => {
      if (Number.isInteger(value) && value >= MIN_TEMPO && value <= MAX_TEMPO) {
        setTempo(value);
      }
    }, []),
    playing: player.playing ? what : null,
    auditioning: player.playing && what === 'audition' ? auditioned : null,
    sounding: player.sounding,
    playable: notes !== null && notes.length > 0,
    slotPlayable: (index) => notes !== null && notesOfSlot(notes, slots, index).length > 0,
    playProgression: () => {
      if (notes !== null) {
        start(progressionPerformance(notes, slots, tempo), 'progression');
      }
    },
    playSlot: (index) => {
      if (notes !== null) {
        start(slotPerformance(notes, slots, index, tempo), 'slot');
      }
    },
    audition: (chord, beats) => trial(chord, slots.length, [...slots, { chord, beats }]),
    auditionIn: (index, chord) =>
      trial(
        chord,
        index,
        slots.map((slot, i) => (i === index ? { chord, beats: slot.beats } : slot)),
      ),
    stop,
  };
}
