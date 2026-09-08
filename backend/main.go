// Command consensus-backend is the entry point for the Consensus API
// server: it loads configuration, connects to MongoDB, wires up the REST
// router and the WebSocket hub, and serves HTTP until interrupted.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"consensus/internal/auth"
	"consensus/internal/config"
	"consensus/internal/db"
	"consensus/internal/handlers"
	"consensus/internal/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if cfg.SessionSecretGenerated {
		log.Println("WARNING: SESSION_SECRET is not set — using a randomly generated one for this process. " +
			"Every login will be invalidated on restart, and this will break auth entirely if you ever run " +
			"more than one instance. Set SESSION_SECRET before deploying (generate one with: openssl rand -base64 32).")
	}
	if len(cfg.AllowedEmails) == 0 {
		log.Println("WARNING: ALLOWED_EMAILS is empty — nobody will be able to sign up (allowlist fails closed).")
	}

	database, err := db.Connect(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongo: failed to connect: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := database.Disconnect(ctx); err != nil {
			log.Printf("mongo: disconnect error: %v", err)
		}
	}()

	authStore := auth.NewStore(database.AuthSessions, cfg.SessionSecret)
	// The session cookie only needs Secure once the frontend (and thus the
	// browser's view of "same site") is served over HTTPS; local dev runs
	// on plain http://localhost.
	cookieSecure := strings.HasPrefix(cfg.FrontendURL, "https://")

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub, database, cfg.FrontendURL)

	h := handlers.New(database, authStore, hub, cfg.AllowedEmails, cookieSecure)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	h.Routes(r)

	// The WebSocket upgrade endpoint sits behind the same RequireAuth
	// middleware as the REST API and reads identity the same way
	// (auth.UserEmail), so a reviewer's cookie is all that's needed —
	// there's no separate WebSocket login step.
	r.Route("/ws", func(r chi.Router) {
		r.Use(auth.RequireAuth(authStore))
		r.Get("/sessions/{sessionID}", wsHandler.ServeSession)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("consensus backend listening on :%s (frontend origin: %s)", cfg.Port, cfg.FrontendURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server: forced shutdown: %v", err)
	}
}
