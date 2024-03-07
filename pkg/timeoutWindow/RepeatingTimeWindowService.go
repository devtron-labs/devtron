package timeoutWindow

import (
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"github.com/go-pg/pg"
	"github.com/samber/lo"
	"time"
)

type RepeatingTimeWindowService interface {
	TimeoutWindowService
	UpdateWindowMappings(windows []*TimeWindow, userId int32, err error, tx *pg.Tx, policyId int) error
	GetActiveWindow(targetTimeWithZone time.Time, windows []*TimeWindow) (bool, time.Time, *TimeWindow)
	GetWindowsForResources(resourceId []int, resourceType repository.ResourceType) (map[int][]*TimeWindow, error)
}

type RepeatingTimeWindowServiceImpl struct {
	*TimeWindowServiceImpl
}

func NewRepeatingTimeWindowServiceImpl(service *TimeWindowServiceImpl) *RepeatingTimeWindowServiceImpl {
	return &RepeatingTimeWindowServiceImpl{TimeWindowServiceImpl: service}
}

func (impl RepeatingTimeWindowServiceImpl) UpdateWindowMappings(windows []*TimeWindow, userId int32, err error, tx *pg.Tx, policyId int) error {

	//TODO validate Windows

	windowExpressions := lo.Map(windows, func(window *TimeWindow, index int) TimeWindowExpression {
		return TimeWindowExpression{
			TimeoutExpression: window.toJsonString(),
			ExpressionFormat:  bean.RecurringTimeRange,
		}
	})

	//create time windows and map
	err = impl.CreateAndMapWithResource(tx, windowExpressions, userId, policyId, repository.DeploymentWindowProfile)
	return err
}

func (impl RepeatingTimeWindowServiceImpl) GetActiveWindow(targetTimeWithZone time.Time, windows []*TimeWindow) (bool, time.Time, *TimeWindow) {
	isActive := false
	maxEndTimeStamp := time.Time{}
	minStartTimeStamp := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	var appliedWindow *TimeWindow
	for _, window := range windows {
		timeRange := window.toTimeRange()
		timestamp, isInside, err := timeRange.GetScheduleSpec(targetTimeWithZone)
		if err != nil {
			impl.logger.Errorw("GetScheduleSpec failed", "timeRange", timeRange, "window", window, "time", targetTimeWithZone)
			continue
		}
		if isInside && !timestamp.IsZero() {
			isActive = true
			if timestamp.After(maxEndTimeStamp) {
				maxEndTimeStamp = timestamp
				appliedWindow = window
			}
		} else if !isInside && !timestamp.IsZero() {
			if timestamp.Before(minStartTimeStamp) {
				minStartTimeStamp = timestamp
				appliedWindow = window
			}
		}
	}
	if isActive {
		return true, maxEndTimeStamp, appliedWindow
	}
	return false, minStartTimeStamp, appliedWindow
}

func (impl RepeatingTimeWindowServiceImpl) GetWindowsForResources(resourceIds []int, resourceType repository.ResourceType) (map[int][]*TimeWindow, error) {
	//get windows
	resourceIdToExpressions, err := impl.GetMappingsForResources(resourceIds, resourceType)
	if err != nil {
		return nil, err
	}

	resourceIdToWindows := make(map[int][]*TimeWindow, 0)

	for resourceId, expressions := range resourceIdToExpressions {
		windows := lo.Map(expressions, func(expr TimeWindowExpression, index int) *TimeWindow {
			window := &TimeWindow{}
			window.setFromJsonString(expr.TimeoutExpression)
			return window
		})
		resourceIdToWindows[resourceId] = windows
	}
	return resourceIdToWindows, nil
}
