import type { Slot, SlotReading } from '@/shared/api/client';
import { labelRows } from './labels';
import { slotName } from './SlotCard';

type Props = { slot: Slot; reading: SlotReading | undefined };

/** The labels of the chosen slot, each with a mark of the kind of knowledge it rests on. */
export function SlotDetail({ slot, reading }: Props) {
  const name = slotName(slot, reading);
  const rows = reading ? labelRows(reading) : [];
  return (
    <section className="slot-detail" aria-labelledby="slot-detail-heading">
      <h2 id="slot-detail-heading">
        {name}
        {reading?.roman && `（${reading.roman.numeral}）`} の札
      </h2>
      {rows.length === 0 ? (
        <p>この枠に札はありません。</p>
      ) : (
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
      )}
    </section>
  );
}
