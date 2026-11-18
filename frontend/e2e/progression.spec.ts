import { expect, test } from '@playwright/test';
import { branches, openScreen, pickBranch, strip } from './screen';

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

test('a progression built from branches plays slot by slot', async ({ page }) => {
  const errors: Error[] = [];
  page.on('pageerror', (error) => errors.push(error));
  await openScreen(page);
  const slots = strip(page).getByRole('listitem');

  // From the start point: focus alone picks nothing; Enter puts the branch in, and the focus
  // goes to the new first branch.
  const c = (await branches(page)).getByRole('button', { name: 'C I', exact: true });
  await c.focus();
  await expect(c).toHaveAttribute('aria-pressed', 'false');
  await c.press('Enter');
  await expect(slots).toHaveCount(1);
  await expect(slots.nth(0)).toContainText('C');
  await expect((await branches(page)).getByRole('button').first()).toBeFocused();

  // Pressing a branch picks it; the context panel puts it in.
  await pickBranch(page, 'F IV');
  await expect(page.getByRole('complementary').getByRole('heading', { level: 2 })).toHaveText(
    'F（2 番目の枠の枝）',
  );
  await page.getByRole('button', { name: 'F を確定する', exact: true }).click();
  await expect(slots).toHaveCount(2);
  await expect(slots.nth(1)).toContainText('IV');

  // A branch of the footprint takes the place of F, by a double click.
  const footprint = page.getByRole('toolbar', { name: '探索の足跡', exact: true });
  await footprint.getByRole('button', { name: 'Dm7 ii7', exact: true }).dblclick();
  await expect(slots).toHaveCount(2);
  await expect(slots.nth(1)).toContainText('Dm7');
  await expect(slots.nth(1)).toContainText('ii7');
  await expect(footprint.getByRole('button', { name: 'F IV', exact: true })).toBeVisible();

  // Tab from a chosen node passes the branches without letting the slot go.
  const node = page.getByRole('button', { name: /^2 Dm7 ii7/ });
  await node.click();
  await node.press('Tab');
  await expect(footprint.getByRole('button').first()).toBeFocused();
  await expect(node).toHaveAttribute('aria-pressed', 'true');
  await expect(page.getByText('試聴中')).toHaveCount(0);

  await page.getByRole('button', { name: '進行を鳴らす', exact: true }).click();
  await expect(slots.nth(0)).toHaveAttribute('aria-current', 'true');
  await expect(slots.nth(0)).toContainText('鳴っています');
  // At 90 BPM a slot of 4 beats lasts about 2.7 seconds; the mark then moves on.
  await expect(slots.nth(1)).toHaveAttribute('aria-current', 'true');
  await expect(slots.nth(1)).toContainText('鳴っています');
  await expect(slots.nth(0)).not.toHaveAttribute('aria-current');

  expect(errors).toEqual([]);
});
