/*
 * Copyright (c) 2024. Devtron Inc.
 */

package appStoreDiscoverRepository

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAppStoreFilterValidation(t *testing.T) {
	t.Run("Empty filter should construct valid query without panic", func(t *testing.T) {
		filter := AppStoreFilter{
			AppStoreName: "",
			ChartRepoId:  []int{},
			RegistryId:   []string{},
		}
		assert.Equal(t, 0, len(filter.ChartRepoId))
		assert.Equal(t, 0, len(filter.RegistryId))
	})

	t.Run("Filter with chart repo and registry IDs", func(t *testing.T) {
		filter := AppStoreFilter{
			AppStoreName: "nginx",
			ChartRepoId:  []int{1, 2},
			RegistryId:   []string{"reg-1"},
			Size:         10,
			Offset:       0,
		}
		assert.Equal(t, "nginx", filter.AppStoreName)
		assert.Equal(t, 2, len(filter.ChartRepoId))
		assert.Equal(t, 1, len(filter.RegistryId))
		assert.Equal(t, 10, filter.Size)
	})
}
