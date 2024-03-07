package constants

type ArtifactPromotionRequestStatus int
type SearchableField string
type SortKey = string
type SortOrder = string

const (
	AWAITING_APPROVAL ArtifactPromotionRequestStatus = 1
	CANCELED          ArtifactPromotionRequestStatus = 2
	PROMOTED          ArtifactPromotionRequestStatus = 3
	STALE             ArtifactPromotionRequestStatus = 4
)

func (status ArtifactPromotionRequestStatus) Status() string {
	switch status {
	case PROMOTED:
		return "promoted"
	case CANCELED:
		return "cancelled"
	case STALE:
		return "stale"
	case AWAITING_APPROVAL:
		return "awaiting for approval"
	}
	return "deleted"
}

type SourceType int

const (
	CI      SourceType = 1
	WEBHOOK SourceType = 2
	CD      SourceType = 3
)

type SourceTypeStr string

const (
	SOURCE_TYPE_CI                      SourceTypeStr = "CI"
	SOURCE_TYPE_WEBHOOK                 SourceTypeStr = "WEBHOOK"
	SOURCE_TYPE_CD                      SourceTypeStr = "ENVIRONMENT"
	ArtifactPromotionRequestNotFoundErr               = "artifact promotion request not found"
	ACTION_PROMOTE                                    = "PROMOTE"
	ACTION_CANCEL                                     = "CANCEL"
	ACTION_APPROVE                                    = "APPROVE"
	PROMOTION_APPROVAL_PENDING_NODE     SourceTypeStr = "PROMOTION_APPROVAL_PENDING_NODE"
	UserCannotCancelRequest                           = "only user who has raised the promotion request can cancel it"
)

func (sourceType SourceTypeStr) GetSourceType() SourceType {
	switch sourceType {
	case SOURCE_TYPE_CI:
		return CI
	case SOURCE_TYPE_WEBHOOK:
		return WEBHOOK
	case SOURCE_TYPE_CD:
		return CD
	}
	return CI
}

func (sourceType SourceType) GetSourceTypeStr() SourceTypeStr {
	switch sourceType {
	case CI:
		return SOURCE_TYPE_CI
	case WEBHOOK:
		return SOURCE_TYPE_WEBHOOK
	case CD:
		return SOURCE_TYPE_CD
	}
	return SOURCE_TYPE_CI
}

const (
	POLICY_NAME_SORT_KEY        SortKey         = "policyName"
	APPROVER_COUNT_SORT_KEY     SortKey         = "approverCount"
	ASC                         SortOrder       = "ASC"
	DESC                        SortOrder       = "DESC"
	APPROVER_COUNT_SEARCH_FIELD SearchableField = "approver_count"
)

const ARTIFACT_ALREADY_PROMOTED PromotionValidationState = "already promoted"
const ALREADY_REQUEST_RAISED PromotionValidationState = "promotion request already raised"
const ERRORED PromotionValidationState = "error occurred"
const EMPTY PromotionValidationState = ""
const PIPELINE_NOT_FOUND PromotionValidationState = "pipeline Not Found"
const POLICY_NOT_CONFIGURED PromotionValidationState = "policy not configured"
const NO_PERMISSION PromotionValidationState = "no permission"
const PROMOTION_SUCCESSFUL PromotionValidationState = "image promoted"
const SENT_FOR_APPROVAL PromotionValidationState = "sent for approval"
const SOURCE_AND_DESTINATION_PIPELINE_MISMATCH PromotionValidationState = "source and destination pipeline order mismatch"
const POLICY_EVALUATION_ERRORED PromotionValidationState = "server unable to evaluate the policy"
const BLOCKED_BY_POLICY PromotionValidationState = "blocked by the policy "
const APPROVED PromotionValidationState = "approved"

type PromotionValidationState string

const ALREADY_APPROVED PromotionValidationState = "you have already approved this"
const ERRORED_APPROVAL PromotionValidationState = "error occurred in submitting the approval"
