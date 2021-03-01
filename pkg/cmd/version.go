package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
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
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mpi v"+Version+"-" + CommitHash + " " + runtime.GOOS + "/" + runtime.GOARCH + " BuildDate=" + BuildAt)
	},
}
