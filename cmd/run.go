package cmd

import (
	"silence/silence"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run series of crawls according to configuration.",
	Long: `Run series of crawls according to configuration.`,
	Run: startApp,
}

func startApp(cmd *cobra.Command, args []string) {
	silence.Run(silence.NewApp(cmd, args))
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
