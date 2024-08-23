/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"silence/silence"

	"github.com/spf13/cobra"
)

// jobCmd represents the job command
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Add empty job file",
	Long: `Placeholder, will get proper name and arguments later.

Add empty job file with default (nil) values.
The file will be job.json`,
	Run: createEmptyJob,
}

func createEmptyJob(cmd *cobra.Command, args []string) {
	job := silence.DefaultJob("")
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		slog.Error("cannot marshal json", slog.String(silence.ErrorKey, err.Error()))
		os.Exit(1)
	}

	_, err = os.Stat(silence.DefaultJobConfigPath)
	if err == nil {
		slog.Error("file alredy exists")
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		slog.Error("error when stating file", slog.String(silence.ErrorKey, err.Error()))
		os.Exit(1)
	}

	err = os.WriteFile(silence.DefaultJobConfigPath, data, 0744)
	if err != nil {
		slog.Error("error when writing file", slog.String(silence.ErrorKey, err.Error()))
	}
}

func init() {
	rootCmd.AddCommand(jobCmd)
}
