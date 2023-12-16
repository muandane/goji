package utils

import (
	"testing"

	"github.com/muandane/goji/pkg/config"
	"github.com/muandane/goji/pkg/models"
)

func TestAskQuestions(t *testing.T) {
	type args struct {
		config *config.Config
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test Case 1",
			args: args{
				config: &config.Config{
					Types: []models.CommitType{
						{
							Name:        "fix",
							Description: "Fix a bug",
							Emoji:       "🐛",
						},
					},
					SkipQuestions: []string{},
				},
			},
			want:    "fix: A bug was fixed",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AskQuestions(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("AskQuestions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AskQuestions() = %v, want %v", got, tt.want)
			}
		})
	}
}
