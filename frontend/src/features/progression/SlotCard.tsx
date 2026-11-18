import type { CSSProperties } from 'react';
import { Fragment } from 'react';
import type { Slot, SlotReading } from '@/shared/api/client';
import { cardMarks, numeralOf } from './labels';
import { ROLE_GROUPS, TERMS } from './terms';
import type { BarTick } from './useProgression';
import { MIN_BEATS } from './useProgression';

/** The lengths from one beat to the most the slot can have; a stored longer length stays chosen. */
function beatChoices(max: number, current: number): number[] {
  const top = Math.max(max, current);
  return Array.from({ length: top - MIN_BEATS + 1 }, (_, i) => MIN_BEATS + i);
}

/** The name a slot shows: the chord as the server spells it, or what kind of empty slot it is. */
export function slotName(slot: Slot, reading: SlotReading | undefined): string {
  if (slot.chord !== undefined) {
    return reading?.content === 'chord' ? reading.chord.name : slot.chord;
  }
  return slot.content === 'none' ? '和音なし' : 'まだ選んでいない';
}

type Props = {
  slot: Slot;
  /** The most beats the slot can have without crossing into the next phrase. */
  maxBeats: number;
  /** The bars that start within the slot. */
  ticks: BarTick[];
  reading: SlotReading | undefined;
  unreadable: boolean;
  selected: boolean;
  /** Whether the slot is sounding now. */
  sounding: boolean;
  onSelect: () => void;
  onRemove: () => void;
  onBeats: (beats: number) => void;
  /** Takes the role chosen, or an empty string for none. */
  onRole: (role: string) => void;
};

/**
 * One slot in the strip, as wide as its beats: the bars and phrases that start
 * in it, its chord, Roman numeral and short marks, its length, its role, and
 * removal.
 */
export function SlotCard({
  slot,
  maxBeats,
  ticks,
  reading,
  unreadable,
  selected,
  sounding,
  onSelect,
  onRemove,
  onBeats,
  onRole,
}: Props) {
  const name = slotName(slot, reading);
  const marks = reading ? cardMarks(reading) : [];
  const lines = [name, numeralOf(reading), marks.join(' '), unreadable && 'この枠を読めません'];
  const shown = lines.filter((line) => typeof line === 'string' && line !== '');
  return (
    <li
      className={['slot-card', selected && 'selected', sounding && 'sounding']
        .filter(Boolean)
        .join(' ')}
      aria-current={sounding || undefined}
      style={{ '--beats': slot.beats } as CSSProperties}
    >
      <span className="slot-ticks">
        {ticks.map((tick) => (
          <span
            key={tick.bar}
            className={tick.phrase === null ? 'tick' : 'tick phrase'}
            style={{ left: `${String(tick.at * 100)}%` }}
          >
            {tick.phrase === null ? `小節 ${String(tick.bar)}` : `フレーズ ${String(tick.phrase)}`}
          </span>
        ))}
      </span>
      {/* The line keeps its place, so that the strip does not move as the playing goes on. */}
      <span className="slot-sounding">{sounding && '鳴っています'}</span>
      <button
        type="button"
        className="slot-select"
        aria-pressed={selected}
        aria-label={shown.join(' ')}
        onClick={onSelect}
      >
        {/* The spaces between the lines keep the visible text inside the name
            (WCAG 2.5.3); the column of lines does not show them. */}
        {shown.map((line, i) => (
          <Fragment key={i}>
            {i > 0 && ' '}
            <span className={i === 0 ? 'slot-name' : undefined}>{line}</span>
          </Fragment>
        ))}
      </button>
      <label>
        長さ
        <select value={slot.beats} onChange={(e) => onBeats(Number(e.target.value))}>
          {beatChoices(maxBeats, slot.beats).map((beats) => (
            <option key={beats} value={beats}>
              {beats} 拍
            </option>
          ))}
        </select>
      </label>
      <label>
        役割
        <select value={slot.role ?? ''} onChange={(e) => onRole(e.target.value)}>
          <option value="">なし</option>
          {ROLE_GROUPS.map((group) => (
            <optgroup key={group.kind} label={group.kind}>
              {group.roles.map((role) => (
                <option key={role} value={role}>
                  {TERMS.role[role]}
                </option>
              ))}
            </optgroup>
          ))}
        </select>
      </label>
      <button type="button" onClick={onRemove} aria-label={`${name} の枠を消す`}>
        消す
      </button>
    </li>
  );
}
