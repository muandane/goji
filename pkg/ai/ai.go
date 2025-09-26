package ai

type CommitResult struct {
	Message string
	Body    string
}

type AIProvider interface {
	GenerateCommitMessage(diff string, commitTypes string, extraContext string) (string, error)
	GenerateDetailedCommit(diff string, commitTypes string, extraContext string) (*CommitResult, error)
	GetModel() string
}
