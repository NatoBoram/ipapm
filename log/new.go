package log

import (
	"log/slog"
	"os"
	"strconv"
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
	nord04 = gchalk.Hex(nord.Nord04.String())
	nord06 = gchalk.Hex(nord.Nord06.String())
	nord11 = gchalk.Hex(nord.Nord11.String())
	nord13 = gchalk.Hex(nord.Nord13.String())
	nord14 = gchalk.Hex(nord.Nord14.String())
	nord15 = gchalk.Hex(nord.Nord15.String())
)

func replace(groups []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.LevelKey {
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

	// Strings
	if attr.Key != slog.MessageKey && attr.Value.Kind() == slog.KindString {
		if value := attr.Value.String(); value != "" {
			attr.Value = slog.StringValue(nord04(value))
			return attr
		}
	}

	// Message
	if attr.Key == slog.MessageKey && attr.Value.Kind() == slog.KindString {
		if value := attr.Value.String(); value != "" {
			attr.Value = slog.StringValue(nord06(value))
			return attr
		}
	}

	// Errors
	if attr.Value.Kind() == slog.KindAny {
		if value, ok := attr.Value.Any().(error); ok {
			attr.Value = slog.StringValue(nord11(value.Error()))
			return attr
		}
	}

	// Numbers
	switch attr.Value.Kind() {
	case slog.KindFloat64:
		value := strconv.FormatFloat(attr.Value.Float64(), 'f', -1, 64)
		attr.Value = slog.StringValue(nord15(value))
		return attr

	case slog.KindInt64:
		value := strconv.FormatInt(attr.Value.Int64(), 10)
		attr.Value = slog.StringValue(nord15(value))
		return attr

	case slog.KindUint64:
		value := strconv.FormatUint(attr.Value.Uint64(), 10)
		attr.Value = slog.StringValue(nord15(value))
		return attr

	}

	return attr
}

func New(c Config) *slog.Logger {
	handler := handler(c)
	handler = slogctx.NewHandler(handler, nil)

	logger := slog.New(handler)
	return logger
}
