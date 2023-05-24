package bitbucket

type users interface {
	Get(username string) (interface{}, error)
	Followers(username string) (interface{}, error)
	Following(username string) (interface{}, error)
	Repositories(username string) (interface{}, error)
}

type user interface {
	Profile() (*User, error)
	Emails() (interface{}, error)
}

type pullrequests interface {
	Create(opt PullRequestsOptions) (interface{}, error)
	Update(opt PullRequestsOptions) (interface{}, error)
	List(opt PullRequestsOptions) (interface{}, error)
	Get(opt PullRequestsOptions) (interface{}, error)
	Activities(opt PullRequestsOptions) (interface{}, error)
	Activity(opt PullRequestsOptions) (interface{}, error)
	Commits(opt PullRequestsOptions) (interface{}, error)
	Patch(opt PullRequestsOptions) (interface{}, error)
	Diff(opt PullRequestsOptions) (interface{}, error)
	Merge(opt PullRequestsOptions) (interface{}, error)
	Decline(opt PullRequestsOptions) (interface{}, error)
}
type workspace interface {
	GetProject(opt ProjectOptions) (*Project, error)
	CreateProject(opt ProjectOptions) (*Project, error)
}

type issues interface {
	Gets(io *IssuesOptions) (interface{}, error)
	Get(io *IssuesOptions) (interface{}, error)
	Delete(io *IssuesOptions) (interface{}, error)
	Update(io *IssuesOptions) (interface{}, error)
	Create(io *IssuesOptions) (interface{}, error)
	GetVote(io *IssuesOptions) (bool, interface{}, error)
	PutVote(io *IssuesOptions) error
	DeleteVote(io *IssuesOptions) error
	GetWatch(io *IssuesOptions) (bool, interface{}, error)
	PutWatch(io *IssuesOptions) error
	DeleteWatch(io *IssuesOptions) error
	GetComments(ico *IssueCommentsOptions) (interface{}, error)
	CreateComment(ico *IssueCommentsOptions) (interface{}, error)
	GetComment(ico *IssueCommentsOptions) (interface{}, error)
	UpdateComment(ico *IssueCommentsOptions) (interface{}, error)
	DeleteComment(ico *IssueCommentsOptions) (interface{}, error)
	GetChanges(ico *IssueChangesOptions) (interface{}, error)
	CreateChange(ico *IssueChangesOptions) (interface{}, error)
	GetChange(ico *IssueChangesOptions) (interface{}, error)
}

type repository interface {
	Get(opt RepositoryOptions) (*Repository, error)
	Create(opt RepositoryOptions) (*Repository, error)
	Delete(opt RepositoryOptions) (interface{}, error)
	ListWatchers(opt RepositoryOptions) (interface{}, error)
	ListForks(opt RepositoryOptions) (interface{}, error)
	ListDefaultReviewers(opt RepositoryOptions) (*DefaultReviewers, error)
	GetDefaultReviewer(opt RepositoryDefaultReviewerOptions) (*DefaultReviewer, error)
	AddDefaultReviewer(opt RepositoryDefaultReviewerOptions) (*DefaultReviewer, error)
	DeleteDefaultReviewer(opt RepositoryDefaultReviewerOptions) (interface{}, error)
	UpdatePipelineConfig(opt RepositoryPipelineOptions) (*Pipeline, error)
	ListPipelineVariables(opt RepositoryPipelineVariablesOptions) (*PipelineVariables, error)
	AddPipelineVariable(opt RepositoryPipelineVariableOptions) (*PipelineVariable, error)
	DeletePipelineVariable(opt RepositoryPipelineVariableDeleteOptions) (interface{}, error)
	AddPipelineKeyPair(opt RepositoryPipelineKeyPairOptions) (*PipelineKeyPair, error)
	UpdatePipelineBuildNumber(opt RepositoryPipelineBuildNumberOptions) (*PipelineBuildNumber, error)
	ListFiles(opt RepositoryFilesOptions) (*[]RepositoryFile, error)
	GetFileBlob(opt RepositoryBlobOptions) (*RepositoryBlob, error)
	ListBranches(opt RepositoryBranchOptions) (*RepositoryBranches, error)
	BranchingModel(opt RepositoryBranchingModelOptions) (*BranchingModel, error)
	ListEnvironments(opt RepositoryEnvironmentsOptions) (*Environments, error)
	AddEnvironment(opt RepositoryEnvironmentOptions) (*Environment, error)
	DeleteEnvironment(opt RepositoryEnvironmentDeleteOptions) (interface{}, error)
	GetEnvironment(opt RepositoryEnvironmentOptions) (*Environment, error)
	ListDeploymentVariables(opt RepositoryDeploymentVariablesOptions) (*DeploymentVariables, error)
	AddDeploymentVariable(opt RepositoryDeploymentVariableOptions) (*DeploymentVariable, error)
	DeleteDeploymentVariable(opt RepositoryDeploymentVariableDeleteOptions) (interface{}, error)
	UpdateDeploymentVariable(opt RepositoryDeploymentVariableOptions) (*DeploymentVariable, error)
}

