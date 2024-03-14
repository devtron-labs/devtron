package timeoutWindow

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

// rename to timewindowservice
type TimeoutWindowService interface {
	GetAllWithIds(ids []int) ([]*repository.TimeoutWindowConfiguration, error)
	UpdateTimeoutExpressionAndFormatForIds(tx *pg.Tx, timeoutExpression string, ids []int, expressionFormat bean.ExpressionFormat, loggedInUserId int32) error
	CreateWithTimeoutExpressionAndFormat(tx *pg.Tx, timeoutExpression string, count int, expressionFormat bean.ExpressionFormat, loggedInUserId int32) ([]*repository.TimeoutWindowConfiguration, error)
	CreateForConfigurationList(tx *pg.Tx, models []*repository.TimeoutWindowConfiguration) ([]*repository.TimeoutWindowConfiguration, error)

	CreateAndMapWithResource(tx *pg.Tx, timeWindows []TimeWindowExpression, userid int32, resourceId int, resourceType repository.ResourceType) error
	GetMappingsForResources(resourceIds []int, resourceType repository.ResourceType) (map[int][]TimeWindowExpression, error)

	UpdateWindowMappings(windows []*TimeWindow, userId int32, err error, tx *pg.Tx, policyId int) error
	GetActiveWindow(targetTimeWithZone time.Time, windows []*TimeWindow) (bool, time.Time, *TimeWindow)
	GetWindowsForResources(resourceId []int, resourceType repository.ResourceType) (map[int][]*TimeWindow, error)
}

func (impl TimeWindowServiceImpl) CreateForConfigurationList(tx *pg.Tx, configurations []*repository.TimeoutWindowConfiguration) ([]*repository.TimeoutWindowConfiguration, error) {
	return impl.timeWindowRepository.CreateInBatch(tx, configurations)
}

type TimeWindowServiceImpl struct {
	logger                      *zap.SugaredLogger
	timeWindowRepository        repository.TimeWindowRepository
	timeWindowMappingRepository repository.TimeoutWindowResourceMappingRepository
}

func NewTimeWindowServiceImpl(logger *zap.SugaredLogger,
	timeWindowRepository repository.TimeWindowRepository,
	timeWindowMappingRepository repository.TimeoutWindowResourceMappingRepository,
) *TimeWindowServiceImpl {
	timeoutWindowServiceImpl := &TimeWindowServiceImpl{
		logger:                      logger,
		timeWindowRepository:        timeWindowRepository,
		timeWindowMappingRepository: timeWindowMappingRepository,
	}
	return timeoutWindowServiceImpl
}

func (impl TimeWindowServiceImpl) GetAllWithIds(ids []int) ([]*repository.TimeoutWindowConfiguration, error) {
	timeWindows, err := impl.timeWindowRepository.GetWithIds(ids)
	if err != nil {
		impl.logger.Errorw("error in GetAllWithIds", "err", err, "timeWindowIds", ids)
		return nil, err
	}
	return timeWindows, err
}

func (impl TimeWindowServiceImpl) UpdateTimeoutExpressionAndFormatForIds(tx *pg.Tx, timeoutExpression string, ids []int, expressionFormat bean.ExpressionFormat, loggedInUserId int32) error {
	err := impl.timeWindowRepository.UpdateTimeoutExpressionAndFormatForIds(tx, timeoutExpression, ids, expressionFormat, loggedInUserId)
	if err != nil {
		impl.logger.Errorw("error in UpdateTimeoutExpressionForIds", "err", err, "timeoutExpression", timeoutExpression)
		return err
	}
	return err
}

func (impl TimeWindowServiceImpl) CreateWithTimeoutExpressionAndFormat(tx *pg.Tx, timeoutExpression string, count int, expressionFormat bean.ExpressionFormat, loggedInUserId int32) ([]*repository.TimeoutWindowConfiguration, error) {
	var models []*repository.TimeoutWindowConfiguration
	for i := 0; i < count; i++ {
		model := &repository.TimeoutWindowConfiguration{
			TimeoutWindowExpression: timeoutExpression,
			ExpressionFormat:        expressionFormat,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: loggedInUserId,
				UpdatedOn: time.Now(),
				UpdatedBy: loggedInUserId,
			},
		}
		models = append(models, model)
	}
	// create in batch
	models, err := impl.timeWindowRepository.CreateInBatch(tx, models)
	if err != nil {
		impl.logger.Errorw("error in CreateWithTimeoutExpression", "err", err, "timeoutExpression", timeoutExpression, "countToBeCreated", count)
		return nil, err
	}
	return models, nil

}

