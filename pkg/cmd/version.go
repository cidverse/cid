package cmd

import (
	"fmt"
	"runtime"

	"github.com/cidverse/cid/pkg/core/version"
	"github.com/spf13/cobra"
)

// Version will be set at build time
var Version string

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

		fmt.Println("cid " + versionPrefix + Version + "-" + CommitHash + " " + runtime.GOOS + "/" + runtime.GOARCH + " BuildDate=" + BuildAt)
	},
}
