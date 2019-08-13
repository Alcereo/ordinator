package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", mainHandler)

	port := 8081
	log.Printf("Server starting on port %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func mainHandler(writer http.ResponseWriter, request *http.Request) {
	_, _ = fmt.Fprintf(writer, "%s %s\n", request.Method, request.URL.String())
	_, _ = fmt.Fprintf(writer, "\n")
	for header, value := range request.Header {
		_, _ = fmt.Fprintf(writer, "%s=%v\n", header, value)
	}
}
