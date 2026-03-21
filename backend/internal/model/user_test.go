package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_TableName(t *testing.T) {
	u := User{}
	assert.Equal(t, "users", u.TableName())
}

func TestUser_IsAdmin_True(t *testing.T) {
	u := &User{Role: RoleAdmin}
	assert.True(t, u.IsAdmin())
}

func TestUser_IsAdmin_False(t *testing.T) {
	u := &User{Role: RoleSales}
	assert.False(t, u.IsAdmin())
}

func TestUser_BeforeCreate_EmptyID(t *testing.T) {
	u := &User{}
	err := u.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, u.ID)
	assert.True(t, strings.Contains(u.ID, "-"), "generated UUID should contain dashes")
}

func TestUser_BeforeCreate_ExistingID(t *testing.T) {
	existingID := "existing-user-uuid"
	u := &User{ID: existingID}
	err := u.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.Equal(t, existingID, u.ID)
}

func TestRole_Constants(t *testing.T) {
	assert.Equal(t, Role("admin"), RoleAdmin)
	assert.Equal(t, Role("sales"), RoleSales)
}
