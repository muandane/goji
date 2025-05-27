package ai

type AIProvider interface {
	GenerateCommitMessage(diff string, commitTypes string) (string, error)
	GetModel() string
}
