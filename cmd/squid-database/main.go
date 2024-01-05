//
// main.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

// This package is runs a RESTApi server used to manage users for squid-proxy server.
//
// You Should use env variables to config the service.
// ex.:
//
//	export DB_PATH=/tmp/squid-vault.json
//	export ADMIN_SECRET='$2a$14$vN59c/ZmesroW/oYaDn3yeAPutg4wkVM5t6n9CNrOcTMJ.zDVtcUm' #secret
//	export ADMIN_ID=admin
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/cropalato/squid-vault-auth/internal/conf"
	"github.com/cropalato/squid-vault-auth/internal/webservices"
)

func main() {
	cfg, err := conf.NewDefaultConfig()
	if err != nil {
		panic(err)
	}

	srv := http.Server{
		Addr:              cfg.Addr,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	handlers, err := webservices.NewHandlers(cfg)
	if err != nil {
		panic(err)
	}
	r := mux.NewRouter()
	r.Use(mux.CORSMethodMiddleware(r))
	r.HandleFunc("/authTest", handlers.AuthHandle)
	r.HandleFunc("/state", handlers.State).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/v1/users", handlers.PutUser).Methods(http.MethodPut, http.MethodOptions)
	r.HandleFunc("/api/v1/users/{user}", handlers.DeleteUser).Methods(http.MethodDelete, http.MethodOptions)
	r.HandleFunc("/api/v1/users/{user}", handlers.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users/{user}", handlers.PatchUser).Methods(http.MethodPatch)
	http.Handle("/", r)
	log.Fatal(srv.ListenAndServe())
}
