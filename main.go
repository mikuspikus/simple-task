package main

import (
	"context"
	"flag"
	"internal/api"
	"log"
)

const (
	defaultConnString = ""
	defaultPort       = 8000
)

func main() {
	connStringPtr := flag.String("conn", defaultConnString, "Connection string for PostgreSQL DB")
	portPtr := flag.Int("port", defaultPort, "Port for HTTP server")

	flag.Parse()

	cfg := api.Config{
		ConnString: *connStringPtr,
		Port:       *portPtr,
	}

	ctx := context.Background()
	s, err := api.NewServer(ctx, cfg.ConnString)
	if err != nil {
		log.Fatal(err)
	}

	defer s.Close(ctx)
	s.Start(cfg)
}