func (impl TimeWindowServiceImpl) GetMappingsForResources(resourceIds []int, resourceType repository.ResourceType) (map[int][]TimeWindowExpression, error) {

	if len(resourceIds) == 0 {
		return make(map[int][]TimeWindowExpression), nil
	}

	resourceMappings, err := impl.timeWindowMappingRepository.GetWindowsForResources(resourceIds, resourceType)
	if err != nil {
		return nil, err
	}

	resourceIdToMappings := make(map[int][]*repository.TimeoutWindowResourceMapping)
	windowIds := make([]int, 0)
	for _, mapping := range resourceMappings {
		resourceIdToMappings[mapping.ResourceId] = append(resourceIdToMappings[mapping.ResourceId], mapping)
		windowIds = append(windowIds, mapping.TimeoutWindowId)
	}

	//resourceIdToMappings = lo.GroupBy(resourceMappings, func(item *repository.TimeoutWindowResourceMapping) int {
	//	return item.ResourceId
	//})

	//windowIds := lo.Map(resourceMappings,
	//	func(mapping *repository.TimeoutWindowResourceMapping, index int) int {
	//		return mapping.TimeoutWindowId
	//	})

	// length check inside

	allConfigurations, err := impl.GetAllWithIds(windowIds)
	if err != nil {
		return nil, err
	}

	windowIdToWindowConfiguration := make(map[int]*repository.TimeoutWindowConfiguration)
	for _, configuration := range allConfigurations {
		windowIdToWindowConfiguration[configuration.Id] = configuration
	}

	resourceIdToTimeWindowExpressions := make(map[int][]TimeWindowExpression)
	for _, resourceId := range resourceIds {
		mappings := resourceIdToMappings[resourceId]
		expressions := make([]TimeWindowExpression, 0)
		for _, mapping := range mappings {
			conf := windowIdToWindowConfiguration[mapping.TimeoutWindowId]
			expressions = append(expressions, TimeWindowExpression{
				TimeoutExpression: conf.TimeoutWindowExpression,
				ExpressionFormat:  conf.ExpressionFormat,
			})
		}
		resourceIdToTimeWindowExpressions[resourceId] = expressions
	}
	return resourceIdToTimeWindowExpressions, nil
}

func (impl TimeWindowServiceImpl) CreateAndMapWithResource(tx *pg.Tx, timeWindows []TimeWindowExpression, userId int32, resourceId int, resourceType repository.ResourceType) error {

	//Delete all existing mappings for the resource
	err := impl.timeWindowMappingRepository.DeleteAllForResource(tx, resourceId, resourceType)
	if err != nil {
		return err
	}

	if len(timeWindows) == 0 {
		return nil
	}
	// Create time window configurations and add new mappings for resource if provided
	configurations := make([]*repository.TimeoutWindowConfiguration, 0)
	for _, window := range timeWindows {
		configurations = append(configurations, window.toTimeWindowDto(userId))
	}

	configurations, err = impl.CreateForConfigurationList(tx, configurations)
	if err != nil {
		return err
	}

	mappings := make([]*repository.TimeoutWindowResourceMapping, 0)
	for _, conf := range configurations {
		mappings = append(mappings, &repository.TimeoutWindowResourceMapping{
			TimeoutWindowId: conf.Id,
			ResourceId:      resourceId,
			ResourceType:    resourceType,
		})
	}

	//mappings := lo.Map(configurations, func(conf *repository.TimeoutWindowConfiguration, index int) *repository.TimeoutWindowResourceMapping {
	//	return &repository.TimeoutWindowResourceMapping{
	//		TimeoutWindowId: conf.Id,
	//		ResourceId:      resourceId,
	//		ResourceType:    resourceType,
	//	}
	//})

	_, err = impl.timeWindowMappingRepository.Create(tx, mappings)
	return err
}

func (impl TimeWindowServiceImpl) UpdateWindowMappings(windows []*TimeWindow, userId int32, err error, tx *pg.Tx, policyId int) error {

	for _, window := range windows {
		err := window.toTimeRange().ValidateTimeRange()
		if err != nil {
			return fmt.Errorf("validation falied %v", err)
		}
	}

	windowExpressions := make([]TimeWindowExpression, 0)
	for _, window := range windows {
		windowExpressions = append(windowExpressions, TimeWindowExpression{
			TimeoutExpression: window.toJsonString(),
			ExpressionFormat:  bean.RecurringTimeRange,
		})
	}

	//create time windows and map
	err = impl.CreateAndMapWithResource(tx, windowExpressions, userId, policyId, repository.DeploymentWindowProfile)
	return err
}

func (impl TimeWindowServiceImpl) GetActiveWindow(targetTimeWithZone time.Time, windows []*TimeWindow) (bool, time.Time, *TimeWindow) {
	isActive := false
	maxEndTimeStamp := time.Time{}
	minStartTimeStamp := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	var appliedWindow *TimeWindow
	for _, window := range windows {
		timeRange := window.toTimeRange()
		timestamp, isInside, err := timeRange.GetTimeRangeWindow(targetTimeWithZone)
		if err != nil {
			impl.logger.Errorw("GetTimeRangeWindow failed", "timeRange", timeRange, "window", window, "time", targetTimeWithZone)
			continue
		}
		if isInside && !timestamp.IsZero() {
			isActive = true
			if timestamp.After(maxEndTimeStamp) {
				maxEndTimeStamp = timestamp
				appliedWindow = window
			}
		} else if !isActive && !isInside && !timestamp.IsZero() {
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

func (impl TimeWindowServiceImpl) GetWindowsForResources(resourceIds []int, resourceType repository.ResourceType) (map[int][]*TimeWindow, error) {
	//get windows
	resourceIdToExpressions, err := impl.GetMappingsForResources(resourceIds, resourceType)
	if err != nil {
		return nil, err
	}

	resourceIdToWindows := make(map[int][]*TimeWindow, 0)

	for resourceId, expressions := range resourceIdToExpressions {
		resourceIdToWindows[resourceId] = impl.toTimeWindow(expressions)
	}
	return resourceIdToWindows, nil
}

func (impl TimeWindowServiceImpl) toTimeWindow(expressions []TimeWindowExpression) []*TimeWindow {

	windows := make([]*TimeWindow, 0)
	for _, expr := range expressions {
		window := &TimeWindow{}
		window.setFromJsonString(expr.TimeoutExpression)
		windows = append(windows, window)
	}
	return windows
}
