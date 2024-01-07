//
// main.go
// Copyright (C) 2024 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cropalato/squid-vault-auth/internal/db"
	"github.com/cropalato/squid-vault-auth/internal/hash"
	"github.com/cropalato/squid-vault-auth/internal/varenv"
)

func scanString(s *bufio.Scanner) (string, error) {
	if s.Scan() {
		return s.Text(), nil
	}
	err := s.Err()
	if err == nil {
		err = io.EOF
	}
	return "", err
}

func main() {
	var user db.UserRecord
	s := bufio.NewScanner(os.Stdin)
	url_base := flag.String("url", varenv.LookupEnvOrString("SQUIDDB_URL", "http://127.0.0.1:8080"), "squid db service URL. format: 'http[s]://(<fqdn>|<ip>)[:<port>]'")
	admin_account := flag.String("admin_user", varenv.LookupEnvOrString("SQUIDDB_USER", "admin"), "admin account used to call squid db service API'")
	admin_pass := flag.String("admin_pass", varenv.LookupEnvOrString("SQUIDDB_PASS", "admin"), "admin password used to call squid db service API")
	flag.Parse()

	for {
		// Set up HTTPS request with basic authorization.
		line, err := scanString(s)
		if err != nil {
			continue
			// log.Fatal(err)
		}

		tokens := strings.Split(line, " ")
		req, err := http.NewRequest(http.MethodGet, strings.TrimRight(*url_base, "/")+"/api/v1/users/"+tokens[0], nil)
		if err != nil {
			log.Fatal(err)
		}
		req.SetBasicAuth(*admin_account, *admin_pass)

		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			log.Fatal(err)
		}
		err = resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		if user.Username == tokens[0] && hash.CheckPasswordHash(tokens[1], user.Password) {
			fmt.Println("OK")
			continue
		}
		fmt.Println("ERR")
	}
}
