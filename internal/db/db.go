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

	"github.com/cropalato/squid-vault-auth/internal/conf"
)

type DB struct {
	cfg   *conf.Config
	Users []UserRecord
}

type UserRecord struct {
	User    string   `json:"user"`
	Pass    string   `json:"pass"`
	Groups  []string `json:"groups"`
	ExpDate int64    `json:"exp_date"`
}

func NewBD(c *conf.Config) (*DB, error) {
	if _, err := os.Stat(c.DbPath); err != nil {
		err := os.WriteFile(c.DbPath, []byte("[]"), 0o600)
		if err != nil {
			return nil, err
		}
	}
	return &DB{cfg: c, Users: nil}, nil
}

func (d *DB) LoadDB() error {
	content, err := os.ReadFile(d.cfg.DbPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &d.Users)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) SaveDB(data []UserRecord) error {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(d.cfg.DbPath, file, 0o600); err != nil {
		return err
	}
	return nil
}

func (d *DB) GetRecord(user string) (UserRecord, error) {
	// TODO
	u := &UserRecord{User: "", Pass: "", Groups: []string{""}, ExpDate: -1}
	return *u, nil
}

func (d *DB) AddRecord(ur UserRecord) error {
	// TODO
	return nil
}

func (d *DB) UpdateRecord(data UserRecord) error {
	// TODO
	return nil
}

func (d *DB) DeleteRecord(user string) error {
	// TODO
	return nil
}
