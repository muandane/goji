// package utils

// import (
// 	"reflect"
// 	"testing"

// 	"github.com/muandane/goji/pkg/config"
// )

// func TestAskQuestions(t *testing.T) {
// 	type args struct {
// 		commitType        string
// 		commitScope       string
// 		commitSubject     string
// 		commitDescription string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want []string
// 	}{
// 		{
// 			name: "both scope and description are filled",
// 			args: args{
// 				commitType:        "feat",
// 				commitScope:       "login",
// 				commitSubject:     "add login functionality",
// 				commitDescription: "added login, signup and forgot password functionalities",
// 			},
// 			want: []string{"feat (login): add login functionality", "added login, signup and forgot password functionalities"},
// 		},
// 		{
// 			name: "only description is filled",
// 			args: args{
// 				commitType:        "feat",
// 				commitScope:       "",
// 				commitSubject:     "add login functionality",
// 				commitDescription: "added login, signup and forgot password functionalities",
// 			},
// 			want: []string{"feat: add login functionality", "added login, signup and forgot password functionalities"},
// 		},
// 		{
// 			name: "only scope is filled",
// 			args: args{
// 				commitType:        "feat",
// 				commitScope:       "login",
// 				commitSubject:     "add login functionality",
// 				commitDescription: "",
// 			},
// 			want: []string{"feat (login): add login functionality"},
// 		},
// 		{
// 			name: "neither scope nor description is filled",
// 			args: args{
// 				commitType:        "feat",
// 				commitScope:       "",
// 				commitSubject:     "add login functionality",
// 				commitDescription: "",
// 			},
// 			want: []string{"feat: add login functionality"},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, _ := AskQuestions(&config.Config{CommitType: tt.args.commitType, CommitScope: tt.args.commitScope, CommitSubject: tt.args.commitSubject, CommitDescription: tt.args.commitDescription})
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("AskQuestions() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
