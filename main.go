package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	Server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	Server.ListenAndServe()

}
