import { useState } from 'react';
import type { ReadingState } from './useProgressionReading';
import { useProgressionReading } from './useProgressionReading';
import { KEYS, useProgression } from './useProgression';
import { SlotCard } from './SlotCard';
import { SlotDetail } from './SlotDetail';
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

/**
 * Screen 1 (a provisional layout, A1 of #148): the key on top, the strip of
 * slots in the middle, the labels of the chosen slot below.
 */
export function ProgressionScreen() {
  const state = useProgression();
  const { progression, selected } = state;
  const reading = useProgressionReading(progression);
  const [chordName, setChordName] = useState('');
  const readings = reading.state === 'read' ? reading.reading.slots : [];
  const unreadable = reading.state === 'refused' ? reading.slot : null;
  const chosen = selected === null ? undefined : progression.slots[selected];

  return (
    <div className="progression-screen">
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

      <ReadingAlert reading={reading} />

      <ol className="slot-strip" role="list" aria-label="進行">
        {progression.slots.map((slot, i) => (
          <SlotCard
            // Slots have no identity of their own; the position is the slot.
            key={i}
            slot={slot}
            reading={readings[i]}
            unreadable={unreadable === i}
            selected={selected === i}
            onSelect={() => state.select(selected === i ? null : i)}
            onRemove={() => state.remove(i)}
            onBeats={(beats) => state.setBeats(i, beats)}
          />
        ))}
      </ol>

      {chosen && selected !== null && <SlotDetail slot={chosen} reading={readings[selected]} />}
    </div>
  );
}
