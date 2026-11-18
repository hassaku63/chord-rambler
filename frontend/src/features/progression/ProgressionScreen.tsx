import { useRef, useState } from 'react';
import type { AudioContextFactory } from './usePlayer';
import type { ReadingState } from './useProgressionReading';
import { useProgressionReading } from './useProgressionReading';
import { KEYS, barTicks, maxBeatsAt, useProgression } from './useProgression';
import { usePlayback } from './usePlayback';
import { PlaybackControls } from './PlaybackControls';
import { Canvas } from './Canvas';
import type { Tab } from './ContextPanel';
import { ContextPanel } from './ContextPanel';
import { SlotCard } from './SlotCard';
import { shortenedNotice } from './terms';
import './progression.css';

function ReadingAlert({ reading }: { reading: ReadingState }) {
  switch (reading.state) {
    case 'refused':
      return <p role="alert">{reading.detail}</p>;
    case 'unreachable':
      return <p role="alert">読みを得られませんでした。</p>;
    default:
      return null;
  }
}

/** The two tabs on the same progression (ADR-0019): exploring from the last slot, and planning. */
const TABS: readonly { id: Tab; word: string }[] = [
  { id: 'explore', word: '探索' },
  { id: 'plan', word: '計画' },
];

/** The tabs, with the arrow keys moving between them (the WAI-ARIA tabs pattern). */
function Tabs({ tab, onTab }: { tab: Tab; onTab: (tab: Tab) => void }) {
  const buttons = useRef<Partial<Record<Tab, HTMLButtonElement | null>>>({});
  return (
    <div className="tabs" role="tablist" aria-label="やり方">
      {TABS.map(({ id, word }) => (
        <button
          key={id}
          ref={(button) => {
            buttons.current[id] = button;
          }}
          type="button"
          role="tab"
          id={`tab-${id}`}
          aria-selected={tab === id}
          aria-controls="workbench-panel"
          tabIndex={tab === id ? 0 : -1}
          onClick={() => onTab(id)}
          onKeyDown={(e) => {
            if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
              // With two tabs, either arrow goes to the other.
              const other = id === 'explore' ? 'plan' : 'explore';
              onTab(other);
              buttons.current[other]?.focus();
            }
          }}
        >
          {word}
        </button>
      ))}
    </div>
  );
}

type Props = {
  /** Makes the AudioContext the playing sounds through; tests pass a stand-in. */
  createAudioContext?: AudioContextFactory;
};

/**
 * The workbench (ADR-0019, docs/product/screen-promise.md): the key, the
 * chord name and the two tabs on top; the canvas and the strip on the left;
 * the context panel, which follows the chosen slot, on the right.
 */
export function ProgressionScreen({ createAudioContext }: Props) {
  const state = useProgression();
  const { progression, selected } = state;
  const reading = useProgressionReading(progression);
  const playback = usePlayback(progression, createAudioContext);
  const [chordName, setChordName] = useState('');
  const [tab, setTab] = useState<Tab>('explore');
  const readings = reading.state === 'read' ? reading.reading.slots : [];
  const unreadable = reading.state === 'refused' ? reading.slot : null;
  const ticks = barTicks(progression.slots);
  const select = (i: number) => state.select(selected === i ? null : i);

  return (
    <div className="workbench">
      <label>
        調
        <select value={progression.key} onChange={(e) => state.setKey(e.target.value)}>
          {KEYS.map((key) => (
            <option key={key} value={key}>
              {key} major
            </option>
          ))}
        </select>
      </label>

      <form
        className="add-slot"
        onSubmit={(e) => {
          e.preventDefault();
          state.addChord(chordName);
          setChordName('');
        }}
      >
        <label>
          和音名
          <input value={chordName} onChange={(e) => setChordName(e.target.value)} />
        </label>
        <button type="submit">枠を足す</button>
        <button type="button" onClick={state.addNoChord}>
          和音なしの枠を足す
        </button>
        <button type="button" onClick={state.addUnchosen}>
          まだ選んでいない枠を足す
        </button>
      </form>

      <Tabs tab={tab} onTab={setTab} />

      <div
        className={`workbench-main ${tab}`}
        id="workbench-panel"
        role="tabpanel"
        aria-labelledby={`tab-${tab}`}
      >
        <div className="workbench-flow">
          <Canvas
            slots={progression.slots}
            readings={readings}
            selected={selected}
            sounding={playback.sounding}
            onSelect={select}
          />

          <section className="strip" aria-labelledby="strip-heading">
            <h2 id="strip-heading" className="area-heading">
              時間と再生
            </h2>
            <ReadingAlert reading={reading} />
            <ol className="slot-strip" role="list" aria-label="進行">
              {progression.slots.map((slot, i) => (
                <SlotCard
                  // Slots have no identity of their own; the position is the slot.
                  key={i}
                  slot={slot}
                  maxBeats={maxBeatsAt(progression.slots, i)}
                  ticks={ticks[i] ?? []}
                  reading={readings[i]}
                  unreadable={unreadable === i}
                  selected={selected === i}
                  sounding={playback.sounding === i}
                  onSelect={() => select(i)}
                  onRemove={() => state.remove(i)}
                  onBeats={(beats) => state.setBeats(i, beats)}
                  onRole={(role) => state.setRole(i, role)}
                />
              ))}
            </ol>

            <PlaybackControls playback={playback} selected={selected} />

            {/* Always there, so that the sentence is read out when it comes (WCAG 4.1.3). */}
            <p role="status" className="slot-notice">
              {shortenedNotice(state.shortened)}
            </p>
          </section>
        </div>

        <ContextPanel tab={tab} state={state} readings={readings} playback={playback} />
      </div>
    </div>
  );
}
