//
// webservices.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package webservices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cropalato/squid-vault-auth/internal/conf"
	"github.com/cropalato/squid-vault-auth/internal/db"
	"github.com/cropalato/squid-vault-auth/internal/hash"
	"github.com/rs/zerolog/log"
)

type HTTPHandlers struct {
	UserDB *db.Database
}

func NewHandlers(cfg *conf.Config) (*HTTPHandlers, error) {
	db, err := db.NewBD(cfg)
	if err != nil {
		log.Err(err)
	}

	err = db.LoadDatabase()
	if err != nil {
		panic(err)
	}
	out, _ := json.MarshalIndent(db.Users, "", " ")
	fmt.Printf("%s\n", out)

	return &HTTPHandlers{UserDB: db}, nil
}

func (h *HTTPHandlers) ValidateCredential(user string, pass string) error {
	if user != h.UserDB.Cfg.AdminID {
		return fmt.Errorf("invalid User %s", user)
	}
	if !hash.CheckPasswordHash(pass, h.UserDB.Cfg.AdminSecret) {
		return fmt.Errorf("invalid password for user %s", user)
	}
	return nil
}

func (h *HTTPHandlers) State(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	// TODO: Add code to validate if service is ready to reply requests
	w.WriteHeader(200)
	_, err := w.Write([]byte("Service is ready"))
	if err != nil {
		log.Err(err)
	}
}

func (h *HTTPHandlers) AuthHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	u, p, ok := r.BasicAuth()
	if !ok {
		fmt.Println("Error parsing basic auth")
		w.WriteHeader(401)
		return
	}
	err := h.ValidateCredential(u, p)
	if err != nil {
		w.WriteHeader(401)
		_, err := w.Write([]byte(fmt.Sprintf("Authentication fail. %s\n", err)))
		if err != nil {
			log.Err(err)
		}
	} else {
		w.WriteHeader(200)
		_, err := w.Write([]byte("Authentication success"))
		if err != nil {
			log.Err(err)
		}
	}
}

func (h *HTTPHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	path := strings.Split(r.URL.Path, "/")
	user := path[len(path)-1]
	j, err := h.UserDB.GetRecord(user)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	data, err := json.Marshal(j)
	if err != nil {
		log.Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
	_, err = w.Write(data)
	if err != nil {
		log.Err(err)
	}
}

func (h *HTTPHandlers) PutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	w.WriteHeader(200)
	_, err := w.Write([]byte("We are working to offer this method ASAP\n"))
	if err != nil {
		log.Err(err)
	}
}

func (h *HTTPHandlers) PatchUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	w.WriteHeader(200)
	_, err := w.Write([]byte("We are working to offer this method ASAP\n"))
	if err != nil {
		log.Err(err)
	}
}

func (h *HTTPHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	w.WriteHeader(200)
	_, err := w.Write([]byte("We are working to offer this method ASAP\n"))
	if err != nil {
		log.Err(err)
	}
}
