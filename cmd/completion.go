package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell|nushell|elvish|ion]",
	Short: "Generate completion script",
	Long: fmt.Sprintf(`To load completions:

Bash:

  $ source <(%[1]s completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
  # macOS:
  $ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ %[1]s completion fish | source

  # To load completions for each session, execute once:
  $ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

PowerShell:

  PS> %[1]s completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> %[1]s completion powershell > %[1]s.ps1
  # and source this file from your PowerShell profile.

Nushell:

  $ %[1]s completion nushell | save --force ~/.cache/carapace/%[1]s.nu

  # Then source it in ~/.config/nushell/config.nu:
  # source ~/.cache/carapace/%[1]s.nu

Elvish:

  $ eval (%[1]s completion elvish | slurp)

  # Add to ~/.elvish/rc.elv:
  # eval (%[1]s completion elvish | slurp)

Ion:

  $ %[1]s completion ion | source

  # Add to your Ion shell init file (e.g., ~/.config/ion/initrc)
`, rootCmd.Name()),
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell", "nushell", "elvish", "ion"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		
		// Delegate all shell completion generation to carapace's _carapace subcommand
		// Use os/exec to call the binary itself with _carapace to avoid recursion
		executable, err := os.Executable()
		if err != nil {
			// Fallback: try to find the binary in PATH
			executable, err = exec.LookPath(os.Args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not determine executable path: %v\n", err)
				os.Exit(1)
			}
		}
		
		// Resolve symlinks to get the actual binary path
		executable, err = filepath.EvalSymlinks(executable)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not resolve executable path: %v\n", err)
			os.Exit(1)
		}
		
		// Execute the binary with _carapace subcommand
		execCmd := exec.Command(executable, "_carapace", shell)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s completion: %v\n", shell, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
