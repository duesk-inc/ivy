package dto

import "time"

// EmailSyncResponse メール同期レスポンス
type EmailSyncResponse struct {
	TotalProcessed    int       `json:"total_processed"`
	NewJobs           int       `json:"new_jobs"`
	NewEngineers      int       `json:"new_engineers"`
	DuplicatesSkipped int       `json:"duplicates_skipped"`
	OtherSkipped      int       `json:"other_skipped"`
	Errors            int       `json:"errors"`
	SyncedAt          time.Time `json:"synced_at"`
}
