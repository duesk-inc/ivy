import { test, expect } from '@playwright/test';

test.describe('認証フロー', () => {
  test('ログインページが表示される', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByText('Ivy')).toBeVisible();
    await expect(page.getByText('SES マッチングツール')).toBeVisible();
    await expect(page.getByLabel('メールアドレス')).toBeVisible();
    await expect(page.getByLabel('パスワード')).toBeVisible();
    await expect(page.getByRole('button', { name: 'ログイン' })).toBeVisible();
  });

  test('未認証時はログインページにリダイレクトされる', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/login/);
  });

  test('ログインしてマッチングページに遷移する', async ({ page }) => {
    await page.goto('/login');

    await page.getByLabel('メールアドレス').fill('admin@duesk.co.jp');
    await page.getByLabel('パスワード').fill('test');
    await page.getByRole('button', { name: 'ログイン' }).click();

    // マッチングページに遷移
    await expect(page.getByRole('heading', { name: 'マッチング実行' })).toBeVisible({ timeout: 10000 });
  });
});