type repositories interface {
	ListForAccount(opt RepositoriesOptions) (interface{}, error)
	ListForTeam(opt RepositoriesOptions) (interface{}, error)
	ListPublic() (interface{}, error)
}

type commits interface {
	GetCommits(opt CommitsOptions) (interface{}, error)
	GetCommit(opt CommitsOptions) (interface{}, error)
	GetCommitComments(opt CommitsOptions) (interface{}, error)
	GetCommitComment(opt CommitsOptions) (interface{}, error)
	GetCommitStatus(opt CommitsOptions) (interface{}, error)
	GiveApprove(opt CommitsOptions) (interface{}, error)
	RemoveApprove(opt CommitsOptions) (interface{}, error)
	CreateCommitStatus(cmo CommitsOptions, cso CommitStatusOptions) (interface{}, error)
}

type branchrestrictions interface {
	Gets(opt BranchRestrictionsOptions) (interface{}, error)
	Get(opt BranchRestrictionsOptions) (interface{}, error)
	Create(opt BranchRestrictionsOptions) (interface{}, error)
	Update(opt BranchRestrictionsOptions) (interface{}, error)
	Delete(opt BranchRestrictionsOptions) (interface{}, error)
}

type diff interface {
	GetDiff(opt DiffOptions) (interface{}, error)
	GetPatch(opt DiffOptions) (interface{}, error)
}

type webhooks interface {
	Gets(opt WebhooksOptions) (interface{}, error)
	Get(opt WebhooksOptions) (interface{}, error)
	Create(opt WebhooksOptions) (interface{}, error)
	Update(opt WebhooksOptions) (interface{}, error)
	Delete(opt WebhooksOptions) (interface{}, error)
}

type teams interface {
	List(role string) (interface{}, error) // [WIP?] role=[admin|contributor|member]
	Profile(teamname string) (interface{}, error)
	Members(teamname string) (interface{}, error)
	Followers(teamname string) (interface{}, error)
	Following(teamname string) (interface{}, error)
	Repositories(teamname string) (interface{}, error)
	Projects(teamname string) (interface{}, error)
}

type pipelines interface {
	List(po *PipelinesOptions) (interface{}, error)
	Get(po *PipelinesOptions) (interface{}, error)
	ListSteps(po *PipelinesOptions) (interface{}, error)
	GetStep(po *PipelinesOptions) (interface{}, error)
	GetLog(po *PipelinesOptions) (string, error)
}

type RepositoriesOptions struct {
	Owner string `json:"owner"`
	Role  string `json:"role"` // role=[owner|admin|contributor|member]
}

type RepositoryOptions struct {
	Uuid     string `json:"uuid"`
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Scm      string `json:"scm"`
	//	Name        string `json:"name"`
	IsPrivate   string `json:"is_private"`
	Description string `json:"description"`
	ForkPolicy  string `json:"fork_policy"`
	Language    string `json:"language"`
	HasIssues   string `json:"has_issues"`
	HasWiki     string `json:"has_wiki"`
	Project     string `json:"project"`
}

type RepositoryForkOptions struct {
	FromOwner string `json:"from_owner"`
	FromSlug  string `json:"from_slug"`
	Owner     string `json:"owner"`
	// TODO: does the API supports specifying  slug on forks?
	// see: https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Bworkspace%7D/%7Brepo_slug%7D/forks#post
	Name        string `json:"name"`
	IsPrivate   string `json:"is_private"`
	Description string `json:"description"`
	ForkPolicy  string `json:"fork_policy"`
	Language    string `json:"language"`
	HasIssues   string `json:"has_issues"`
	HasWiki     string `json:"has_wiki"`
	Project     string `json:"project"`
}

type RepositoryFilesOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Ref      string `json:"ref"`
	Path     string `json:"path"`
}

type RepositoryBlobOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Ref      string `json:"ref"`
	Path     string `json:"path"`
}

// Based on https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Bworkspace%7D/%7Brepo_slug%7D/src#post
type RepositoryBlobWriteOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
	Author   string `json:"author"`
	Message  string `json:"message"`
	Branch   string `json:"branch"`
}

