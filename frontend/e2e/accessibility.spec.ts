import AxeBuilder from '@axe-core/playwright';
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

// WCAG 2.0, 2.1 and 2.2 at levels A and AA, as axe-core tags them. axe-core
// 4.14 has no wcag22a tag: none of its rules checks a level A criterion new
// in 2.2.
const WCAG_A_AA = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'];

/** The violations, each with the elements it was found on, so that a failure says where. */
async function violations(page: Page) {
  const results = await new AxeBuilder({ page }).withTags(WCAG_A_AA).analyze();
  return results.violations.map((v) => ({
    rule: v.id,
    help: v.help,
    targets: v.nodes.map((node) => node.target.join(' ')),
  }));
}

test('the empty progression has no violations', async ({ page }) => {
  await openScreen(page);
  await tailCandidates(page);
  expect(await violations(page)).toEqual([]);
});

test('a chosen slot in the plan tab, with judgments on the lines, has no violations', async ({
  page,
}) => {
  await openScreen(page);
  for (const chord of ['C', 'F', 'G7', 'C']) {
    await addByName(page, chord);
  }
  await expect(page.getByRole('list', { name: '本線' })).toContainText('解決 正格終止');
  await planTab(page);
  const choose = strip(page)
    .getByRole('button', { name: /^C I\b/ })
    .first();
  await choose.click();
  await expect(choose).toHaveAttribute('aria-pressed', 'true');
  const fill = page.getByRole('region', { name: /^選んだ枠を埋める候補/ });
  await expect(fill).toHaveAttribute('aria-busy', 'false');
  await expect(page.getByRole('list', { name: /変形/ })).toBeVisible();
  expect(await violations(page)).toEqual([]);
});

test('a picked branch and the footprint have no violations', async ({ page }) => {
  await openScreen(page);
  await (await branches(page)).getByRole('button', { name: 'C I', exact: true }).press('Enter');
  await pickBranch(page, 'F IV');
  await expect(page.getByRole('toolbar', { name: '探索の足跡' })).toBeVisible();
  expect(await violations(page)).toEqual([]);
});

test('the candidates opened have no violations', async ({ page }) => {
  await openScreen(page);
  await addByName(page, 'C');
  const candidates = await tailCandidates(page);
  // A summary has no role of its own; its text is what a user finds it by.
  for (const group of ['セカンダリー・ドミナント', '借用和音']) {
    await candidates.getByText(new RegExp(`^${group}（\\d+）$`)).click();
  }
  const first = candidates.getByRole('listitem').first();
  await first.getByText('詳しい札', { exact: true }).click();
  await expect(first.getByRole('group')).toHaveAttribute('open', '');
  await expect(candidates.getByRole('button', { name: 'E7 を足す', exact: true })).toBeVisible();
  await expect(candidates.getByRole('button', { name: 'Fm を足す', exact: true })).toBeVisible();
  expect(await violations(page)).toEqual([]);
});
