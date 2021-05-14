package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Product struct {
	ID    int64  `jsonapi:"primary,products"`
	Name  string `jsonapi:"attr,name"`
	Price int    `jsonapi:"attr,price"`
}

var mysqlDB *sql.DB

func Server() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", handleHome).Methods("GET")
	router.HandleFunc("/api/products", BrowseProduct).Methods("GET")
	router.HandleFunc("/api/products", CreateProduct).Methods("POST")
	router.HandleFunc("/api/products/{id}", DeleteProduct).Methods("DELETE")
	router.HandleFunc("/api/products/{id}", ShowProduct).Methods("GET")
	router.HandleFunc("/api/products/{id}", UpdateProduct).Methods("PUT")
	return  router
}

func main() {
	mysqlDB = connect()
	defer mysqlDB.Close()
	router := Server()
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleHome(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	json.NewEncoder(writer).Encode(map[string]interface{}{
		"status":200,
		"message":"Hello World",
	})
}

func BrowseProduct(writer http.ResponseWriter, _ *http.Request) {
	rows, err := mysqlDB.Query("SELECT id, name, price FROM products")
	if err != nil {
		renderJson(writer, map[string]interface{}{
			"message": "Not Found",
		})
	}

	var products []*Product

	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			log.Print(err)
		} else {
			products = append(products, &product)
		}
	}

	renderJson(writer, products)
}

func CreateProduct(writer http.ResponseWriter, request *http.Request) {
	var product Product
	err := jsonapi.UnmarshalPayload(request.Body, &product)
	if err != nil {
		log.Print(err)
		return
	}

	query, err := mysqlDB.Prepare("INSERT INTO products (name, price) values (?, ?)")
	if err != nil {
		log.Print(err)
		return
	}
	result, err := query.Exec(product.Name, product.Price)
	productID, err := result.LastInsertId()
	if err != nil {
		log.Print(err)
		return
	}

	product.ID = productID
	writer.WriteHeader(http.StatusCreated)
	renderJson(writer, &product)
}

func DeleteProduct(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	productID := mux.Vars(request)["id"]

	rows, err := mysqlDB.Exec("DELETE FROM products WHERE id = ?", productID)
	if err != nil {
		renderJson(writer, map[string]interface{}{
			"message": "Not Found",
		})
	}

	affected, err := rows.RowsAffected()
	if affected == 0 {
		writer.WriteHeader(http.StatusNotFound)
		jsonapi.MarshalErrors(writer, []*jsonapi.ErrorObject{{
			Title: "NotFound",
			Status: strconv.Itoa(http.StatusNotFound),
			Detail: fmt.Sprintf("Product with id %s not found", productID),
		}})
	}

	writer.WriteHeader(http.StatusNoContent)
}

func ShowProduct(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	productID := mux.Vars(request)["id"]

	rows, err := mysqlDB.Query("SELECT id, name, price FROM products WHERE id = " + productID)
	if err != nil {
		renderJson(writer, map[string]interface{}{
			"message": "Not Found",
		})
	}

	var product Product
	for rows.Next() {
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			log.Print(err)
		}
	}

	renderJson(writer, &product)
}

func UpdateProduct(writer http.ResponseWriter, request *http.Request) {
	productID := mux.Vars(request)["id"]
	var product Product
	err := jsonapi.UnmarshalPayload(request.Body, &product)
	if err != nil {
		writer.Header().Set("Content-Type", jsonapi.MediaType)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(writer, []*jsonapi.ErrorObject {{
			Title: "ValidationError",
			Detail: "Given request is invalid",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
		}})
		return
	}

	query, err := mysqlDB.Prepare("UPDATE products SET name = ?, price = ? WHERE id = ?")
	query.Exec(product.Name, product.Price, productID)
	if err != nil {
		log.Print(err)
		return
	}

	product.ID, _ = strconv.ParseInt(productID, 10, 64)
	renderJson(writer, &product)
}