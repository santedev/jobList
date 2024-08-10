package main

import (
	"jobList/config"
	h "jobList/handlers"
	"jobList/handlers/render"
	"jobList/services/auth"
	"jobList/store"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	db, err := store.NewSQLStorage(store.Sqlconfig())
	if err != nil {
		log.Fatal(err)
	}
	s := store.NewStore(db)
	store.InitStorage(db)
	store.DB = s

	r := chi.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*indeed.com", "*computrabajo.com", "*linkedin.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           15,
	})

	sessionStore := auth.NewCookieStore(auth.SessionOptions{
		CookiesKey: config.Envs.CookiesAuthSecret,
		MaxAge:     config.Envs.CookiesAuthAgeInSeconds,
		HttpOnly:   config.Envs.CookiesAuthIsHttpOnly,
		Secure:     config.Envs.CookiesAuthIsSecure,
	})
	authService := auth.NewAuthService(sessionStore)

	authHandler := h.New(authService)

	r.Use(c.Handler)
	r.Use(middleware.Logger)
	r.Handle("/*", public())

	r.Get("/", render.Make(h.HandleHome))
	r.Post("/jobs/get", render.Make(h.HandleJobs))
	r.Get("/jobs/saved/page", auth.RequireSession(render.Make(h.HandleSavedJobs), authService))
	r.Get("/jobs/get/saved", render.Make(h.HandleGetSavedJobs))
	r.Post("/job/save", render.Make(h.HandleSaveJobs))
	r.Post("/job/unsave", render.Make(h.HandleUnsaveJobs))

	r.Get("/auth/{provider}", render.Make(authHandler.HandleProviderLogin))
	r.Get("/auth/{provider}/callback", render.Make(authHandler.HandleAuthCallback))
	r.Get("/auth/logout/{provider}", render.Make(authHandler.HandleAuthLogout))
	r.Get("/login", render.Make(h.HandleLogin))

	listenAddr := ":" + config.Envs.Port
	slog.Info("HTTP server started", "listenAddr", listenAddr)
	http.ListenAndServe(listenAddr, r)
}