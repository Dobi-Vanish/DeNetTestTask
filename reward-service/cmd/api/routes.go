package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(middleware.Heartbeat("/ping"))

	mux.Group(func(r chi.Router) {
		r.Use(app.authTokenMiddleware("some_secret_key")) //

		r.Get("/users/leaderboard", app.GetLeaderboard)
		r.Get("/users/{id}/status", app.retrieveOne)
		r.Post("/users/{id}/task/complete", app.completeTask)
		r.Post("/users/{id}/task/telegramSign", app.completeTelegramSign)
		r.Post("/users/{id}/task/XSign", app.completeXSign)
		r.Post("/users/{id}/referrer", app.redeemReferrer)
	})

	mux.Post("/authenticate", app.Authenticate)
	mux.Post("/registrate", app.Registrate)

	return mux
}
