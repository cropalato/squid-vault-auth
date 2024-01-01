//
// main.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//


package squid

import (
    dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func New() (interface{}, error) {
    db, err := newDatabase()
    if err != nil {
        return nil, err
    }

    // This middleware isn't strictly required, but highly recommended to prevent accidentally exposing
    // values such as passwords in error messages. An example of this is included below
    db = dbplugin.NewDatabaseErrorSanitizerMiddleware(db, db.secretValues)
    return db, nil
}

type MyDatabase struct {
    // Variables for the database
    password string
}

func newDatabase() (MyDatabase, error) {
    // ...
    db := &MyDatabase{
        // ...
    }
    return *db, nil
}

func (db *MyDatabase) secretValues() map[string]string {
    return map[string]string{
        db.password: "[password]",
    }
}

