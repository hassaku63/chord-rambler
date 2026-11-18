import { candidateMarks, candidateRows, footprintRows } from './labels';
import { LabelList } from './LabelList';
import type { PickedBranch } from './useBranches';

/**
 * The branch picked, in the context panel: its short marks and labels, in
 * place of the words of #204 until they are decided
 * (docs/product/screen-promise.md, 5), and putting it in.
 */
export function BranchDetail({
  picked,
  onConfirm,
}: {
  picked: PickedBranch;
  onConfirm: () => void;
}) {
  const { candidate, place } = picked;
  const name = candidate.chord.name;
  return (
    <section className="branch-detail" aria-labelledby="branch-detail-heading">
      <h3 id="branch-detail-heading">
        {name}（{candidate.roman.numeral}） の札
      </h3>
      <p>{candidateMarks(candidate).join(' ')}</p>
      <LabelList rows={place === 'tail' ? candidateRows(candidate) : footprintRows(candidate)} />
      <button type="button" aria-label={`${name} を確定する`} onClick={onConfirm}>
        確定する
      </button>
    </section>
  );
}
