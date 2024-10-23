package sl

import (
	"fmt"
	"log/slog"
	"time"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func Method(method string) slog.Attr {
	return slog.Attr{
		Key:   "method",
		Value: slog.StringValue(method),
	}
}

func Module(module string) slog.Attr {
	return slog.Attr{
		Key:   "module",
		Value: slog.StringValue(module),
	}
}

func Millis(key string, duration time.Duration) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(fmt.Sprintf("%dms", duration.Milliseconds())),
	}
}
