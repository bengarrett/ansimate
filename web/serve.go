//go:build ignore

package main

import (
	"log"
	"net/http"
)

const port = "8070"

func main() {
	log.Println("Serving web/ at http://localhost:" + port + " (Ctrl+C to stop) ...")
	log.Fatal(http.ListenAndServe(":"+port, http.FileServer(http.Dir("web"))))
}
