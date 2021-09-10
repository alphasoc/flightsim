// Package doh provides general DNS-over-HTTPS functionality.
package doh

import (
	"encoding/json"
	"net/http"
)

// Question section of DoH query.
type Question struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

// Answer section of DoH query.
type Answer struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

// Response to a DoH query.
type Response struct {
	Status   int        `json:"Status"`
	TC       bool       `json:"TC"`
	RD       bool       `json:"RD"`
	RA       bool       `json:"RA"`
	AD       bool       `json:"AD"`
	CD       bool       `json:"CD"`
	Question []Question `json:"Question"`
	Answer   []Answer   `json:"Answer"`
	Comment  string     `json:"Comment"`
}

// Decode decodes a DoH response, returning a pointer to a Response and an error.
// NOTE: Currently we don't care about parsing responses, but perhaps in the future.
func Decode(r *http.Response) (*Response, error) {
	var resp Response
	err := json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil

}

// IsValidResponse returns a boolean indicating if the the *http.Response r was a 200OK.
func IsValidResponse(r *http.Response) bool {
	return r != nil && r.StatusCode == 200
}
