package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetting_TableName(t *testing.T) {
	s := Setting{}
	assert.Equal(t, "settings", s.TableName())
}

func TestSetting_BeforeCreate_EmptyID(t *testing.T) {
	s := &Setting{}
	err := s.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, s.ID)
	assert.True(t, strings.Contains(s.ID, "-"), "generated UUID should contain dashes")
}

func TestSetting_BeforeCreate_ExistingID(t *testing.T) {
	existingID := "existing-setting-uuid"
	s := &Setting{ID: existingID}
	err := s.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.Equal(t, existingID, s.ID)
}

func TestParseMarginSetting(t *testing.T) {
	data := json.RawMessage(`{"type":"fixed","amount":50000}`)
	result, err := ParseMarginSetting(data)

	assert.NoError(t, err)
	assert.Equal(t, "fixed", result.Type)
	assert.Equal(t, 50000, result.Amount)
}

func TestParseMarginSetting_Percentage(t *testing.T) {
	data := json.RawMessage(`{"type":"percentage","amount":15}`)
	result, err := ParseMarginSetting(data)

	assert.NoError(t, err)
	assert.Equal(t, "percentage", result.Type)
	assert.Equal(t, 15, result.Amount)
}

func TestParseMarginSetting_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`{invalid`)
	result, err := ParseMarginSetting(data)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseMarginSetting_EmptyJSON(t *testing.T) {
	data := json.RawMessage(`{}`)
	result, err := ParseMarginSetting(data)

	assert.NoError(t, err)
	assert.Equal(t, "", result.Type)
	assert.Equal(t, 0, result.Amount)
}

func TestParseAIModelSetting(t *testing.T) {
	data := json.RawMessage(`{"model":"claude-sonnet-4-20250514"}`)
	result, err := ParseAIModelSetting(data)

	assert.NoError(t, err)
	assert.Equal(t, "claude-sonnet-4-20250514", result.Model)
}

func TestParseAIModelSetting_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`not-json`)
	result, err := ParseAIModelSetting(data)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseDataRetentionSetting(t *testing.T) {
	data := json.RawMessage(`{"jobs_days":90,"engineers_days":180,"matchings_days":365}`)
	result, err := ParseDataRetentionSetting(data)

	assert.NoError(t, err)
	assert.Equal(t, 90, result.JobsDays)
	assert.Equal(t, 180, result.EngineersDays)
	assert.Equal(t, 365, result.MatchingsDays)
}

func TestParseDataRetentionSetting_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`{bad}`)
	result, err := ParseDataRetentionSetting(data)

	assert.Error(t, err)
	assert.Nil(t, result)
}
