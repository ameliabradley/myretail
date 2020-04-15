package productaggregate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

// ProductNameRepository handles product names
type ProductNameRepository interface {
	Get(productID int) (string, error)
}

type targetResponse struct {
	Product struct {
		Item struct {
			ProductDescription struct {
				Title string
			} `json:"product_description"`
		}
	}
}

// TargetProductNameRepository handles product name fetching from Target's API
type TargetProductNameRepository struct {
	httpClient *http.Client
}

// NewTargetProductNameRepository creates a new TargetProductNameRepository
func NewTargetProductNameRepository() TargetProductNameRepository {
	return TargetProductNameRepository{
		httpClient: getClient(),
	}
}

func getClient() *http.Client {
	// See: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
}

// Get fetches a product's name by id
func (t TargetProductNameRepository) Get(productID int) (string, error) {
	url := fmt.Sprintf("https://redsky.target.com/v2/pdp/tcin/%d?excludes=taxonomy,price,promotion,bulk_ship,rating_and_review_reviews,rating_and_review_statistics,question_answer_statistics", productID)
	log.Printf("Making request to %s", url)

	response, err := t.httpClient.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return t.readTitle(body)
}

func (t TargetProductNameRepository) readTitle(data []byte) (string, error) {
	var tr targetResponse
	err := json.Unmarshal(data, &tr)
	if err != nil {
		return "", err
	}

	return tr.Product.Item.ProductDescription.Title, nil
}
