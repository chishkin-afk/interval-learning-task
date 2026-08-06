package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

func NewHandler(env string) slog.Handler {
	switch env {
	case EnvLocal, EnvDev:
		return newPrettyHandler(os.Stdout)

	case EnvProd:
		return newJSONHandler(os.Stdout)

	default:
		return newPrettyHandler(os.Stdout)
	}
}

func newPrettyHandler(w io.Writer) slog.Handler {
	return tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: "15:04:05",
		AddSource:  true,
		NoColor:    false,
	})
}

func newJSONHandler(w io.Writer) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
}
