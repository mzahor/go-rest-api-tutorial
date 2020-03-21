package main

import (
	"database/sql"
	"errors"
)

type product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func (p *product) get(db *sql.DB) error {
	return errors.New("Not Implemented")
}

func (p *product) create(db *sql.DB) error {
	return errors.New("Not Implemented")
}

func (p *product) update(db *sql.DB) error {
	return errors.New("Not Implemented")
}

func (p *product) delete(db *sql.DB) error {
	return errors.New("Not Implemented")
}

func getProducts(db *sql.DB, skip, take int) ([]product, error) {
	return nil, errors.New("Not impemented")
}
