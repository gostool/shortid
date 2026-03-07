//go:build goframe

package gfhttp

import (
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gostool/shortid"
)

// BindRoutes registers shortid SDK routes on a GoFrame server.
// It is intentionally thin so the SDK remains framework-agnostic.
func BindRoutes(server *ghttp.Server, generator *shortid.Generator) {
	BindRoutesWithEndpoint(server, shortid.NewEndpoint(generator))
}

// BindRoutesWithEndpoint registers shortid routes with a transport-agnostic endpoint.
func BindRoutesWithEndpoint(server *ghttp.Server, endpoint *shortid.Endpoint) {
	server.BindHandler("/nextid", func(r *ghttp.Request) {
		if r.Request.Method != http.MethodGet && r.Request.Method != http.MethodPost {
			r.Response.WriteStatus(http.StatusMethodNotAllowed)
			r.Response.WriteJson(map[string]any{
				"error": "method not allowed",
			})
			return
		}

		id, err := endpoint.NextID(r.Context())
		if err != nil {
			r.Response.WriteStatus(http.StatusInternalServerError)
			r.Response.WriteJson(map[string]any{
				"error": gerror.Wrap(err, "generate id failed").Error(),
			})
			return
		}

		r.Response.WriteJson(map[string]any{
			"id": id,
		})
	})

	server.BindHandler("/health", func(r *ghttp.Request) {
		if err := endpoint.Health(r.Context()); err != nil {
			r.Response.WriteStatus(http.StatusServiceUnavailable)
			r.Response.WriteJson(map[string]any{
				"status": "error",
				"error":  gerror.Wrap(err, "shortid health check failed").Error(),
			})
			return
		}

		r.Response.WriteJson(map[string]any{
			"status": "ok",
		})
	})
}
