package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/mattn/go-isatty"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

const specialChar = "%"

var manDescription = `
Goji CLI Documentation

# NAME
goji - A CLI tool for generating conventional commits with emojis.

# SYNOPSIS
**goji** [**--help**|**-h**]

**goji init** [**--global**|**--repo**]

**goji** [**--type** <type>] [**--message** <message>] [**--scope** <scope>]

# DESCRIPTION
Goji is a command-line interface designed to facilitate conventional commit messages with the addition of emojis. It can be used in interactive mode or with flags to quickly generate commit messages.

# OPTIONS
**--help**, **-h**
* Display the help message and exit.

**init**
* Initialize a Goji configuration file. Use **--global** to apply settings globally or **--repo** to apply to the current repository.

**--type** <type>
* Specify the type of commit (e.g., feat, fix, docs).

**--message** <message>
* The commit message to be used.

**--scope** <scope>
* Optional scope for the commit message.

# EXAMPLES
**goji init --global**
* Initialize a global Goji configuration file in the user's home path.

**goji init --repo**
* Add a Goji configuration file to the current repository.

**goji**
* Enter interactive mode to generate a commit message.

**goji --type feat --message "Add login feature" --scope auth**
* Generate a commit message for adding a feature without entering interactive mode.

# ENVIRONMENT
**GOJI_CONFIG**
* The path to the Goji configuration file. If not set, defaults to the home directory or the current repository's root.

`

var manBugs = "See GitHub Issues: <https://github.com/muandane/goji/issues>"
var manAuthor = "Zine el abidine Moualhi <zineelabidinemoualhi@gmail.com>"

// manCmd generates the man pages.
var manCmd = &cobra.Command{
	Use:     "manual",
	Aliases: []string{"man"},
	Short:   "Generate man pages",
	Args:    cobra.NoArgs,
	Hidden:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithStandardStyle("dark"),
			)
			if err != nil {
				return err
			}
			// Render the manual in the terminal using glamour.
			out, err := renderer.Render(markdownManual())
			if err != nil {
				return err
			}
			fmt.Print(out)
		} else {
			// Generate the man page using mango-cobra.
			manPage, err := mcobra.NewManPage(1, rootCmd)
			if err != nil {
				return err
			}

			manPage = manPage.
				WithLongDescription(sanitizeSpecial(manDescription)).
				WithSection("Bugs", sanitizeSpecial(manBugs)).
				WithSection("Author", sanitizeSpecial(manAuthor)).
				WithSection("Copyright", "Copyright (C) 2024 Moualhi Zine El Abidine")

			fmt.Println(manPage.Build(roff.NewDocument()))
			return nil
		}
		return nil
	},
}

// Generate markdown for the manual.
func markdownManual() string {
	return fmt.Sprint(
		"# MANUAL\n", sanitizeMarkdown(manDescription),
		"# BUGS\n", manBugs,
		"\n# AUTHOR\n", manAuthor,
	)
}

func sanitizeMarkdown(input string) string {
	escaped := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"`", "&#96;", // backtick
	).Replace(input)
	return strings.ReplaceAll(escaped, specialChar, "`")
}

func sanitizeSpecial(s string) string {
	return strings.ReplaceAll(s, specialChar, "")
}

func init() {
	rootCmd.AddCommand(manCmd)
}
