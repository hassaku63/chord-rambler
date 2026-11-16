import type { Slot, SlotReading } from '@/shared/api/client';
import { labelRows, numeralOf } from './labels';
import { LabelList } from './LabelList';
import { slotName } from './SlotCard';

type Props = { slot: Slot; reading: SlotReading | undefined };

/** The labels of the chosen slot, each with a mark of the kind of knowledge it rests on. */
export function SlotDetail({ slot, reading }: Props) {
  const name = slotName(slot, reading);
  const rows = reading ? labelRows(reading) : [];
  const numeral = numeralOf(reading);
  return (
    <section className="slot-detail" aria-labelledby="slot-detail-heading">
      <h2 id="slot-detail-heading">
        {name}
        {numeral !== undefined && `（${numeral}）`} の札
      </h2>
      {rows.length === 0 ? <p>この枠に札はありません。</p> : <LabelList rows={rows} />}
    </section>
  );
}
