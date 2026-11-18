import { useState } from 'react';
import type { SlotReading } from '@/shared/api/client';
import type { AudioContextFactory } from './usePlayer';
import type { ReadingState } from './useProgressionReading';
import { useProgressionReading } from './useProgressionReading';
import type { ProgressionState } from './useProgression';
import { KEYS, barTicks, maxBeatsAt, useProgression } from './useProgression';
import type { PlaybackState } from './usePlayback';
import { triedAt, usePlayback } from './usePlayback';
import { useNextCandidates } from './useNextCandidates';
import type { BranchesState } from './useBranches';
import { useBranches } from './useBranches';
import { Branches } from './Branches';
import { BranchDetail } from './BranchDetail';
import { CandidateList } from './CandidateList';
import { candidateMarks, candidateRows } from './labels';
import { SlotFill } from './SlotFill';
import { useVariations } from './useVariations';
import { VariationList } from './VariationList';
import { PlaybackControls } from './PlaybackControls';
import { Canvas } from './Canvas';
import { SlotCard, slotName } from './SlotCard';
import { SlotDetail } from './SlotDetail';
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
type Tab = 'explore' | 'plan';

const TABS: readonly { id: Tab; word: string }[] = [
  { id: 'explore', word: '探索' },
  { id: 'plan', word: '計画' },
];

/** The tabs, with the arrow keys moving between them (the WAI-ARIA tabs pattern). */
function Tabs({ tab, onTab }: { tab: Tab; onTab: (tab: Tab) => void }) {
  return (
    <div className="tabs" role="tablist" aria-label="やり方">
      {TABS.map(({ id, word }) => (
        <button
          key={id}
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
              const other =
                e.currentTarget.nextElementSibling ?? e.currentTarget.previousElementSibling;
              onTab(id === 'explore' ? 'plan' : 'explore');
              if (other instanceof HTMLElement) {
                other.focus();
              }
            }
          }}
        >
          {word}
        </button>
      ))}
    </div>
  );
}

/** The key, and adding a slot by its chord name or with no chord. */
function Controls({ state }: { state: ProgressionState }) {
  const [chordName, setChordName] = useState('');
  return (
    <>
      <label>
        調
        <select value={state.progression.key} onChange={(e) => state.setKey(e.target.value)}>
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
    </>
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
  // Asked for even while hidden, so that the last answer is there when the slot is let go.
  const next = useNextCandidates(progression);
  const branches = useBranches(state, next, playback);
  const [tab, setTab] = useState<Tab>('explore');
  const readings = reading.state === 'read' ? reading.reading.slots : [];
  const unreadable = reading.state === 'refused' ? reading.slot : null;
  const ticks = barTicks(progression.slots);
  const select = (i: number) => {
    branches.unpick();
    state.select(selected === i ? null : i);
  };
  const explore = tab === 'explore';

  return (
    <div className="workbench">
      <Controls state={state} />

      <Tabs tab={tab} onTab={setTab} />

      <div
        className="workbench-main"
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
            branches={
              explore && (
                <Branches
                  group={branches.tail}
                  label="枝"
                  busy={branches.busy}
                  toolbarRef={branches.tailRef}
                />
              )
            }
            footprint={
              explore && branches.footprint
                ? {
                    at: branches.footprint.at,
                    element: <Branches group={branches.footprint} label="探索の足跡" footprint />,
                  }
                : null
            }
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

        <ContextPanel
          tab={tab}
          state={state}
          readings={readings}
          playback={playback}
          next={next}
          branches={branches}
        />
      </div>
    </div>
  );
}

/** The labels and the variations of the chosen slot, if any. */
function ChosenSlot({ state, readings }: { state: ProgressionState; readings: SlotReading[] }) {
  const { progression, selected } = state;
  const chosen = selected === null ? undefined : progression.slots[selected];
  const chord = chosen?.chord;
  // A slot added from the candidates keeps its variations from that chord (#115).
  const base = chord === undefined || selected === null ? null : (state.bases[selected] ?? chord);
  const variations = useVariations(progression.key, base);

  return (
    <>
      {chosen && selected !== null && <SlotDetail slot={chosen} reading={readings[selected]} />}

      {chord !== undefined &&
        selected !== null &&
        base !== null &&
        variations.state === 'given' && (
          <VariationList
            base={base}
            current={chord}
            variations={variations.variations}
            onChoose={(name) => state.replaceChord(selected, name)}
          />
        )}
    </>
  );
}

type PanelProps = {
  tab: Tab;
  state: ProgressionState;
  readings: SlotReading[];
  playback: PlaybackState;
  next: ReturnType<typeof useNextCandidates>;
  branches: BranchesState;
};

/** What the context panel names as chosen: a slot, a branch, or nothing. */
function chosenWords(state: ProgressionState, readings: SlotReading[], branches: BranchesState) {
  const { progression, selected } = state;
  const chosen = selected === null ? undefined : progression.slots[selected];
  if (chosen !== undefined && selected !== null) {
    return `${slotName(chosen, readings[selected])}（${String(selected + 1)} 番目の枠）`;
  }
  const { picked } = branches;
  return picked === null
    ? '枠を選んでいません'
    : `${picked.candidate.chord.name}（${String(picked.at + 1)} 番目の枠の枝）`;
}

/**
 * The context panel, which follows what is chosen: always which slot or
 * branch it is; for a chosen slot its labels and variations, and in the plan
 * tab the candidates to fill it; in the explore tab, the labels of the branch
 * picked and, with no slot chosen, all the candidates for the slot after the
 * last, as the other candidates.
 */
function ContextPanel({ tab, state, readings, playback, next, branches }: PanelProps) {
  const { progression, selected } = state;
  const { beats, candidates, busy } = next;
  // The candidate groups opened or closed, kept here so that a new answer keeps them.
  const [opened, setOpened] = useState<Partial<Record<string, boolean>>>({});
  const { picked } = branches;

  return (
    <aside className="context-panel" aria-labelledby="context-heading">
      <p className="context-where">{TABS.find(({ id }) => id === tab)?.word}のタブ</p>
      <h2 id="context-heading">{chosenWords(state, readings, branches)}</h2>

      <ChosenSlot state={state} readings={readings} />

      {tab === 'explore' && picked !== null && (
        <BranchDetail
          picked={picked}
          auditioning={triedAt(playback.auditioning, picked.at) === picked.candidate.chord.name}
          onAudition={() => branches.audition(picked)}
          onStop={playback.stop}
          onConfirm={() => branches.confirm(picked.place, picked.candidate.chord.name)}
        />
      )}

      <SlotFill
        progression={progression}
        selected={tab === 'plan' ? selected : null}
        playback={playback}
        onFill={state.fillSlot}
      />

      {tab === 'plan' && selected === null && <p>枠を選ぶと、その枠を埋める候補が出ます。</p>}

      {tab === 'explore' && selected === null && candidates.state === 'given' && (
        <CandidateList
          words={{
            heading: `他の候補（末尾に足す、${beats} 拍）`,
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
