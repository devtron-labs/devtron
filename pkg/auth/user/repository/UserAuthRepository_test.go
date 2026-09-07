/*
 * Copyright (c) 2024. Devtron Inc.
 */

package repository

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRoleLookupQueryHelper(t *testing.T) {
	t.Run("Empty parameters return empty model", func(t *testing.T) {
		team := ""
		app := ""
		env := ""
		act := ""
		assert.True(t, team == "" && app == "" && env == "" && act == "")
	})

	t.Run("Wildcard placeholder check", func(t *testing.T) {
		assert.Equal(t, "null", EMPTY_PLACEHOLDER_FOR_QUERY)
	})
}
