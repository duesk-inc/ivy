package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateGrade(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		expected string
	}{
		{"score 100 -> A", 100, "A"},
		{"score 80 -> A", 80, "A"},
		{"score 79 -> B", 79, "B"},
		{"score 60 -> B", 60, "B"},
		{"score 59 -> C", 59, "C"},
		{"score 40 -> C", 40, "C"},
		{"score 39 -> D", 39, "D"},
		{"score 25 -> D", 25, "D"},
		{"score 0 -> D", 0, "D"},
		{"score -1 -> D", -1, "D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateGrade(tt.score)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMatching_GradeLabel(t *testing.T) {
	tests := []struct {
		name     string
		grade    string
		expected string
	}{
		{"grade A", "A", "提案推奨"},
		{"grade B", "B", "提案検討可"},
		{"grade C", "C", "条件次第で検討"},
		{"grade D", "D", "提案非推奨"},
		{"unknown grade X", "X", ""},
		{"empty grade", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Matching{Grade: tt.grade}
			assert.Equal(t, tt.expected, m.GradeLabel())
		})
	}
}

func TestMatching_TableName(t *testing.T) {
	m := Matching{}
	assert.Equal(t, "matchings", m.TableName())
}

func TestMatching_BeforeCreate_EmptyID(t *testing.T) {
	m := &Matching{}
	err := m.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, m.ID)
	assert.True(t, strings.Contains(m.ID, "-"), "generated UUID should contain dashes")
}

func TestMatching_BeforeCreate_ExistingID(t *testing.T) {
	existingID := "existing-uuid-1234"
	m := &Matching{ID: existingID}
	err := m.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.Equal(t, existingID, m.ID)
}
