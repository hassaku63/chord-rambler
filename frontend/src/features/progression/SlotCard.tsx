import type { Slot, SlotReading } from '@/shared/api/client';
import { cardMarks } from './labels';
import { ROLE_GROUPS, TERMS } from './terms';
import { MAX_BEATS, MIN_BEATS } from './useProgression';

const BEAT_CHOICES = Array.from({ length: MAX_BEATS - MIN_BEATS + 1 }, (_, i) => MIN_BEATS + i);

/** The name a slot shows: the chord as the server spells it, or what kind of empty slot it is. */
export function slotName(slot: Slot, reading: SlotReading | undefined): string {
  if (slot.chord !== undefined) {
    return reading?.chord?.name ?? slot.chord;
  }
  return slot.content === 'none' ? '和音なし' : 'まだ選んでいない';
}

type Props = {
  slot: Slot;
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
 * One slot in the strip: its chord, Roman numeral and short marks, its length,
 * its role, and removal.
 */
export function SlotCard({
  slot,
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
  const lines = [
    name,
    reading?.roman?.numeral,
    marks.join(' '),
    unreadable && 'この枠を読めません',
  ];
  const shown = lines.filter((line) => typeof line === 'string' && line !== '');
  return (
    <li
      className={['slot-card', selected && 'selected', sounding && 'sounding']
        .filter(Boolean)
        .join(' ')}
      aria-current={sounding || undefined}
    >
      {sounding && <span className="slot-sounding">鳴っています</span>}
      <button
        type="button"
        className="slot-select"
        aria-pressed={selected}
        aria-label={shown.join(' ')}
        onClick={onSelect}
      >
        {shown.map((line, i) => (
          <span key={i} className={i === 0 ? 'slot-name' : undefined}>
            {line}
          </span>
        ))}
      </button>
      <label>
        長さ
        <select value={slot.beats} onChange={(e) => onBeats(Number(e.target.value))}>
          {BEAT_CHOICES.map((beats) => (
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
