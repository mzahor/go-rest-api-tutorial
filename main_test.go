package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a App

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

func addProducts(count int) {
	if count < 1 {
		log.Fatal("Can't create less than one product")
	}
	for i := 0; i < count; i++ {
		_, err := a.DB.Exec("insert into products(name, price) values ($1, $2)", fmt.Sprintf("Product %d", i+1), (i+1)*10)
		if err != nil {
			log.Fatalf("Failed to add products for testing. %v", err)
		}
	}
}

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

func TestGetNonExistant(t *testing.T) {
	clearTable()
	req, err := http.NewRequest("GET", "/products/11", nil)
	if err != nil {
		log.Fatal(err)
	}
	res := execReq(req)
	checkResponseCode(t, http.StatusNotFound, res.Code)
	var body map[string]string
	json.Unmarshal(res.Body.Bytes(), &body)
	if body == nil {
		t.Errorf("Expected to get json body back. Got '%s'", res.Body.String())
	}
	if body["error"] != "Product not found" {
		t.Errorf("Expected error prop to = 'Product not found'. Got '%s'", body["error"])
	}
}

func TestCreate(t *testing.T) {
	clearTable()
	reqBody := `{"name": "laptop", "price": 2999.99}`
	req, err := http.NewRequest("POST", "/products", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-type", "application/json")
	res := execReq(req)
	var m map[string]interface{}
	err = json.Unmarshal(res.Body.Bytes(), &m)
	if err != nil {
		t.Errorf("Failed to parse response json. %v", err)
		// return
	}
	if m["name"] != "laptop" {
		t.Errorf("Expected name to be laptop. Got '%v'", m["name"])
	}
	if m["price"] != 2999.99 {
		t.Errorf("Expected price to be 2999.99. Got '%v'", m["price"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected id to be 1. Got: '%v'", m["id"])
	}
}

func TestGetExistant(t *testing.T) {
	clearTable()
	addProducts(1)
	req, _ := http.NewRequest("GET", "/products/1", nil)
	res := execReq(req)
	var m map[string]interface{}
	err := json.Unmarshal(res.Body.Bytes(), &m)
	if err != nil {
		t.Errorf("Failed to parse response json. %v", err)
	}
	if m["name"] != "Product 1" {
		t.Errorf("Expected name to be 'Product 1'. Got '%v'", m["name"])
	}
	if m["price"] != 10.0 {
		t.Errorf("Expected price to be 10.0. Got '%v'", m["price"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected id to be 1. Got: '%v'", m["id"])
	}
}

func TestUpdate(t *testing.T) {
	clearTable()
	addProducts(1)
	updBody := `{"name": "updated name", "price": 20}`
	req, _ := http.NewRequest("PUT", "/products/1", bytes.NewBufferString(updBody))
	res := execReq(req)
	var m map[string]interface{}
	err := json.Unmarshal(res.Body.Bytes(), &m)
	checkResponseCode(t, http.StatusOK, res.Code)
	if err != nil {
		t.Errorf("Failed to parse response json. %v", err)
	}
	if m["name"] != "updated name" {
		t.Errorf("Expected name to be 'Product 1'. Got '%v'", m["name"])
	}
	if m["price"] != 20.0 {
		t.Errorf("Expected price to be 20.0. Got '%v'", m["price"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected id to be 1. Got: '%v'", m["id"])
	}
}

func TestDelete(t *testing.T) {
	clearTable()
	addProducts(1)
	req, _ := http.NewRequest("GET", "/products/1", nil)
	res := execReq(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, _ = http.NewRequest("DELETE", "/products/1", nil)
	res = execReq(req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, _ = http.NewRequest("GET", "/products/1", nil)
	res = execReq(req)
	checkResponseCode(t, http.StatusNotFound, res.Code)
}
