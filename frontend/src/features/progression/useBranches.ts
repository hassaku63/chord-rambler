import type { KeyboardEvent, RefObject } from 'react';
import { useEffect, useRef, useState } from 'react';
import type { Candidate, CandidateGroups, Progression } from '@/shared/api/client';
import type { CandidatesState } from './useNextCandidates';
import type { PlaybackState } from './usePlayback';
import { triedAt } from './usePlayback';
import type { Footprint, ProgressionState } from './useProgression';

/**
 * How many branches grow from the last slot: the interim of
 * docs/product/screen-promise.md, 5, until #203 decides the rule.
 */
export const BRANCH_COUNT = 5;

/** The branches: the first candidates of the answer, in the order of its groups. */
export function branchesOf(groups: CandidateGroups['groups'], count = BRANCH_COUNT): Candidate[] {
  return groups.flatMap((group) => group.candidates).slice(0, count);
}

/** Where a branch grows: after the last slot, or beside the slot of the footprint. */
export type BranchPlace = 'tail' | 'footprint';

/** The branch picked: its candidate, where it grows, and the slot it would go in. */
export type PickedBranch = { candidate: Candidate; place: BranchPlace; at: number };

/** What a group of branches does when touched, confirmed or moved in. */
export type BranchHandlers = {
  onPointerDown: () => void;
  onFocus: (candidate: Candidate) => void;
  onClick: (candidate: Candidate) => void;
  onDoubleClick: (candidate: Candidate) => void;
  onKeyDown: (event: KeyboardEvent<HTMLElement>, candidate: Candidate) => void;
};

/** A group of branches as the canvas shows it. */
export type BranchGroup = {
  /** The slot the branches would go in. */
  at: number;
  candidates: Candidate[];
  /** The chord of the branch picked in this group, or null. */
  picked: string | null;
  /** The chord of the branch on trial in this group, or null. */
  auditioning: string | null;
  handlers: BranchHandlers;
};

export type BranchesState = {
  /** The branches after the last slot. While busy, those of the last answer, which do nothing. */
  tail: BranchGroup;
  /** The branches of the step before that were not put in, beside the slot; null when none. */
  footprint: BranchGroup | null;
  busy: boolean;
  picked: PickedBranch | null;
  /** The group of the branches after the last slot, where the focus goes after a confirm. */
  tailRef: RefObject<HTMLDivElement | null>;
  audition: (picked: PickedBranch) => void;
  /** Puts the branch in: after the last slot, or in the slot of the footprint. */
  confirm: (place: BranchPlace, chord: string) => void;
  unpick: () => void;
};

type Picked = { chord: string; place: BranchPlace; for: Progression };

type Next = { beats: number; candidates: CandidatesState; busy: boolean };

/** Where the branches of each group would go, and which; no footprint when none is left out. */
function placesOf(
  progression: Progression,
  footprint: Footprint | null,
  candidates: CandidatesState,
) {
  const tail = candidates.state === 'given' ? branchesOf(candidates.groups) : [];
  const placed = footprint && progression.slots[footprint.at]?.chord;
  const unchosen = footprint?.branches.filter((c) => c.chord.name !== placed) ?? [];
  return {
    tail: { at: progression.slots.length, candidates: tail },
    footprint: footprint && unchosen.length > 0 ? { at: footprint.at, candidates: unchosen } : null,
  };
}

/** The branch picked, while the progression it was picked in holds and the branch is there. */
function pickOf(
  picked: Picked | null,
  progression: Progression,
  places: ReturnType<typeof placesOf>,
): PickedBranch | null {
  const where = picked?.for === progression ? places[picked.place] : null;
  const found = where?.candidates.find((c) => c.chord.name === picked?.chord);
  return picked && where && found ? { candidate: found, place: picked.place, at: where.at } : null;
}

