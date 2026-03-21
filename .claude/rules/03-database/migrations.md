---
paths: backend/migrations/**/*.sql
---

# マイグレーション規約

## 初期開発フェーズ限定ルール

> **重要**: このセクションは初回リリース後に無効となります

### テーブル定義変更時の方針

**初期開発中は ALTER 文を追加せず、直接 CREATE 文を修正すること**

```sql
-- ✅ 正しい: 元の CREATE 文を直接修正
-- 000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(255) PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  new_column VARCHAR(100),  -- 新カラムを直接追加
  -- ...
);

-- ❌ 禁止（初期開発中）: 別マイグレーションで ALTER
-- 000999_add_new_column_to_users.up.sql
-- ALTER TABLE users ADD COLUMN new_column VARCHAR(100);
```

### 理由

1. **スキーマの見通しが良い**: テーブル定義が1ファイルに集約される
2. **dirty 状態を回避**: マイグレーションの失敗・やり直しが容易
3. **初期段階では本番データがない**: 破壊的変更のリスクがない

### 修正手順

1. 対象の `.up.sql` ファイルを直接編集
2. 対応する `.down.sql` も必要に応じて修正
3. ローカルDBをリセット: `docker compose down -v && docker compose up -d`
4. マイグレーション実行: `docker compose exec backend ./entrypoint.sh migrate`

### 初回リリース後

- このルールは無効となり、通常の ALTER 文による追加方式に移行
- 既存データの移行を考慮した段階的なマイグレーションが必要

---

## ファイル命名規則

### 番号帯の意味

| 番号帯 | 用途 | 例 |
|--------|------|-----|
| `000XXX` | 初期スキーマ（テーブル作成） | `000001_create_users_table` |
| `100XXX` | 初期シードデータ | `100000_seed_initial_data` |
| `200XXX` | スキーマ追加・修正 | `200001_create_user_roles_table` |
| `300XXX` | 認証・ユーザー関連 | `300000_seed_cognito_users` |
| `400XXX` | 機能追加 | `400001_create_candidate_tables` |
| `500XXX` | 大規模変更・クリーンアップ | `500100_notification_phase1_cleanup` |

### 命名フォーマット

```
{番号}_{動詞}_{対象}_{詳細}.up.sql
{番号}_{動詞}_{対象}_{詳細}.down.sql
```

**動詞一覧:**
- `create` - テーブル/インデックス作成
- `drop` - テーブル/インデックス削除
- `add` - カラム/制約追加
- `alter` - カラム変更
- `seed` - 初期データ投入
- `update` - データ更新
- `configure` - 設定変更

---

## 外部キー制約の順序（重要）

### dirty 状態を防ぐための鉄則

**親テーブルを先に、子テーブルを後に作成すること**

```
作成順序:
1. users          (親)
2. profiles       (users を参照)
3. user_roles     (users を参照)
4. weekly_reports (users を参照)
5. daily_records  (weekly_reports を参照)
```

### 依存関係の確認方法

```sql
-- 外部キー制約の確認
SELECT
    tc.table_name AS child_table,
    kcu.column_name AS child_column,
    ccu.table_name AS parent_table,
    ccu.column_name AS parent_column
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY';
```

### 外部キー制約の書き方

```sql
-- ✅ 推奨: 制約名を明示し、ON DELETE/UPDATE を指定
CONSTRAINT fk_user_roles_user_id
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE

-- ❌ 避ける: 暗黙の制約名
FOREIGN KEY (user_id) REFERENCES users(id)
```

### 循環参照への対処

```sql
-- DEFERRABLE を使用して循環参照を解決
ALTER TABLE table_a
ADD CONSTRAINT fk_table_a_table_b
    FOREIGN KEY (table_b_id)
    REFERENCES table_b(id)
    DEFERRABLE INITIALLY DEFERRED;
```

---

## dirty 状態の対処

### dirty 状態とは

マイグレーション実行中にエラーが発生し、`schema_migrations` テーブルの `dirty` フラグが `true` になった状態。

### 確認方法

