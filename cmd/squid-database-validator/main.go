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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cropalato/squid-vault-auth/internal/db"
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
top:
	for {
		// Set up HTTPS request with basic authorization.
		line, err := scanString(s)
		if err != nil {
			continue
			// log.Fatal(err)
		}

		tokens := strings.Split(line, " ")
		req, err := http.NewRequest(http.MethodGet, "http://10.0.0.81:8080/api/v1/users/"+tokens[0], nil)
		if err != nil {
			log.Fatal(err)
		}
		req.SetBasicAuth("admin", "secret")

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
		if user.User == tokens[0] {
			for _, g := range user.Groups {
				if g == tokens[1] {
					fmt.Println("OK")
					continue top
				}
			}
		}
		fmt.Println("ERR")
	}
}
