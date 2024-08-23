package cmd

import (
	"silence/silence"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run series of crawls according to configuration.",
	Long:  `Run series of crawls according to configuration.`,
	Run:   runApp,
}

var app = new(silence.App)

func runApp(cmd *cobra.Command, args []string) {
	silence.Run(app.InitCommand(cmd, args))
}

func init() {
	rootCmd.AddCommand(runCmd)

	workDirFlag := runCmd.Flags().String("work-dir", "", "Sets working directory")
	app.WorkDirFlag = workDirFlag
}
