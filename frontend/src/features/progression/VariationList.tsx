import type { Variation } from '@/shared/api/client';
import { variationMarks } from './labels';

type Props = {
  /** The chord the list comes from: the slot's base, or its chord when it has none. */
  base: string;
  /** The chord in the slot now. */
  current: string;
  variations: Variation[];
  onChoose: (chord: string) => void;
};

/**
 * The variations of the chosen slot's chord. Choosing one puts it in the
 * slot; the one in the slot now is shown pressed.
 */
export function VariationList({ base, current, variations, onChoose }: Props) {
  return (
    <section className="variations" aria-labelledby="variations-heading">
      <h2 id="variations-heading">{base} の変形</h2>
      <ul className="variation-list" role="list" aria-label={`${base} の変形`}>
        {variations.map((variation) => {
          const name = variation.chord.name;
          // The list names the base as the server spells it; a chord chosen from it, by its name.
          const chosen = base === current ? variation.isInput : name === current;
          return (
            <li key={name}>
              <button type="button" aria-pressed={chosen} onClick={() => onChoose(name)}>
                {name}
              </button>{' '}
              {/* Beside the button, not in its name, as a candidate's marks are. */}
              <span>{variationMarks(variation).join(' ')}</span>
            </li>
          );
        })}
      </ul>
    </section>
  );
}
