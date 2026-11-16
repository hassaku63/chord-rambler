import type { Locator } from '@playwright/test';
import { expect, test } from '@playwright/test';
import { addByName, openScreen, tailCandidates } from './screen';

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
