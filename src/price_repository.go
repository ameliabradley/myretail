package productaggregate

import (
	"context"
	"strconv"

	"cloud.google.com/go/datastore"
)

// ProductPriceRepository handles product prices
type ProductPriceRepository interface {
	Get(productID int) (*ProductPrice, error)
	Put(ProductPrice) error
}

// GoogleProductPriceRepository gets product prices from Google Cloud
type GoogleProductPriceRepository struct {
	datastoreID string
	client      *datastore.Client
	ctx         context.Context
}

// NewGoogleProductPriceRepository creates a new GoogleProductPriceRepository
func NewGoogleProductPriceRepository(ctx context.Context, projectID string, datastoreID string) (*GoogleProductPriceRepository, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &GoogleProductPriceRepository{
		datastoreID: datastoreID,
		client:      client,
		ctx:         ctx,
	}, nil
}

func (p GoogleProductPriceRepository) keyFromProductID(productID int) string {
	return "product_" + strconv.Itoa(productID)
}

// Get fetches a product price by id
func (p GoogleProductPriceRepository) Get(productID int) (*ProductPrice, error) {
	key := p.keyFromProductID(productID)
	datastoreKey := datastore.NameKey(p.datastoreID, key, nil)

	newdata := &ProductPrice{}
	if err := p.client.Get(p.ctx, datastoreKey, newdata); err != nil {
		return &ProductPrice{}, err
	}

	return newdata, nil
}

// Put updates a product price
func (p GoogleProductPriceRepository) Put(product ProductPrice) error {
	key := p.keyFromProductID(product.ProductID)

	// Make a key to map to datastore
	datastoreKey := datastore.NameKey(p.datastoreID, key, nil)

	if _, err := p.client.Put(p.ctx, datastoreKey, &product); err != nil {
		return err
	}

	return nil
}
