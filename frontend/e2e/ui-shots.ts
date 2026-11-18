import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import type { Page } from '@playwright/test';
import { expect, test } from '@playwright/test';
import {
  addByName,
  branches,
  openScreen,
  pickBranch,
  planTab,
  strip,
  tailCandidates,
} from './screen';

// Shots for the ui-review skill (.claude/skills/ui-review/SKILL.md): each
// state at each width, as an image of the whole page and as the
// accessibility tree. Not a check: the name is outside the default testMatch,
// so `make e2e` never runs it; `npm run ui:shots` does. The files go to
// UI_SHOTS_DIR (default ui-review-out/, ignored by git).

const out = process.env.UI_SHOTS_DIR ?? 'ui-review-out';

/** The widths of the layout check (layout.spec.ts). */
const widths = [360, 768, 1280];

/** The states the screen is reviewed in, each reached from a fresh page. */
const states: { name: string; reach: (page: Page) => Promise<void> }[] = [
  {
    name: '1-first',
    reach: async (page) => {
      await openScreen(page);
      await tailCandidates(page);
    },
  },
  {
    name: '2-added',
    reach: async (page) => {
      await openScreen(page);
      await addByName(page, 'C');
      await tailCandidates(page);
    },
  },
  {
    name: '3-chosen',
    reach: async (page) => {
      await openScreen(page);
      await addByName(page, 'C');
      await tailCandidates(page);
      await planTab(page);
      const choose = strip(page).getByRole('button', { name: /^C I\b/ });
      await choose.click();
      await expect(choose).toHaveAttribute('aria-pressed', 'true');
      const fill = page.getByRole('region', { name: /^選んだ枠を埋める候補/ });
      await expect(fill).toHaveAttribute('aria-busy', 'false');
      await expect(page.getByRole('list', { name: /変形/ })).toBeVisible();
    },
  },
  {
    name: '4-explored',
    reach: async (page) => {
      await openScreen(page);
      for (const chord of ['C', 'Am', 'F', 'G7', 'C', 'Am', 'Dm', 'G7']) {
        await addByName(page, chord);
      }
      await tailCandidates(page);
      const choose = page
        .getByRole('list', { name: '本線' })
        .getByRole('button', { name: /^4 G7/ });
      await choose.click();
      await expect(choose).toHaveAttribute('aria-pressed', 'true');
      await expect(page.getByRole('list', { name: /変形/ })).toBeVisible();
    },
  },
  {
    name: '5-branch-picked',
    reach: async (page) => {
      await openScreen(page);
      await addByName(page, 'C');
      await pickBranch(page, 'F IV');
    },
  },
  {
    name: '6-footprint',
    reach: async (page) => {
      await openScreen(page);
      for (const chord of ['C', 'Am']) {
        await addByName(page, chord);
      }
      await pickBranch(page, 'F IV');
      await page.getByRole('button', { name: 'F を確定する', exact: true }).click();
      await branches(page);
      await expect(page.getByRole('toolbar', { name: '探索の足跡' })).toBeVisible();
    },
  },
  {
    name: '7-plan',
    reach: async (page) => {
      await openScreen(page);
      await planTab(page);
      const roles = ['安定', '準備', '高まり', '解決'];
      for (const role of roles) {
        await page.getByRole('button', { name: 'まだ選んでいない枠を足す' }).click();
        await strip(page)
          .getByRole('listitem')
          .last()
          .getByRole('combobox', { name: '役割' })
          .selectOption({ label: role });
      }
      const card = (i: number) =>
        strip(page).getByRole('listitem').nth(i).getByRole('button').first();
      await card(0).click();
      const fill = page.getByRole('region', { name: /^選んだ枠を埋める候補/ });
      await expect(fill).toHaveAttribute('aria-busy', 'false');
      await page.getByRole('button', { name: 'C をこの枠に入れる', exact: true }).click();
      // Filling goes on to the next slot not chosen yet, whose candidates come.
      await expect(card(1)).toHaveAttribute('aria-pressed', 'true');
      await expect(fill).toHaveAttribute('aria-busy', 'false');
    },
  },
];

for (const width of widths) {
  test.describe(`${String(width)} px`, () => {
    test.use({ viewport: { width, height: 800 } });
    for (const state of states) {
      test(state.name, async ({ page }) => {
        await state.reach(page);
        mkdirSync(out, { recursive: true });
        const base = join(out, `${state.name}-${String(width)}`);
        await page.screenshot({ path: `${base}.png`, fullPage: true });
        writeFileSync(`${base}.aria.yaml`, await page.locator('body').ariaSnapshot());
      });
    }
  });
}
