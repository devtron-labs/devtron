/*
 * Copyright (c) 2024. Devtron Inc.
 */

package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetRoleForOtherEntityWithoutMatchingFilter(t *testing.T) {
	repository := UserAuthRepositoryImpl{Logger: zap.NewNop().Sugar()}
	for _, tc := range []struct {
		name string
		team string
		app  string
	}{
		{name: "empty filter"},
		{name: "incomplete filter", team: "team-1", app: "app-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			role, err := repository.GetRoleForOtherEntity(tc.team, tc.app, "", "", "", false)

			require.NoError(t, err)
			require.Equal(t, RoleModel{}, role)
		})
	}
}

func TestGetRoleGroupsByIdsWithoutIds(t *testing.T) {
	repository := RoleGroupRepositoryImpl{}

	groups, err := repository.GetRoleGroupsByIds(nil)

	require.NoError(t, err)
	require.Nil(t, groups)
}
