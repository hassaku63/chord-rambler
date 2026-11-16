import { useCallback, useEffect, useState } from 'react';
import type { Note, Progression } from '@/shared/api/client';
import { getProgressionNotes } from '@/shared/api/client';
import {
  DEFAULT_TEMPO,
  MAX_TEMPO,
  MIN_TEMPO,
  notesOfSlot,
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
  /** What is playing: the whole progression, one slot, or nothing. */
  playing: 'progression' | 'slot' | null;
  /** The slot sounding now, or null. */
  sounding: number | null;
  /** Whether the progression has notes to play. */
  playable: boolean;
  /** Whether the slot has notes to play. */
  slotPlayable: (index: number) => boolean;
  playProgression: () => void;
  playSlot: (index: number) => void;
  stop: () => void;
};

/**
 * Plays the progression, or one slot of it, at the tempo the screen holds.
 * A change to the progression stops the playing; a change to the tempo takes
 * effect at the next play.
 */
export function usePlayback(
  progression: Progression,
  create?: AudioContextFactory,
  get: typeof getProgressionNotes = getProgressionNotes,
): PlaybackState {
  const notes = useNotes(progression, get);
  const player = usePlayer(create);
  const { play, stop } = player;
  const [tempo, setTempo] = useState(DEFAULT_TEMPO);
  const [what, setWhat] = useState<'progression' | 'slot'>('progression');
  const { slots } = progression;

  useEffect(() => stop, [progression, stop]);

  return {
    tempo,
    setTempo: useCallback((value: number) => {
      if (Number.isInteger(value) && value >= MIN_TEMPO && value <= MAX_TEMPO) {
        setTempo(value);
      }
    }, []),
    playing: player.playing ? what : null,
    sounding: player.sounding,
    playable: notes !== null && notes.length > 0,
    slotPlayable: (index) => notes !== null && notesOfSlot(notes, slots, index).length > 0,
    playProgression: () => {
      if (notes !== null) {
        play(progressionPerformance(notes, slots, tempo));
        setWhat('progression');
      }
    },
    playSlot: (index) => {
      if (notes !== null) {
        play(slotPerformance(notes, slots, index, tempo));
        setWhat('slot');
      }
    },
    stop,
  };
}