```bash
# マイグレーションバージョン確認
docker compose exec backend migrate \
  -path ./migrations \
  -database "postgresql://postgres:postgres@postgres:5432/monstera?sslmode=disable" \
  version
```

### 解決方法

```bash
# 1. dirty フラグを強制リセット
docker compose exec backend migrate \
  -path ./migrations \
  -database "postgresql://..." \
  force {バージョン番号}

# 2. または DB をリセット（開発環境のみ）
docker compose down -v
docker compose up -d
```

### 予防策

1. **トランザクション内で実行**: DDL は可能な限りトランザクション内で
2. **小さな単位で分割**: 1マイグレーション = 1つの論理的変更
3. **down.sql のテスト**: ロールバックが正常に動作するか確認
4. **外部キー順序の遵守**: 親テーブル → 子テーブルの順

---

## up.sql / down.sql の書き方

### up.sql のテンプレート

```sql
-- テーブル作成
CREATE TABLE IF NOT EXISTS table_name (
    id VARCHAR(36) PRIMARY KEY,
    -- カラム定義...
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 作成日時
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 更新日時
    deleted_at TIMESTAMPTZ NULL -- 削除日時（論理削除が必要な場合）
);

-- インデックス作成
CREATE INDEX IF NOT EXISTS idx_table_name_column ON table_name(column);

-- 外部キー制約（テーブル作成後）
ALTER TABLE table_name
ADD CONSTRAINT fk_table_name_parent_id
    FOREIGN KEY (parent_id) REFERENCES parent_table(id)
    ON DELETE CASCADE;

-- コメント追加（推奨）
COMMENT ON TABLE table_name IS 'テーブルの説明';
COMMENT ON COLUMN table_name.column IS 'カラムの説明';
```

### down.sql のテンプレート

```sql
-- 外部キー制約の削除（先に）
ALTER TABLE table_name DROP CONSTRAINT IF EXISTS fk_table_name_parent_id;

-- インデックス削除
DROP INDEX IF EXISTS idx_table_name_column;

-- テーブル削除
DROP TABLE IF EXISTS table_name;

-- ENUM型削除（必要な場合）
DROP TYPE IF EXISTS custom_enum_type;
```

### IF EXISTS / IF NOT EXISTS の使用

```sql
-- ✅ 推奨: 冪等性を確保
CREATE TABLE IF NOT EXISTS ...
CREATE INDEX IF NOT EXISTS ...
DROP TABLE IF EXISTS ...
DROP INDEX IF EXISTS ...

-- ❌ 避ける: 再実行時にエラー
CREATE TABLE ...
DROP TABLE ...
```

---

## シードデータ

### 命名規則

```
100XXX_seed_{対象}.up.sql  -- 初期データ
200XXX_seed_{対象}.up.sql  -- 追加データ
```

### 冪等なシードの書き方

```sql
-- ✅ 推奨: ON CONFLICT で重複を回避
INSERT INTO leave_types (id, code, name)
VALUES
    ('1', 'PAID', '有給休暇'),
    ('2', 'SICK', '病気休暇')
ON CONFLICT (id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name;

-- ✅ 推奨: 存在チェック
INSERT INTO settings (key, value)
SELECT 'setting_key', 'setting_value'
WHERE NOT EXISTS (
    SELECT 1 FROM settings WHERE key = 'setting_key'
);
```

---

## タイムスタンプカラム設計

### 統一パターン（必須）

すべてのテーブルで以下のパターンを使用すること：

```sql
-- ✅ 正しい: 統一されたタイムスタンプパターン
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 作成日時
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 更新日時
deleted_at TIMESTAMPTZ NULL, -- 削除日時（論理削除が必要な場合のみ）
```

### 禁止パターン

以下のパターンは使用禁止：

```sql
-- ❌ 禁止: TIMESTAMP 型（タイムゾーン情報なし）
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

-- ❌ 禁止: TIMESTAMP(3)（精度指定）
created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP

-- ❌ 禁止: タイムゾーン指定の DEFAULT
created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Tokyo')

-- ❌ 禁止: CURRENT_TIMESTAMP（NOW() を使用すること）
created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP

-- ❌ 禁止: NOT NULL なしの created_at/updated_at
created_at TIMESTAMPTZ DEFAULT NOW()
```

