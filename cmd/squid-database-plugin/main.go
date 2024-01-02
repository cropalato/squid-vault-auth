//
// main.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"log"
	"os"

	"github.com/cropalato/squid-vault-auth/internal/squid"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func main() {
	err := Run()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

// Run instantiates a squid object, and runs the RPC server for the plugin
func Run() error {
	dbplugin.ServeMultiplex(squid.New)

	return nil
}
