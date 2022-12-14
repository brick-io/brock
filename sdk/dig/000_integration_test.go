//go:build integration

package sdkdig_test

import (
	"context"
	"net"
	"testing"

	. "github.com/brick-io/brock/sdk/dig"
)

func testDig(t *testing.T) {
	if t.Skipped() {
		return
	}

	ctx := context.Background()
	res, dig := net.DefaultResolver, Dig{}
	name := "onebrick.io"

	if mxs, err := res.LookupMX(ctx, name); err == nil {
		for i, mx := range mxs {
			t.Log(i, mx)
		}
	}

	if mxs, err := dig.MX(ctx, name); err == nil {
		for i, mx := range mxs {
			t.Log(i, mx)
		}
	}
}
