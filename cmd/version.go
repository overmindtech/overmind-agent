package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const AgentVersion = "0.12.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns the agent version",
	Long:  `Returns the agent version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(AgentVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
