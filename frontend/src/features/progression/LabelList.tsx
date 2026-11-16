import type { LabelRow } from './labels';

/** Labels as terms and words, each word with a mark of the kind of knowledge it rests on. */
export function LabelList({ rows }: { rows: LabelRow[] }) {
  return (
    <dl>
      {rows.map((row) => (
        <div key={row.term} className="label-row">
          <dt>{row.term}</dt>
          <dd>
            {row.word} <span className="basis">〔{row.basis}〕</span>
          </dd>
        </div>
      ))}
    </dl>
  );
}
