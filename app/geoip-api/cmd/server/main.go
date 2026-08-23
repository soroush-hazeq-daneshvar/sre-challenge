package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sre-challenge/geoip-api/internal/config"
	"github.com/sre-challenge/geoip-api/internal/handler"
	"github.com/sre-challenge/geoip-api/internal/metrics"
	"github.com/sre-challenge/geoip-api/internal/provider"
	"github.com/sre-challenge/geoip-api/internal/repository"
	"github.com/sre-challenge/geoip-api/internal/service"
)

func main() {
	cfg := config.Load()

	metrics.Register()

	repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(context.Background()); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	geoProvider := provider.NewIPAPIProvider(cfg.GeoIPProviderURL, cfg.GeoIPProviderTimeout)
	svc := service.NewGeoIPService(repo, geoProvider)
	h := handler.NewGeoIPHandler(svc)

	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods(http.MethodGet)
	r.HandleFunc("/ready", h.Ready).Methods(http.MethodGet)
	r.HandleFunc("/country", h.GetCountry).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("starting geoip-api on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
