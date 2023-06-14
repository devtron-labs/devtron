package bean

type HistoryOfAction int

const (
	HISTORY_OF_ACTION_CREATE HistoryOfAction = iota
	HISTORY_OF_ACTION_UPDATE
	HISTORY_OF_ACTION_DELETE
)
