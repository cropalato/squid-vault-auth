//
// webservices.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

// Package with all handler functions
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

// NewHandlers create a new HTTPHandlers class
func NewHandlers(cfg *conf.Config) (*HTTPHandlers, error) {
	db, err := db.NewBD(cfg)
	if err != nil {
		log.Err(err)
	}

	err = db.LoadDatabase()
	if err != nil {
		panic(err)
	}

	return &HTTPHandlers{UserDB: db}, nil
}

// ValidateCredential can be use to be sure the user/password is valid.
func (h *HTTPHandlers) ValidateCredential(user string, pass string) error {
	if user != h.UserDB.Cfg.AdminID {
		return fmt.Errorf("invalid User %s", user)
	}
	if !hash.CheckPasswordHash(pass, h.UserDB.Cfg.AdminSecret) {
		return fmt.Errorf("invalid password for user %s", user)
	}
	return nil
}

// State is used to check is the service is running and health.
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

// AuthHandle expose a simple entrypoint to test user authentication
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

// GetUser return json with user details.
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

// PutUser create new user.
// If user exist it will upgrade the user record.
func (h *HTTPHandlers) PutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	var user db.UserRecord
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	up, err := hash.HashPassword(user.Pass)
	if err != nil {
		log.Err(err)
		http.Error(w, "failed processing request", http.StatusInternalServerError)
	}
	user.Pass = up
	err = h.UserDB.AddRecord(user)
	if err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmp, err := json.Marshal(user)
	if err != nil {
		log.Err(err)
	}
	log.Debug().Msg(string(tmp))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, err = w.Write([]byte("{ \"msg\": \"Added new user record, username=" + user.User + "\" }\n"))
	if err != nil {
		log.Err(err)
	}
}

// PatchUser upgrade user record
func (h *HTTPHandlers) PatchUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	var user db.UserRecord
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	up, err := hash.HashPassword(user.Pass)
	if err != nil {
		log.Err(err)
		http.Error(w, "failed processing request", http.StatusInternalServerError)
	}
	user.Pass = up

	err = h.UserDB.UpdateRecord(user)
	if err != nil {
		log.Err(err)
		w.Header().Set("Content-Type", "application/json")
		if err.Error() == "user not found" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, err = w.Write([]byte("{ \"msg\": \"" + err.Error() + "\" }\n"))
		if err != nil {
			log.Err(err)
		}
		return
	}
	tmp, err := json.Marshal(user)
	if err != nil {
		log.Err(err)
	}
	log.Debug().Msg(string(tmp))
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte("{ \"msg\": \"Updated user record, username=" + user.User + "\" }\n"))
	if err != nil {
		log.Err(err)
	}
}

// DeleteUser remove user record
func (h *HTTPHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", h.UserDB.Cfg.CorsOrigin)
	if r.Method == http.MethodOptions {
		return
	}
	path := strings.Split(r.URL.Path, "/")
	user := path[len(path)-1]
	err := h.UserDB.DeleteRecord(user)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(200)
	_, err = w.Write([]byte("{ \"msg\": \"Deleted user record, username=" + user + "\" }\n"))
	if err != nil {
		log.Err(err)
	}
}
