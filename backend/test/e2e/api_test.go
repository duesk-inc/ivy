//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8081"

// --- Helper functions ---

func apiGet(t *testing.T, path string) *http.Response {
	resp, err := http.Get(baseURL + path)
	require.NoError(t, err)
	return resp
}

func apiPost(t *testing.T, path string, body interface{}) *http.Response {
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	return resp
}

func apiPostRaw(t *testing.T, path string, rawBody []byte) *http.Response {
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewReader(rawBody))
	require.NoError(t, err)
	return resp
}

func apiPut(t *testing.T, path string, body interface{}) *http.Response {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("PUT", baseURL+path, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func apiPutRaw(t *testing.T, path string, rawBody []byte) *http.Response {
	req, _ := http.NewRequest("PUT", baseURL+path, bytes.NewReader(rawBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func apiDelete(t *testing.T, path string) *http.Response {
	req, _ := http.NewRequest("DELETE", baseURL+path, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func parseJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	defer resp.Body.Close()
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

// createMatchingRequest returns a standard matching request body
func createMatchingRequest() map[string]interface{} {
	return map[string]interface{}{
		"job_text": `【案件概要】
案件名: ECサイトリニューアル
スキル: Go, React, AWS
期間: 2026年4月〜長期
単価: 70万円〜85万円
勤務地: リモート可
備考: 設計から参画可能な方`,
		"engineer_text": `【エンジニア情報】
名前: 山田太郎
経験年数: 5年
スキル: Go(3年), React(2年), AWS(2年), Docker
希望単価: 75万円
稼働: 即日可能
備考: ECサイト開発経験あり`,
		"supplement": map[string]interface{}{
			"affiliation_type": "duesk",
			"rate":             75,
			"nationality":     "japanese",
			"employment_type":  "employee",
			"available_from":   "2026-04",
		},
	}
}

// --- Test cases ---

func TestE2E_HealthCheck(t *testing.T) {
	resp := apiGet(t, "/health")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := parseJSON(t, resp)
	assert.Equal(t, "ok", body["status"])

	checks, ok := body["checks"].(map[string]interface{})
	require.True(t, ok, "checks should be a map")
	assert.Equal(t, "ok", checks["database"])
	assert.Equal(t, "ok", checks["redis"])
}

func TestE2E_AuthFlow(t *testing.T) {
	t.Run("Login", func(t *testing.T) {
		resp := apiPost(t, "/api/v1/auth/login", map[string]interface{}{
			"email":    "admin@duesk.co.jp",
			"password": "test",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		assert.NotEmpty(t, body["access_token"], "access_token should be present")
	})

	t.Run("Me", func(t *testing.T) {
		resp := apiGet(t, "/api/v1/me")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		assert.NotEmpty(t, body["id"], "user id should be present")
		assert.NotEmpty(t, body["role"], "user role should be present")
	})

	t.Run("Logout", func(t *testing.T) {
		resp := apiPost(t, "/api/v1/auth/logout", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		assert.NotEmpty(t, body["message"])
	})
}

func TestE2E_SettingsFlow(t *testing.T) {
	t.Run("GetAll", func(t *testing.T) {
		resp := apiGet(t, "/api/v1/settings")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		settings, ok := body["settings"].([]interface{})
		require.True(t, ok, "settings should be an array")
		require.GreaterOrEqual(t, len(settings), 1, "should have at least 1 setting")

		keys := make([]string, 0)
		for _, s := range settings {
			item := s.(map[string]interface{})
			keys = append(keys, item["key"].(string))
		}
		assert.Contains(t, keys, "margin")
		assert.Contains(t, keys, "ai_model")
		assert.Contains(t, keys, "data_retention")
	})

	t.Run("UpdateMargin", func(t *testing.T) {
		resp := apiPut(t, "/api/v1/settings/margin", map[string]interface{}{
			"value": map[string]interface{}{
				"type":   "fixed",
				"amount": 60000,
			},
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("VerifyMarginUpdated", func(t *testing.T) {
		resp := apiGet(t, "/api/v1/settings")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		settings := body["settings"].([]interface{})

		var marginValue map[string]interface{}
		for _, s := range settings {
			item := s.(map[string]interface{})
			if item["key"] == "margin" {
				err := json.Unmarshal([]byte(fmt.Sprintf("%v", item["value"])), &marginValue)
				if err != nil {
					raw, _ := json.Marshal(item["value"])
					err = json.Unmarshal(raw, &marginValue)
					require.NoError(t, err)
				}
				break
			}
		}
		require.NotNil(t, marginValue, "margin setting should exist")
		assert.Equal(t, "fixed", marginValue["type"])
		assert.Equal(t, float64(60000), marginValue["amount"])
	})

	t.Run("RestoreMargin", func(t *testing.T) {
		resp := apiPut(t, "/api/v1/settings/margin", map[string]interface{}{
			"value": map[string]interface{}{
				"type":   "fixed",
				"amount": 50000,
			},
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestE2E_MatchingFlow(t *testing.T) {
	var matchingID string

	t.Run("Execute", func(t *testing.T) {
		reqBody := createMatchingRequest()
		resp := apiPost(t, "/api/v1/matchings", reqBody)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		assert.NotEmpty(t, body["id"], "id should be present")
		assert.Equal(t, float64(72), body["total_score"], "total_score should be 72")
		assert.Equal(t, "B", body["grade"], "grade should be B")
		assert.NotEmpty(t, body["grade_label"], "grade_label should be present")
		assert.NotNil(t, body["result"], "result should be present")

		matchingID = body["id"].(string)
	})

	t.Run("List", func(t *testing.T) {
		resp := apiGet(t, "/api/v1/matchings")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		items, ok := body["items"].([]interface{})
		require.True(t, ok, "items should be an array")
		assert.GreaterOrEqual(t, len(items), 1, "should have at least 1 item")

		total, ok := body["total"].(float64)
		require.True(t, ok, "total should be a number")
		assert.GreaterOrEqual(t, total, float64(1))
	})

	t.Run("GetByID", func(t *testing.T) {
		require.NotEmpty(t, matchingID, "matchingID must be set from Execute test")

		resp := apiGet(t, "/api/v1/matchings/"+matchingID)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		assert.Equal(t, matchingID, body["id"])
		assert.Equal(t, float64(72), body["total_score"])
		assert.Equal(t, "B", body["grade"])
		assert.NotEmpty(t, body["grade_label"])
		assert.NotNil(t, body["result"])
		assert.NotEmpty(t, body["job_text"])
		assert.NotEmpty(t, body["engineer_text"])
	})

	t.Run("Delete", func(t *testing.T) {
		require.NotEmpty(t, matchingID, "matchingID must be set from Execute test")

		resp := apiDelete(t, "/api/v1/matchings/"+matchingID)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		assert.NotEmpty(t, body["message"])
	})

	t.Run("GetByID_AfterDelete", func(t *testing.T) {
		require.NotEmpty(t, matchingID, "matchingID must be set from Execute test")

		resp := apiGet(t, "/api/v1/matchings/"+matchingID)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestE2E_MatchingValidation(t *testing.T) {
	t.Run("EmptyBody", func(t *testing.T) {
		resp := apiPost(t, "/api/v1/matchings", map[string]interface{}{})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("NoEngineerInfo", func(t *testing.T) {
		resp := apiPost(t, "/api/v1/matchings", map[string]interface{}{
			"job_text": "test job description",
		})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("NoJobText", func(t *testing.T) {
		resp := apiPost(t, "/api/v1/matchings", map[string]interface{}{
			"engineer_text": "test engineer info",
		})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestE2E_MatchingListPagination(t *testing.T) {
	createdIDs := make([]string, 0, 3)

	// Create 3 matchings
	for i := 0; i < 3; i++ {
		reqBody := createMatchingRequest()
		resp := apiPost(t, "/api/v1/matchings", reqBody)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		id, ok := body["id"].(string)
		require.True(t, ok, "id should be a string")
		createdIDs = append(createdIDs, id)
	}

	t.Run("Page1", func(t *testing.T) {
		resp := apiGet(t, "/api/v1/matchings?page=1&page_size=2")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		total, ok := body["total"].(float64)
		require.True(t, ok)
		assert.GreaterOrEqual(t, total, float64(3), "total should be at least 3")

		assert.Equal(t, float64(2), body["page_size"])

		items, ok := body["items"].([]interface{})
		require.True(t, ok)
		assert.LessOrEqual(t, len(items), 2, "page should have at most 2 items")
	})

	t.Run("Page2", func(t *testing.T) {
		resp := apiGet(t, "/api/v1/matchings?page=2&page_size=2")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := parseJSON(t, resp)
		assert.Equal(t, float64(2), body["page"])
	})

	// Cleanup
	for _, id := range createdIDs {
		resp := apiDelete(t, "/api/v1/matchings/"+id)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestE2E_SettingsUpdateValidation(t *testing.T) {
	t.Run("NonexistentKey", func(t *testing.T) {
		resp := apiPut(t, "/api/v1/settings/nonexistent_key", map[string]interface{}{
			"value": map[string]interface{}{},
		})
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		resp := apiPutRaw(t, "/api/v1/settings/margin", []byte(`{invalid json`))
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}
