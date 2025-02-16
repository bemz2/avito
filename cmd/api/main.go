package main

import (
	"avito-shop/internal/api"
	"avito-shop/internal/api/handlers"
	"avito-shop/internal/api/middleware"
	"avito-shop/internal/config"
	"avito-shop/internal/repository"
	"avito-shop/internal/service"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.NewConfig()

	db, err := sql.Open("postgres", cfg.GetDBConnString())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, cfg.JWTSecret)
	handler := handlers.NewHandler(svc)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	router := api.NewRouter(handler, authMiddleware)

	log.Printf("Сервер запущен на порту :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
