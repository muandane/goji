package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(manCmd)
}

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Display manual for Goji CLI",
	Run: func(cmd *cobra.Command, args []string) {
		manPage := `
GOJI(1)                      General Commands Manual                     GOJI(1)

NAME
       goji - A CLI tool for generating conventional commits with emojis.

SYNOPSIS
       goji [--help|-h]

       goji init [--global|--repo]

       goji [--type <type>] [--message <message>] [--scope <scope>]
            [--no-verify|-n] [--amend] [--add|-a] [--git-flag <flag>]

DESCRIPTION
       Goji is a command-line interface designed to facilitate conventional commit messages with the addition of emojis.
       It can be used in interactive mode or with flags to quickly generate commit messages.

OPTIONS
       --help, -h
              Display the help message and exit.

       init
              Initialize a Goji configuration file. Use --global to apply settings globally or --repo to apply to the current repository.

       --type <type>, -t <type>
              Specify the type of commit (e.g., feat, fix, docs).

       --message <message>, -m <message>
              The commit message to be used.

       --scope <scope>, -s <scope>
              Optional scope for the commit message.

       --no-verify, -n
              Bypass pre-commit and commit-msg hooks.

       --version, -v
              Display version information.

       --amend
              Amend the last commit.

       --add, -a
              Automatically stage files that have been modified and deleted.

       --git-flag <flag>
              Additional Git flags (can be used multiple times).

EXAMPLES
       goji init --global
              Initialize a global Goji configuration file in the user's home path.

       goji init --repo
              Add a Goji configuration file to the current repository.

       goji
              Enter interactive mode to generate a commit message.

       goji --type feat --message "Add login feature" --scope auth
              Generate a commit message for adding a feature without entering interactive mode.

       goji --type fix --message "Fix bug" --no-verify
              Create a commit with the "fix" type and bypass pre-commit hooks.

       goji --type docs --message "Update README" --amend
              Amend the last commit with a new message.

       goji --type feat --message "Add feature" --add
              Stage changes and create a commit.

       goji --type chore --message "Update dependencies" --git-flag "--signoff"
              Add a sign-off trailer to the commit.

AUTHOR
       Written by Zine el abidine Moualhi.

COPYRIGHT
       Copyright (C) 2024 Moualhi Zine El Abidine. This is free software; see the source for copying conditions.
`

		fmt.Println(manPage)
	},
}
