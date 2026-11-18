import type { ReactNode } from 'react';
import { Fragment } from 'react';
import type { Slot, SlotReading } from '@/shared/api/client';
import { edgeWords, numeralOf } from './labels';
import { slotName } from './SlotCard';

type Props = {
  slots: Slot[];
  readings: SlotReading[];
  /** The slot chosen, counted from 0, or null. */
  selected: number | null;
  /** The slot sounding now, or null. */
  sounding: number | null;
  onSelect: (index: number) => void;
  /** The branches after the last slot, in the explore tab. */
  branches?: ReactNode;
  /** The footprint and the slot it is beside, in the explore tab. */
  footprint?: { at: number; element: ReactNode } | null;
};

/** The judgments on a line, each a small label of its own, so that none breaks into the next. */
function Edge({ words }: { words: string[] }) {
  return (
    <span className="edge">
      <span className="edge-words">
        {words.map((word, k) => (
          <Fragment key={word}>
            {k > 0 && ' '}
            <span className="edge-word">{word}</span>
          </Fragment>
        ))}
      </span>
      <span className="edge-arrow" aria-hidden="true">
        →
      </span>
    </span>
  );
}

/**
 * The flow of the progression: a node of the main line for each slot, with its
 * place, chord and Roman numeral, and between the nodes the judgments of the
 * reading in words, not by colour alone. Choosing a node chooses its slot, as
 * the strip does. In the explore tab, the branches grow from the last node,
 * or from a start point when there is none, and the footprint lies beside
 * the node it was left at.
 */
export function Canvas({
  slots,
  readings,
  selected,
  sounding,
  onSelect,
  branches,
  footprint,
}: Props) {
  return (
    <section className="canvas" aria-labelledby="canvas-heading">
      <h2 id="canvas-heading" className="area-heading">
        流れ
      </h2>
      <div className="canvas-flow">
        {slots.length === 0 ? (
          branches ? (
            <span className="canvas-start">始まり</span>
          ) : (
            <p className="canvas-empty">枠を足すと、和音のつながりがここに出ます。</p>
          )
        ) : (
          <ol className="canvas-line" role="list" aria-label="本線">
            {slots.map((slot, i) => {
              const reading = readings[i];
              const numeral = numeralOf(reading);
              return (
                // Slots have no identity of their own; the position is the slot.
                <li key={i} className="node-item" aria-current={sounding === i || undefined}>
                  {i > 0 && <Edge words={edgeWords(readings[i - 1], reading)} />}
                  <span className="node-wrap">
                    <span className="node-sounding">{sounding === i && '鳴っています'}</span>
                    <button
                      type="button"
                      className={slot.chord === undefined ? 'node empty' : 'node'}
                      aria-pressed={selected === i}
                      onClick={() => onSelect(i)}
                    >
                      {/* The spaces keep the visible text inside the name (WCAG 2.5.3). */}
                      <span className="node-index">{i + 1}</span>{' '}
                      <span className="node-name">{slotName(slot, reading)}</span>
                      {numeral !== undefined && (
                        <>
                          {' '}
                          <span className="node-numeral">{numeral}</span>
                        </>
                      )}
                    </button>
                    {footprint?.at === i && footprint.element}
                  </span>
                </li>
              );
            })}
          </ol>
        )}
        {branches}
      </div>
    </section>
  );
}
