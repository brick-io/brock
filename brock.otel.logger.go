package brock

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// -----------------------------------------------------------------------------
// Logger
// -----------------------------------------------------------------------------

func (o_t) NewLogger(ctx context.Context, c ...io.Writer) *Logger {
	if l, ok := ctx.Value(ctx_key_logger{}).(*Logger); ok && l != nil {
		return l
	}

	ws, z := make([]io.Writer, 0), new(zerolog.Logger)
	hook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) { Nop(e, level, message) })

	for _, v := range c {
		if v == nil || v == io.Discard {
			continue
		} else if v == os.Stdout || v == os.Stderr {
			v = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
				w.Out = v
				if !isatty.IsTerminal(os.Stdout.Fd()) {
					w.NoColor = true
				}
			})
		} else if f, ok := v.(*os.File); ok && f != nil {
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
		*z = Val(zz)
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

func (l *Logger) Context(ctx context.Context) context.Context {
	if l_, ok := ctx.Value(ctx_key_logger{}).(*Logger); ok && l == l_ {
		return ctx
	}

	return context.WithValue(ctx, ctx_key_logger{}, l)
}

func (x *Logger) ParseLevel(level string) (zerolog.Level, logrus.Level) {
	level = strings.ToLower(level)
	z, _ := zerolog.ParseLevel(level)
	l, _ := logrus.ParseLevel(level)
	return z, l
}
