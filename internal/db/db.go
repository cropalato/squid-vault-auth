//
// db.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

// Package to manage json file with user records.
package db

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/cropalato/squid-vault-auth/internal/conf"
	"github.com/rs/zerolog/log"
)

type Database struct {
	Cfg   *conf.Config
	Users []UserRecord
	sync.Mutex
}

type UserRecord struct {
	User    string   `json:"user"`
	Pass    string   `json:"pass"`
	Groups  []string `json:"groups"`
	ExpDate int64    `json:"exp_date"`
}

// NewBD load json file.
// If file doesn't exist, we will create the file.
func NewBD(c *conf.Config) (*Database, error) {
	log.Debug().Msg("Creating database object")
	if _, err := os.Stat(c.DbPath); err != nil {
		log.Debug().Msg("database file doesn't exist. Creating file " + c.DbPath)
		err := os.WriteFile(c.DbPath, []byte("[]"), 0o600)
		if err != nil {
			log.Panic().Err(err)
		}
	}
	log.Debug().Msg("Created database object")
	return &Database{Cfg: c, Users: nil}, nil
}

// LoadDatabase read and parse json file.
func (d *Database) LoadDatabase() error {
	d.Lock()
	defer d.Unlock()
	content, err := os.ReadFile(d.Cfg.DbPath)
	if err != nil {
		log.Err(err)
		return err
	}
	err = json.Unmarshal(content, &d.Users)
	if err != nil {
		log.Err(err)
		return err
	}
	return nil
}

// SaveDatabase upgrade json file.
func (d *Database) SaveDatabase() error {
	file, err := json.MarshalIndent(d.Users, "", "  ")
	if err != nil {
		log.Err(err)
		return err
	}
	if err := os.WriteFile(d.Cfg.DbPath, file, 0o600); err != nil {
		log.Err(err)
		return err
	}
	return nil
}

// GetRecord returns a user record.
func (d *Database) GetRecord(user string) (*UserRecord, error) {
	d.Lock()
	defer d.Unlock()
	for _, r := range d.Users {
		if r.User == user {
			u := &UserRecord{User: r.User, Pass: r.Pass, Groups: r.Groups, ExpDate: r.ExpDate}
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

// AddRecord insert new user record.
func (d *Database) AddRecord(ur UserRecord) error {
	d.Lock()
	defer d.Unlock()
	for _, r := range d.Users {
		if r.User == ur.User {
			return errors.New("user already exist")
			/*
			   d.Users[i].Pass = ur.Pass
			   d.Users[i].Groups = ur.Groups
			   d.Users[i].ExpDate = ur.ExpDate
			   return d.SaveDatabase()
			*/
		}
	}
	d.Users = append(d.Users, UserRecord{User: ur.User, Pass: ur.Pass, Groups: ur.Groups, ExpDate: ur.ExpDate})
	return d.SaveDatabase()
}

// UpdateRecord update user record with new data
func (d *Database) UpdateRecord(ur UserRecord) error {
	d.Lock()
	defer d.Unlock()
	for i, r := range d.Users {
		if r.User == ur.User {
			d.Users[i].Pass = ur.Pass
			d.Users[i].Groups = ur.Groups
			d.Users[i].ExpDate = ur.ExpDate
			return d.SaveDatabase()
		}
	}
	return errors.New("user not found")
}

// DeleteRecord remove user record if it exist
func (d *Database) DeleteRecord(user string) error {
	d.Lock()
	defer d.Unlock()
	var indexes []int
	for i, r := range d.Users {
		if r.User == user {
			indexes = append(indexes, i)
		}
	}
	for _, index := range indexes {
		d.Users = append(d.Users[:index], d.Users[index+1:]...)
	}
	return d.SaveDatabase()
}
