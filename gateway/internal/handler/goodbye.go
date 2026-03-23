package handler

import (
	"fmt"
	"net/http"
)

func HandleGoodbye(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Goodbye, world!")
}
