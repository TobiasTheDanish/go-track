package server

import (
	"fmt"
	"go-track/cmd/web"
	"go-track/internal/db"
	"go-track/internal/github"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port       int
	webHandler *web.Handler
}

func NewServer() *http.Server {
	db, err := db.New()
	if err != nil {
		log.Fatalf("Creating DatabaseFacade failed! %e", err)
	}
	gh, err := github.New()
	if err != nil {
		log.Fatalf("Creating GithubService failed! %e", err)
	}

	webHandler := web.NewHandler(db, gh)

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port:       port,
		webHandler: webHandler,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
