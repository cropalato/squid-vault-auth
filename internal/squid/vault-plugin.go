//
// main.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package squid

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-secure-stdlib/strutil"
	"github.com/hashicorp/vault/sdk/database/helper/dbutil"
	"github.com/hashicorp/vault/sdk/helper/template"

	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

const (
	squidDbTypeName         = "squidDB"
	defaultUserNameTemplate = `{{ printf "v_%s_%s_%s_%s" (.DisplayName | truncate 15) (.RoleName | truncate 15) (random 20) (unix_time) | truncate 100 | replace "-" "_" | lowercase }}`
)

type SquidDatabase struct {
	*squidConnectionProducer
	// Variables for the database
	usernameProducer template.StringTemplate
}

func New() (interface{}, error) {
	db := new()

	// This middleware isn't strictly required, but highly recommended to prevent accidentally exposing
	// values such as passwords in error messages. An example of this is included below
	dbType := dbplugin.NewDatabaseErrorSanitizerMiddleware(db, db.SecretValues)
	return dbType, nil
}

func new() *SquidDatabase {
	connProducer := &squidConnectionProducer{}
	connProducer.Type = squidDbTypeName

	return &SquidDatabase{
		squidConnectionProducer: connProducer,
	}
}

// Type returns the TypeName for this backend
func (s *SquidDatabase) Type() (string, error) {
	return squidDbTypeName, nil
}

func (s *SquidDatabase) Initialize(ctx context.Context, req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
	usernameTemplate, err := strutil.GetString(req.Config, "username_template")
	if err != nil {
		return dbplugin.InitializeResponse{}, fmt.Errorf("failed to retrieve username_template: %w", err)
	}
	if usernameTemplate == "" {
		usernameTemplate = defaultUserNameTemplate
	}

	up, err := template.NewTemplate(template.Template(usernameTemplate))
	if err != nil {
		return dbplugin.InitializeResponse{}, fmt.Errorf("unable to initialize username template: %w", err)
	}
	s.usernameProducer = up

	_, err = s.usernameProducer.Generate(dbplugin.UsernameMetadata{})
	if err != nil {
		return dbplugin.InitializeResponse{}, fmt.Errorf("invalid username template: %w", err)
	}

	return s.squidConnectionProducer.Initialize(ctx, req)
}

func (s *SquidDatabase) NewUser(ctx context.Context, req dbplugin.NewUserRequest) (dbplugin.NewUserResponse, error) {
	s.Lock()
	defer s.Unlock()

	var result *multierror.Error

	// defaultUserCreationIFQL := "{\"username\": \"{{username}}\", \"password\": \"{{password}}\", \"group\": \"" + req.UsernameConfig.RoleName + "\"}"
	defaultUserCreationIFQL := "{\"username\": \"{{username}}\", \"password\": \"{{password}}\", \"groups\": [\"{{rolename}}\"], \"exp_date\": {{exp_date}} }"

	creationIFQL := req.Statements.Commands
	if len(creationIFQL) == 0 {
		creationIFQL = []string{defaultUserCreationIFQL}
	}

	username, err := s.usernameProducer.Generate(req.UsernameConfig)
	if err != nil {
		return dbplugin.NewUserResponse{}, err
	}

	m := map[string]string{
		"username": username,
		"password": req.Password,
		"rolename": req.UsernameConfig.RoleName,
		"exp_date": string(fmt.Sprintf("%d", req.Expiration.Unix())),
	}
	data := []byte(dbutil.QueryHelper(creationIFQL[0], m))

	url := strings.TrimRight(s.ConnectionURL, " /") + "/api/v1/users"
	err = addUser(url, data, s.Username, s.Password)
	if err != nil {
		result = multierror.Append(result, err)
	}

	if result.ErrorOrNil() != nil {
		return dbplugin.NewUserResponse{}, fmt.Errorf("unable to create user cleanly: %w", result.ErrorOrNil())
	}
	resp := dbplugin.NewUserResponse{
		Username: username,
	}
	return resp, nil
}

func delUser(url string, user string, pass string) error {
	// create a new HTTP client
	client := &http.Client{}

	// create a new DELETE request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(user, pass)

	// send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read the response body
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *SquidDatabase) DeleteUser(ctx context.Context, req dbplugin.DeleteUserRequest) (dbplugin.DeleteUserResponse, error) {
	s.Lock()
	defer s.Unlock()

	var result *multierror.Error

	url := strings.TrimRight(s.ConnectionURL, " /") + "/api/v1/users/" + req.Username
	err := delUser(url, s.Username, s.Password)
	if err != nil {
		result = multierror.Append(result, err)
	}

	if result.ErrorOrNil() != nil {
		return dbplugin.DeleteUserResponse{}, fmt.Errorf("failed to delete user cleanly: %w", result.ErrorOrNil())
	}
	return dbplugin.DeleteUserResponse{}, nil
}

func addUser(url string, data []byte, user string, pass string) error {
	// create a new HTTP client
	client := &http.Client{}

	// create a new DELETE request
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, pass)

	// send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read the response body
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *SquidDatabase) UpdateUser(ctx context.Context, req dbplugin.UpdateUserRequest) (dbplugin.UpdateUserResponse, error) {
	if req.Password == nil && req.Expiration == nil {
		return dbplugin.UpdateUserResponse{}, fmt.Errorf("no changes requested")
	}

	s.Lock()
	defer s.Unlock()

	var result *multierror.Error
	if req.Password != nil {
		tpl := "{\"password\": \"{{password}}\"}"

		m := map[string]string{
			"password": req.Password.NewPassword,
		}
		data := []byte(dbutil.QueryHelper(tpl, m))

		url := strings.TrimRight(s.ConnectionURL, " /") + "/api/v1/users/" + req.Username
		err := changeUserPassword(url, data)
		if err != nil {
			result = multierror.Append(result, err)
		}

		if result.ErrorOrNil() != nil {
			return dbplugin.UpdateUserResponse{}, fmt.Errorf("unable to create user cleanly: %w", result.ErrorOrNil())
		}
	}
	// Expiration is a no-op
	return dbplugin.UpdateUserResponse{}, nil
}

func changeUserPassword(url string, data []byte) error {
	// create a new HTTP client
	client := &http.Client{}

	// create a new DELETE request
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read the response body
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}
