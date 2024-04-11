package utils

// func TestAskQuestions(t *testing.T) {
// 	cfg := &config.Config{
// 		Types: []models.CommitType{
// 			{Name: "feat", Emoji: "🎉", Description: "Introduces a new feature"},
// 			{Name: "fix", Emoji: "🐛", Description: "Fixes a bug"},
// 		},
// 		Scopes: []string{"ci", "api", "parser"},
// 	}

// 	tests := []struct {
// 		name     string
// 		expected []string
// 	}{
// 		{
// 			name: "feat with scope and description",
// 			expected: []string{
// 				"feat (ci): new commit",
// 				"new description",
// 			},
// 		},
// 		{
// 			name: "feat with description only",
// 			expected: []string{
// 				"feat: new commit",
// 				"new description",
// 			},
// 		},
// 		{
// 			name: "feat with scope only",
// 			expected: []string{
// 				"feat (ci): new commit",
// 				"",
// 			},
// 		},
// 		{
// 			name: "feat without scope and description",
// 			expected: []string{
// 				"feat: new commit",
// 				"",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result, err := AskQuestions(cfg)
// 			if err != nil {
// 				t.Fatalf("AskQuestions failed: %v", err)
// 			}

// 			if !reflect.DeepEqual(result, tt.expected) {
// 				t.Errorf("expected %v, got %v", tt.expected, result)
// 			}
// 		})
// 	}
// }
