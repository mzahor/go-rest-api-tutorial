package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("USER"),
		os.Getenv("PASSWORD"),
		os.Getenv("DBNAME"),
	)
	createTableIfNotExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()
	req, err := http.NewRequest("GET", "/products", nil)
	if err != nil {
		log.Fatal(err)
	}
	res := execReq(req)
	checkResponseCode(t, 200, res.Code)

	if body := res.Body.String(); body != "[]" {
		t.Errorf("Expected empty array in the body. Got '%s'", body)
	}
}

const createTableQuery = `create table if not exists products (
	id serial,
	name text not null,
	price numeric(10,2) not null default 0.00,
	constraint products_pkey primary key (id)
)`

func createTableIfNotExists() {
	if _, err := a.DB.Exec(createTableQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("truncate table products")
	a.DB.Exec("alter sequence products_id_seq restart with 1")
}

func execReq(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected int, actual int) {
	if expected != actual {
		t.Errorf("Expected status code to be %d. Got %d", expected, actual)
	}
}
