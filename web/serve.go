//go:build ignore

package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Serving web/ at http://localhost:8080 (Ctrl+C to stop) ...")
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("web"))))
}
