package main

import (
	"log"
	"net/http"

	"github.com/muesli/smolder"
)

func main() {
	context := &Context{}

	// Setup web-service
	smolderConfig := smolder.APIConfig{
		BaseURL:    "localhost:8080",
		PathPrefix: "",
	}

	wsContainer := smolder.NewSmolderContainer(smolderConfig, nil, nil)
	(&HelloResource{}).Register(wsContainer, smolderConfig, context)

	// GlobalLog("Starting polly web-api...")
	server := &http.Server{Addr: ":8080", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
