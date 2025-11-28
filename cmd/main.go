package main

import (
	"log"
	server "ne_noy/internal"
	"ne_noy/internal/config"
	"ne_noy/internal/database"
)

func main() {
	cfg, err := config.Load("/Users/konstantinusolcev/GolandProjects/awesomeProject/configs/app.dev.yaml")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	srv := server.New(db, cfg.Secret, cfg.AppId)
	if err := srv.Run(cfg.Server.Host, cfg.Server.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
