package productaggregate

import (
	"errors"
	"testing"

	"cloud.google.com/go/datastore"
	"golang.org/x/net/context"
)

type testDatastoreClient struct {
	getErr error
	putErr error
}

func newTestDatastoreClient(getErr error, putErr error) (DatastoreClient, error) {
	return testDatastoreClient{
		getErr: getErr,
		putErr: putErr,
	}, nil
}

func (t testDatastoreClient) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	return nil, t.putErr
}

func (t testDatastoreClient) Get(ctx context.Context, key *datastore.Key, dst interface{}) (err error) {
	return t.getErr
}

func newTestDatastoreClientCreator(createErr error, getErr error, putErr error) NewDatastoreClient {
	return func(ctx context.Context) (DatastoreClient, error) {
		if createErr != nil {
			return nil, createErr
		}

		return newTestDatastoreClient(getErr, putErr)
	}
}

type priceRepositoryCreateIn struct {
	productID int
	err       error
}

type priceRepositoryCreateWant struct {
	hasError bool
}

var priceRepositoryCreateTests = []struct {
	name string
	in   priceRepositoryCreateIn
	want priceRepositoryCreateWant
}{
	{
		name: "Create basic test, no error",
		in: priceRepositoryCreateIn{
			productID: 10,
		},
		want: priceRepositoryCreateWant{
			hasError: false,
		},
	},
	{
		name: "Create throwing error",
		in: priceRepositoryCreateIn{
			productID: 10,
			err:       errors.New("Error creating repository"),
		},
		want: priceRepositoryCreateWant{
			hasError: true,
		},
	},
}

func TestProductPriceRepositoryCreate(t *testing.T) {
	for _, tt := range priceRepositoryCreateTests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			creator := newTestDatastoreClientCreator(tt.in.err, nil, nil)
			_, err := NewGCPProductPriceRepository(ctx, creator, "test")

			if err != nil && !tt.want.hasError {
				t.Errorf("expected no error. error thrown: %+v", err)
			}

			if err == nil && tt.want.hasError {
				t.Error("expected error. none found")
			}
		})
	}
}

type priceRepositoryGetIn struct {
	productID int
	err       error
}

type priceRepositoryGetWant struct {
	hasError bool
}

var priceRepositoryGetTests = []struct {
	name string
	in   priceRepositoryGetIn
	want priceRepositoryGetWant
}{
	{
		name: "Get basic test, no error",
		in: priceRepositoryGetIn{
			productID: 10,
		},
		want: priceRepositoryGetWant{
			hasError: false,
		},
	},
	{
		name: "Get throws error",
		in: priceRepositoryGetIn{
			productID: 10,
			err:       errors.New("Error getting data"),
		},
		want: priceRepositoryGetWant{
			hasError: true,
		},
	},
}

func TestProductPriceRepositoryGet(t *testing.T) {
	for _, tt := range priceRepositoryGetTests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			creator := newTestDatastoreClientCreator(nil, tt.in.err, nil)
			repository, createErr := NewGCPProductPriceRepository(ctx, creator, "test")
			if createErr != nil {
				t.Errorf("unexpected error: %+v", createErr)
			}

			_, err := repository.Get(tt.in.productID)

			if err != nil && !tt.want.hasError {
				t.Errorf("expected no error. error thrown: %+v", err)
			}

			if err == nil && tt.want.hasError {
				t.Error("expected error. none found")
			}
		})
	}
}

type priceRepositoryPutIn struct {
	price ProductPrice
	err   error
}

type priceRepositoryPutWant struct {
	hasError bool
}

var priceRepositoryPutTests = []struct {
	name string
	in   priceRepositoryPutIn
	want priceRepositoryPutWant
}{
	{
		name: "Put basic test, no error",
		in: priceRepositoryPutIn{
			price: ProductPrice{
				ProductID: 10,
			},
		},
		want: priceRepositoryPutWant{
			hasError: false,
		},
	},
	{
		name: "Put throws error",
		in: priceRepositoryPutIn{
			price: ProductPrice{
				ProductID: 10,
			},
			err: errors.New("Error putting data"),
		},
		want: priceRepositoryPutWant{
			hasError: true,
		},
	},
}

func TestProductPriceRepositoryPut(t *testing.T) {
	for _, tt := range priceRepositoryPutTests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			creator := newTestDatastoreClientCreator(nil, nil, tt.in.err)
			repository, createErr := NewGCPProductPriceRepository(ctx, creator, "test")
			if createErr != nil {
				t.Errorf("unexpected error: %+v", createErr)
			}

			err := repository.Put(tt.in.price)

			if err != nil && !tt.want.hasError {
				t.Errorf("expected no error. error thrown: %+v", err)
			}

			if err == nil && tt.want.hasError {
				t.Error("expected error. none found")
			}
		})
	}
}
