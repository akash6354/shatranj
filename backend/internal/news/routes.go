package news

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler, adminOnly func(http.Handler) http.Handler) {
	mux.HandleFunc("GET /api/v1/news", handler.List)
	mux.HandleFunc("GET /api/v1/news/{slug}", handler.Get)
	mux.Handle("GET /api/v1/admin/news", adminOnly(http.HandlerFunc(handler.AdminList)))
	mux.Handle("POST /api/v1/admin/news", adminOnly(http.HandlerFunc(handler.Create)))
	mux.Handle("PUT /api/v1/admin/news/{postID}", adminOnly(http.HandlerFunc(handler.Update)))
	mux.Handle("PUT /api/v1/admin/news/{postID}/status", adminOnly(http.HandlerFunc(handler.SetStatus)))
	mux.Handle("DELETE /api/v1/admin/news/{postID}", adminOnly(http.HandlerFunc(handler.Delete)))
}
