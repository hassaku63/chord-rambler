import { useId, useRef } from 'react';
import type { Candidate } from '@/shared/api/client';
import type { LabelRow } from './labels';
import { LabelList } from './LabelList';
import { TERMS, wordFor } from './terms';

/** The words of a list of candidates, which tell one list from another on the screen. */
export type CandidateListWords = {
  heading: string;
  /** The name of a group's list, from the word of its kind. */
  list: (kind: string) => string;
  /** The add button: 足す, この枠に入れる. Its name puts the chord before it (F を足す). */
  add: string;
  /** The name of a candidate's trial button. */
  trial: (chord: string) => string;
};

/** How a list shows its candidates: their short marks and their labels. */
type Shown<C> = { marksOf: (candidate: C) => string[]; rowsOf: (candidate: C) => LabelRow[] };

type ItemProps<C> = Shown<C> & {
  candidate: C;
  words: CandidateListWords;
  auditioning: boolean;
  /** Adding waits for the next answer, so that the same candidate is not added twice. */
  busy: boolean;
  onAdd: () => void;
  onAudition: () => void;
};

/**
 * One candidate: its chord, Roman numeral and short marks, adding and trying,
 * and its labels. The one on trial says so in words, not by colour alone.
 */
function CandidateItem<C extends Candidate>({
  candidate,
  words,
  marksOf,
  rowsOf,
  auditioning,
  busy,
  onAdd,
  onAudition,
}: ItemProps<C>) {
  const name = candidate.chord.name;
  return (
    <li
      className={auditioning ? 'candidate auditioning' : 'candidate'}
      aria-current={auditioning || undefined}
    >
      {auditioning && <span className="candidate-auditioning">試聴中</span>}
      <span className="slot-name">{name}</span>
      <span>{candidate.roman.numeral}</span>
      <span>{marksOf(candidate).join(' ')}</span>
      <span className="candidate-actions">
        <button type="button" aria-label={`${name} を${words.add}`} disabled={busy} onClick={onAdd}>
          {words.add}
        </button>
        <button
          type="button"
          aria-label={words.trial(name)}
          aria-pressed={auditioning}
          onClick={onAudition}
        >
          試聴
        </button>
      </span>
      <details>
        <summary>詳しい札</summary>
        <LabelList rows={rowsOf(candidate)} />
      </details>
    </li>
  );
}

type Props<C> = Shown<C> & {
  words: CandidateListWords;
  groups: { kind: string; candidates: C[] }[];
  /** Whether the answer to the latest progression is still awaited: the groups are the last answer. */
  busy: boolean;
  /** The groups the user opened or closed, by kind; the others keep their first state. */
  opened: Partial<Record<string, boolean>>;
  onToggle: (kind: string, open: boolean) => void;
  /** The chord on trial, or null. */
  auditioning: string | null;
  onAdd: (chord: string) => void;
  onAudition: (chord: string) => void;
  onStop: () => void;
};

/**
 * The candidates for a slot, by group (C1 of #112): the diatonic group open,
 * the others opened to be seen, each in the order the server gives. The last
 * answer stays while the next is awaited, and adding a candidate puts the
 * focus on the heading, so that neither the focus nor the open groups are lost.
 */
export function CandidateList<C extends Candidate>({
  words,
  groups,
  marksOf,
  rowsOf,
  busy,
  opened,
  onToggle,
  auditioning,
  onAdd,
  onAudition,
  onStop,
}: Props<C>) {
  const heading = useRef<HTMLHeadingElement>(null);
  const id = useId();
  return (
    <section className="candidates" aria-labelledby={id} aria-busy={busy}>
      <h2 id={id} ref={heading} tabIndex={-1}>
        {words.heading}
      </h2>
      {/* The line keeps its place when empty, so that the list does not move. */}
      <p className="candidates-wait">{busy ? '新しい候補を待っています。' : ''}</p>
      {groups.map((group) => {
        const kind = wordFor(TERMS.kind, group.kind) ?? group.kind;
        return (
          <details
            key={group.kind}
            open={opened[group.kind] ?? group.kind === 'diatonic'}
            onToggle={(e) => onToggle(group.kind, e.currentTarget.open)}
          >
            <summary>
              {kind}（{group.candidates.length}）
            </summary>
            <ul className="candidate-list" role="list" aria-label={words.list(kind)}>
              {group.candidates.map((candidate) => {
                const name = candidate.chord.name;
                return (
                  <CandidateItem
                    key={name}
                    candidate={candidate}
                    words={words}
                    marksOf={marksOf}
                    rowsOf={rowsOf}
                    auditioning={auditioning === name}
                    busy={busy}
                    onAdd={() => {
                      onAdd(name);
                      heading.current?.focus();
                    }}
                    onAudition={() => (auditioning === name ? onStop() : onAudition(name))}
                  />
                );
              })}
            </ul>
          </details>
        );
      })}
    </section>
  );
}
