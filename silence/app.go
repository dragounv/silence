package silence

import (
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

type App struct {
	cmd *cobra.Command
	args []string

	Log *slog.Logger
	WorkDir string
}

func NewApp(cmd *cobra.Command, args []string) *App {
	return &App{
		cmd: cmd,
		args: args,
	}
}

func (app *App) InitLogger(w io.Writer, level slog.Level) {
	app.Log = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
}

// If path is empty then curent working directoty will be used
func (app *App) InitWorkDir(path string, log *slog.Logger) error {
	var err error
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			log.Error(
				"Unable to get working directory",
				slog.String("error", err.Error()),
				slog.String("returned_from", "os.Getwd"),
				return "FixME"
			)
		}
	}
	

}