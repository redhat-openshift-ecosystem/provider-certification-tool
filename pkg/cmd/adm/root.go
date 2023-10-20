package adm

import (
	"os"

	"github.com/spf13/cobra"
)

var admCmd = &cobra.Command{
	Use:   "adm",
	Short: "Administrative commands.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	admCmd.AddCommand(parseMetricsCmd)
	admCmd.AddCommand(parseEtcdLogsCmd)
}

func NewCmdAdm() *cobra.Command {
	return admCmd
}
