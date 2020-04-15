package productaggregate

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

var readTitleTests = []struct {
	in          string
	out         string
	expectError bool
}{
	{"lebowski.json", "The Big Lebowski (Blu-ray)", false},
	{"spongebob.json", "SpongeBob SquarePants: SpongeBob's Frozen Face-off", false},
	{"notfound.json", "", false},
	{"baddata", "", true},
}

func helperLoadBytes(t *testing.T, subdir string, name string) []byte {
	path := filepath.Join("testdata", subdir, name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestNameRepositoryReadTitle(t *testing.T) {
	repository := TargetProductNameRepository{}

	for _, tt := range readTitleTests {
		t.Run(tt.in, func(t *testing.T) {
			data := helperLoadBytes(t, "products", tt.in)
			result, err := repository.readTitle(data)
			if result != tt.out {
				t.Errorf("got %s, want %s", result, tt.out)
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

func TestNewTargetNameRepository(t *testing.T) {
	NewTargetProductNameRepository()
}

func TestTargetNameRepositoryGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "BAD DATA")
	}))
	ts.EnableHTTP2 = true
	defer ts.Close()

	repository := TargetProductNameRepository{
		httpClient: ts.Client(),
	}

	resp, err := repository.Get(123)
	if resp != "" {
		t.Errorf("Expected '', got '%s'", resp)
	}

	if err != nil {
		t.Errorf("Expected no error, got '%+v'", err)
	}
}
