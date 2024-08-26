package silence

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	CommandKey = "command"
	ErrorKey   = "error"
	StatusKey  = "status"
)

const (
	ErrorStatus = 1
)

func Run(app *App) {
	err := app.initApp()
	if err != nil {
		app.Log.Error(
			"failed to inicialize, exiting with error status",
			slog.Int(StatusKey, ErrorStatus),
			slog.String(ErrorKey, err.Error()),
		)
		os.Exit(ErrorStatus)
	}

	// TODO: Add locking mechanism to prevent two instances running at once.

	app.Log.Debug("app is inicialized")

	job, err := NewJob(app, DefaultJobConfigPath)
	if err != nil {
		app.Log.Error(
			"failed to create job, exiting with error status",
			slog.Int(StatusKey, ErrorStatus),
			slog.String(ErrorKey, err.Error()),
		)
		os.Exit(ErrorStatus)
	}

	if job.Name == "" {
		app.Log.Error(
			"job name must be set",
			slog.Int(StatusKey, ErrorStatus),
		)
		os.Exit(ErrorStatus)
	}
	app.Log.Info(fmt.Sprintf("job %s was inicialized", job.Name))

	err = job.run(app)
	if err != nil {
		app.Log.Error(
			"fatal error, exiting with error status",
			slog.Int(StatusKey, ErrorStatus),
			slog.String(ErrorKey, err.Error()),
		)
		os.Exit(ErrorStatus)
	}
}
