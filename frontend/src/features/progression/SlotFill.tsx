import { useState } from 'react';
import type { Progression } from '@/shared/api/client';
import { CandidateList } from './CandidateList';
import { slotCandidateMarks, slotCandidateRows } from './labels';
import type { PlaybackState } from './usePlayback';
import { triedAt } from './usePlayback';
import { useSlotCandidates } from './useSlotCandidates';

type Props = {
  progression: Progression;
  /** The slot chosen, counted from 0, or null. */
  selected: number | null;
  playback: PlaybackState;
  onFill: (index: number, chord: string) => void;
};

/** The slot to fill: the chosen one when it is not chosen yet or holds a chord, else null. */
function slotToFill({ slots }: Progression, selected: number | null): number | null {
  const slot = selected === null ? undefined : slots[selected];
  return slot === undefined || slot.content === 'none' ? null : selected;
}

/**
 * The candidates for the chosen slot (screen 4), in the list of the
 * candidates for the slot after the last: filling keeps the slot's length
 * and role, and a trial plays the slots around it. A slot with no chord
 * (none) shows none.
 */
export function SlotFill({ progression, selected, playback, onFill }: Props) {
  const index = slotToFill(progression, selected);
  const { candidates, busy } = useSlotCandidates(progression, index);
  // The groups opened or closed, kept here so that a new answer keeps them.
  const [opened, setOpened] = useState<Partial<Record<string, boolean>>>({});
  const slot = index === null ? undefined : progression.slots[index];
  if (index === null || slot === undefined || candidates.state !== 'given') {
    return null;
  }
  return (
    <CandidateList
      words={{
        heading: `選んだ枠を埋める候補（${slot.beats} 拍）`,
        list: (kind) => `この枠の${kind}の候補`,
        add: 'この枠に入れる',
        trial: (chord) => `${chord} をこの枠で試聴する`,
      }}
      groups={candidates.groups}
      marksOf={slotCandidateMarks}
      rowsOf={slotCandidateRows}
      busy={busy}
      opened={opened}
      onToggle={(kind, open) => setOpened((o) => ({ ...o, [kind]: open }))}
      auditioning={triedAt(playback.auditioning, index)}
      onAdd={(chord) => onFill(index, chord)}
      onAudition={(chord) => playback.auditionIn(index, chord)}
      onStop={playback.stop}
    />
  );
}