// RepositoryRefOptions represents the options for describing a repository's refs (i.e.
// tags and branches). The field BranchFlg is a boolean that is indicates whether a specific
// RepositoryRefOptions instance is meant for Branch specific set of api methods.
type RepositoryRefOptions struct {
	Owner     string `json:"owner"`
	RepoSlug  string `json:"repo_slug"`
	Query     string `json:"query"`
	Sort      string `json:"sort"`
	PageNum   int    `json:"page"`
	Pagelen   int    `json:"pagelen"`
	MaxDepth  int    `json:"max_depth"`
	Name      string `json:"name"`
	BranchFlg bool
}

type RepositoryBranchOptions struct {
	Owner      string `json:"owner"`
	RepoSlug   string `json:"repo_slug"`
	Query      string `json:"query"`
	Sort       string `json:"sort"`
	PageNum    int    `json:"page"`
	Pagelen    int    `json:"pagelen"`
	MaxDepth   int    `json:"max_depth"`
	BranchName string `json:"branch_name"`
}

type RepositoryBranchCreationOptions struct {
	Owner    string                 `json:"owner"`
	RepoSlug string                 `json:"repo_slug"`
	Name     string                 `json:"name"`
	Target   RepositoryBranchTarget `json:"target"`
}

type RepositoryBranchDeleteOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	RepoUUID string `json:"uuid"`
	RefName  string `json:"name"`
	RefUUID  string `json:uuid`
}

type RepositoryBranchTarget struct {
	Hash string `json:"hash"`
}

type RepositoryTagOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Query    string `json:"q"`
	Sort     string `json:"sort"`
	PageNum  int    `json:"page"`
	Pagelen  int    `json:"pagelen"`
	MaxDepth int    `json:"max_depth"`
}

type RepositoryTagCreationOptions struct {
	Owner    string              `json:"owner"`
	RepoSlug string              `json:"repo_slug"`
	Name     string              `json:"name"`
	Target   RepositoryTagTarget `json:"target"`
}

type RepositoryTagTarget struct {
	Hash string `json:"hash"`
}

type PullRequestsOptions struct {
	ID                string   `json:"id"`
	CommentID         string   `json:"comment_id"`
	Owner             string   `json:"owner"`
	RepoSlug          string   `json:"repo_slug"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	CloseSourceBranch bool     `json:"close_source_branch"`
	SourceBranch      string   `json:"source_branch"`
	SourceRepository  string   `json:"source_repository"`
	DestinationBranch string   `json:"destination_branch"`
	DestinationCommit string   `json:"destination_repository"`
	Message           string   `json:"message"`
	Reviewers         []string `json:"reviewers"`
	States            []string `json:"states"`
	Query             string   `json:"query"`
	Sort              string   `json:"sort"`
}

type PullRequestCommentOptions struct {
	Owner         string `json:"owner"`
	RepoSlug      string `json:"repo_slug"`
	PullRequestID string `json:"id"`
	Content       string `json:"content"`
	CommentId     string `json:"-"`
}

type IssuesOptions struct {
	ID        string   `json:"id"`
	Owner     string   `json:"owner"`
	RepoSlug  string   `json:"repo_slug"`
	States    []string `json:"states"`
	Query     string   `json:"query"`
	Sort      string   `json:"sort"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	State     string   `json:"state"`
	Kind      string   `json:"kind"`
	Milestone string   `json:"milestone"`
	Component string   `json:"component"`
	Priority  string   `json:"priority"`
	Version   string   `json:"version"`
	Assignee  string   `json:"assignee"`
}

type IssueCommentsOptions struct {
	IssuesOptions
	Query          string `json:"query"`
	Sort           string `json:"sort"`
	CommentContent string `json:"comment_content"`
	CommentID      string `json:"comment_id"`
}

type IssueChangesOptions struct {
	IssuesOptions
	Query    string `json:"query"`
	Sort     string `json:"sort"`
	Message  string `json:"message"`
	ChangeID string `json:"change_id"`
	Changes  []struct {
		Type     string
		NewValue string
	} `json:"changes"`
}

type CommitsOptions struct {
	Owner       string `json:"owner"`
	RepoSlug    string `json:"repo_slug"`
	Revision    string `json:"revision"`
	Branchortag string `json:"branchortag"`
	Include     string `json:"include"`
	Exclude     string `json:"exclude"`
	CommentID   string `json:"comment_id"`
}

