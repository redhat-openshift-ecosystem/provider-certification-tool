package adm

import (
	"fmt"

	"github.com/spf13/cobra"
)

var admCmd = &cobra.Command{
	Use:   "adm",
	Short: "Administrative commands.",
	Run:   runGet,
}

func init() {
	admCmd.AddCommand(parseMetricsCmd)
}

func NewCmdAdm() *cobra.Command {
	return admCmd
}

func runGet(cmd *cobra.Command, args []string) {
	fmt.Println("Nothing to do. See -h for more options.")
}
