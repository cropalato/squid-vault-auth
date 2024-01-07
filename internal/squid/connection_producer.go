//
// connection_producer.go
// Copyright (C) 2024 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package squid

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-secure-stdlib/parseutil"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
	"github.com/hashicorp/vault/sdk/database/helper/connutil"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
)

type squidConnectionProducer struct {
	ConnectionURL     string      `json:"connection_url"   mapstructure:"connection_url"  structs:"connection_url"`
	Username          string      `json:"username"         mapstructure:"username"        structs:"username"`
	Password          string      `json:"password"         mapstructure:"password"        structs:"password"`
	ConnectTimeoutRaw interface{} `json:"connect_timeout" structs:"connect_timeout" mapstructure:"connect_timeout"`

	rawConfig      map[string]interface{}
	connectTimeout time.Duration
	Initialized    bool
	Type           string
	sync.Mutex
}

func (s *squidConnectionProducer) Initialize(ctx context.Context, req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
	s.rawConfig = req.Config

	err := mapstructure.WeakDecode(req.Config, s)
	if err != nil {
		return dbplugin.InitializeResponse{}, err
	}

	if s.ConnectTimeoutRaw == nil {
		s.ConnectTimeoutRaw = "5s"
	}

	s.connectTimeout, err = parseutil.ParseDurationSecond(s.ConnectTimeoutRaw)
	if err != nil {
		return dbplugin.InitializeResponse{}, fmt.Errorf("invalid connect_timeout: %w", err)
	}

	switch {
	case len(s.ConnectionURL) == 0:
		return dbplugin.InitializeResponse{}, fmt.Errorf("connection_url cannot be empty")
	case len(s.Username) == 0:
		return dbplugin.InitializeResponse{}, fmt.Errorf("username cannot be empty")
	case len(s.Password) == 0:
		return dbplugin.InitializeResponse{}, fmt.Errorf("password cannot be empty")
	}

	s.Initialized = true

	if req.VerifyConnection {
		if _, err := s.Connection(ctx); err != nil {
			return dbplugin.InitializeResponse{}, fmt.Errorf("error verifying connection: %w", err)
		}
	}

	resp := dbplugin.InitializeResponse{
		Config: req.Config,
	}

	return resp, nil
}

func (c *squidConnectionProducer) Connection(ctx context.Context) (interface{}, error) {
	if !c.Initialized {
		return nil, connutil.ErrNotInitialized
	}

	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(c.ConnectionURL, "/")+"/authTest", nil)
	if err != nil {
		log.Fatal().Err(err)
	}
	req.SetBasicAuth(c.Username, c.Password)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Err(err)
		return nil, err
	}
	readErr := res.Body.Close()
	if readErr != nil {
		log.Err(readErr)
		return nil, err
	}
	if res.StatusCode > 299 || res.StatusCode < 200 {
		log.Error().Msg(fmt.Sprintf("Response failed with status code: %d and body: %s", res.StatusCode, body))
		return nil, fmt.Errorf("response failed with status code: %d and body: %s", res.StatusCode, body)
	}

	return nil, nil
}

// Close attempts to close the connection
func (c *squidConnectionProducer) Close() error {
	// Grab the write lock
	c.Lock()
	defer c.Unlock()
	// Nothing to do here, It is an HTTP connection

	return nil
}

func (c *squidConnectionProducer) SecretValues() map[string]string {
	return map[string]string{
		c.Password: "[password]",
	}
}
