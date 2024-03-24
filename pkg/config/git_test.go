package config

import "testing"

func TestGitRepo(t *testing.T) {
	t.Run("repo dir is found", func(t *testing.T) {
		_, err := GitRepo()
		if err != nil {
			t.Errorf("Error finding git root directory: %v", err)

		}
	})
}
