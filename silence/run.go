package silence

import (
	"log/slog"
	"os"
)

func Run(app *App) {
	app.InitLogger(os.Stdout, slog.LevelInfo)
	app.Log.Info("Silence is starting", slog.String("command", app.cmd.Name()))

	app.WorkDir = os.
}

// Try to create a file, that will signal that an instace of this app is running.
// If the file alredy exists, then the function should return an error.
// Do not continue program execution if non nil error is returned.
//
// If the file exists even when no instance is running, then it is safe to delete it.
// TODO: The file should contain timestamp and PID of the other process,
// so we can check if it still exists.
func CreateLock() error {

}