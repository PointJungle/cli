package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the CLI's version",
		Args:  noArgs,
		RunE: func(*cobra.Command, []string) error {
			a.result(a.Build, fmt.Sprintf("pj %s (%s, %s)", a.Build.Version, a.Build.Commit, a.Build.Date))
			return nil
		},
	}
}
