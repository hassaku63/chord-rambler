import { useCallback, useEffect, useRef, useState } from 'react';
import type { Note, Progression } from '@/shared/api/client';
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

export type PlaybackState = {
  tempo: number;
  setTempo: (tempo: number) => void;
  /** What is playing: the whole progression, one slot, a candidate on trial, or nothing. */
  playing: 'progression' | 'slot' | 'audition' | null;
  /** The chord on trial while one plays, or null. */
  auditioning: string | null;
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
  stop: () => void;
};

/**
 * Plays the progression, one slot of it, or a candidate after the last slot,
 * at the tempo the screen holds. A change to the progression stops the
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
  const [auditioned, setAuditioned] = useState('');
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
    audition: (chord, beats) => {
      pending.current.trial?.abort();
      const trial = new AbortController();
      pending.current.trial = trial;
      const withChord = [...slots, { chord, beats }];
      const last = slots.length - 1;
      const from = slots[last]?.chord === undefined ? last + 1 : last;
      get({ ...progression, slots: withChord }, trial.signal)
        .then((answer) => {
          if (!trial.signal.aborted && answer.ok) {
            start(
              passagePerformance(answer.notes, withChord, from, withChord.length, tempo),
              'audition',
            );
            setAuditioned(chord);
          }
        })
        .catch(() => undefined);
    },
    stop,
  };
}
