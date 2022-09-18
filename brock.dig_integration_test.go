//go:build integration

package brock_test

import (
	"context"
	"testing"

	"go.onebrick.io/brock"
)

func test_dig(t *testing.T) {
	ctx := context.Background()
	dig := brock.Dig{}

	mxs, err := dig.MX(ctx, "onebrick.io")
	t.Log(err)
	for i, mx := range mxs {
		t.Log(i, mx)
	}
}
