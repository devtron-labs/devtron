package timeoutWindow

// Move to TimeWindowService
//type RepeatingTimeWindowService interface {
//	TimeoutWindowService
//	//UpdateWindowMappings(windows []*TimeWindow, userId int32, err error, tx *pg.Tx, policyId int) error
//	//GetActiveWindow(targetTimeWithZone time.Time, windows []*TimeWindow) (bool, time.Time, *TimeWindow)
//	//GetWindowsForResources(resourceId []int, resourceType repository.ResourceType) (map[int][]*TimeWindow, error)
//}
//
//type RepeatingTimeWindowServiceImpl struct {
//	*TimeWindowServiceImpl
//}
//
//func NewRepeatingTimeWindowServiceImpl(service *TimeWindowServiceImpl) *RepeatingTimeWindowServiceImpl {
//	return &RepeatingTimeWindowServiceImpl{TimeWindowServiceImpl: service}
//}
//
