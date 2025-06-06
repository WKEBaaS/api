package middlewares

import "github.com/danielgtaylor/huma/v2"

func TLSMiddleware(ctx huma.Context, next func(huma.Context)) {
	ctx = huma.WithValue(ctx, "TLS", ctx.TLS() != nil)
	next(ctx)
}
