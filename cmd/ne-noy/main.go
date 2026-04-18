package main

import (
	"context"
	"flag"
	"log"
	server "ne_noy/internal"
	"ne_noy/internal/config"
	"ne_noy/internal/database"
)

//	@title			Ne-Noy-Api
//	@version		1.0
//	@description	API к ИС Не-Ной

//	@host		https://simply-funny-spearfish.cloudpub.ru
//	@BasePath	/api

// @securityDefinitions.apikey	VkAuth
// @in							header
// @name						Authorization
func main() {
	ctx := context.Background()
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}
	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	srv := server.New(db, *cfg)
	if err := srv.Run(cfg.Server.Host, cfg.Server.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
