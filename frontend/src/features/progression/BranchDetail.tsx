import { candidateMarks, candidateRows, footprintRows } from './labels';
import { LabelList } from './LabelList';
import type { PickedBranch } from './useBranches';

type Props = {
  picked: PickedBranch;
  /** Whether the branch is on trial now. */
  auditioning: boolean;
  onAudition: () => void;
  onStop: () => void;
  onConfirm: () => void;
};

/**
 * The branch picked, in the context panel: its short marks and labels, in
 * place of the words of #204 until they are decided
 * (docs/product/screen-promise.md, 5), trying it again, and putting it in.
 */
export function BranchDetail({ picked, auditioning, onAudition, onStop, onConfirm }: Props) {
  const { candidate, place } = picked;
  const name = candidate.chord.name;
  const rows = place === 'tail' ? candidateRows(candidate) : footprintRows(candidate);
  return (
    <section className="branch-detail" aria-labelledby="branch-detail-heading">
      <h3 id="branch-detail-heading">
        {name}（{candidate.roman.numeral}） の札
      </h3>
      <p>{candidateMarks(candidate).join(' ')}</p>
      <LabelList rows={rows} />
      <p className="candidate-actions">
        <button type="button" aria-label={`${name} を確定する`} onClick={onConfirm}>
          確定する
        </button>{' '}
        <button
          type="button"
          aria-label={`${name} をもう一度試聴する`}
          aria-pressed={auditioning}
          onClick={auditioning ? onStop : onAudition}
        >
          もう一度試聴
        </button>
      </p>
    </section>
  );
}
