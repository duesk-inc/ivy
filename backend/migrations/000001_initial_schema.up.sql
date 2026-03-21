-- Ivy Initial Schema
-- Users (JIT provisioned from Cognito)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cognito_sub VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'sales',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_cognito_sub ON users(cognito_sub);

-- Job Groups (同一案件の複数経路をグルーピング)
CREATE TABLE IF NOT EXISTS job_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(500) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_job_groups_user_id ON job_groups(user_id);

-- Matchings (matching results)
CREATE TABLE IF NOT EXISTS matchings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    job_group_id UUID REFERENCES job_groups(id),
    job_text TEXT NOT NULL,
    engineer_text TEXT NOT NULL,
    engineer_file_key VARCHAR(500),
    supplement JSONB DEFAULT '{}',
    supply_chain_level INTEGER NOT NULL DEFAULT 0,
    supply_chain_source VARCHAR(255),
    total_score INTEGER NOT NULL,
    grade VARCHAR(1) NOT NULL,
    result JSONB NOT NULL,
    model_used VARCHAR(100) NOT NULL,
    tokens_used INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_matchings_user_id ON matchings(user_id);
CREATE INDEX idx_matchings_created_at ON matchings(created_at DESC);
CREATE INDEX idx_matchings_grade ON matchings(grade);
CREATE INDEX idx_matchings_job_group_id ON matchings(job_group_id);

-- Settings (key-value store)
CREATE TABLE IF NOT EXISTS settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) NOT NULL UNIQUE,
    value JSONB NOT NULL,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Development user (for local development with COGNITO_ENABLED=false)
INSERT INTO users (id, cognito_sub, email, name, role) VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'admin@duesk.co.jp', '開発ユーザー', 'admin')
ON CONFLICT (cognito_sub) DO NOTHING;

-- Initial settings
INSERT INTO settings (key, value) VALUES
    ('margin', '{"type": "fixed", "amount": 50000}'),
    ('ai_model', '{"model": "claude-haiku-4-5-20251001"}'),
    ('data_retention', '{"jobs_days": 90, "engineers_days": 180, "matchings_days": 365}')
ON CONFLICT (key) DO NOTHING;
