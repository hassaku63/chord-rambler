import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';

/** Opens the screen and waits until it has heard from the API server. */
export async function openScreen(page: Page): Promise<void> {
  await page.goto('/');
  await expect(page.getByText(/サーバーにつながっています/)).toHaveRole('status');
}

/** The strip of slots. */
export function strip(page: Page): Locator {
  return page.getByRole('list', { name: '進行', exact: true });
}

/** The candidates for the slot after the last, once the answer to the latest progression is in. */
export async function tailCandidates(page: Page): Promise<Locator> {
  const region = page.getByRole('region', { name: /^末尾に足す候補/ });
  await expect(region).toHaveAttribute('aria-busy', 'false');
  return region;
}

/** Adds a slot by typing its chord name. */
export async function addByName(page: Page, chord: string): Promise<void> {
  await page.getByRole('textbox', { name: '和音名' }).fill(chord);
  await page.getByRole('button', { name: '枠を足す', exact: true }).click();
}
