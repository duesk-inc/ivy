-- Phase 2: Gmail連携 + N:Nマッチング用テーブル

-- jobs (メールから抽出した案件情報)
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_hash VARCHAR(64) NOT NULL,
    source_email_id VARCHAR(255),
    raw_text TEXT NOT NULL,
    parsed JSONB NOT NULL DEFAULT '{}',
    start_month VARCHAR(7), -- YYYY-MM
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active / archived
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 作成日時
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 更新日時
    expires_at TIMESTAMPTZ -- 自動削除日
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_jobs_content_hash ON jobs(content_hash);
CREATE INDEX IF NOT EXISTS idx_jobs_start_month ON jobs(start_month);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_expires_at ON jobs(expires_at);

-- engineer_profiles (メールから抽出した人材情報)
CREATE TABLE IF NOT EXISTS engineer_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_hash VARCHAR(64) NOT NULL,
    source_email_id VARCHAR(255),
    raw_text TEXT NOT NULL,
    file_key VARCHAR(500),
    parsed JSONB NOT NULL DEFAULT '{}',
    start_month VARCHAR(7), -- YYYY-MM
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active / archived
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 作成日時
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 更新日時
    expires_at TIMESTAMPTZ -- 自動削除日
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_engineer_profiles_content_hash ON engineer_profiles(content_hash);
CREATE INDEX IF NOT EXISTS idx_engineer_profiles_start_month ON engineer_profiles(start_month);
CREATE INDEX IF NOT EXISTS idx_engineer_profiles_status ON engineer_profiles(status);
CREATE INDEX IF NOT EXISTS idx_engineer_profiles_expires_at ON engineer_profiles(expires_at);

-- processed_emails (処理済みメール追跡)
CREATE TABLE IF NOT EXISTS processed_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_hash VARCHAR(64) NOT NULL,
    gmail_message_id VARCHAR(255) NOT NULL,
    classification VARCHAR(20) NOT NULL, -- job / engineer / other
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW() -- 処理日時
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_processed_emails_content_hash ON processed_emails(content_hash);
CREATE INDEX IF NOT EXISTS idx_processed_emails_gmail_message_id ON processed_emails(gmail_message_id);

-- gmail_sync_state (Gmail同期位置管理)
CREATE TABLE IF NOT EXISTS gmail_sync_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    last_history_id BIGINT NOT NULL DEFAULT 0,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 最終同期日時
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW() -- 更新日時
);

-- batch_matchings (N:Nバッチ実行状態)
CREATE TABLE IF NOT EXISTS batch_matchings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    batch_type VARCHAR(20) NOT NULL DEFAULT 'n_to_n', -- n_to_n / job_to_engineers / engineer_to_jobs
    start_month_from VARCHAR(7) NOT NULL,
    start_month_to VARCHAR(7) NOT NULL,
    total_pairs INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'running', -- running / completed / failed
    results JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 作成日時
    completed_at TIMESTAMPTZ -- 完了日時
);

CREATE INDEX IF NOT EXISTS idx_batch_matchings_user_id ON batch_matchings(user_id);
CREATE INDEX IF NOT EXISTS idx_batch_matchings_status ON batch_matchings(status);
CREATE INDEX IF NOT EXISTS idx_batch_matchings_created_at ON batch_matchings(created_at DESC);

-- data_retention設定にprocessed_emails_daysを追加
UPDATE settings SET value = value || '{"processed_emails_days": 180}'
WHERE key = 'data_retention' AND NOT (value ? 'processed_emails_days');
