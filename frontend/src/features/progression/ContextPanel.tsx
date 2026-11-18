import { useState } from 'react';
import type { Slot, SlotReading } from '@/shared/api/client';
import { CandidateList } from './CandidateList';
import { candidateMarks, candidateRows } from './labels';
import type { PlaybackState } from './usePlayback';
import { triedAt } from './usePlayback';
import type { ProgressionState } from './useProgression';
import { useNextCandidates } from './useNextCandidates';
import { slotName } from './SlotCard';
import { SlotDetail } from './SlotDetail';
import { SlotFill } from './SlotFill';
import { useVariations } from './useVariations';
import { VariationList } from './VariationList';

/** The two tabs on the same progression (ADR-0019): exploring from the last slot, and planning. */
export type Tab = 'explore' | 'plan';

type Props = {
  tab: Tab;
  state: ProgressionState;
  readings: SlotReading[];
  playback: PlaybackState;
};

/** The labels and the variations of the chosen slot. */
function ChosenSlot({
  state,
  index,
  slot,
  reading,
}: {
  state: ProgressionState;
  index: number;
  slot: Slot;
  reading: SlotReading | undefined;
}) {
  const chord = slot.chord;
  // A slot added from the candidates keeps its variations from that chord (#115).
  const base = chord === undefined ? null : (state.bases[index] ?? chord);
  const variations = useVariations(state.progression.key, base);
  return (
    <>
      <SlotDetail slot={slot} reading={reading} />
      {chord !== undefined && base !== null && variations.state === 'given' && (
        <VariationList
          base={base}
          current={chord}
          variations={variations.variations}
          onChoose={(name) => state.replaceChord(index, name)}
        />
      )}
    </>
  );
}

/**
 * The context panel, which follows what is chosen: always which slot it is;
 * for a chosen slot its labels and variations, and in the plan tab the
 * candidates to fill it; in the explore tab with nothing chosen, the
 * candidates for the slot after the last, as the other candidates (the
 * branches come in a later slice).
 */
export function ContextPanel({ tab, state, readings, playback }: Props) {
  const { progression, selected } = state;
  // Asked for even while hidden, so that the last answer is there when the slot is let go.
  const { beats, candidates, busy } = useNextCandidates(progression);
  // The candidate groups opened or closed, kept here so that a new answer keeps them.
  const [opened, setOpened] = useState<Partial<Record<string, boolean>>>({});
  const chosen = selected === null ? undefined : progression.slots[selected];

  return (
    <aside className="context-panel" aria-labelledby="context-heading">
      <p className="context-where">{tab === 'explore' ? '探索' : '計画'}のタブ</p>
      <h2 id="context-heading">
        {chosen === undefined || selected === null
          ? '枠を選んでいません'
          : `${slotName(chosen, readings[selected])}（${String(selected + 1)} 番目の枠）`}
      </h2>

      {chosen && selected !== null && (
        <ChosenSlot state={state} index={selected} slot={chosen} reading={readings[selected]} />
      )}

      {tab === 'plan' && (
        <SlotFill
          progression={progression}
          selected={selected}
          playback={playback}
          onFill={state.fillSlot}
        />
      )}

      {tab === 'plan' && selected === null && <p>枠を選ぶと、その枠を埋める候補が出ます。</p>}

      {tab === 'explore' && selected === null && candidates.state === 'given' && (
        <CandidateList
          words={{
            heading: `他の候補（末尾に足す、${String(beats)} 拍）`,
            list: (kind) => `${kind}の候補`,
            add: '足す',
            trial: (chord) => `${chord} を試聴する`,
          }}
          groups={candidates.groups}
          marksOf={candidateMarks}
          rowsOf={candidateRows}
          busy={busy}
          opened={opened}
          onToggle={(kind, open) => setOpened((o) => ({ ...o, [kind]: open }))}
          auditioning={triedAt(playback.auditioning, progression.slots.length)}
          onAdd={(name) => state.addCandidate(name, beats)}
          onAudition={(name) => playback.audition(name, beats)}
          onStop={playback.stop}
        />
      )}
    </aside>
  );
}
