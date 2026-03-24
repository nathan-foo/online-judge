package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/router"
)

const serverPort = 8080

func main() {
	mux := router.NewMux()
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", serverPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Server running on port :%d", serverPort)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Error running http server: %v", err)
	}
}
