package ai

type AIProvider interface {
	GenerateCommitMessage(diff string, commitTypes string, extraContext string) (string, error)
	GetModel() string
}
