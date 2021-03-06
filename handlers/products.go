// Package classification of Product API
//
//  Documentation for Product API
//
// Schemes:	http
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dev-muser/learning_go/data"
	"github.com/gorilla/mux"
)

// A list of products returns in the response
// swagger:response productResponse
type productsResponseWrapper struct {
	// All products in the system
	// in: body
	Body []data.Product
}

// swagger:response noContent
type productsNoContent struct {
}

// swagger:parameters deleteProduct
type productIDParameterWrapper struct {
	// The id of the product to delete from the datastore
	// in: path
	// required: true
	ID int `json:id`
}

// Products is a http.Handler
type Products struct {
	l *log.Logger
}

func NewProducts(l *log.Logger) *Products {
	return &Products{l}
}

// func (p *Products) GetProducts(rw http.ResponseWriter, r *http.Request) {
// 	p.l.Println("Handle GET Product")
// 	listOfProducts := data.GetProducts()
// 	err := listOfProducts.ToJSON(rw)
// 	if err != nil {
// 		http.Error(rw, "Unable to marshal json", http.StatusInternalServerError)
// 	}
// }

func (p *Products) AddProduct(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle POST Product")

	prod := r.Context().Value(KeyProduct{}).(data.Product) //safe to cast it
	data.AddProduct(&prod)
}

func (p *Products) UpdateProducts(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(rw, "Unable to convert id", http.StatusBadRequest)
		return
	}
	p.l.Println("Handle PUT Products", id)
	prod := r.Context().Value(KeyProduct{}).(data.Product) //safe to cast it

	err = data.UpdateProduct(id, &prod)
	if err == data.ErrProductNotFound {
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Product not found", http.StatusInternalServerError)
		return
	}
}

// the way to use a context is to define a key (type - preffered or string)
type KeyProduct struct{}

// to avoid duplicate code, like prod.FromJSON, we can put it on a function or use a Middleware
// USE function in Gorilla allows you apply Middleware.
// MiddleWare is a http handler and can chain multiplers handlers together (like Cors - validate requests)

//MiddlewareProductValidation validates the product in the request and calls next if ok
func (p Products) MiddlewareProductValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prod := data.Product{}

		err := prod.FromJSON(r.Body)
		if err != nil {
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		//validate/sanitize the product
		err = prod.Validate()
		if err != nil {
			p.l.Println("[ERROR] validating product", err)
			http.Error(
				rw,
				fmt.Sprintf("Error validating prodcut: %s", err),
				http.StatusBadRequest,
			)
			return
		}

		// add the product to the context to use it later on.
		// Request has Context.
		ctx := context.WithValue(r.Context(), KeyProduct{}, prod)
		req := r.WithContext(ctx)

		// call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, req)
	})
}

// GenericError is a generic error message returned by a server
type GenericError struct {
	Message string `json:"message"`
}

// getProductID return the product ID from the URL
// Panics if cannot convert the id into an integer
// thins should never happen as the router ensures that
// this is a valid number
func getProductID(r *http.Request) int {
	// parse the product id from the URL
	vars := mux.Vars(r)

	// convert the id into the integer and return
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		// should never happen
		panic(err)
	}

	return id
}
