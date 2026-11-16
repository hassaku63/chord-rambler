import { useRef } from 'react';
import type { Candidate, CandidateGroups } from '@/shared/api/client';
import { candidateMarks, candidateRows } from './labels';
import { LabelList } from './LabelList';
import { TERMS, wordFor } from './terms';

type ItemProps = {
  candidate: Candidate;
  auditioning: boolean;
  onAdd: () => void;
  onAudition: () => void;
};

/**
 * One candidate: its chord, Roman numeral and short marks, adding and trying,
 * and its labels. The one on trial says so in words, not by colour alone.
 */
function CandidateItem({ candidate, auditioning, onAdd, onAudition }: ItemProps) {
  const name = candidate.chord.name;
  return (
    <li
      className={auditioning ? 'candidate auditioning' : 'candidate'}
      aria-current={auditioning || undefined}
    >
      {auditioning && <span className="candidate-auditioning">試聴中</span>}
      <span className="slot-name">{name}</span>
      <span>{candidate.roman.numeral}</span>
      <span>{candidateMarks(candidate).join(' ')}</span>
      <span className="candidate-actions">
        <button type="button" aria-label={`${name} を足す`} onClick={onAdd}>
          足す
        </button>
        <button
          type="button"
          aria-label={`${name} を試聴する`}
          aria-pressed={auditioning}
          onClick={onAudition}
        >
          試聴
        </button>
      </span>
      <details>
        <summary>詳しい札</summary>
        <LabelList rows={candidateRows(candidate)} />
      </details>
    </li>
  );
}

type Props = {
  groups: CandidateGroups['groups'];
  /** The beats of the slot a candidate goes in. */
  beats: number;
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
 * The candidates for the slot after the last, by group (C1 of #112): the
 * diatonic group open, the others opened to be seen. The last answer stays
 * while the next is awaited, and adding a candidate puts the focus on the
 * heading, so that neither the focus nor the open groups are lost.
 */
export function CandidateList({
  groups,
  beats,
  busy,
  opened,
  onToggle,
  auditioning,
  onAdd,
  onAudition,
  onStop,
}: Props) {
  const heading = useRef<HTMLHeadingElement>(null);
  return (
    <section className="candidates" aria-labelledby="candidates-heading" aria-busy={busy}>
      <h2 id="candidates-heading" ref={heading} tabIndex={-1}>
        末尾に足す候補（{beats} 拍）
      </h2>
      {busy && <p>新しい候補を待っています。</p>}
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
            <ul className="candidate-list" role="list" aria-label={`${kind}の候補`}>
              {group.candidates.map((candidate) => {
                const name = candidate.chord.name;
                return (
                  <CandidateItem
                    key={name}
                    candidate={candidate}
                    auditioning={auditioning === name}
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
