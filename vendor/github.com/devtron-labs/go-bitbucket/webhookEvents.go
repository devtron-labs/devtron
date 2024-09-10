package bitbucket

const (
	RepoPushEvent                  string = "repo:push"
	RepoForkEvent                  string = "repo:fork"
	RepoUpdatedEvent               string = "repo:updated"
	RepoCommitCommentCreatedEvent  string = "repo:commit_comment_created"
	RepoCommitStatusCreatedEvent   string = "repo:commit_status_created"
	RepoCommitStatusUpdatedEvent   string = "repo:commit_status_updated"
	IssueCreatedEvent              string = "issue:created"
	IssueUpdatedEvent              string = "issue:updated"
	IssueCommentCreatedEvent       string = "issue:comment_created"
	PullRequestCreatedEvent        string = "pullrequest:created"
	PullRequestUpdatedEvent        string = "pullrequest:updated"
	PullRequestApprovedEvent       string = "pullrequest:approved"
	PullRequestUnapprovedEvent     string = "pullrequest:unapproved"
	PullRequestMergedEvent         string = "pullrequest:fulfilled"
	PullRequestDeclinedEvent       string = "pullrequest:rejected"
	PullRequestCommentCreatedEvent string = "pullrequest:comment_created"
	PullRequestCommentUpdatedEvent string = "pullrequest:comment_updated"
	PullRequestCommentDeletedEvent string = "pullrequest:comment_deleted"
)
