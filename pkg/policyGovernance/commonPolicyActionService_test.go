/*
 * Copyright (c) 2024. Devtron Inc.
 */

package policyGovernance

import (
	"fmt"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddApplyEventObserver(t *testing.T) {
	cps := NewCommonPolicyActionsService(nil,
		nil,
		nil, nil,
		nil, nil, nil)

	observerNames := []string{"observer1", "observer1"}
	observer1 := func(tx *pg.Tx, commaSeperatedAppEnvIds [][]int) error {
		return fmt.Errorf("%s", observerNames[0])
	}

	observer2 := func(tx *pg.Tx, commaSeperatedAppEnvIds [][]int) error {
		return fmt.Errorf("%s", observerNames[1])
	}

	added := cps.AddApplyEventObserver(ImagePromotion, observer1)
	assert.Equal(t, true, added)
	added = cps.AddApplyEventObserver(ImagePromotion, observer2)
	assert.Equal(t, true, added)
	observers := cps.applyEventObservers[ImagePromotion]
	assert.Equal(t, len(observers), len(observerNames))
	containsObserver1 := util2.Contains(observers, func(observer ApplyObserver) bool {
		err := observer(nil, nil)
		return observerNames[0] == err.Error()
	})
	assert.Equal(t, true, containsObserver1)

	containsObserver2 := util2.Contains(observers, func(observer ApplyObserver) bool {
		err := observer(nil, nil)
		return observerNames[1] == err.Error()
	})
	assert.Equal(t, true, containsObserver2)

}
