import { test, expect } from '@playwright/test';

async function loginAsAdmin(page: any) {
  await page.goto('/login');
  await page.getByLabel('メールアドレス').fill('admin@duesk.co.jp');
  await page.getByLabel('パスワード').fill('test');
  await page.getByRole('button', { name: 'ログイン' }).click();
  await expect(page.getByRole('heading', { name: 'マッチング実行' })).toBeVisible({ timeout: 10000 });
}

test.describe('マッチング履歴', () => {
  test('履歴ページに遷移できる', async ({ page }) => {
    await loginAsAdmin(page);
    await page.getByRole('button', { name: '履歴' }).click();
    await expect(page.getByRole('heading', { name: 'マッチング履歴' })).toBeVisible({ timeout: 5000 });
  });

  test('履歴一覧にテーブルが表示される', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/history');
    await expect(page.getByRole('heading', { name: 'マッチング履歴' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('columnheader', { name: '日時' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: '案件名' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'スコア' })).toBeVisible();
  });
});
