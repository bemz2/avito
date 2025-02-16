package api

import (
	"avito-shop/internal/api/handlers"
	"avito-shop/internal/api/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(handler *handlers.Handler, authMiddleware *middleware.AuthMiddleware) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/auth", handler.Auth).Methods(http.MethodPost)

	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(authMiddleware.Authenticate)

	protected.HandleFunc("/info", handler.GetInfo).Methods(http.MethodGet)
	protected.HandleFunc("/sendCoin", handler.SendCoins).Methods(http.MethodPost)
	protected.HandleFunc("/buy/{item}", handler.BuyItem).Methods(http.MethodPost)

	return r
}
