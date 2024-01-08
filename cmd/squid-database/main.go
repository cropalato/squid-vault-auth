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
//	export SQUIDDB_PATH=/tmp/squid-vault.json
//	export SQUIDDB_PASS='$2a$14$vN59c/ZmesroW/oYaDn3yeAPutg4wkVM5t6n9CNrOcTMJ.zDVtcUm' #secret
//	export SQUIDDB_USER=admin
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/cropalato/squid-vault-auth/internal/conf"
	"github.com/cropalato/squid-vault-auth/internal/varenv"
	"github.com/cropalato/squid-vault-auth/internal/webservices"
)

func main() {
	listen := flag.String("listen", varenv.LookupEnvOrString("SQUIDDB_LISTEN", ":8080"), "IP and port used by squid db service. format: '[<ip>]:<port>'. default: ':8080'")
	admin_account := flag.String("admin_user", varenv.LookupEnvOrString("SQUIDDB_USER", "admin"), "admin account used to call squid db service API'")
	// the dafault password is 'admin'. ypu can use create a new one using
	// python -c 'import bcrypt; print(bcrypt.hashpw(b"PASSWORD", bcrypt.gensalt(rounds=15)).decode("ascii"))'
	admin_pass := flag.String("admin_pass", varenv.LookupEnvOrString("SQUIDDB_PASS", "$2b$15$QjL.GaBkHXXTifvFFQo2eOVPqzHpQQ7y/axXslpylNACTpeCYR.t6"), "admin password used to call squid db service API")
	db_path := flag.String("db_path", varenv.LookupEnvOrString("SQUIDDB_PATH", "/etc/squid-vault.json"), "squid db file path")
	cors := flag.String("cors_origin", varenv.LookupEnvOrString("SQUIDDB_CORS", "*"), "configure Access-Control-Allow-Origin header")
	debug := flag.Bool("debug", varenv.LookupEnvOrBool("SQUIDDB_DEBUG", false), "activate debug mode")
	flag.Parse()

	cfg := &conf.Config{Debug: *debug, Addr: *listen, AdminID: *admin_account, AdminSecret: *admin_pass, DbPath: *db_path, CorsOrigin: *cors}

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
