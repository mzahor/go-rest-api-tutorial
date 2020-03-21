package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	DB     *sql.DB
	Router *mux.Router
}

func (a *App) Initialize(user, password, dbname string) {
	connString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)
	var err error
	a.DB, err = sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {

}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/products", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/products/{id:[0-9]+}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/products/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/products/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
}

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid id: %s", vars["id"]))
	}
	p := product{ID: id}
	if err := p.get(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondError(w, http.StatusNotFound, "Product not found")
		default:
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get product. %s", err.Error()))
		}
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	skip, _ := strconv.Atoi(r.FormValue("skip"))
	take, _ := strconv.Atoi(r.FormValue("take"))
	if skip < 0 {
		skip = 0
	}
	if take > 10 || take < 1 {
		take = 10
	}

	products, err := getProducts(a.DB, skip, take)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := p.create(a.DB); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, p)
}

func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid id")
	}
	var p product
	d := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err = d.Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
	}
	p.ID = id

	if err = p.update(a.DB); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
	}
	p := product{ID: id}
	if err := p.delete(a.DB); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
	}
	respondJSON(w, http.StatusOK, map[string]string{"result": "deleted"})
}

func respondError(w http.ResponseWriter, status int, message string) {
	payload := map[string]string{"error": message}
	respondJSON(w, status, payload)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal to json %v", payload)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}
