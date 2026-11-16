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

/** One candidate: its chord, Roman numeral and short marks, adding and trying, and its labels. */
function CandidateItem({ candidate, auditioning, onAdd, onAudition }: ItemProps) {
  const name = candidate.chord.name;
  return (
    <li className="candidate">
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
  /** The chord on trial, or null. */
  auditioning: string | null;
  onAdd: (chord: string) => void;
  onAudition: (chord: string) => void;
  onStop: () => void;
};

/**
 * The candidates for the slot after the last, by group (C1 of #112): the
 * diatonic group open, the others opened to be seen.
 */
export function CandidateList({ groups, beats, auditioning, onAdd, onAudition, onStop }: Props) {
  return (
    <section className="candidates" aria-labelledby="candidates-heading">
      <h2 id="candidates-heading">末尾に足す候補（{beats} 拍）</h2>
      {groups.map((group) => {
        const kind = wordFor(TERMS.kind, group.kind) ?? group.kind;
        return (
          <details key={group.kind} open={group.kind === 'diatonic'}>
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
                    onAdd={() => onAdd(name)}
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
