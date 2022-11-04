package sdkotel_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	sdkotel "github.com/brick-io/brock/sdk/otel"
)

func Test_sdkotel(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "test-*")
	o := os.Stdout

	f, o = o, f

	buf := new(bytes.Buffer)
	_ = sdkotel.Log(ctx)
	_ = sdkotel.Log(ctx, nil)
	_ = sdkotel.Log(ctx, buf)
	ctx = sdkotel.Log(ctx, buf, f, o, io.Discard).Context(ctx)
	log := sdkotel.Log(ctx)
	ctx = log.Context(ctx)
	_ = ctx

	log.Info().Err(err).Int("3", 3).Msg("msg")
	log.Warn().Err(err).Int("3", 3).Msg("msg")

	f, o = o, f
	_, _ = f, o
}
