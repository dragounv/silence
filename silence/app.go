package silence

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

type App struct {
	cmd  *cobra.Command
	args []string

	WorkDirFlag *string

	Log *slog.Logger
	// WorkDir string
}

func (app *App) initApp() error {
	app.initLogger(os.Stdout, slog.LevelInfo)
	app.Log.Info("app is inicializing", slog.String(CommandKey, app.cmd.Name()))

	if err := app.setWorkingDirectory(*app.WorkDirFlag); err != nil {
		return err
	}

	return nil
}

func (app *App) InitCommand(cmd *cobra.Command, args []string) *App {
	app.cmd = cmd
	app.args = args
	return app
}

func (app *App) initLogger(w io.Writer, level slog.Level) {
	app.Log = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
}

func (app *App) setWorkingDirectory(path string) error {
	var err error
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			app.Log.Error(
				"os.Getwd could not provide current working directory",
				slog.String(ErrorKey, err.Error()),
			)
			return err
		}
	}

	err = os.Chdir(path)
	if err != nil {
		var reason string

		if errors.Is(err, fs.ErrNotExist) {
			reason = fmt.Sprintf("%s does not exist", path)
		} else if errors.Is(err, fs.ErrPermission) {
			reason = fmt.Sprintf("insufficent permissions for path %s", path)
		} else {
			reason = fmt.Sprintf("error acessing path %s", path)
		}

		app.Log.Error(reason, slog.String(ErrorKey, err.Error()))
		return err
	}

	return nil
}
