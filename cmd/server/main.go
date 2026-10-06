package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/svanegmond/agentic-eng-oct-6/internal/api"
	"github.com/svanegmond/agentic-eng-oct-6/internal/auth"
	"github.com/svanegmond/agentic-eng-oct-6/internal/store"
)

func main() {
	dbDriver := flag.String("db", "sqlite", "Database driver (sqlite or postgres)")
	dbDSN := flag.String("dsn", "identity.db", "Database DSN or file path")
	port := flag.String("port", "8080", "HTTP server port")
	jwtSecret := flag.String("jwt-secret", "dev-jwt-secret-interview-mock-long-enough", "JWT signing secret key")
	seed := flag.Bool("seed", false, "Seed database with demo profile and credentials")
	flag.Parse()

	// Environment variable overrides
	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}
	if envDB := os.Getenv("DB"); envDB != "" {
		*dbDriver = envDB
	}
	if envDSN := os.Getenv("DSN"); envDSN != "" {
		*dbDSN = envDSN
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("[server] opening %s database at %s...", *dbDriver, *dbDSN)
	dao, err := store.Open(ctx, store.DBConfig{
		Driver: *dbDriver,
		DSN:    *dbDSN,
	})
	if err != nil {
		log.Fatalf("[server] failed to open database: %v", err)
	}
	defer dao.Close()

	if *seed {
		seedDemoData(context.Background(), dao)
	}

	authSvc := auth.NewService(*jwtSecret, dao)
	router := api.NewRouter(dao, authSvc)

	srv := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Graceful shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[server] identity service listening on http://localhost:%s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] server error: %v", err)
		}
	}()

	<-stop
	log.Println("[server] shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[server] error during shutdown: %v", err)
	}
	log.Println("[server] stopped.")
}

func seedDemoData(ctx context.Context, dao store.DAO) {
	aliceID := "11111111-1111-1111-1111-111111111111"
	_, err := dao.GetProfileByID(ctx, aliceID)
	if err == nil {
		log.Printf("[server] demo data already present, skipping seed")
		return
	}

	now := time.Now().UTC()
	aliceProfile := &store.UserProfile{
		ID:        aliceID,
		Name:      "Alice Smith",
		Address:   "123 Market St, San Francisco, CA 94105",
		Phone:     "+15551234567",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := dao.CreateProfile(ctx, aliceProfile); err != nil {
		log.Printf("[server] warning: failed to seed alice profile: %v", err)
	}

	aliceCred := &store.UserCredential{
		ID:        uuid.NewString(),
		UserID:    aliceID,
		Username:  "alice",
		Method:    "password",
		Password:  "password123",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := dao.CreateCredential(ctx, aliceCred); err != nil {
		log.Printf("[server] warning: failed to seed alice credential: %v", err)
	}

	// Also seed Bob for search demonstrations
	bobID := "22222222-2222-2222-2222-222222222222"
	bobProfile := &store.UserProfile{
		ID:        bobID,
		Name:      "Bob Smith",
		Address:   "456 Castro St, Mountain View, CA 94041",
		Phone:     "+15559876543",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := dao.CreateProfile(ctx, bobProfile); err != nil {
		log.Printf("[server] warning: failed to seed bob profile: %v", err)
	}

	log.Printf("[server] seeded demo users: alice (password: password123), bob")
}
