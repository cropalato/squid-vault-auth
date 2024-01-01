//
// webservices.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package webservices

import (
  "fmt"
  "net/http"
  "github.com/cropalato/squid-vault-auth/internal/conf"
  "github.com/cropalato/squid-vault-auth/internal/hash"
	"github.com/rs/zerolog/log"
)

type HTTPHandlers struct {
  cfg *conf.Config
}

func NewHandlers(cfg *conf.Config) (*HTTPHandlers, error) {
  return &HTTPHandlers{cfg: cfg}, nil
}

func (h *HTTPHandlers) ValidateCredential(user string, pass string) error{
	if user != h.cfg.AdminID {
		return fmt.Errorf("Invalid User %s", user)
	}
	if ! hash.CheckPasswordHash(pass, h.cfg.AdminSecret) {
		return fmt.Errorf("Invalid password for user %s", user)
	}
		return nil
}

func (h *HTTPHandlers) AuthHandle(w http.ResponseWriter, r *http.Request) {

  w.Header().Set("Access-Control-Allow-Origin", h.cfg.CorsOrigin)
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

  w.Header().Set("Access-Control-Allow-Origin", h.cfg.CorsOrigin)
  if r.Method == http.MethodOptions {
    return
  }
  w.WriteHeader(200)
  _, err := w.Write([]byte("We are working to offer this method ASAP\n"))
  if err != nil {
    log.Err(err)
  }
}

func (h *HTTPHandlers) PutUser(w http.ResponseWriter, r *http.Request) {

  w.Header().Set("Access-Control-Allow-Origin", h.cfg.CorsOrigin)
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

  w.Header().Set("Access-Control-Allow-Origin", h.cfg.CorsOrigin)
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

  w.Header().Set("Access-Control-Allow-Origin", h.cfg.CorsOrigin)
  if r.Method == http.MethodOptions {
    return
  }
  w.WriteHeader(200)
  _, err := w.Write([]byte("We are working to offer this method ASAP\n"))
  if err != nil {
    log.Err(err)
  }
}

