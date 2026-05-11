package main

import (
	"context"
	"log"
	"net/http"

	"template/internal/config"
	"template/internal/gateway"
	"template/internal/platform/db"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("database handle: %v", err)
	}
	defer sqlDB.Close()

	handler, cleanup, err := gateway.New(ctx, cfg, database)
	if err != nil {
		log.Fatalf("setup gateway: %v", err)
	}
	defer cleanup()

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}

	log.Printf("listening on http://%s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
