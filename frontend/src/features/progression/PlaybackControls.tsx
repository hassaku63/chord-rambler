import { MAX_TEMPO, MIN_TEMPO } from './schedule';
import type { PlaybackState } from './usePlayback';

type Props = {
  playback: PlaybackState;
  /** The chosen slot, which "この枠だけ鳴らす" plays, or null. */
  selected: number | null;
};

/**
 * The playing of the progression: the whole, the chosen slot alone, stop, and
 * the tempo. A pressed play button stops when pressed again.
 */
export function PlaybackControls({ playback, selected }: Props) {
  const { playing, tempo } = playback;
  const slotPlayable = selected !== null && playback.slotPlayable(selected);
  return (
    <div className="playback" role="group" aria-label="再生">
      <button
        type="button"
        aria-pressed={playing === 'progression'}
        disabled={!playback.playable}
        onClick={playing === 'progression' ? playback.stop : playback.playProgression}
      >
        進行を鳴らす
      </button>
      <button
        type="button"
        aria-pressed={playing === 'slot'}
        disabled={!slotPlayable}
        onClick={() => {
          if (playing === 'slot') {
            playback.stop();
          } else if (selected !== null) {
            playback.playSlot(selected);
          }
        }}
      >
        この枠だけ鳴らす
      </button>
      <button type="button" disabled={playing === null} onClick={playback.stop}>
        止める
      </button>
      <label>
        テンポ
        <input
          type="range"
          min={MIN_TEMPO}
          max={MAX_TEMPO}
          step={1}
          value={tempo}
          aria-valuetext={`${tempo} BPM`}
          onChange={(e) => playback.setTempo(Number(e.target.value))}
        />
        <span>{tempo} BPM</span>
      </label>
    </div>
  );
}
