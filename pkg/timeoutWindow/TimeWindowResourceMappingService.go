package timeoutWindow

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"github.com/go-pg/pg"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"time"
)

type TimeoutWindowResourceMappingService interface {
	CreateAndMapWithResource(tx *pg.Tx, timeWindows []TimeWindowExpression, userid int32, resourceId int, resourceType repository.ResourceType) error
	GetMappingsForResources(resourceIds []int, resourceType repository.ResourceType) (map[int][]TimeWindowExpression, error)
}

type TimeoutWindowResourceMappingServiceImpl struct {
	logger                      *zap.SugaredLogger
	timeWindowMappingRepository repository.TimeoutWindowResourceMappingRepository
	timeWindowService           TimeoutWindowService
}

func NewTimeoutWindowResourceMappingServiceImpl(logger *zap.SugaredLogger, timeWindowMappingRepository repository.TimeoutWindowResourceMappingRepository, timeWindowService TimeoutWindowService) *TimeoutWindowResourceMappingServiceImpl {
	return &TimeoutWindowResourceMappingServiceImpl{logger: logger, timeWindowMappingRepository: timeWindowMappingRepository, timeWindowService: timeWindowService}
}

type TimeWindowExpression struct {
	TimeoutExpression string
	ExpressionFormat  bean.ExpressionFormat
}

func (impl TimeoutWindowResourceMappingServiceImpl) GetMappingsForResources(resourceIds []int, resourceType repository.ResourceType) (map[int][]TimeWindowExpression, error) {
	resourceMappings, err := impl.timeWindowMappingRepository.GetWindowsForResources(resourceIds, resourceType)
	if err != nil {
		return nil, err
	}

	resourceIdToMappings := lo.GroupBy(resourceMappings, func(item *repository.TimeoutWindowResourceMapping) int {
		return item.ResourceId
	})

	windowIds := lo.Map(resourceMappings,
		func(mapping *repository.TimeoutWindowResourceMapping, index int) int {
			return mapping.TimeoutWindowId
		})

	// length check inside

	allConfigurations, err := impl.timeWindowService.GetAllWithIds(windowIds)
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

func (impl TimeoutWindowResourceMappingServiceImpl) CreateAndMapWithResource(tx *pg.Tx, timeWindows []TimeWindowExpression, userId int32, resourceId int, resourceType repository.ResourceType) error {

	//Delete all existing mappings for the resource
	err := impl.timeWindowMappingRepository.DeleteAllForResource(tx, resourceId, resourceType)
	if err != nil {
		return err
	}

	if len(timeWindows) == 0 {
		return nil
	}
	// Create time window configurations and add new mappings for resource if provided
	configurations := lo.Map(timeWindows,
		func(timeWindow TimeWindowExpression, index int) *repository.TimeoutWindowConfiguration {
			return timeWindow.toTimeWindowDto(userId)
		})

	configurations, err = impl.timeWindowService.CreateForConfigurationList(tx, configurations)
	if err != nil {
		return err
	}

	mappings := lo.Map(configurations, func(conf *repository.TimeoutWindowConfiguration, index int) *repository.TimeoutWindowResourceMapping {
		return &repository.TimeoutWindowResourceMapping{
			TimeoutWindowId: conf.Id,
			ResourceId:      resourceId,
			ResourceType:    resourceType,
		}
	})

	_, err = impl.timeWindowMappingRepository.Create(tx, mappings)
	return err
}

func (expr TimeWindowExpression) toTimeWindowDto(userId int32) *repository.TimeoutWindowConfiguration {
	return &repository.TimeoutWindowConfiguration{
		TimeoutWindowExpression: expr.TimeoutExpression,
		ExpressionFormat:        expr.ExpressionFormat,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
}
