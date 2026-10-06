package admin

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler, adminOnly func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/admin/dashboard", adminOnly(http.HandlerFunc(handler.Dashboard)))
	mux.Handle("GET /api/v1/admin/users", adminOnly(http.HandlerFunc(handler.Users)))
	mux.Handle("PUT /api/v1/admin/users/{userID}/status", adminOnly(http.HandlerFunc(handler.SetUserStatus)))
	mux.Handle("GET /api/v1/admin/payments", adminOnly(http.HandlerFunc(handler.Payments)))
	mux.Handle("GET /api/v1/admin/payment-flags", adminOnly(http.HandlerFunc(handler.PaymentFlags)))
	mux.Handle("POST /api/v1/admin/payments/{paymentID}/flags", adminOnly(http.HandlerFunc(handler.FlagPayment)))
	mux.Handle("PUT /api/v1/admin/payment-flags/{flagID}/resolve", adminOnly(http.HandlerFunc(handler.ResolvePaymentFlag)))
	mux.Handle("PUT /api/v1/admin/coaches/{userID}/status", adminOnly(http.HandlerFunc(handler.SetCoachStatus)))
}