/** The keys that move between the branches of a group, and where each goes. */
const MOVES: Record<string, (index: number, count: number) => number> = {
  ArrowDown: (i, n) => (i + 1) % n,
  ArrowRight: (i, n) => (i + 1) % n,
  ArrowUp: (i, n) => (i - 1 + n) % n,
  ArrowLeft: (i, n) => (i - 1 + n) % n,
  Home: () => 0,
  End: (_, n) => n - 1,
};

/** Moves the focus to another button of the group, as the key says. */
function move(event: KeyboardEvent<HTMLElement>): void {
  const go = MOVES[event.key];
  const group = event.currentTarget.closest('[role="toolbar"]');
  if (go === undefined || group === null) {
    return;
  }
  event.preventDefault();
  const buttons = Array.from(group.querySelectorAll('button'));
  buttons[go(buttons.indexOf(event.currentTarget as HTMLButtonElement), buttons.length)]?.focus();
}

/**
 * The branches of the explore tab (docs/product/screen-promise.md, 3 and 4):
 * touching one picks it and plays the slot before with it; Enter, a double
 * click or the context panel puts it in; the arrows move between them. After
 * a branch is put in, the branches of that step stay beside it as the
 * footprint, and the focus goes to the first of the new branches.
 */
export function useBranches(
  state: ProgressionState,
  next: Next,
  playback: PlaybackState,
): BranchesState {
  const { progression, footprint } = state;
  const { beats, candidates, busy } = next;
  const [picked, setPicked] = useState<Picked | null>(null);
  const tailRef = useRef<HTMLDivElement>(null);
  // Whether the focus goes to the first new branch when the answer comes, and
  // whether the next focus comes from a pointer, whose click plays instead.
  const pending = useRef({ focus: false, pointer: false });
  const places = placesOf(progression, footprint, candidates);
  const tail = places.tail.candidates;
  const pickedBranch = pickOf(picked, progression, places);

  const audition = ({ candidate, place, at }: PickedBranch) => {
    const chord = candidate.chord.name;
    if (place === 'tail') {
      playback.audition(chord, beats);
    } else {
      playback.auditionIn(at, chord);
    }
  };

  const touch = (place: BranchPlace, candidate: Candidate) => {
    const group = places[place];
    if (group === null || (place === 'tail' && busy)) {
      return;
    }
    state.select(null);
    setPicked({ chord: candidate.chord.name, place, for: progression });
    audition({ candidate, place, at: group.at });
  };

  const confirm = (place: BranchPlace, chord: string) => {
    if (place === 'tail') {
      if (busy) {
        return;
      }
      state.addCandidate(chord, beats, tail);
    } else {
      state.replaceFromFootprint(chord);
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

  const handlers = (place: BranchPlace): BranchHandlers => ({
    onPointerDown: () => {
      pending.current.pointer = true;
    },
    onFocus: (candidate) => {
      if (!pending.current.pointer) {
        touch(place, candidate);
      }
      pending.current.pointer = false;
    },
    onClick: (candidate) => {
      pending.current.pointer = false;
      touch(place, candidate);
    },
    onDoubleClick: (candidate) => confirm(place, candidate.chord.name),
    onKeyDown: (event, candidate) => {
      if (event.key === 'Enter') {
        // No click after it: the branch is put in, not picked again.
        event.preventDefault();
        confirm(place, candidate.chord.name);
      } else {
        move(event);
      }
    },
  });

  const group = (
    place: BranchPlace,
    where: { at: number; candidates: Candidate[] },
  ): BranchGroup => ({
    ...where,
    picked: pickedBranch?.place === place ? pickedBranch.candidate.chord.name : null,
    auditioning: triedAt(playback.auditioning, where.at),
    handlers: handlers(place),
  });

  return {
    tail: group('tail', places.tail),
    footprint: places.footprint && group('footprint', places.footprint),
    busy,
    picked: pickedBranch,
    tailRef,
    audition,
    confirm,
    unpick: () => setPicked(null),
  };
}
