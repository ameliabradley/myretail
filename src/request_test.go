package productaggregate

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var parseTests = []struct {
	in          string
	outInt      int
	expectError bool
}{
	{"/123", 123, false},
	{"/123.00", 0, true},
}

func TestParseProductID(t *testing.T) {
	for _, tt := range parseTests {
		t.Run(tt.in, func(t *testing.T) {
			result, err := parseProductID(tt.in)
			if result != tt.outInt {
				t.Errorf("got %d, want %d", result, tt.outInt)
			}

			haveError := err != nil
			if haveError && !tt.expectError {
				t.Errorf("received unexpected error: %s", err)
			}

			if !haveError && tt.expectError {
				t.Errorf("expected error. did not receive error")
			}
		})
	}
}

type httpWant struct {
	code int
	body string
}

type priceGetResult struct {
	price *ProductPrice
	err   error
}

type nameResult struct {
	name string
	err  error
}

type handlerIn struct {
	request *http.Request
	pgr     priceGetResult
	ppr     error
	nr      nameResult
}

func dummyRequest(method string, data string) *http.Request {
	return httptest.NewRequest(method, "http://example.com/123", strings.NewReader(data))
}

var handlerTests = []struct {
	name string
	in   handlerIn
	want httpWant
}{
	{
		name: "GET Basic test",
		in: handlerIn{
			request: httptest.NewRequest("GET", "http://example.com/123", nil),
		},
		want: httpWant{
			code: http.StatusOK,
			body: `{"product_id":123}`,
		},
	},
	{
		name: "GET No product number",
		in: handlerIn{
			request: httptest.NewRequest("GET", "http://example.com/", nil),
		},
		want: httpWant{
			code: http.StatusBadRequest,
			body: "Invalid product ID\n",
		},
	},
	{
		name: "PUT No product number",
		in: handlerIn{
			request: httptest.NewRequest("PUT", "http://example.com/", nil),
		},
		want: httpWant{
			code: http.StatusBadRequest,
			body: "Invalid product ID\n",
		},
	},
	{
		name: "POST Unsupported method",
		in: handlerIn{
			request: httptest.NewRequest("POST", "http://example.com/", nil),
		},
		want: httpWant{
			code: http.StatusMethodNotAllowed,
			body: "Unsupported method\n",
		},
	},
	{
		name: "GET Expect error on bad product id",
		in: handlerIn{
			request: httptest.NewRequest("GET", "http://example.com/1-23", nil),
		},
		want: httpWant{
			code: http.StatusBadRequest,
			body: "Invalid product ID\n",
		},
	},
	{
		name: "GET Expect only price if only price found",
		in: handlerIn{
			request: httptest.NewRequest("GET", "http://example.com/123", nil),
			pgr: priceGetResult{
				price: &ProductPrice{
					ProductID:    123,
					Price:        100,
					CurrencyCode: "USD",
				},
			},
			nr: nameResult{
				err: errors.New("Could not fetch name"),
			},
		},
		want: httpWant{
			code: http.StatusOK,
			body: `{"product_id":123,"current_price":{"value":100,"currency_code":"USD"}}`,
		},
	},
	{
		name: "GET Expect only name if only name found",
		in: handlerIn{
			request: httptest.NewRequest("GET", "http://example.com/123", nil),
			pgr: priceGetResult{
				price: nil,
				err:   errors.New("Could not fetch price"),
			},
			nr: nameResult{
				name: "Picard",
			},
		},
		want: httpWant{
			code: http.StatusOK,
			body: `{"product_id":123,"name":"Picard"}`,
		},
	},
	{
		name: "PUT With empty data",
		in: handlerIn{
			request: dummyRequest("PUT", ""),
		},
		want: httpWant{
			code: http.StatusBadRequest,
			body: "Request body must not be empty\n",
		},
	},
	{
		name: "PUT Valid",
		in: handlerIn{
			request: dummyRequest("PUT", `{"value":100,"currency_code":"USD"}`),
		},
		want: httpWant{
			code: http.StatusOK,
			body: "Product updated",
		},
	},
	{
		name: "PUT With unknown field",
		in: handlerIn{
			request: dummyRequest("PUT", `{"foo":1}`),
		},
		want: httpWant{
			code: http.StatusBadRequest,
			body: "Request body contains unknown field \"foo\"\n",
		},
	},
	{
		name: "PUT With multiple json blobs",
		in: handlerIn{
			request: dummyRequest("PUT", `{"value":1,"currency_code":"USD"}{"foo2":2}`),
		},
		want: httpWant{
			code: http.StatusBadRequest,
			body: "Request body must only contain a single JSON object\n",
		},
	},
	{
		name: "PUT Ran into datastore error",
		in: handlerIn{
			request: dummyRequest("PUT", `{"value":100,"currency_code":"USD"}`),
			ppr:     errors.New("Datastore put error"),
		},
		want: httpWant{
			code: http.StatusInternalServerError,
			body: "Error updating product\n",
		},
	},
}

type StubPriceRepository struct {
	pgr priceGetResult
	ppr error
}

func (s StubPriceRepository) Get(productID int) (*ProductPrice, error) {
	return s.pgr.price, s.pgr.err
}

func (s StubPriceRepository) Put(price ProductPrice) error {
	return s.ppr
}

type StubNameRepository struct {
	nr nameResult
}

func (s StubNameRepository) Get(productID int) (string, error) {
	return s.nr.name, s.nr.err
}

func TestRequestHandler(t *testing.T) {
	for _, tt := range handlerTests {
		t.Run(tt.name, func(t *testing.T) {
			rh := RequestHandler{
				priceRepository: StubPriceRepository{
					ppr: tt.in.ppr,
					pgr: tt.in.pgr,
				},
				nameRepository: StubNameRepository{
					nr: tt.in.nr,
				},
			}

			w := httptest.NewRecorder()
			rh.HandleRequest(w, tt.in.request)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			if string(body) != tt.want.body {
				t.Errorf("got %s, want %s", string(body), tt.want.body)
			}

			if resp.StatusCode != tt.want.code {
				t.Errorf("got %d, want %d", resp.StatusCode, tt.want.code)
			}
		})
	}
}
