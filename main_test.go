package main

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Home(t *testing.T) {
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	Server().ServeHTTP(response, request)
	expectedResponse := `{"message":"Hello World","status":200}`
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Log(err)
	}
	assert.Equal(t, 200, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(responseBody)))
	//t.Log(response.Body.String())
}

func Test_BrowseProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Log(err)
	}
	rows := sqlmock.NewRows([]string{"id","name","price"}).
		AddRow(2, "Tempe", 1500)
	mock.ExpectQuery("SELECT id, name, price FROM products").
		WillReturnRows(rows)

	mysqlDB = db
	request, _ := http.NewRequest("GET", "/api/products", nil)
	response := httptest.NewRecorder()

	Server().ServeHTTP(response, request)
	expectedResponse := `{"data":[{"type":"products","id":"2","attributes":{"name":"Tempe","price":1500}}],"meta":{"total":1}}`
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Log(err)
	}
	assert.Equal(t, 200, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(responseBody)))
	t.Log(response.Body.String())
}

func Test_CreateProducts(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Log(err)
	}

	mock.ExpectPrepare("INSERT INTO products (name, price) values (?, ?)")
	mock.ExpectExec("INSERT INTO products (name, price) values (?, ?)").
		WithArgs("tempe", 1500).
		WillReturnResult(sqlmock.NewResult(2, 1))

	mysqlDB = db
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "products",
			"attributes": map[string]interface{}{
				"name":  "tempe",
				"price": 1500,
			},
		},
	}
	requestBody, _ := json.Marshal(data)
	request, _ := http.NewRequest("POST", "/api/products", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	Server().ServeHTTP(response, request)
	expectedResponse := `{"data":{"type":"products","id":"2","attributes":{"name":"tempe","price":1500}}}`
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Log(err)
	}
	assert.Equal(t, http.StatusCreated, response.Code, "Invalid response code")
	assert.Equal(t, expectedResponse, string(bytes.TrimSpace(responseBody)))
	t.Log(response.Body.String())
}

func Test_DeleteProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Log(err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM products WHERE id = ?").
		WillReturnResult(sqlmock.NewResult(2,1))

	mysqlDB = db
	request, _ := http.NewRequest("DELETE", "/api/products/1", nil)
	response := httptest.NewRecorder()
	Server().ServeHTTP(response, request)

	assert.Equal(t, http.StatusNoContent, response.Code, "Invalid response code")
}