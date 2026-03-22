-- Phase 2 rollback

DROP TABLE IF EXISTS batch_matchings;
DROP TABLE IF EXISTS gmail_sync_state;
DROP TABLE IF EXISTS processed_emails;
DROP TABLE IF EXISTS engineer_profiles;
DROP TABLE IF EXISTS jobs;

-- data_retention設定からprocessed_emails_daysを除去
UPDATE settings SET value = value - 'processed_emails_days'
WHERE key = 'data_retention';
