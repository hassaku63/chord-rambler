import type { KeyboardEvent } from 'react';
import { useEffect, useRef, useState } from 'react';
import type { Candidate, CandidateGroups, Progression } from '@/shared/api/client';
import { useNextCandidates } from './useNextCandidates';
import type { PlaybackState } from './usePlayback';
import { triedAt } from './usePlayback';
import type { ProgressionState } from './useProgression';

/** How many branches grow: the interim of docs/product/screen-promise.md, 5, until #203. */
export const BRANCH_COUNT = 5;

/** The branches: the first candidates of the answer, in the order of its groups. */
export function branchesOf(groups: CandidateGroups['groups'], count = BRANCH_COUNT): Candidate[] {
  return groups.flatMap((group) => group.candidates).slice(0, count);
}

/** Where a branch grows: after the last slot, or beside the slot of the footprint. */
export type BranchPlace = 'tail' | 'footprint';

/** A group of branches: the slot they would go in, the chords picked and on trial, and its events. */
export type BranchGroup = {
  at: number;
  candidates: Candidate[];
  picked: string | null;
  auditioning: string | null;
  on: {
    pointerDown: () => void;
    focus: (candidate: Candidate) => void;
    click: (candidate: Candidate) => void;
    confirm: (candidate: Candidate) => void;
    keyDown: (event: KeyboardEvent<HTMLElement>, candidate: Candidate) => void;
  };
};

/** The branch picked: its candidate, where it grows, and the slot it would go in. */
export type PickedBranch = { candidate: Candidate; place: BranchPlace; at: number };

type Picked = { chord: string; place: BranchPlace; for: Progression };

/** Moves the focus to another branch of the toolbar with the arrows (WAI-ARIA toolbar). */
function move(event: KeyboardEvent<HTMLElement>): void {
  const step = { ArrowDown: 1, ArrowRight: 1, ArrowUp: -1, ArrowLeft: -1 }[event.key];
  const toolbar = event.currentTarget.closest('[role="toolbar"]');
  const buttons = Array.from(toolbar?.querySelectorAll('button') ?? []);
  if (step !== undefined) {
    event.preventDefault();
    const at = buttons.indexOf(event.currentTarget as HTMLButtonElement) + step;
    buttons[(at + buttons.length) % buttons.length]?.focus();
  }
}

/**
 * The branches of the explore tab (docs/product/screen-promise.md, 3 and 4):
 * touching one picks it and plays the slot before with it; Enter, a double
 * click or the context panel puts it in, and the focus goes to the first of
 * the new branches. The other branches of the step stay as the footprint.
 */
export function useBranches(state: ProgressionState, playback: PlaybackState) {
  const { progression, footprint } = state;
  // Asked for even while hidden, so that the last answer is there when the slot is let go.
  const next = useNextCandidates(progression);
  const { beats, candidates, busy } = next;
  const [picked, setPicked] = useState<Picked | null>(null);
  const tailRef = useRef<HTMLDivElement>(null);
  // Whether the focus goes to the first new branch when the answer comes, and
  // whether the next focus comes from a pointer, whose click plays instead.
  const pending = useRef({ focus: false, pointer: false });

  const tail = candidates.state === 'given' ? branchesOf(candidates.groups) : [];
  const placed = footprint && progression.slots[footprint.at]?.chord;
  const unchosen = footprint?.branches.filter((c) => c.chord.name !== placed) ?? [];
  const places = {
    tail: { at: progression.slots.length, candidates: tail },
    footprint: footprint && unchosen.length > 0 ? { at: footprint.at, candidates: unchosen } : null,
  };
  // A pick holds only for the progression it was made in.
  const where = picked?.for === progression ? places[picked.place] : null;
  const found = where?.candidates.find((c) => c.chord.name === picked?.chord);
  const pickedBranch: PickedBranch | null =
    picked && where && found ? { candidate: found, place: picked.place, at: where.at } : null;

  const touch = (place: BranchPlace, candidate: Candidate) => {
    const group = places[place];
    if (group === null || (place === 'tail' && busy)) {
      return;
    }
    pending.current.focus = false;
    state.select(null);
    const chord = candidate.chord.name;
    setPicked({ chord, place, for: progression });
    if (place === 'tail') {
      playback.audition(chord, beats);
    } else {
      playback.auditionIn(group.at, chord);
    }
  };

  const confirm = (place: BranchPlace, chord: string) => {
    if (place === 'tail' && !busy) {
      state.addCandidate(chord, beats);
      state.setFootprint({ at: progression.slots.length, branches: tail });
    } else if (place === 'footprint' && footprint !== null) {
      state.fillSlot(footprint.at, chord);
      state.setFootprint(footprint);
    } else {
      return;
    }
    setPicked(null);
    pending.current.focus = true;
  };

  useEffect(() => {
    if (!pending.current.focus || busy) {
      return;
    }
    pending.current.focus = false;
    const first = tailRef.current?.querySelector('button');
    // Focusing the first branch plays it; one that has the focus already is given it again.
    if (first === document.activeElement) {
      first?.blur();
    }
    first?.focus();
  }, [busy, candidates]);

  const on = (place: BranchPlace): BranchGroup['on'] => ({
    pointerDown: () => {
      pending.current.pointer = true;
    },
    focus: (candidate) => {
      if (!pending.current.pointer) {
        touch(place, candidate);
      }
      pending.current.pointer = false;
    },
    click: (candidate) => {
      pending.current.pointer = false;
      touch(place, candidate);
    },
    confirm: (candidate) => confirm(place, candidate.chord.name),
    keyDown: (event, candidate) => {
      if (event.key === 'Enter') {
        // No click after it: the branch is put in, not picked again.
        event.preventDefault();
        confirm(place, candidate.chord.name);
      }
      move(event);
    },
  });

  const group = (place: BranchPlace, at: { at: number; candidates: Candidate[] }): BranchGroup => ({
    ...at,
    picked: pickedBranch?.place === place ? pickedBranch.candidate.chord.name : null,
    auditioning: triedAt(playback.auditioning, at.at),
    on: on(place),
  });

  return {
    next,
    tail: group('tail', places.tail),
    footprint: places.footprint && group('footprint', places.footprint),
    busy,
    picked: pickedBranch,
    tailRef,
    confirm,
    unpick: () => setPicked(null),
  };
}

export type BranchesState = ReturnType<typeof useBranches>;
