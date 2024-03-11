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
	SOURCE_TYPE_CI                  SourceTypeStr = "CI"
	SOURCE_TYPE_WEBHOOK             SourceTypeStr = "WEBHOOK"
	SOURCE_TYPE_CD                  SourceTypeStr = "ENVIRONMENT"
	PROMOTION_APPROVAL_PENDING_NODE SourceTypeStr = "PROMOTION_APPROVAL_PENDING_NODE"
)

type RequestAction string

const (
	ACTION_PROMOTE RequestAction = "PROMOTE"
	ACTION_CANCEL  RequestAction = "CANCEL"
	ACTION_APPROVE RequestAction = "APPROVE"
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

type PromotionValidationState string

const INFO PromotionValidationState = "INFO"
const PENDING PromotionValidationState = "PENDING"
const SUCCESS PromotionValidationState = "SUCCESS"
const ERROR PromotionValidationState = "ERROR"

type PromotionValidationMsg string

// info msgs
const ARTIFACT_ALREADY_PROMOTED PromotionValidationMsg = "already promoted"
const ALREADY_REQUEST_RAISED PromotionValidationMsg = "promotion request already raised"
const ALREADY_APPROVED PromotionValidationMsg = "you have already approved this"

// error messages
const ERRORED PromotionValidationMsg = "error occurred"
const PIPELINE_NOT_FOUND PromotionValidationMsg = "pipeline Not Found"
const POLICY_NOT_CONFIGURED PromotionValidationMsg = "policy not configured"
const NO_PERMISSION PromotionValidationMsg = "no permission"
const SOURCE_AND_DESTINATION_PIPELINE_MISMATCH PromotionValidationMsg = "source and destination pipeline order mismatch"
const POLICY_EVALUATION_ERRORED PromotionValidationMsg = "server unable to evaluate the policy"
const BLOCKED_BY_POLICY PromotionValidationMsg = "blocked by the policy "
const ERRORED_APPROVAL PromotionValidationMsg = "error occurred in submitting the approval"

const SENT_FOR_APPROVAL PromotionValidationMsg = "sent for approval"

const APPROVED PromotionValidationMsg = "approved"
const PROMOTION_SUCCESSFUL PromotionValidationMsg = "image promoted"
const EMPTY PromotionValidationMsg = ""

func (pvm PromotionValidationMsg) GetValidationState() PromotionValidationState {
	switch pvm {
	case ARTIFACT_ALREADY_PROMOTED, ALREADY_REQUEST_RAISED, ALREADY_APPROVED:
		return INFO
	case ERRORED_APPROVAL, BLOCKED_BY_POLICY, POLICY_EVALUATION_ERRORED, SOURCE_AND_DESTINATION_PIPELINE_MISMATCH, POLICY_NOT_CONFIGURED, PIPELINE_NOT_FOUND, ERRORED:
		return ERROR
	case SENT_FOR_APPROVAL:
		return PENDING
	case APPROVED, PROMOTION_SUCCESSFUL, EMPTY:
		return SUCCESS
	default:
		return ERROR
	}
}

const BUILD_TRIGGER_USER_CANNOT_APPROVE_MSG = "User who has built the image cannot approve promotion request for this environment"
const PROMOTION_REQUESTED_BY_USER_CANNOT_APPROVE_MSG = "User who has raised the promotion request cannot approve for this environment"
const USER_DOES_NOT_HAVE_ARTIFACT_PROMOTER_ACCESS = "user does not have image promoter access for given app and env"
const ARTIFACT_NOT_FOUND_ERR = "artifact not found for given id"
const ArtifactPromotionRequestNotFoundErr = "artifact promotion request not found"
const UserCannotCancelRequest = "only user who has raised the promotion request can cancel it"
