import type { RefObject } from 'react';
import { useId } from 'react';
import type { BranchGroup } from './useBranches';

type Props = {
  group: BranchGroup;
  /** The name of the group, which the footprint also shows in words. */
  label: string;
  footprint?: boolean;
  /** While the branches of the latest progression are awaited, those shown do nothing. */
  busy?: boolean;
  toolbarRef?: RefObject<HTMLDivElement | null>;
};

/**
 * A group of branches, each a button with its chord and Roman numeral: one
 * Tab stop, the arrows moving between them (the toolbar pattern of WAI-ARIA).
 * The one picked is pressed; the one on trial says so in words and by a
 * dashed line, not by motion or colour alone.
 */
export function Branches({ group, label, footprint = false, busy = false, toolbarRef }: Props) {
  const id = useId();
  const { candidates, picked, auditioning, handlers } = group;
  // The Tab stop: the branch picked, or the first.
  const stop = candidates.some((c) => c.chord.name === picked) ? picked : candidates[0]?.chord.name;
  return (
    <div className={footprint ? 'branches footprint' : 'branches'}>
      {footprint && (
        <span id={id} className="footprint-caption">
          {label}
        </span>
      )}
      <div
        ref={toolbarRef}
        role="toolbar"
        aria-orientation="vertical"
        {...(footprint ? { 'aria-labelledby': id } : { 'aria-label': label })}
        aria-busy={busy}
        className="branch-list"
      >
        {candidates.map((candidate) => {
          const name = candidate.chord.name;
          const trial = auditioning === name;
          return (
            <span key={name} className="branch-item">
              <button
                type="button"
                className="branch"
                aria-pressed={picked === name}
                aria-current={trial || undefined}
                tabIndex={name === stop ? 0 : -1}
                onPointerDown={handlers.onPointerDown}
                onFocus={() => handlers.onFocus(candidate)}
                onClick={() => handlers.onClick(candidate)}
                onDoubleClick={() => handlers.onDoubleClick(candidate)}
                onKeyDown={(e) => handlers.onKeyDown(e, candidate)}
              >
                <span className="branch-name">{name}</span>{' '}
                <span className="branch-numeral">{candidate.roman.numeral}</span>
              </button>
              {trial && <span className="branch-auditioning">試聴中</span>}
            </span>
          );
        })}
      </div>
    </div>
  );
}
