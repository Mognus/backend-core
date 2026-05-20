package main

import (
	"log"
	"net/http"

	"template/internal/about"
	"template/internal/authclient"
	"template/internal/config"
	"template/internal/platform/db"
	"template/internal/router"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("database handle: %v", err)
	}
	defer sqlDB.Close()

	authClient, err := authclient.New(cfg.Auth.ServiceAddr)
	if err != nil {
		log.Fatalf("setup auth client: %v", err)
	}
	defer authClient.Close()

	handler := router.New(cfg, router.Deps{
		Services: router.Services{
			About: about.NewService(database),
		},
		Clients: router.Clients{
			Auth: authClient.Auth,
		},
	})

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