type CommitStatusOptions struct {
	Key         string `json:"key"`
	Url         string `json:"url"`
	State       string `json:"state"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BranchRestrictionsOptions struct {
	Owner    string            `json:"owner"`
	RepoSlug string            `json:"repo_slug"`
	ID       string            `json:"id"`
	Groups   map[string]string `json:"groups"`
	Pattern  string            `json:"pattern"`
	Users    []string          `json:"users"`
	Kind     string            `json:"kind"`
	FullSlug string            `json:"full_slug"`
	Name     string            `json:"name"`
	Value    interface{}       `json:"value"`
}

type DiffOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Spec     string `json:"spec"`
}

type DiffStatOptions struct {
	Owner      string `json:"owner"`
	RepoSlug   string `json:"repo_slug"`
	Spec       string `json:"spec"`
	Whitespace bool   `json:"ignore_whitespace"`
	Merge      bool   `json:"merge"`
	Path       string `json:"path"`
	Renames    bool   `json:"renames"`
	PageNum    int    `json:"page"`
	Pagelen    int    `json:"pagelen"`
	MaxDepth   int    `json:"max_depth"`
	Fields     []string
}

type WebhooksOptions struct {
	Owner       string   `json:"owner"`
	RepoSlug    string   `json:"repo_slug"`
	Uuid        string   `json:"uuid"`
	Description string   `json:"description"`
	Url         string   `json:"url"`
	Active      bool     `json:"active"`
	Events      []string `json:"events"` // EX: {'repo:push','issue:created',..} REF: https://bit.ly/3FjRHHu
}

type RepositoryPipelineOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Enabled  bool   `json:"has_pipelines"`
}

type RepositoryDefaultReviewerOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Username string `json:"username"`
}

type RepositoryPipelineVariablesOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Query    string `json:"q"`
	Sort     string `json:"sort"`
	PageNum  int    `json:"page"`
	Pagelen  int    `json:"pagelen"`
	MaxDepth int    `json:"max_depth"`
}

type RepositoryPipelineVariableOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Uuid     string `json:"uuid"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Secured  bool   `json:"secured"`
}

type RepositoryPipelineVariableDeleteOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Uuid     string `json:"uuid"`
}

type RepositoryPipelineKeyPairOptions struct {
	Owner      string `json:"owner"`
	RepoSlug   string `json:"repo_slug"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

type RepositoryPipelineBuildNumberOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Next     int    `json:"next"`
}

type RepositoryBranchingModelOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
}

type DownloadsOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
}

type PageRes struct {
	Page     int32 `json:"page"`
	PageLen  int32 `json:"pagelen"`
	MaxDepth int32 `json:"max_depth"`
	Size     int32 `json:"size"`
}

type PipelinesOptions struct {
	Owner    string `json:"owner"`
	Page     int    `json:"page"`
	RepoSlug string `json:"repo_slug"`
	Query    string `json:"query"`
	Sort     string `json:"sort"`
	IDOrUuid string `json:"ID"`
	StepUuid string `json:"StepUUID"`
}

type RepositoryEnvironmentsOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
}

type RepositoryEnvironmentTypeOption int

const (
	Test RepositoryEnvironmentTypeOption = iota
	Staging
	Production
)

func (e RepositoryEnvironmentTypeOption) String() string {
	return [...]string{"Test", "Staging", "Production"}[e]
}

type RepositoryEnvironmentOptions struct {
	Owner           string                          `json:"owner"`
	RepoSlug        string                          `json:"repo_slug"`
	Uuid            string                          `json:"uuid"`
	Name            string                          `json:"name"`
	EnvironmentType RepositoryEnvironmentTypeOption `json:"environment_type"`
	Rank            int                             `json:"rank"`
}

type RepositoryEnvironmentDeleteOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Uuid     string `json:"uuid"`
}

type RepositoryDeploymentVariablesOptions struct {
	Owner       string       `json:"owner"`
	RepoSlug    string       `json:"repo_slug"`
	Environment *Environment `json:"environment"`
	Query       string       `json:"q"`
	Sort        string       `json:"sort"`
	PageNum     int          `json:"page"`
	Pagelen     int          `json:"pagelen"`
	MaxDepth    int          `json:"max_depth"`
}

type RepositoryDeploymentVariableOptions struct {
	Owner       string       `json:"owner"`
	RepoSlug    string       `json:"repo_slug"`
	Environment *Environment `json:"environment"`
	Uuid        string       `json:"uuid"`
	Key         string       `json:"key"`
	Value       string       `json:"value"`
	Secured     bool         `json:"secured"`
}

type RepositoryDeploymentVariableDeleteOptions struct {
	Owner       string       `json:"owner"`
	RepoSlug    string       `json:"repo_slug"`
	Environment *Environment `json:"environment"`
	Uuid        string       `json:"uuid"`
}

type DeployKeyOptions struct {
	Owner    string `json:"owner"`
	RepoSlug string `json:"repo_slug"`
	Id       int    `json:"id"`
	Label    string `json:"label"`
	Key      string `json:"key"`
}
