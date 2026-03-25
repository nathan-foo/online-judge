package handler

import (
	"log"
	"net/http"

	"github.com/nathan-foo/online-judge/gateway/internal/service"
)

func HandleHello(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	res, err := service.GetHello()

	if err != nil {
		log.Printf("Service error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(res))
}
