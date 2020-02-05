package context

import (
	"context"
	"tgwabr/api"
)

type (
	tgCtx    struct{}
	waCtx    struct{}
	dbCtx    struct{}
	cacheCtx struct{}
)

func NewTG(ctx context.Context, impl api.TG) context.Context {
	return context.WithValue(ctx, tgCtx{}, impl)
}

func FromTG(ctx context.Context) (impl api.TG, ok bool) {
	v := ctx.Value(tgCtx{})
	if v != nil {
		impl, ok = v.(api.TG)
		return
	}
	return
}

func NewWA(ctx context.Context, impl api.WA) context.Context {
	return context.WithValue(ctx, waCtx{}, impl)
}

func FromWA(ctx context.Context) (impl api.WA, ok bool) {
	v := ctx.Value(waCtx{})
	if v != nil {
		impl, ok = v.(api.WA)
		return
	}
	return
}

func NewDB(ctx context.Context, impl api.Store) context.Context {
	return context.WithValue(ctx, dbCtx{}, impl)
}

func FromDB(ctx context.Context) (impl api.Store, ok bool) {
	v := ctx.Value(dbCtx{})
	if v != nil {
		impl, ok = v.(api.Store)
		return
	}
	return
}

func NewCache(ctx context.Context, impl api.Cache) context.Context {
	return context.WithValue(ctx, cacheCtx{}, impl)
}

func FromCache(ctx context.Context) (impl api.Cache, ok bool) {
	v := ctx.Value(cacheCtx{})
	if v != nil {
		impl, ok = v.(api.Cache)
		return
	}
	return
}
