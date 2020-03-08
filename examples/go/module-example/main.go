package main

import (
	"net/http"

	"github.com/golang/example/stringutil"
)

// Handler is the entry point for this function
func Handler(w http.ResponseWriter, r *http.Request) {
	msg := stringutil.Reverse(stringutil.Reverse("Vendor Example Test"))
	w.Write([]byte(msg))
}
