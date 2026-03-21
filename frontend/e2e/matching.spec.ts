import { test, expect } from '@playwright/test';

async function loginAsAdmin(page: any) {
  await page.goto('/login');
  await page.getByLabel('メールアドレス').fill('admin@duesk.co.jp');
  await page.getByLabel('パスワード').fill('test');
  await page.getByRole('button', { name: 'ログイン' }).click();
  await expect(page.getByRole('heading', { name: 'マッチング実行' })).toBeVisible({ timeout: 10000 });
}

test.describe('マッチング実行', () => {
  test('案件テキストとエンジニアテキストでマッチング実行', async ({ page }) => {
    await loginAsAdmin(page);

    // 案件情報を入力（プレースホルダーで特定）
    await page.getByPlaceholder('案件名、必須スキル').fill('【案件】Java開発エンジニア\n必須スキル: Java 3年以上\n単価: 65-75万円');

    // エンジニア情報を入力
    await page.getByPlaceholder('スキル、経験年数').fill('T.Y. 30歳 男性\nJava 4年\n希望単価: 60万円');

    // マッチング実行
    await page.getByRole('button', { name: 'マッチング実行' }).click();

    // 結果表示を確認（モックAIなので72点B判定）
    await expect(page.getByText('72 / 100')).toBeVisible({ timeout: 30000 });
    await expect(page.getByText('B判定')).toBeVisible();
  });

  test('結果にスコア詳細とアドバイスが表示される', async ({ page }) => {
    await loginAsAdmin(page);

    await page.getByPlaceholder('案件名、必須スキル').fill('Java案件');
    await page.getByPlaceholder('スキル、経験年数').fill('Java 4年');

    await page.getByRole('button', { name: 'マッチング実行' }).click();
    await expect(page.getByText('72 / 100')).toBeVisible({ timeout: 30000 });

    await expect(page.getByText('スコア詳細')).toBeVisible();
    await expect(page.getByText('スキル適合')).toBeVisible();
    await expect(page.getByText('アドバイス')).toBeVisible();
    await expect(page.getByText('追加確認のヒント')).toBeVisible();
  });

  test('案件情報なしでエラー', async ({ page }) => {
    await loginAsAdmin(page);

    await page.getByPlaceholder('スキル、経験年数').fill('Java 4年');
    await page.getByRole('button', { name: 'マッチング実行' }).click();

    await expect(page.getByRole('alert')).toBeVisible({ timeout: 3000 });
  });
});
