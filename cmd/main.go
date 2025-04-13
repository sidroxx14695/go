package main

import (
	"entitlements/internal/router"

	"log"

	"net/http"
)

func main() {

	r := router.NewRouter()

	log.Println("Starting server on :8081")

	log.Fatal(http.ListenAndServe(":8081", r))

}
