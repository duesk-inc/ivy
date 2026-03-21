---
description: Ivyプロジェクトの概要。実装時に必ず参照する。
globs: ["**/*"]
---

# Ivy プロジェクト概要

## 何のアプリか
SES営業向けの案件×エンジニアマッチングWebアプリ。
Claude APIでマッチ度をスコアリングし、営業の判断を支援する。
自社エンジニアの案件探し・BP人材の案件マッチングの両方に対応。

## 設計書（実装の根拠）
- 設計書: `/Users/daichirouesaka/Documents/duesk-company/products/matching-tool/DESIGN.md`
- AIプロンプト: `/Users/daichirouesaka/Documents/duesk-company/products/matching-tool/matching_prompt.md`

## 参考プロジェクト
- Monstera: `/Users/daichirouesaka/dev/monstera`
- 流用: config, middleware, common/logger, s3_service
- 非流用: freee連携, 週報, スキルシート生成, Slack通知

## 技術スタック
- Backend: Go 1.24 + Gin + GORM + PostgreSQL 16 + Redis 7
- Frontend: React + Vite + TypeScript + MUI v7（Next.jsではない）
- AI: Claude API (Anthropic) — インターフェースでモック切替可能
- Auth: AWS Cognito (MonsteraとUser Pool共有、JITプロビジョニング)

## Monsteraとの差分で注意すること
- フロントエンドはNext.jsではなくReact SPA（Vite）
- ポートはMonstera+1（Backend:8081, DB:5433, Redis:6380, Cognito:9230）
- ユーザー管理はIvyに持たない（Monsteraで一元管理、JITで自動作成）
- Monsteraのfreee連携・週報等のコードはIvyに持ち込まない
