---
paths: backend/**/*.go
---

# Docker操作規約（バックエンド）

## 重要：バックエンド変更後のリビルド必須

**バックエンドのGoコードに変更を加えた後は、必ずDockerコンテナをリビルドすること。**

### 理由

Dockerコンテナはビルド時に生成されたバイナリを使用しています。`docker compose restart` だけでは古いバイナリが使い続けられ、コード変更が反映されません。

### 変更後の必須コマンド

```bash
# バックエンドコンテナをリビルドして再起動
docker compose up -d --build backend
```

### リビルドが必要なケース

| 変更内容 | リビルド必須 |
|---------|------------|
| Handler/Service/Repository の実装変更 | ✅ 必須 |
| ルーティング（routes/*.go）の変更 | ✅ 必須 |
| Model/DTO の変更 | ✅ 必須 |
| Config の変更 | ✅ 必須 |
| マイグレーションファイルの追加のみ | ❌ 不要（マウント） |
| 環境変数の変更のみ | ❌ 不要（restart で可） |

### リビルド後の確認

```bash
# ルーティングが正しく登録されたか確認
docker compose logs backend | grep -i "GIN-debug" | grep "<追加したエンドポイント>"

# エラーがないか確認
docker compose logs backend --tail=50 | grep -i "error\|panic"
```

---

## よくある問題

### 症状：新しいAPIエンドポイントが404を返す

**原因**: コンテナがリビルドされていない

**解決**:
```bash
docker compose up -d --build backend
```

### 症状：コード変更が反映されない

**原因**: 古いバイナリが使用されている

**解決**:
```bash
# キャッシュを使わずにリビルド
docker compose build --no-cache backend
docker compose up -d backend
```

---

## チェックリスト

バックエンドに変更を加えた後：

- [ ] `docker compose up -d --build backend` を実行したか
- [ ] `docker compose logs backend` でエラーがないか確認したか
- [ ] 新しいエンドポイントの場合、ルーティングログで登録を確認したか
