package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for bash, zsh, fish, or powershell.

To load completions:

Bash:
  $ source <(kbvault completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ kbvault completion bash | sudo tee /etc/bash_completion.d/kbvault
  # macOS:
  $ kbvault completion bash | sudo tee /usr/local/etc/bash_completion.d/kbvault

Zsh:
  # If shell completion is not already enabled in your environment you will need
  # to enable it.  You can execute the following once:
  $ echo "autoload -Uz compinit && compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ kbvault completion zsh | sudo tee /usr/share/zsh/site-functions/_kbvault
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ kbvault completion fish | source
  # To load completions for each session, execute once:
  $ kbvault completion fish | sudo tee /usr/share/fish/vendor_completions.d/kbvault.fish

PowerShell:
  PS> kbvault completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> kbvault completion powershell > $PROFILE
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
			return fmt.Errorf("invalid shell type: %s", args[0])
		},
	}

	return cmd
}
