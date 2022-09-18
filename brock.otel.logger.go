package brock

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// -----------------------------------------------------------------------------
// Logger
// -----------------------------------------------------------------------------

func (open_telemetry) NewLogger(ctx context.Context, c ...io.Writer) *Logger {
	if l, ok := ctx.Value(ctx_key_logger{}).(*Logger); ok && l != nil {
		return l
	}

	ws, z := make([]io.Writer, 0), new(zerolog.Logger)
	hook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {})

	for _, v := range c {
		if v != nil && v != io.Discard {
			switch v {
			case os.Stdout, os.Stderr:
				v = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
					w.Out = v
				})
			}
			if f, ok := v.(*os.File); ok && f != nil {
				v = &lumberjack.Logger{
					Filename:   f.Name(),
					MaxSize:    100, // megabytes
					MaxAge:     10,  // days
					MaxBackups: 10,  // num
					Compress:   true,
				}
			}
			ws = append(ws, v)
		}
	}

	if len(c) > 0 {
		switch len(ws) {
		case 0:
			*z = zerolog.Nop()
		case 1:
			*z = zerolog.New(ws[0]).Hook(hook)
		default:
			*z = zerolog.New(zerolog.MultiLevelWriter(ws...)).Hook(hook)
		}
	} else if zz := zerolog.Ctx(ctx); zz != nil {
		*z = *zz
	}

	*z = z.With().Timestamp().Logger()

	return &Logger{z, log.New(z, "", log.LstdFlags), &logrus.Logger{Out: z}}
}

type Logger struct {
	*zerolog.Logger
	Log    *log.Logger
	Logrus *logrus.Logger
}

type ctx_key_logger struct{}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	if l_, ok := ctx.Value(ctx_key_logger{}).(*Logger); ok {
		if l == l_ {
			return ctx
		}
	}

	return context.WithValue(ctx, ctx_key_logger{}, l)
}

func (x *Logger) Level(level string) *Logger {
	lv, err := zerolog.ParseLevel(strings.ToLower(level))
	if err == nil {
		*x.Logger = x.Logger.Level(lv)
		if x.Log != nil {
			x.Log.SetOutput(x.Logger)
		}
		if x.Logrus != nil {
			x.Logrus.SetOutput(x.Logger)
			switch lv {
			case zerolog.TraceLevel:
				x.Logrus.Level = logrus.TraceLevel
			case zerolog.DebugLevel:
				x.Logrus.Level = logrus.DebugLevel
			case zerolog.InfoLevel:
				x.Logrus.Level = logrus.InfoLevel
			case zerolog.WarnLevel:
				x.Logrus.Level = logrus.WarnLevel
			case zerolog.ErrorLevel:
				x.Logrus.Level = logrus.ErrorLevel
			case zerolog.FatalLevel:
				x.Logrus.Level = logrus.FatalLevel
			case zerolog.PanicLevel:
				x.Logrus.Level = logrus.PanicLevel
			}
		}
	}

	return x
}
