package api

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Server struct {
	DB     DB
	Router *mux.Router
}

type Config struct {
	ConnString string
	Port       int
}

func NewServer(ctx context.Context, connString string) (*Server, error) {
	db, err := NewDB(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &Server{
		DB:     db,
		Router: mux.NewRouter(),
	}, nil
}

func (s *Server) Close(ctx context.Context) {
	s.DB.Close(ctx)
}

func setContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/vnd.api+json")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start(cfg Config) {
	s.Router.Use(setContentType)
	s.routes()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.Router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	server.Shutdown(ctx)
	log.Println("HTTP server is shutting down")
	os.Exit(0)
}
