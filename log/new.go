package log

import (
	"log/slog"
	"os"
	"time"

	"github.com/NatoBoram/ipapm/env"
	"github.com/NatoBoram/ipapm/nord"
	"github.com/NatoBoram/ipapm/progress"
	"github.com/jwalton/gchalk"
	"github.com/jwalton/go-supportscolor"
	"github.com/lmittmann/tint"
	slogctx "github.com/veqryn/slog-context"
)

type Config struct {
	GO_ENV   env.Environment
	Progress *progress.Pool
}

func handler(c Config) slog.Handler {
	if c.GO_ENV == env.Production {
		return slog.NewJSONHandler(os.Stdout, new(slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return tint.NewTextHandler(c.Progress, new(tint.Options{
		AddSource:   false,
		Level:       slog.LevelDebug,
		NoColor:     !supportscolor.Stderr().SupportsColor,
		ReplaceAttr: replace,
		TimeFormat:  time.Kitchen,
	}))
}

var (
	nord03 = gchalk.Hex(nord.Nord03.String())
	nord04 = gchalk.Hex(nord.Nord04.String())
	nord11 = gchalk.Hex(nord.Nord11.String())
	nord13 = gchalk.Hex(nord.Nord13.String())
	nord14 = gchalk.Hex(nord.Nord14.String())
	nord15 = gchalk.Hex(nord.Nord15.String())
)

func replace(groups []string, attr slog.Attr) slog.Attr {
	switch attr.Key {

	case slog.LevelKey:
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

	case slog.MessageKey:
		if value := attr.Value.String(); value != "" {
			attr.Value = slog.StringValue(nord04(value))
			return attr
		}

	case slog.TimeKey:
		return attr
	case slog.SourceKey:
		return attr

	}

	if attr.Value.Kind() != slog.KindGroup {
		if value := attr.Value.String(); value != "" {
			attr.Value = slog.StringValue(nord03(value))
			return attr
		}
	}

	return attr
}

func New(c Config) *slog.Logger {
	handler := handler(c)
	handler = slogctx.NewHandler(handler, nil)

	logger := slog.New(handler)
	return logger
}
