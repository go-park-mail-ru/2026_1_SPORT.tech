package app

import "net/http"

func Run() error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: NewRouter(),
	}

	return server.ListenAndServe()
}
