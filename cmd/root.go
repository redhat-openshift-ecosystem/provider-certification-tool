package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vmware-tanzu/sonobuoy/cmd/sonobuoy/app"

	"github.com/redhat-openshift-ecosystem/provider-certification-tool/pkg/cmd/adm"
	"github.com/redhat-openshift-ecosystem/provider-certification-tool/pkg/cmd/get"
	"github.com/redhat-openshift-ecosystem/provider-certification-tool/pkg/destroy"
	"github.com/redhat-openshift-ecosystem/provider-certification-tool/pkg/report"
	"github.com/redhat-openshift-ecosystem/provider-certification-tool/pkg/retrieve"
	"github.com/redhat-openshift-ecosystem/provider-certification-tool/pkg/run"
	"github.com/redhat-openshift-ecosystem/provider-certification-tool/pkg/status"
	"github.com/redhat-openshift-ecosystem/provider-certification-tool/pkg/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opct",
	Short: "OPCT",
	Long:  `OpenShift/OKD Provider Compatibility Tool is used to validate an OpenShift installation on a provider or hardware using standard conformance suites`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error

		// Validate logging level
		loglevel := viper.GetString("log-level")
		logrusLevel, err := log.ParseLevel(loglevel)
		if err != nil {
			log.Fatal(err)
		}
		log.SetLevel(logrusLevel)

		// Additional log options
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initBindFlag(flag string) {
	err := viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag))
	if err != nil {
		log.Warnf("Unable to bind flag %s\n", flag)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("kubeconfig", "", "kubeconfig for target OpenShift cluster")
	rootCmd.PersistentFlags().String("log-level", "info", "logging level")
	initBindFlag("kubeconfig")
	initBindFlag("log-level")

	// Link in child commands
	rootCmd.AddCommand(destroy.NewCmdDestroy())
	rootCmd.AddCommand(retrieve.NewCmdRetrieve())
	rootCmd.AddCommand(run.NewCmdRun())
	rootCmd.AddCommand(status.NewCmdStatus())
	rootCmd.AddCommand(version.NewCmdVersion())
	rootCmd.AddCommand(report.NewCmdReport())
	rootCmd.AddCommand(get.NewCmdGet())
	rootCmd.AddCommand(adm.NewCmdAdm())

	// Link in child commands direct from Sonobuoy
	rootCmd.AddCommand(app.NewSonobuoyCommand())
	rootCmd.AddCommand(app.NewCmdResults())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}
