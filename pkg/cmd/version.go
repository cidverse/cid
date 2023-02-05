package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cidverse/cidverseutils/pkg/version"
	"github.com/spf13/cobra"
)

// Version will be set at build time
var Version string

// RepositoryStatus will be set at build time
var RepositoryStatus string

// CommitHash will be set at build time
var CommitHash string

// BuildAt will be set at build time
var BuildAt string

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of cid",
	Long:  `All software has versions. This is cid's`,
	Run: func(cmd *cobra.Command, args []string) {
		versionPrefix := ""
		if version.IsStable(Version) {
			versionPrefix = "v"
		}

		fmt.Fprintf(os.Stdout, "GitVersion:    %s\n", versionPrefix+Version)
		fmt.Fprintf(os.Stdout, "GitCommit:     %s\n", CommitHash)
		fmt.Fprintf(os.Stdout, "GitTreeState:  %s\n", RepositoryStatus)
		fmt.Fprintf(os.Stdout, "BuildDate:     %s\n", BuildAt)
		fmt.Fprintf(os.Stdout, "GoVersion:     %s\n", runtime.Version())
		fmt.Fprintf(os.Stdout, "Compiler:      %s\n", runtime.Compiler)
		fmt.Fprintf(os.Stdout, "Platform:      %s\n", runtime.GOOS+"/"+runtime.GOARCH)
	},
}
