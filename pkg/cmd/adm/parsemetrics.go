package adm

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/redhat-openshift-ecosystem/provider-certification-tool/internal/openshift/mustgathermetrics"
)

type parseMetricsInput struct {
	input  string
	output string
}

var parseMetricsArgs parseMetricsInput
var parseMetricsCmd = &cobra.Command{
	Use:     "parse-metrics",
	Example: "opct adm parse-metrics --input ./metrics.tar.xz --output /tmp/metrics",
	Short:   "Process the metrics and create a HTML graph.",
	Run:     parseMetricsRun,
}

func init() {
	parseMetricsCmd.Flags().StringVar(&parseMetricsArgs.input, "input", "", "Input metrics file. Example: metrics.tar.xz")
	parseMetricsCmd.Flags().StringVar(&parseMetricsArgs.output, "output", "", "Output directory. Example: /tmp/metrics")
}

func parseMetricsRun(cmd *cobra.Command, args []string) {
	fi, err := os.Open(parseMetricsArgs.input)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	// make a read buffer
	r := bufio.NewReader(fi)

	buf := &bytes.Buffer{}
	buf.ReadFrom(r)

	htmlFile := "/metrics.html"
	mgm, err := mustgathermetrics.NewMustGatherMetrics(parseMetricsArgs.output, htmlFile, buf)
	if err != nil {
		panic(err)
	}
	err = mgm.Process()
	if err != nil {
		fmt.Printf("ERROR: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Success! HTML report created at %s/%s\n", parseMetricsArgs.output, htmlFile)
}
