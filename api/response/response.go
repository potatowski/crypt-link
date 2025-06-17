package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON writes the given data as a JSON response with the specified HTTP status code.
// It sets the "Content-Type" header to "application/json" and encodes the data into the response writer.
// If data is nil, no body is written. If encoding fails, the function logs the error and terminates the application.
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Fatal(err)
		}
	}
}

// Error writes a JSON-formatted error response with the specified HTTP status code.
// The response body contains an "error" field with the error message.
// It sets the "Content-Type" header to "application/json".
func Error(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	JSON(w, statusCode, struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
