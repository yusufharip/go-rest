package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/jsonapi"
	"log"
	"net/http"
	"reflect"
)

func connect() *sql.DB {
	user := "root"
	password := "root"
	host := "127.0.0.1"
	port := "3333"
	database := "go_products"

	// root:secret@tcp(127.0.0.1:3333)/products
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)

	db, err := sql.Open("mysql", connection)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func renderJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	if payload, err := jsonapi.Marshal(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		payloads, ok := payload.(*jsonapi.ManyPayload)
		if ok {
			val := reflect.ValueOf(data)
			payloads.Meta = &jsonapi.Meta{
				"total": val.Len(),
			}
			json.NewEncoder(w).Encode(payloads)
		} else {
			json.NewEncoder(w).Encode(payload)
		}
	}
}