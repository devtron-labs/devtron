package resourceFilter

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"time"
)

type FilterHistoryObject struct {
	FilterHistoryId int         `json:"filter_history_id"`
	State           FilterState `json:"state"`
	Message         string      `json:"message"`
}

type FilterEvaluationAuditService interface {
	CreateFilterEvaluation(subjectType SubjectType, subjectId int, refType ReferenceType, refId int, filters []*FilterMetaDataBean, filterIdVsState map[int]FilterState) (*ResourceFilterEvaluationAudit, error)
	UpdateFilterEvaluationAuditRef(id int, refType ReferenceType, refId int) error
	GetFilterEvaluationAudits()
}

type FilterEvaluationAuditServiceImpl struct {
	logger                    *zap.SugaredLogger
	filterEvaluationAuditRepo FilterEvaluationAuditRepository
	filterAuditRepo           FilterAuditRepository
}

func NewFilterEvaluationAuditServiceImpl(logger *zap.SugaredLogger,
	filterEvaluationAuditRepo FilterEvaluationAuditRepository,
	filterAuditRepo FilterAuditRepository) *FilterEvaluationAuditServiceImpl {
	return &FilterEvaluationAuditServiceImpl{
		logger:                    logger,
		filterEvaluationAuditRepo: filterEvaluationAuditRepo,
		filterAuditRepo:           filterAuditRepo,
	}
}

func (impl *FilterEvaluationAuditServiceImpl) CreateFilterEvaluation(subjectType SubjectType, subjectId int, refType ReferenceType, refId int, filters []*FilterMetaDataBean, filterIdVsState map[int]FilterState) (*ResourceFilterEvaluationAudit, error) {
	filterHistoryObjectsStr, err := impl.extractFilterHistoryObjects(filters, filterIdVsState)
	if err != nil {
		impl.logger.Errorw("error in extracting filter history objects", "err", err, "filters", filters, "filterIdVsState", filterIdVsState)
		return nil, err
	}

	currentTime := time.Now()
	auditLog := sql.AuditLog{
		CreatedOn: currentTime,
		UpdatedOn: currentTime,
	}

	filterEvaluationAudit := NewResourceFilterEvaluationAudit(&refType, refId, filterHistoryObjectsStr, &subjectType, subjectId, auditLog)
	savedFilterEvaluationAudit, err := impl.filterEvaluationAuditRepo.Create(&filterEvaluationAudit)
	if err != nil {
		impl.logger.Errorw("error in saving resource filter evaluation result in resource_filter_evaluation_audit table", "err", err, "filterEvaluationAudit", filterEvaluationAudit)
		return savedFilterEvaluationAudit, err
	}
	return savedFilterEvaluationAudit, nil
}

func (impl *FilterEvaluationAuditServiceImpl) UpdateFilterEvaluationAuditRef(id int, refType ReferenceType, refId int) error {
	return impl.filterEvaluationAuditRepo.UpdateRefTypeAndRefId(id, refType, refId)
}

func (impl *FilterEvaluationAuditServiceImpl) GetFilterEvaluationAudits() {

}

func (impl *FilterEvaluationAuditServiceImpl) extractFilterHistoryObjects(filters []*FilterMetaDataBean, filterIdVsState map[int]FilterState) (string, error) {
	filterIds := make([]int, 0)
	//store filtersMap here, later will help to identify filters that doesn't have filterAudit
	filtersMap := make(map[int]*FilterMetaDataBean)
	filterHistoryObjectMap := make(map[int]*FilterHistoryObject)
	for _, filter := range filters {
		filterIds = append(filterIds, filter.Id)
		filtersMap[filter.Id] = filter
		message := ""
		for _, condition := range filter.Conditions {
			message = fmt.Sprintf("\n%s conditionType : %v , errorMsg : %v", message, condition.ConditionType, condition.ErrorMsg)
		}
		filterHistoryObjectMap[filter.Id] = &FilterHistoryObject{
			State:   filterIdVsState[filter.Id],
			Message: message,
		}
	}

	resourceFilterEvaluationAudits, err := impl.filterAuditRepo.GetLatestResourceFilterAuditByFilterIds(filterIds)
	if err != nil {
		impl.logger.Errorw("error in getting latest resource filter audits for given filter id's", "filterIds", filterIds, "err", err)
		return "", err
	}

	for _, resourceFilterEvaluationAudit := range resourceFilterEvaluationAudits {
		if filterHistoryObject, ok := filterHistoryObjectMap[resourceFilterEvaluationAudit.FilterId]; ok {
			filterHistoryObject.FilterHistoryId = resourceFilterEvaluationAudit.Id

			//delete filter from filtersMap for which we found filter audit
			delete(filtersMap, resourceFilterEvaluationAudit.FilterId)
		}
	}

	//if filtersMap is not empty ,there are some filters for which we never stored audit entry, so create filter audit for those
	if len(filtersMap) > 0 {
		filterHistoryObjectMap, err = impl.createFilterAuditForMissingFilters(filtersMap, filterHistoryObjectMap)
		if err != nil {
			impl.logger.Errorw("error in creating filter audit data for missing filters", "missingFiltersMap", filtersMap, "err", err)
			return "", err
		}
	}

	filterHistoryObjects := make([]*FilterHistoryObject, 0, len(filterHistoryObjectMap))
	for _, val := range filterHistoryObjectMap {
		filterHistoryObjects = append(filterHistoryObjects, val)
	}
	jsonStr, err := getJsonStringFromFilterHistoryObjects(filterHistoryObjects)
	if err != nil {
		impl.logger.Errorw("error in getting json string for filter history objects", "filterHistoryObjects", filterHistoryObjects, "err", err)
		return "", err
	}
	return jsonStr, err

}

// createFilterAuditForMissingFilters will create snapshot of filter data in filter audit table and gets updated filterHistoryObjectMap.
// this function exists because filter auditing is added later, so there is possibility that filters exist without any auditing data
func (impl *FilterEvaluationAuditServiceImpl) createFilterAuditForMissingFilters(filtersMap map[int]*FilterMetaDataBean, filterHistoryObjectMap map[int]*FilterHistoryObject) (map[int]*FilterHistoryObject, error) {
	tx, err := impl.filterAuditRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting db transaction", "err", err)
		return filterHistoryObjectMap, err
	}

	defer impl.filterAuditRepo.RollbackTx(tx)

	for _, filter := range filtersMap {
		conditionsStr, err := getJsonStringFromResourceCondition(filter.Conditions)
		if err != nil {
			impl.logger.Errorw("error in getting json string from filter conditions", "err", err, "filterConditions", filter.Conditions)
			return filterHistoryObjectMap, err
		}
		action := Create
		userId := int32(1) //system user
		filterAudit := NewResourceFilterAudit(filter.Id, conditionsStr, filter.TargetObject, &action, userId)
		savedFilterAudit, err := impl.filterAuditRepo.CreateResourceFilterAudit(tx, &filterAudit)
		if err != nil {
			impl.logger.Errorw("error in creating filter audit for missing filters", "err", err, "filterAudit", filterAudit)
			return filterHistoryObjectMap, err
		}

		if filterHistoryObject, ok := filterHistoryObjectMap[savedFilterAudit.FilterId]; ok {
			filterHistoryObject.FilterHistoryId = savedFilterAudit.Id
		}

	}
	err = impl.filterAuditRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing db transaction", "err", err)
		return filterHistoryObjectMap, err
	}

	return filterHistoryObjectMap, err
}
