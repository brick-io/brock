package brock

import (
	"context"
	"time"
)

type Metadata struct {
	Ok bool

	Environment string

	BuildTime time.Time
	Namespace string
	Version   string
	Commit    string
}

type ctx_key_metadata struct{}

func (x Metadata) Save(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctx_key_metadata{}, x)
}

func (x Metadata) Load(ctx context.Context) Metadata {
	x, x.Ok = ctx.Value(ctx_key_metadata{}).(Metadata)
	return x
}
