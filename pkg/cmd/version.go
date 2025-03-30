package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cidverseutils/version"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of cid",
		Long:  `All software has versions. This is cid's`,
		Run: func(cmd *cobra.Command, args []string) {
			versionPrefix := ""
			if version.IsStable(constants.Version) {
				versionPrefix = "v"
			}

			_, _ = fmt.Fprintf(os.Stdout, "GitVersion:    %s\n", versionPrefix+constants.Version)
			_, _ = fmt.Fprintf(os.Stdout, "GitCommit:     %s\n", constants.CommitHash)
			_, _ = fmt.Fprintf(os.Stdout, "GitTreeState:  %s\n", constants.RepositoryStatus)
			_, _ = fmt.Fprintf(os.Stdout, "BuildDate:     %s\n", constants.BuildAt)
			_, _ = fmt.Fprintf(os.Stdout, "GoVersion:     %s\n", runtime.Version())
			_, _ = fmt.Fprintf(os.Stdout, "Compiler:      %s\n", runtime.Compiler)
			_, _ = fmt.Fprintf(os.Stdout, "Platform:      %s\n", runtime.GOOS+"/"+runtime.GOARCH)
		},
	}
}