### ルールの詳細

| 項目 | ルール | 理由 |
|------|--------|------|
| **型** | `TIMESTAMPTZ` を使用 | タイムゾーン情報を保持し、国際化対応が容易 |
| **NOT NULL** | `created_at`, `updated_at` は必須 | レコードの監査・追跡に必要 |
| **deleted_at** | `NULL` 許容（NOT NULL 不要） | 論理削除のため NULL = 未削除 |
| **DEFAULT** | `NOW()` を使用 | `CURRENT_TIMESTAMP` より短く可読性が高い |
| **コメント** | 日本語コメント必須 | `-- 作成日時`, `-- 更新日時`, `-- 削除日時` |

### updated_at 自動更新トリガー

`updated_at` カラムの自動更新には、共通のトリガー関数を使用：

```sql
-- トリガー関数（既に存在する場合は作成不要）
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 各テーブルにトリガーを設定
DROP TRIGGER IF EXISTS update_table_name_updated_at ON table_name;
CREATE TRIGGER update_table_name_updated_at
    BEFORE UPDATE ON table_name
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### その他のタイムスタンプカラム

`created_at`, `updated_at`, `deleted_at` 以外のタイムスタンプカラムも同様に `TIMESTAMPTZ` を使用：

```sql
-- ✅ 正しい
executed_at TIMESTAMPTZ NULL,
scheduled_at TIMESTAMPTZ NOT NULL,
completed_at TIMESTAMPTZ NULL,

-- ❌ 禁止
executed_at TIMESTAMP NULL,
scheduled_at TIMESTAMP(3) NOT NULL,
```

---

## インデックス設計

### 命名規則

```sql
-- 単一カラム
idx_{テーブル名}_{カラム名}

-- 複合インデックス
idx_{テーブル名}_{カラム1}_{カラム2}

-- ユニーク制約
uq_{テーブル名}_{カラム名}
```

### 例

```sql
CREATE INDEX IF NOT EXISTS idx_weekly_reports_user_id ON weekly_reports(user_id);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_year_week ON weekly_reports(year, week_number);
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email ON users(email);
```

---

## チェックリスト

### マイグレーション作成時

- [ ] 番号帯は適切か（000XXX〜500XXX）
- [ ] ファイル名は命名規則に従っているか
- [ ] up.sql と down.sql の両方を作成したか
- [ ] IF EXISTS / IF NOT EXISTS を使用しているか
- [ ] 外部キーの参照先テーブルは先に作成されるか
- [ ] 制約名を明示しているか
- [ ] インデックス名は命名規則に従っているか
- [ ] タイムスタンプカラムは `TIMESTAMPTZ NOT NULL DEFAULT NOW()` を使用しているか
- [ ] タイムスタンプカラムに日本語コメント（`-- 作成日時` 等）を付けているか

### 初期開発フェーズ（リリース前）

- [ ] ALTER 文ではなく CREATE 文を直接修正しているか
- [ ] 修正後に DB リセット + マイグレーション実行を確認したか

### テスト

- [ ] up → down → up が正常に動作するか
- [ ] dirty 状態にならないか
- [ ] 既存データへの影響は考慮されているか（初回リリース後）

---

## トラブルシューティング

### よくあるエラー

| エラー | 原因 | 対処 |
|--------|------|------|
| `relation "xxx" does not exist` | 外部キー参照先が未作成 | マイグレーション順序を確認 |
| `duplicate key value violates unique constraint` | シードデータの重複 | ON CONFLICT を使用 |
| `Dirty database version` | マイグレーション途中で失敗 | `force` コマンドで解決 |
| `column "xxx" does not exist` | カラム名の不一致 | スペルミスを確認 |

### ログ確認

```bash
# バックエンドログでSQL実行状況を確認
docker compose logs -f backend | grep -i "migration\|error\|sql"
```
