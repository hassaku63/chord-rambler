import type { Locator } from '@playwright/test';
import { expect, test } from '@playwright/test';
import { addByName, openScreen, strip, tailCandidates } from './screen';

/** Fails unless the element is visible and lies within the width of the page. */
async function expectWithin(control: Locator, width: number) {
  await control.scrollIntoViewIfNeeded();
  await expect(control).toBeVisible();
  const box = await control.boundingBox();
  expect(box).not.toBeNull();
  if (box !== null) {
    expect(box.x).toBeGreaterThanOrEqual(0);
    expect(box.x + box.width).toBeLessThanOrEqual(width);
  }
}

for (const width of [360, 768, 1280]) {
  test(`at ${width} px the page does not scroll sideways and the main controls show`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 800 });
    await openScreen(page);
    await addByName(page, 'C');
    const candidates = await tailCandidates(page);

    const sideways = await page.evaluate(() => {
      const root = document.documentElement;
      return { scroll: root.scrollWidth, client: root.clientWidth };
    });
    expect(sideways.scroll).toBeLessThanOrEqual(sideways.client);

    await expectWithin(page.getByRole('combobox', { name: '調' }), width);
    await expectWithin(page.getByRole('textbox', { name: '和音名' }), width);
    await expectWithin(page.getByRole('group', { name: '再生' }), width);
    await expectWithin(page.getByRole('button', { name: '進行を鳴らす', exact: true }), width);
    await expectWithin(candidates.getByRole('button', { name: 'G7 を足す', exact: true }), width);
  });
}

test('at 1280 by 800 the canvas, the strip and the context panel show together', async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await openScreen(page);
  for (const chord of ['C', 'Am', 'F', 'G7']) {
    await addByName(page, chord);
  }
  await tailCandidates(page);
  await strip(page)
    .getByRole('button', { name: /^Am vi\b/ })
    .click();
  await expect(page.getByRole('list', { name: /変形/ })).toBeVisible();

  const height = await page.evaluate(() => document.documentElement.scrollHeight);
  expect(height).toBeLessThanOrEqual(800);
  await expect(page.getByRole('region', { name: '流れ' })).toBeInViewport({ ratio: 1 });
  await expect(page.getByRole('region', { name: '時間と再生' })).toBeInViewport({ ratio: 1 });
  await expect(page.getByRole('complementary')).toBeInViewport({ ratio: 1 });
});
