import { expect, test } from '@playwright/test';
import { addByName, openScreen, strip, tailCandidates } from './screen';

// No stand-in for the AudioContext: headless Chromium makes one without a
// sound device, so the screen plays as it does for a user. The first test
// checks that, so that a browser without it fails here, plainly, and not in
// the middle of the scenario.
test('headless Chromium can make and resume an AudioContext', async ({ page }) => {
  await openScreen(page);
  const state = await page.evaluate(async () => {
    const context = new AudioContext();
    await context.resume();
    const resumed = context.state;
    await context.close();
    return resumed;
  });
  expect(state).toBe('running');
});

test('a progression built from a typed chord and a candidate plays slot by slot', async ({
  page,
}) => {
  const errors: Error[] = [];
  page.on('pageerror', (error) => errors.push(error));
  await openScreen(page);
  const slots = strip(page).getByRole('listitem');

  await addByName(page, 'C');
  await expect(slots).toHaveCount(1);
  await expect(slots.nth(0)).toContainText('C');
  await expect(slots.nth(0)).toContainText('I');

  const candidates = await tailCandidates(page);
  await candidates.getByRole('button', { name: 'G7 を足す', exact: true }).click();
  await expect(slots).toHaveCount(2);
  await expect(slots.nth(1)).toContainText('G7');
  await expect(slots.nth(1)).toContainText('V7');

  await page.getByRole('button', { name: '進行を鳴らす', exact: true }).click();
  await expect(slots.nth(0)).toHaveAttribute('aria-current', 'true');
  await expect(slots.nth(0)).toContainText('鳴っています');
  // At 90 BPM a slot of 4 beats lasts about 2.7 seconds; the mark then moves on.
  await expect(slots.nth(1)).toHaveAttribute('aria-current', 'true');
  await expect(slots.nth(1)).toContainText('鳴っています');
  await expect(slots.nth(0)).not.toHaveAttribute('aria-current');

  expect(errors).toEqual([]);
});
