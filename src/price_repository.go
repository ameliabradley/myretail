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

// GCPProductPriceRepository gets product prices from Google Cloud
type GCPProductPriceRepository struct {
	datastoreID string
	client      DatastoreClient
	ctx         context.Context
}

type DatastoreClient interface {
	Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error)
	Get(ctx context.Context, key *datastore.Key, dst interface{}) (err error)
}

type NewDatastoreClient func(ctx context.Context) (DatastoreClient, error)

func NewGCPDatastoreClientCreator(projectID string) NewDatastoreClient {
	return func(ctx context.Context) (DatastoreClient, error) {
		return datastore.NewClient(ctx, projectID)
	}
}

// NewGCPProductPriceRepository creates a new GCPProductPriceRepository
func NewGCPProductPriceRepository(ctx context.Context, newClient NewDatastoreClient, datastoreID string) (*GCPProductPriceRepository, error) {
	client, err := newClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCPProductPriceRepository{
		datastoreID: datastoreID,
		client:      client,
		ctx:         ctx,
	}, nil
}

func (p GCPProductPriceRepository) keyFromProductID(productID int) string {
	return "product_" + strconv.Itoa(productID)
}

// Get fetches a product price by id
func (p GCPProductPriceRepository) Get(productID int) (*ProductPrice, error) {
	key := p.keyFromProductID(productID)
	datastoreKey := datastore.NameKey(p.datastoreID, key, nil)

	newdata := &ProductPrice{}
	if err := p.client.Get(p.ctx, datastoreKey, newdata); err != nil {
		return &ProductPrice{}, err
	}

	return newdata, nil
}

// Put updates a product price
func (p GCPProductPriceRepository) Put(product ProductPrice) error {
	key := p.keyFromProductID(product.ProductID)

	// Make a key to map to datastore
	datastoreKey := datastore.NameKey(p.datastoreID, key, nil)

	if _, err := p.client.Put(p.ctx, datastoreKey, &product); err != nil {
		return err
	}

	return nil
}
