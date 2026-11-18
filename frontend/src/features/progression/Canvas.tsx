import type { Slot, SlotReading } from '@/shared/api/client';
import { Fragment } from 'react';
import { edgeWords, numeralOf } from './labels';
import { slotName } from './SlotCard';
import { Branches } from './Branches';
import type { BranchesState } from './useBranches';

type Props = {
  slots: Slot[];
  readings: SlotReading[];
  /** The slot chosen, counted from 0, or null. */
  selected: number | null;
  /** The slot sounding now, or null. */
  sounding: number | null;
  onSelect: (index: number) => void;
  /** The branches and the footprint, in the explore tab. */
  branches: BranchesState | null;
};

/**
 * The flow of the progression: a node of the main line for each slot, with its
 * place, chord and Roman numeral, and between the nodes the judgments of the
 * reading in words, not by colour alone. Choosing a node chooses its slot, as
 * the strip does. In the explore tab, the branches grow from the last node,
 * or from a start point, and the footprint lies under the node put in.
 */
export function Canvas({ slots, readings, selected, sounding, onSelect, branches }: Props) {
  return (
    <section className="canvas" aria-labelledby="canvas-heading">
      <h2 id="canvas-heading" className="area-heading">
        流れ
      </h2>
      {slots.length === 0 && branches ? (
        <span className="canvas-start">始まり</span>
      ) : slots.length === 0 ? (
        <p className="canvas-empty">枠を足すと、和音のつながりがここに出ます。</p>
      ) : (
        <ol className="canvas-line" role="list" aria-label="本線">
          {slots.map((slot, i) => {
            const reading = readings[i];
            const numeral = numeralOf(reading);
            return (
              // Slots have no identity of their own; the position is the slot.
              <li key={i} className="node-item" aria-current={sounding === i || undefined}>
                {i > 0 && (
                  <span className="edge">
                    <span className="edge-words">
                      {edgeWords(readings[i - 1], reading).map((word, k) => (
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
                )}
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
                  {branches?.footprint?.at === i && (
                    <Branches group={branches.footprint} label="探索の足跡" footprint />
                  )}
                </span>
              </li>
            );
          })}
        </ol>
      )}
      {branches && branches.tail.candidates.length > 0 && (
        <Branches
          group={branches.tail}
          label="枝"
          busy={branches.busy}
          toolbarRef={branches.tailRef}
        />
      )}
    </section>
  );
}
