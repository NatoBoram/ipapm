package log

import (
	"log/slog"
	"os"
	"time"

	"github.com/NatoBoram/ipapm/env"
	"github.com/NatoBoram/ipapm/nord"
	"github.com/jwalton/gchalk"
	"github.com/jwalton/go-supportscolor"
	"github.com/lmittmann/tint"
	slogctx "github.com/veqryn/slog-context"
)

type Config struct {
	GO_ENV env.Environment
}

func handler(c Config) slog.Handler {
	if c.GO_ENV == env.Production {
		return slog.NewJSONHandler(os.Stdout, new(slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	w := os.Stderr
	return tint.NewTextHandler(w, new(tint.Options{
		Level:       slog.LevelDebug,
		NoColor:     !supportscolor.Stderr().SupportsColor,
		ReplaceAttr: replace,
		TimeFormat:  time.Kitchen,
	}))
}

var (
	nord11 = gchalk.Hex(nord.Nord11.String())
	nord13 = gchalk.Hex(nord.Nord13.String())
	nord14 = gchalk.Hex(nord.Nord14.String())
	nord15 = gchalk.Hex(nord.Nord15.String())
)

func replace(groups []string, attr slog.Attr) slog.Attr {
	if attr.Key != slog.LevelKey {
		return attr
	}

	switch attr.Value.Any().(slog.Level) {
	case slog.LevelDebug:
		attr.Value = slog.StringValue(nord15("DEBUG"))
	case slog.LevelInfo:
		attr.Value = slog.StringValue(nord14("INFO"))
	case slog.LevelWarn:
		attr.Value = slog.StringValue(nord13("WARN"))
	case slog.LevelError:
		attr.Value = slog.StringValue(nord11("ERROR"))
	}
	return attr
}

func New(c Config) *slog.Logger {
	handler := handler(c)
	handler = slogctx.NewHandler(handler, nil)

	logger := slog.New(handler)
	return logger
}
