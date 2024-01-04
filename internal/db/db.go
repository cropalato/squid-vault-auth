//
// db.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package db

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/cropalato/squid-vault-auth/internal/conf"
)

type Database struct {
	Cfg   *conf.Config
	Users []UserRecord
	*sync.Mutex
}

type UserRecord struct {
	User    string   `json:"user"`
	Pass    string   `json:"pass"`
	Groups  []string `json:"groups"`
	ExpDate int64    `json:"exp_date"`
}

func NewBD(c *conf.Config) (*Database, error) {
	if _, err := os.Stat(c.DbPath); err != nil {
		err := os.WriteFile(c.DbPath, []byte("[]"), 0o600)
		if err != nil {
			return nil, err
		}
	}
	return &Database{Cfg: c, Users: nil}, nil
}

func (d *Database) LoadDatabase() error {
	d.Lock()
	defer d.Unlock()
	content, err := os.ReadFile(d.Cfg.DbPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &d.Users)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) SaveDatabase(data []UserRecord) error {
	d.Lock()
	defer d.Unlock()
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(d.Cfg.DbPath, file, 0o600); err != nil {
		return err
	}
	return nil
}

func (d *Database) GetRecord(user string) (UserRecord, error) {
	d.Lock()
	defer d.Unlock()
	// TODO
	u := &UserRecord{User: "", Pass: "", Groups: []string{""}, ExpDate: -1}
	return *u, nil
}

func (d *Database) AddRecord(ur UserRecord) error {
	d.Lock()
	defer d.Unlock()
	// TODO
	return nil
}

func (d *Database) UpdateRecord(data UserRecord) error {
	d.Lock()
	defer d.Unlock()
	// TODO
	return nil
}

func (d *Database) DeleteRecord(user string) error {
	d.Lock()
	defer d.Unlock()
	// TODO
	return nil
}
