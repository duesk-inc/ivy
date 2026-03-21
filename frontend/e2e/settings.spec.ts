import { test, expect } from '@playwright/test';

async function loginAsAdmin(page: any) {
  await page.goto('/login');
  await page.getByLabel('メールアドレス').fill('admin@duesk.co.jp');
  await page.getByLabel('パスワード').fill('test');
  await page.getByRole('button', { name: 'ログイン' }).click();
  await expect(page.getByRole('heading', { name: 'マッチング実行' })).toBeVisible({ timeout: 10000 });
}

test.describe('設定画面', () => {
  test('管理者は設定画面にアクセスできる', async ({ page }) => {
    await loginAsAdmin(page);
    await page.getByRole('button', { name: '設定' }).click();
    await expect(page.getByRole('heading', { name: '設定', exact: true })).toBeVisible({ timeout: 5000 });
  });

  test('マージン設定とAIモデル設定が表示される', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/settings');
    await expect(page.getByText('マージン設定')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('AIモデル設定')).toBeVisible();
    await expect(page.getByText('ユーザー管理')).toBeVisible();
  });
});
