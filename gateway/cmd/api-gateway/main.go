package main

import (
	"fmt"
	"net/http"

	"github.com/nathan-foo/online-judge/gateway/internal/router"
)

func main() {
	mux := router.NewMux()

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", mux)
}
