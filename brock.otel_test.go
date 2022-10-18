package brock_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"go.onebrick.io/brock"
)

func TestOtel(t *testing.T) {
	ctx := context.Background()
	otel := brock.OpenTelemetry

	f, err := os.CreateTemp("", "test-*")
	o := os.Stdout

	f, o = o, f

	buf := new(bytes.Buffer)
	_ = otel.NewLogger(ctx)
	_ = otel.NewLogger(ctx, nil)
	_ = otel.NewLogger(ctx, buf)
	ctx = otel.NewLogger(ctx, buf, f, o, io.Discard).Context(ctx)
	log := otel.NewLogger(ctx)
	ctx = log.Context(ctx)

	log.Info().Err(err).Int("3", 3).Msg("msg")
	log.Warn().Err(err).Int("3", 3).Msg("msg")

	f, o = o, f

	brock.Nop(f, o, ctx)
}
