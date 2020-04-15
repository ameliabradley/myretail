package productaggregate

import (
	"log"
	"net/http"
)

// StartCloudFunction starts the product handler in Google Cloud
func StartCloudFunction(w http.ResponseWriter, r *http.Request) {
	handler, err := NewRequestHandler()
	if err != nil {
		msg := "Could not process request"
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Initialization failure: %s", err)
		return
	}

	handler.HandleRequest(w, r)
}
