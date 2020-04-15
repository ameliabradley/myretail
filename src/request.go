package productaggregate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/gddo/httputil/header"
)

// RequestHandler handles incoming product requests
type RequestHandler struct {
	priceRepository ProductPriceRepository
	nameRepository  ProductNameRepository
}

// NewRequestHandler creates a new RequestHandler
func NewRequestHandler() (RequestHandler, error) {
	projectID := os.Getenv("PROJECT_ID")
	datastoreID := os.Getenv("DATASTORE_ID")

	log.Printf("Created request handler { PROJECT_ID: %s DATASTORE_ID: %s }", projectID, datastoreID)

	ctx := context.Background()

	priceRepository, err := NewGoogleProductPriceRepository(ctx, projectID, datastoreID)
	if err != nil {
		return RequestHandler{}, err
	}

	nameRepository := NewTargetProductNameRepository()

	return RequestHandler{
		priceRepository: priceRepository,
		nameRepository:  nameRepository,
	}, nil
}

// HandleRequest is the main entrypoint for http requests
func (rh RequestHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request { PATH: %s METHOD: %s }", r.URL.Path, r.Method)

	switch r.Method {
	case "GET":
		rh.HandleGet(w, r)

	case "PUT":
		rh.HandlePut(w, r)

	default:
		msg := "Unsupported method"
		http.Error(w, msg, http.StatusMethodNotAllowed)
	}
}

// Product represents the core product data returned by this service
type Product struct {
	ProductID    int           `json:"product_id"`
	Name         string        `json:"name,omitempty"`
	CurrentPrice *ProductPrice `json:"current_price,omitempty"`
}

// ProductPrice represents the product price information in the datastore
type ProductPrice struct {
	ProductID int `datastore:"product_id" json:"-"`

	// I'd switch to big.Rat for currency if performing price operations
	Price float64 `datastore:"price" json:"value"`

	CurrencyCode string `datastore:"currency_code" json:"currency_code"`
}

func parseProductID(path string) (int, error) {
	runes := []rune(path)             // Convert to rune
	withoutSlash := string(runes[1:]) // Eliminate the slash at the beginning
	return strconv.Atoi(withoutSlash)
}

// HandlePut handles product PUT requests
func (rh RequestHandler) HandlePut(w http.ResponseWriter, r *http.Request) {
	productID, err := parseProductID(r.URL.Path)
	if err != nil {
		msg := "Invalid product ID"
		http.Error(w, msg, http.StatusBadRequest)
		log.Printf("Could not parse product ID '%s'", r.URL.Path)
		return
	}

	// If the Content-Type header is present, check that it has the value
	// application/json. Note that we are using the gddo/httputil/header
	// package to parse and extract the value here, so the check works
	// even if the client includes additional charset or boundary
	// information in the header.
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination".
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var price ProductPrice
	err = dec.Decode(&price)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by the request body being too large. Again
		// there is an open issue regarding turning this into a sentinel
		// error at https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Check that the request body only contained a single JSON object.
	if dec.More() {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	price.ProductID = productID

	err = rh.priceRepository.Put(price)
	if err != nil {
		msg := "Error updating product"
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Failed updating product price: %s", err)
		return
	}

	fmt.Fprint(w, "Product updated")
}

// HandleGet handles product GET requests
func (rh RequestHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	productID, err := parseProductID(r.URL.Path)
	if err != nil {
		msg := "Invalid product ID"
		http.Error(w, msg, http.StatusBadRequest)
		log.Printf("Could not parse product ID '%s'", r.URL.Path)
		return
	}

	product := Product{
		ProductID:    productID,
		Name:         "",
		CurrentPrice: nil,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		price, err := rh.priceRepository.Get(productID)
		if err != nil {
			log.Printf("Failed fetching from price repository: %s", err)
			return
		}

		product.CurrentPrice = price
	}()

	go func() {
		defer wg.Done()
		name, err := rh.nameRepository.Get(productID)
		if err != nil {
			log.Printf("Failed fetching from name repository: %s", err)
			return
		}

		product.Name = name
	}()
	wg.Wait()

	json, err := json.Marshal(product)
	if err != nil {
		msg := "Could not process request"
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Error marshalling product %d", productID)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}
